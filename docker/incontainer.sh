#!/bin/sh
# Entry point of the Compose tool services; runs inside the containers, so the
# host needs only Docker. Logic lives in core.sh.
#
#   start     setup (build if needed, first-use DB init, TLS), wait for web,
#             readiness, baseline until it has passed once   (run --rm start)
#   verify    baseline check on demand                       (run --rm verify)
#   rebuild   [tool...]  (default: all tools)                (run --rm rebuild)
#   watch     rebuild on source change                       (run --rm watch)
#   inspect   [duplicates | search "<term>"]  read-only views (run --rm inspect)

set -eu

C_HINT="docker compose -f docker/compose.yaml"
sql() { psql -v ON_ERROR_STOP=1 -qtAX -c "$1"; }      # PGHOST etc. from compose env
ws()  ( mkdir -p /work/run && cd /work/run && exec "$@" )
# The certificate names localhost; connect to the web service but keep
# verifying the certificate against that name.
webcurl() { curl --cacert /work/tls/current/ca.pem --connect-to localhost:8443:web:8443 "$@"; }
web_logs() { echo "  See: $C_HINT logs --tail 50 web"; }

. /src/docker/core.sh

# Wait until web (whose supervisor waits for binary + certificate) answers.
wait_web_ready() {
    # A web container first created while host port 8443 was busy can be left
    # by Docker without a network: it runs but is unreachable from here.
    i=0
    until getent hosts web >/dev/null 2>&1; do
        i=$((i + 1))
        [ "$i" -lt 10 ] || fail "the web container is not on the Compose network.
  This happens when it was first created while host port 8443 was busy.
  Recreate it (no data is lost), then run start again:
    $C_HINT up -d --force-recreate web"
        sleep 0.5
    done
    i=0
    until web_up; do
        i=$((i + 1))
        if [ "$i" -ge "$(( $1 * 2 ))" ]; then
            web_logs
            fail "web did not answer within $1s"
        fi
        sleep 0.5
    done
}

usage() {
    echo "SSTorytime Docker environment (run from the repository root):"
    echo "  $C_HINT run --rm start      provision or resume"
    echo "  $C_HINT run --rm verify     original doors.n4l baseline"
    echo "  $C_HINT run --rm rebuild [tool...]   (default: all; tools: $SST_TOOLS)"
    echo "  $C_HINT run --rm watch      rebuild on save"
    echo "  $C_HINT run --rm inspect [duplicates | search \"<term>\"]   read-only views"
}

cmd=${1:-}; [ "$#" -gt 0 ] && shift
case "$cmd" in
    start)
        stage services "database"
        wait_db 90
        stage dirs "preparing /work"
        ensure_dirs
        if [ "${SST_REBUILD:-0}" = "1" ] || ! binaries_present; then
            stage build "compiling into /work/bin/.staging"
            build_binaries $SST_TOOLS || fail "build failed; existing binaries in /work/bin are unchanged"
        else
            stage build "binaries present; skipping (run --rm rebuild to recompile)"
        fi
        stage initialize "checking database state"
        initialize_db
        stage tls "checking certificates"
        ensure_tls
        stage web "waiting for the web server"
        wait_web_ready 60
        web_sync 30
        stage ready "checking the web server can reach PostgreSQL"
        web_probe
        info "db: reachable   web: answering, search endpoint working"
        baseline_if_needed "${SST_VERIFY:-0}"
        printf '\nSSTorytime is running: https://localhost:%s\n  %s\n' "${SST_WEB_PORT:-8443}" "$(tls_note)"
        printf '  First step: open that address (https://, not http://) and search for: door\n'
        printf '  Tutorial examples: see "Loading the tutorial examples" in docker/README.md\n\n'
        printf '  Shell:      %s exec workspace bash\n' "$C_HINT"
        printf '  Rebuild:    %s run --rm rebuild [tool...]   (default: all)\n' "$C_HINT"
        printf '  Watch:      %s run --rm watch\n' "$C_HINT"
        printf '  Verify:     %s run --rm verify\n' "$C_HINT"
        printf '  Stop:       %s stop      (keeps all data)\n' "$C_HINT"
        printf '  FULL RESET: %s down -v   (DELETES all data and certificates)\n' "$C_HINT"
        ;;
    verify)
        stage services "checking web"
        web_up || fail "web is not reachable; start with: $C_HINT run --rm start"
        verify_baseline ;;
    rebuild)
        # shellcheck disable=SC2086
        [ "$#" -gt 0 ] || set -- $SST_TOOLS
        for t in "$@"; do
            is_tool "$t" || { echo "unknown tool: $t (tools: $SST_TOOLS)" >&2; exit 2; }
        done
        rebuild_cmd "$@" ;;
    watch)
        watch_loop /src "${SST_WATCH_INTERVAL:-1}" rebuild_cmd ;;
    inspect)
        . /src/docker/inspect.sh
        case "${1:-}" in
            "")         inspect_overview ;;
            duplicates) inspect_duplicates ;;
            search)     shift; inspect_search "$*" ;;
            *)          echo "usage: inspect [duplicates | search \"<term>\"]" >&2; exit 2 ;;
        esac ;;
    *)
        usage; exit 2 ;;
esac
