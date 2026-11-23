#!/bin/sh
# Entry point of the `web` service: runs http_server and restarts it when
#  - /work/bin/http_server is replaced (rebuild publishes by atomic rename,
#    so a new inode means a new binary), or
#  - the active certificate changes (/work/tls/current: mkcert or self-signed).
# Records "<binary inode> <cert sha256>" in /work/run/web.running so the
# rebuild/start tools can wait for the reload.
# If http_server exits on its own, the container exits too (visible failure).

set -eu
BIN=/work/bin/http_server
TLS=/work/tls/current
# On a fresh install `run --rm start` builds the binary, initializes the
# database and creates the certificate while this container is already up:
# wait for all three. The database must carry the install marker
# (fixture=seeded): http_server alters the schema of whatever database it
# opens, so it must never start against an unrecognized one.
db_ready() {
    [ "$(psql -qtAX -c "select case when to_regclass('sst_install_test.state') is null then ''
        else coalesce((select value from sst_install_test.state where key='fixture'),'') end" 2>/dev/null)" = seeded ]
}
waited=0
until [ -e "$BIN" ] && [ -r "$TLS/cert.pem" ] && [ -r "$TLS/key.pem" ] && db_ready; do
    [ "$waited" = 1 ] || echo "web-supervisor: waiting for setup (binary, certificate, initialized database); run: docker compose -f docker/compose.yaml run --rm start"
    waited=1
    sleep 1
done
mkdir -p /work/run
cd /work/run

state() { echo "$(stat -c %i "$BIN" 2>/dev/null) $(sha256sum < "$TLS/cert.pem" 2>/dev/null | cut -c1-64)"; }

pid=""
start_server() {
    running=$(state)
    "$BIN" -https :8443 -cert "$TLS/cert.pem" -key "$TLS/key.pem" -resources /src/examples &
    pid=$!
    echo "$running" > /work/run/web.running
    echo "web-supervisor: started http_server (pid $pid, $(readlink "$TLS") certificate)"
}
stop_server() {
    [ -n "$pid" ] || return 0
    kill -TERM "$pid" 2>/dev/null || true
    wait "$pid" 2>/dev/null || true
    pid=""
}
trap 'stop_server; rm -f /work/run/web.running; exit 0' TERM INT

start_server
while :; do
    sleep 1 &
    wait $! || true
    if ! kill -0 "$pid" 2>/dev/null; then
        rc=0; wait "$pid" || rc=$?
        rm -f /work/run/web.running
        echo "web-supervisor: http_server exited (status $rc)" >&2
        exit "${rc:-1}"
    fi
    # Only reload onto a complete, readable new state (ignore half-written files).
    [ -e "$BIN" ] && [ -r "$TLS/cert.pem" ] && [ -r "$TLS/key.pem" ] || continue
    cur=$(state)
    if [ "$cur" != "$running" ]; then
        echo "web-supervisor: new binary or certificate detected; restarting"
        stop_server
        start_server
    fi
done
