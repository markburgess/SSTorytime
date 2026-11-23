#!/bin/sh
# Automated fresh-installation test for the Docker environment (maintainers/CI).
#
#   ./docker/test-fresh-install.sh [--keep]
#
# Provisions a brand-new, isolated Compose project (own volumes, network and
# host port; default port 18443, override with SST_TEST_PORT), then checks:
#   1. first start: build, one-time fixture import, TLS, web, baseline
#   2. baseline verification on demand
#   3. own data added, stop, resume: no re-import, data/certificate/marker kept
# The test project is removed at the end (down -v of THAT project only);
# --keep leaves it for inspection. An existing environment is never touched.
#
# Needs a POSIX shell and the Docker CLI on the host (unlike normal use, which
# needs only Docker): it has to orchestrate Compose from the outside.

set -eu
cd "$(dirname -- "$0")/.."

KEEP=0
for a in "$@"; do
    case "$a" in
        --keep) KEEP=1 ;;
        -h|--help) sed -n '2,15p' "$0"; exit 0 ;;
        *) echo "unknown option: $a" >&2; exit 2 ;;
    esac
done

PROJECT="sst-test-$(date +%Y%m%d%H%M%S)-$$"
PORT=${SST_TEST_PORT:-18443}
export COMPOSE_PROJECT_NAME="$PROJECT" SST_WEB_PORT="$PORT"
C="docker compose -f docker/compose.yaml"
LOG=$(mktemp -d "${TMPDIR:-/tmp}/sst-test.XXXXXX")
PASS=0; FAIL=0

case "$PROJECT" in sst-test-*) ;; *) echo "refusing: unexpected project name $PROJECT" >&2; exit 2 ;; esac

cleanup() {
    rc=$?
    if [ "$KEEP" = 1 ]; then
        echo "kept: COMPOSE_PROJECT_NAME=$PROJECT SST_WEB_PORT=$PORT $C ...   (logs: $LOG)"
    else
        $C --profile tools down -v --remove-orphans >/dev/null 2>&1 || true
        rm -rf "$LOG"
    fi
    exit "$rc"
}
trap cleanup EXIT
trap 'exit 130' INT TERM

ok()   { PASS=$((PASS + 1)); printf '  ok    %s\n' "$1"; }
bad()  { FAIL=$((FAIL + 1)); printf '  FAIL  %s\n' "$1"; [ -z "${2:-}" ] || printf '%s\n' "$2" | sed 's/^/        /' | tail -n 30; }
check() { desc=$1; shift; if "$@" >/dev/null 2>&1; then ok "$desc"; else bad "$desc"; fi; }
eq()   { if [ "$2" = "$3" ]; then ok "$1 ($2)"; else bad "$1: expected '$3', got '$2'"; fi; }
has()  { if grep -qE -- "$2" "$3"; then ok "$1"; else bad "$1: no match for /$2/" "$(cat "$3")"; fi; }
hasnt(){ if grep -qE -- "$2" "$3"; then bad "$1: unexpected /$2/" "$(grep -E -- "$2" "$3")"; else ok "$1"; fi; }
sql()  { $C exec -T db psql -h 127.0.0.1 -U sstoryline -d sstoryline -qtAX -c "$1"; }
ws()   { $C exec -T -w /work/run workspace "$@"; }
run_tool() {  # run_tool NAME LOGFILE -> exit code; output in LOGFILE
    rc=0; $C run --rm "$1" > "$2" 2>&1 || rc=$?; return "$rc"
}
section() { printf '\n== %s\n' "$*"; }

echo "fresh-install test: project $PROJECT, host port $PORT"
[ -z "$($C ps -a -q 2>/dev/null)" ] || { echo "project $PROJECT already exists" >&2; exit 2; }

# ------------------------------------------------------------------ 1
section "1. first start on empty volumes"
t0=$(date +%s)
if run_tool start "$LOG/start1.log"; then ok "start exits 0 ($(( $(date +%s) - t0 ))s)"
else bad "start exits 0" "$(cat "$LOG/start1.log")"; echo; echo "RESULT: FAIL (first start broken)"; exit 1; fi
has   "binaries built"                       '^\[build\] compiling'                      "$LOG/start1.log"
has   "fixture imported and confirmed"       'fixture imported and confirmed'            "$LOG/start1.log"
has   "certificate prepared"                 'mkcert certificate|self-signed'             "$LOG/start1.log"
has   "web search endpoint answering"        'search endpoint working'                   "$LOG/start1.log"
has   "baseline verified and recorded"       'first baseline pass recorded'              "$LOG/start1.log"
eq    "marker fixture"   "$(sql "select value from sst_install_test.state where key='fixture'")"  seeded
eq    "marker baseline"  "$(sql "select value from sst_install_test.state where key='baseline'")" verified
nodes1=$(sql "select count(*) from node")
check "doors.n4l nodes present ($nodes1)"   test "$nodes1" -gt 0
check "web published on 127.0.0.1:$PORT only" sh -c "$C port web 8443 | grep -qx '127.0.0.1:$PORT'"
if command -v curl >/dev/null 2>&1; then
    $C exec -T workspace cat /work/tls/current/ca.pem > "$LOG/ca.pem"
    check "host HTTPS request to https://localhost:$PORT (certificate verified)" \
        curl -fsS --cacert "$LOG/ca.pem" -o /dev/null "https://localhost:$PORT/"
