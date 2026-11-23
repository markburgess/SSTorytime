# Shared logic for the SSTorytime Docker environment (POSIX sh; sourced by
# incontainer.sh, which runs inside the containers).
#
# The caller must define these adapters before calling anything here:
#   sql "QUERY"      run SQL against the test database, print tuples only
#   ws CMD [ARG...]  run a command in the workspace environment, cwd /work/run
#   webcurl ARG...   curl against https://localhost:8443 with explicit cert trust
#   web_logs         print where to find recent web server logs

C_HINT=${C_HINT:-"docker compose -f docker/compose.yaml"}
STAGE=${STAGE:-init}
SERVER_CHANGED=0

stage() { STAGE=$1; printf '\n[%s] %s\n' "$1" "${2:-}"; }
info()  { printf '  %s\n' "$*"; }
warn()  { printf '  WARNING: %s\n' "$*" >&2; }
fail()  {
    printf '\nFAILED at stage [%s]: %s\n' "$STAGE" "$*" >&2
    printf 'Containers and volumes were left as they are for diagnosis.\n' >&2
    exit 1
}

# ---------------------------------------------------------------- marker
# sst_install_test.state: fixture = seeding|seeded, baseline = verified.
# Initialization history only, not proof that the fixture still exists.

marker_get() {
    sql "select case when to_regclass('sst_install_test.state') is null then ''
         else coalesce((select value from sst_install_test.state where key='$1'),'') end"
}

marker_set() {
    sql "insert into sst_install_test.state(key,value,updated_at) values ('$1','$2',now())
         on conflict (key) do update set value=excluded.value, updated_at=now()" >/dev/null
}

# The independent fixture check: doors.n4l's "door (fwd) passage" link exists.
fixture_present() {
    n=$(sql "select count(*) from node d cross join lateral unnest(d.il1) as l
             join node p on p.s = 'passage' and p.nptr = l.dst
             where d.s = 'door'" 2>/dev/null) || return 1
    [ "${n:-0}" -gt 0 ]
}

# ---------------------------------------------------------------- database
# wait_db SECONDS: TCP connection AND init scripts done (unaccent present).
wait_db() {
    info "waiting for PostgreSQL (TCP) and the unaccent extension"
    i=0
    until [ "$(sql "select count(*) from pg_extension where extname='unaccent'" 2>/dev/null || true)" = "1" ]; do
        i=$((i + 1))
        [ "$i" -lt "$(( $1 * 2 ))" ] || fail "PostgreSQL not ready (or unaccent missing) after $1s"
        sleep 0.5
    done
    info "PostgreSQL ready"
}

# First-use initialization. Imports examples/doors.n4l ONLY into a
# confirmed-fresh database; stops on anything it does not recognize.
initialize_db() {
    has_schema=$(sql "select count(*) from pg_namespace where nspname='sst_install_test'")

    if [ "$has_schema" != "0" ]; then
        if [ "$(sql "select to_regclass('sst_install_test.state') is not null")" != "t" ]; then
            fail "invalid marker: schema sst_install_test exists but table state is missing.
  Inspect the database; if you want a clean start: $C_HINT down -v"
        fi
        fixture=$(marker_get fixture)
        case "$fixture" in
            seeded)
                when=$(sql "select to_char(updated_at,'YYYY-MM-DD HH24:MI') from sst_install_test.state where key='fixture'")
                info "database already initialized ($when); not importing anything"
                return 0 ;;
            seeding)
                fail "a previous initialization did not complete (marker fixture=seeding).
  The database may be partially populated. Nothing was wiped.
  Inspect: $C_HINT exec db psql -U sstoryline
  Start over (DESTROYS all test data): $C_HINT down -v" ;;
            *)
                fail "invalid marker: fixture='${fixture}' is not a recognized state. Nothing was changed." ;;
        esac
    fi

    ntables=$(sql "select count(*) from pg_tables where schemaname='public'")
    if [ "$ntables" != "0" ]; then
        fail "unrecognized database: no install marker but $ntables table(s) exist in schema public.
  Refusing to seed or to start SSTorytime against it (the application alters its schema on open).
  Inspect: $C_HINT exec db psql -U sstoryline"
    fi

    info "fresh database confirmed; marking as seeding"
    sql "begin;
         create schema sst_install_test;
         create table sst_install_test.state(key text primary key, value text not null,
                                             updated_at timestamptz not null default now());
         insert into sst_install_test.state(key,value) values ('fixture','seeding');
         commit;" >/dev/null

    info "importing examples/doors.n4l (N4L -u -wipe)"
    if ! out=$(ws /work/bin/N4L -u -wipe /src/examples/doors.n4l 2>&1); then
        printf '%s\n' "$out" >&2
        fail "N4L import exited nonzero; marker left at fixture=seeding"
    fi
    if ! fixture_present; then
        printf '%s\n' "$out" >&2
        fail "N4L finished but the expected 'door -> passage' link is not in PostgreSQL; marker left at fixture=seeding"
    fi
    marker_set fixture seeded
    info "fixture imported and confirmed in PostgreSQL (fixture=seeded)"
}