else
    echo "  skip  host HTTPS request (no curl on host)"
fi

# ------------------------------------------------------------------ 2
section "2. baseline verification on demand"
if run_tool verify "$LOG/verify.log"; then ok "verify exits 0"; else bad "verify exits 0" "$(cat "$LOG/verify.log")"; fi
has   "CLI baseline"                         'CLI result contains door -> passage'        "$LOG/verify.log"
has   "HTTPS baseline"                       'HTTPS result: Response=Orbits'              "$LOG/verify.log"
eq    "verify leaves node count unchanged"   "$(sql "select count(*) from node")"         "$nodes1"

# ------------------------------------------------------------------ 3
section "3. own data, stop, resume"
ws sh -c 'printf -- "- fresh install test\n\n sst_test_alpha (fwd) sst_test_omega\n" > /work/run/sst-test.n4l'
if ws /work/bin/N4L -u /work/run/sst-test.n4l > "$LOG/import.log" 2>&1; then ok "import own data (no -wipe)"
else bad "import own data (no -wipe)" "$(cat "$LOG/import.log")"; fi
nodes2=$(sql "select count(*) from node")
check "node count grew ($nodes1 -> $nodes2)" test "$nodes2" -gt "$nodes1"
cert1=$(ws sha256sum /work/tls/current/cert.pem | cut -c1-64)
seeded1=$(sql "select updated_at from sst_install_test.state where key='fixture'")

if $C stop > "$LOG/stop.log" 2>&1; then ok "stop"; else bad "stop" "$(cat "$LOG/stop.log")"; fi
check "nothing running after stop" sh -c "[ -z \"\$($C ps -q)\" ]"

t0=$(date +%s)
if run_tool start "$LOG/start2.log"; then ok "resume: start exits 0 ($(( $(date +%s) - t0 ))s)"
else bad "resume: start exits 0" "$(cat "$LOG/start2.log")"; fi
has   "resume recognizes initialized database" 'database already initialized'           "$LOG/start2.log"
hasnt "resume does not import again"          'fixture imported|N4L -u -wipe'            "$LOG/start2.log"
hasnt "resume does not rebuild"               '^\[build\] compiling'                    "$LOG/start2.log"
has   "resume skips baseline (passed before)" '^\[baseline\] skipped'                    "$LOG/start2.log"
eq    "node count retained"                   "$(sql "select count(*) from node")"       "$nodes2"
eq    "fixture marker unchanged"              "$(sql "select updated_at from sst_install_test.state where key='fixture'")" "$seeded1"
eq    "certificate unchanged"                 "$(ws sha256sum /work/tls/current/cert.pem | cut -c1-64)" "$cert1"
if ws /work/bin/searchN4L sst_test_alpha > "$LOG/search.log" 2>&1 && grep -q 'sst_test_omega' "$LOG/search.log"
then ok "own data searchable after resume (CLI)"; else bad "own data searchable after resume (CLI)" "$(cat "$LOG/search.log")"; fi
if $C exec -T workspace curl -fsS --cacert /work/tls/current/ca.pem --connect-to localhost:8443:web:8443 \
       --data-urlencode name=sst_test_alpha https://localhost:8443/searchN4L 2>&1 | grep -q sst_test_omega
then ok "own data searchable after resume (HTTPS)"; else bad "own data searchable after resume (HTTPS)"; fi
if run_tool verify "$LOG/verify2.log"; then ok "baseline still verifies after resume"; else bad "baseline still verifies after resume" "$(cat "$LOG/verify2.log")"; fi

# ------------------------------------------------------------------ result
printf '\nRESULT: %s  (%d passed, %d failed)\n' "$([ "$FAIL" = 0 ] && echo PASS || echo FAIL)" "$PASS" "$FAIL"
[ "$FAIL" = 0 ]