# ---------------------------------------------------------------- workspace
ensure_dirs() {
    ws mkdir -p /work/bin/.staging /work/tls /work/run /work/cache || fail "cannot create /work directories"
}

# All SSTorytime tools built from cmd/ (cmd/demo_pocs are internal test
# programs and are not built). Binary name = directory name, except
# cmd/server -> http_server.
SST_TOOLS="N4L searchN4L http_server text2N4L removeN4L notes pathsolve graph_report API_EXAMPLE_1 API_EXAMPLE_2 API_EXAMPLE_3 API_EXAMPLE_4"
tool_dir() { case $1 in http_server) echo server ;; *) echo "$1" ;; esac; }
is_tool()  { case " $SST_TOOLS " in *" $1 "*) return 0 ;; esac; return 1; }

binaries_present() {
    ws sh -c 'for b in "$@"; do test -x /work/bin/$b || exit 1; done' sh $SST_TOOLS
}

# build_binaries NAME...   (NAME: one of $SST_TOOLS)
# Compiles all requested tools in one parallel `go build` into a fresh
# /work/bin/.staging. If anything fails, nothing is published. Otherwise
# changed binaries replace /work/bin/NAME by atomic rename (safe while the old
# one runs). Sets SERVER_CHANGED=1 if http_server was replaced.
build_binaries() {
    SERVER_CHANGED=0
    pkgs=""
    for name in "$@"; do
        is_tool "$name" || { echo "  unknown tool: $name" >&2; return 1; }
        pkgs="$pkgs ./cmd/$(tool_dir "$name")"
    done
    info "go build ($#):$pkgs"
    # shellcheck disable=SC2086
    result=$(ws sh -c '
        set -e
        rm -rf /work/bin/.staging && mkdir -p /work/bin/.staging
        pkgs=$1; shift
        go build -C /src -o /work/bin/.staging/ $pkgs || {
            echo "  go build failed; existing binaries in /work/bin are unchanged" >&2
            exit 1; }
        [ ! -e /work/bin/.staging/server ] || mv /work/bin/.staging/server /work/bin/.staging/http_server
        for name in "$@"; do
            new=$(sha256sum "/work/bin/.staging/$name" | cut -d" " -f1)
            old=$(sha256sum "/work/bin/$name" 2>/dev/null | cut -d" " -f1 || true)
            if [ "$new" = "$old" ]; then
                rm -f "/work/bin/.staging/$name"; echo "$name:unchanged"
            else
                mv -f "/work/bin/.staging/$name" "/work/bin/$name"; echo "$name:updated"
            fi
        done' sh "$pkgs" "$@") || return 1
    changed=""; same=0
    for r in $result; do
        case "$r" in
            *:updated) changed="$changed ${r%%:*}" ;;
            *) same=$((same + 1)) ;;
        esac
        [ "$r" = "http_server:updated" ] && SERVER_CHANGED=1
    done
    info "updated:${changed:- none}   unchanged: $same"
    return 0
}

# ---------------------------------------------------------------- tls
# Two certificate sources; /work/tls/current points at the active one and every
# consumer (http_server, healthcheck, verify) reads current/{cert,key,ca}.pem:
#   mkcert  docker/tls/{cert,key,rootCA}.pem in the checkout (/src/docker/tls),
#           signed by a CA the host browser trusts after `mkcert -install`.
#           Used whenever those files exist; validated, then snapshotted to
#           /work/tls/mkcert-<sha>/ so web only ever loads a checked pair.
#   self    /work/tls/self/: generated self-signed fallback (browser warning).
#           Generated once, never overwritten.
TLS_USER_DIR=/src/docker/tls
TLS_MODE=""

ensure_tls() {
    tls_state=$(ws sh -c '
        u='"$TLS_USER_DIR"'
        if [ -e $u/rootCA-key.pem ]; then echo cakey; exit; fi
        n=0; for f in cert.pem key.pem rootCA.pem; do [ -e $u/$f ] && n=$((n+1)); done
        if [ $n -eq 0 ]; then echo none; exit; fi
        if [ $n -ne 3 ]; then echo incomplete; exit; fi
        c=$u/cert.pem; k=$u/key.pem; ca=$u/rootCA.pem
        if ! openssl x509 -in $c -noout -checkend 0 >/dev/null 2>&1; then echo expired
        elif [ "$(openssl x509 -in $c -noout -pubkey 2>/dev/null)" != "$(openssl pkey -in $k -pubout 2>/dev/null)" ]; then echo mismatch
        elif ! openssl x509 -in $c -noout -checkhost localhost 2>/dev/null | grep -q "does match"; then echo nohost
        elif cmp -s $c $ca; then echo notca
        elif ! openssl verify -CAfile $ca $c >/dev/null 2>&1; then echo unsigned
        else echo ok; fi')
    case "$tls_state" in
        ok)
            # Validated snapshot in its own directory, then an atomic switch, so
            # web never sees a half-edited pair from docker/tls.
            set_dir=$(ws sh -c 'set -e
                u='"$TLS_USER_DIR"'
                d=/work/tls/mkcert-$(sha256sum < $u/cert.pem | cut -c1-12)
                if [ ! -d $d ]; then
                    mkdir -p $d.tmp
                    cp $u/cert.pem $d.tmp/cert.pem; cp $u/key.pem $d.tmp/key.pem; cp $u/rootCA.pem $d.tmp/ca.pem
                    chmod 600 $d.tmp/key.pem
                    mv $d.tmp $d
                fi
                basename $d') || fail "could not snapshot the docker/tls certificate"
            tls_select "$set_dir"
            TLS_MODE=mkcert
            until=$(ws openssl x509 -in "/work/tls/$set_dir/cert.pem" -noout -enddate | cut -d= -f2)
            info "using mkcert certificate from docker/tls (valid until $until)"
            return 0 ;;
        none) ;;
        cakey)      fail "docker/tls/rootCA-key.pem found: that is mkcert's CA PRIVATE KEY. Remove it from the checkout now; it must stay only in \`mkcert -CAROOT\`." ;;
        incomplete) fail "docker/tls needs all of cert.pem, key.pem and rootCA.pem (see docker/tls/README.md), or none of them for the self-signed fallback." ;;
        expired)    fail "docker/tls/cert.pem has expired. Regenerate it with mkcert (see docker/tls/README.md)." ;;
        mismatch)   fail "docker/tls/cert.pem and key.pem do not belong together. Regenerate both with mkcert." ;;
        nohost)     fail "docker/tls/cert.pem is not valid for 'localhost'. Generate it with: mkcert -cert-file cert.pem -key-file key.pem localhost 127.0.0.1 ::1" ;;
        notca)      fail "docker/tls/rootCA.pem is the server certificate. Copy mkcert's CA certificate instead: \$(mkcert -CAROOT)/rootCA.pem" ;;
        unsigned)   fail "docker/tls/cert.pem is not signed by docker/tls/rootCA.pem. Copy rootCA.pem from \`mkcert -CAROOT\` of the mkcert that generated the certificate." ;;
        *)          fail "could not inspect docker/tls ($tls_state)" ;;
    esac

    # Self-signed fallback. One-time move of the pre-mkcert layout
    # (/work/tls/{cert,key}.pem) into /work/tls/self without regenerating.
    ws sh -c 'set -e
        if [ ! -d /work/tls/self ] && [ -f /work/tls/cert.pem ] && [ -f /work/tls/key.pem ]; then
            mkdir -p /work/tls/self
            mv /work/tls/cert.pem /work/tls/key.pem /work/tls/self/
            cp /work/tls/self/cert.pem /work/tls/self/ca.pem
            echo "  moved existing self-signed certificate to /work/tls/self"
        fi' || fail "could not migrate existing certificate"

    self_state=$(ws sh -c '
        d=/work/tls/self; c=$d/cert.pem; k=$d/key.pem
        if [ ! -e $c ] && [ ! -e $k ]; then echo absent
        elif [ ! -e $c ] || [ ! -e $k ]; then echo partial
        elif ! openssl x509 -in $c -noout -checkend 0 >/dev/null 2>&1; then echo expired
        elif [ "$(openssl x509 -in $c -noout -pubkey 2>/dev/null)" != "$(openssl pkey -in $k -pubout 2>/dev/null)" ]; then echo mismatch
        else echo ok; fi')
    case "$self_state" in
        absent)
            info "generating self-signed localhost certificate from cmd/server/localhost.conf"
            ws sh -c '
                set -e; mkdir -p /work/tls/self; cd /work/tls/self
                openssl req -x509 -newkey rsa:4096 -nodes -days 365 \
                    -config /src/cmd/server/localhost.conf \
                    -keyout key.pem.tmp -out cert.pem.tmp 2>/dev/null
                chmod 644 key.pem.tmp cert.pem.tmp
                cp cert.pem.tmp ca.pem
                mv key.pem.tmp key.pem && mv cert.pem.tmp cert.pem' \
                || fail "certificate generation failed" ;;
        ok) ws sh -c '[ -e /work/tls/self/ca.pem ] || cp /work/tls/self/cert.pem /work/tls/self/ca.pem'
            info "reusing existing self-signed certificate" ;;
        partial)  fail "only one of /work/tls/self/cert.pem, key.pem exists. Remove the leftover file (in the workspace shell) to regenerate." ;;
        expired)  fail "/work/tls/self/cert.pem has expired. Delete /work/tls/self in the workspace shell to regenerate." ;;
        mismatch) fail "/work/tls/self/cert.pem and key.pem do not match. Delete /work/tls/self in the workspace shell to regenerate." ;;
        *) fail "could not inspect TLS files ($self_state)" ;;
    esac
    tls_select self
    info "self-signed: browsers will warn. For a trusted certificate see docker/tls/README.md"
}

# tls_select DIR: atomically repoint /work/tls/current to /work/tls/DIR and
# drop mkcert snapshots that are no longer active.
tls_select() {
    TLS_MODE=$1
    ws sh -c 'set -e; cd /work/tls
        ln -sfn "$1" current.tmp && mv -T current.tmp current
        for d in mkcert-*; do [ -d "$d" ] && [ "$d" != "$1" ] && rm -rf "$d"; done; true' sh "$1" \
        || fail "could not activate the $1 certificate"
}

# ---------------------------------------------------------------- web
web_up() { webcurl -fsS -o /dev/null https://localhost:8443/ 2>/dev/null; }

tls_note() {
    case "${TLS_MODE:-$(ws readlink /work/tls/current 2>/dev/null)}" in mkcert*) m=1 ;; *) m=0 ;; esac
    if [ "$m" = 1 ]; then
        echo "(mkcert certificate: trusted by browsers where mkcert -install ran)"
    else
        echo "(self-signed certificate: expect a browser warning; see docker/tls/README.md to avoid it)"
    fi
}

# Fixture-independent check that the database-backed search answers. Any
# search opens the database; the server exits if it cannot. Requires HTTP 200
# and a JSON body. No data is imported or changed.
web_probe() {
    resp=$(webcurl -sS -w '\n%{http_code}' --data-urlencode name=sst-readiness-probe \
           https://localhost:8443/searchN4L 2>&1) || resp="$resp
000"
    code=$(printf '%s\n' "$resp" | tail -n 1)
    if [ "$code" != "200" ] || ! printf '%s\n' "$resp" | sed '$d' | ws jq -e .Response >/dev/null 2>&1; then
        printf '%s\n' "$resp" | head -c 1000 >&2; echo >&2
        web_logs >&2 || true
        fail "web is up but its database-backed search endpoint did not answer"
    fi
}

# The web service runs http_server under docker/web-supervisor.sh, which
# restarts it when /work/bin/http_server is replaced or the active certificate
# changes, and records "<binary inode> <cert sha256>" in /work/run/web.running.
web_want() {
    ws sh -c 'echo "$(stat -c %i /work/bin/http_server) $(sha256sum < /work/tls/current/cert.pem | cut -c1-64)"'
}
web_state() { ws cat /work/run/web.running 2>/dev/null || true; }

# wait_web_reloaded SECONDS: until web runs the current binary + certificate.
wait_web_reloaded() {
    want=$(web_want)
    i=0
    until [ "$(web_state)" = "$want" ] && web_up; do
        i=$((i + 1))
        if [ "$i" -ge "$(( $1 * 2 ))" ]; then
            web_logs >&2 || true
            fail "web did not pick up the new binary/certificate within $1s"
        fi
        sleep 0.5
    done
    web_probe
}

# web_sync SECONDS: if a running web is behind (binary or certificate), wait for its reload.
web_sync() {
    st=$(web_state)
    if [ -n "$st" ] && [ "$st" != "$(web_want)" ]; then
        info "web is reloading the new binary/certificate"
        wait_web_reloaded "$1"
    fi
}

# rebuild_cmd NAME...: build, then wait for web to reload if http_server changed.
rebuild_cmd() {
    STAGE=build
    build_binaries "$@" || fail "build failed; existing binaries in /work/bin are unchanged"
    if [ "$SERVER_CHANGED" = "1" ]; then
        STAGE=web
        if [ -n "$(web_state)" ]; then
            info "waiting for web to reload the new http_server"
            wait_web_reloaded 30
            info "web reloaded and answering"
        else
            info "web is not running; it will use the new binary when started"
        fi
    fi
}

# ---------------------------------------------------------------- baseline
# Checks that the original examples/doors.n4l fixture answers via CLI and HTTPS.
# No -wipe, import, upload, container stop or volume change. (The application
# runs idempotent schema setup, e.g. CREATE OR REPLACE FUNCTION, on open.)
verify_baseline() {
    stage cli "searchN4L door"
    if ! cli_out=$(ws /work/bin/searchN4L door 2>&1); then
        printf '%s\n' "$cli_out" >&2
        fail "searchN4L exited nonzero"
    fi
    # Expect the orbit of "door" to list:   -    (leads to) - passage
    if ! { printf '%s\n' "$cli_out" | grep -q '^0: "door"' &&
           printf '%s\n' "$cli_out" | grep -Eq '\(leads to\) - passage[[:space:]]'; }; then
        printf '%s\n' "$cli_out" | head -60 >&2
        fixture_hint
        fail "searchN4L output does not show the door -> passage relationship"
    fi
    info "CLI result contains door -> passage"

    stage https "web search for door over HTTPS (web service, inside Docker)"
    if ! resp=$(webcurl -sS -w '\n%{http_code}' --data-urlencode name=door https://localhost:8443/searchN4L 2>&1); then
        printf '%s\n' "$resp" >&2
        web_logs >&2 || true
        fail "HTTPS request failed"
    fi
    code=$(printf '%s\n' "$resp" | tail -n 1)
    body=$(printf '%s\n' "$resp" | sed '$d')
    if [ "$code" != "200" ]; then
        printf 'HTTP %s\n%s\n' "$code" "$(printf '%s' "$body" | head -c 2000)" >&2
        web_logs >&2 || true
        fail "expected HTTP 200"
    fi
    check=$(printf '%s' "$body" | ws jq -r '
        if .Response != "Orbits" then "response=\(.Response)"
        elif ([.. | strings | select(test("^passage$"))] | length) > 0 then "ok"
        else "no-passage" end' 2>&1) || check="invalid-json: $check"
    if [ "$check" != "ok" ]; then
        printf 'HTTP %s\n%s\n' "$code" "$(printf '%s' "$body" | head -c 2000)" >&2
        fixture_hint
        fail "web search result check failed ($check)"
    fi
    info "HTTPS result: Response=Orbits, contains passage"
    stage done "baseline verified (CLI + HTTPS)"
}

fixture_hint() {
    if [ "$(marker_get fixture 2>/dev/null || true)" = "seeded" ] && ! fixture_present; then
        warn "baseline fixture no longer present (door -> passage missing in PostgreSQL)."
        warn "The environment may still be healthy; this only checks the original example."
    fi
}

# baseline_if_needed FORCE(0|1): run until it has passed once, then on request.
baseline_if_needed() {
    if [ "$1" = "1" ] || [ "$(marker_get baseline)" != "verified" ]; then
        stage baseline "checking the original doors.n4l baseline"
        ( verify_baseline ) || { STAGE=baseline; fail "baseline verification failed (see above)"; }
        if [ "$(marker_get baseline)" != "verified" ]; then
            marker_set baseline verified
            info "first baseline pass recorded (baseline=verified)"
        fi
    else
        stage baseline "skipped (passed before); run the verify command to check anytime"
    fi
}

# ---------------------------------------------------------------- watch
# watch_loop SRC_ROOT INTERVAL REBUILD_FUNCTION
# Polls SRC_ROOT and calls REBUILD_FUNCTION with the affected binaries. A
# change is a file with a newer timestamp OR a difference in the inventory of
# watched files: deletions, additions and renames (mv keeps the timestamp)
# would otherwise go unnoticed. A failed build is reported; the loop goes on.
watch_loop() {
    src=$1; interval=$2; rebuild_fn=$3
    tmp=$(mktemp -d "${TMPDIR:-/tmp}/sst-watch.XXXXXX")
    trap 'rm -rf "$tmp"' EXIT
    trap 'exit 130' INT TERM

    watch_dirs=""
    for x in $SST_TOOLS; do watch_dirs="$watch_dirs cmd/$(tool_dir "$x")"; done
    # watched_files [find tests...]: watched source files, sorted, one per line.
    watched_files() {
        ( cd "$src" && find pkg $watch_dirs go.mod go.sum -type f "$@" \
            \( -name '*.go' -o -name go.mod -o -name go.sum -o -path 'cmd/server/public/*' \) \
            ! -name '*~' ! -name '.#*' ! -name '*.swp' ! -name '*.swx' ! -name '*.tmp' \
            2>/dev/null | LC_ALL=C sort ) || true
    }
    # changes: "~ path" (modified), "+ path" (added), "- path" (removed)
    # relative to stamp and inventory; also writes the new inventory to inv.next.
    changes() {
        watched_files > "$tmp/inv.next"
        LC_ALL=C comm -13 "$tmp/inv" "$tmp/inv.next" | sed 's/^/+ /'
        LC_ALL=C comm -23 "$tmp/inv" "$tmp/inv.next" | sed 's/^/- /'
        watched_files -newer "$tmp/stamp" | LC_ALL=C comm -12 - "$tmp/inv" | sed 's/^/~ /'
    }

    touch "$tmp/stamp"
    watched_files > "$tmp/inv"

    echo "watching pkg/ and the cmd/ folders of all $(echo $SST_TOOLS | wc -w) tools (every ${interval}s); Ctrl-C to stop"
    while :; do
        sleep "$interval"
        [ -n "$(changes | head -n 1)" ] || continue

        sleep 0.5                       # let a burst of saves settle
        touch "$tmp/next"
        changes > "$tmp/changes"
        mv "$tmp/next" "$tmp/stamp"
        mv "$tmp/inv.next" "$tmp/inv"

        targets=""
        while read -r _ f; do
            case "$f" in
                pkg/*|go.mod|go.sum) t=$SST_TOOLS ;;
                cmd/server/*)        t=http_server ;;
                cmd/*/*)             d=${f#cmd/}; d=${d%%/*}; if is_tool "$d"; then t=$d; else t=""; fi ;;
                *) t="" ;;
            esac
            for x in $t; do
                case " $targets " in *" $x "*) ;; *) targets="$targets $x" ;; esac
            done
        done < "$tmp/changes"
        [ -n "$targets" ] || continue

        printf '\n[%s] changed: %s\n' "$(date +%H:%M:%S)" "$(head -n 5 "$tmp/changes" | tr '\n' ' ')"
        # shellcheck disable=SC2086
        if ( "$rebuild_fn" $targets ); then
            printf '[%s] ok:%s\n' "$(date +%H:%M:%S)" "$targets"
        else
            printf '[%s] BUILD FAILED - fix the error and save again (running binaries unchanged)\n' "$(date +%H:%M:%S)"
        fi
    done
}
