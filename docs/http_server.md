
# `http_server` — operator guide

The `http_server` binary is the HTTPS entry point to SSTorytime's graph. It
serves JSON to web clients, files to browsers, and (via a proxy) requests
from LLMs. This page is for people running the server; for the wire
protocol see [Web API](WebAPI.md), and for LLM wiring see
[MCP-SST](http-api/mcp-sst.md).

## Starting the server

After `make` from the repository root, the binary lives at
`src/bin/http_server`. It takes a single flag:

<pre>
./bin/http_server -resources /data/directory
</pre>

`-resources` is the root directory backing the `/Resources/` URL prefix
— any file under that directory becomes reachable at the matching path
(useful for images, PDFs, and media referenced from your N4L notes). The
default is `/mnt`. If you share `/mnt/Recordings` as the resource root,
then the file

<pre>
/mnt/Recordings/Rush/Presto/Folder.jpg
</pre>

is served at

<pre>
https://localhost:8443/Resources/Rush/Presto/Folder.jpg
</pre>

Use `-v` for verbose logging. There is no daemon mode — run it under
`systemd`, `supervisord`, a container orchestrator, or a terminal
multiplexer, whichever your operations story prefers.

## Ports and TLS

The server binds **HTTPS on port 8443** as its primary endpoint. A secondary
listener on port 8080 issues a 301 redirect to the HTTPS URL, so any plaintext
HTTP client is transparently upgraded. See the listener setup at
[`src/server/http_server.go:123-166`](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go#L123-L166).

TLS certificates are auto-generated on first build of `src/server/` by the
[`make_certificate`](https://github.com/markburgess/SSTorytime/blob/main/src/server/make_certificate)
script (self-signed RSA-4096, 365-day validity). The running server reads them
from `../server/cert.pem` / `../server/key.pem` relative to the binary at
`src/bin/http_server`, so the default working directory for launching the
server is `src/bin/`.

## Rotating the TLS certificate

The certificate expires 365 days after issue. To replace it:

1. Stop the server (`Ctrl-C` or `systemctl stop …`).
2. Delete the old pair: `rm src/server/cert.pem src/server/key.pem`.
3. Regenerate: `cd src/server && sh make_certificate`.
4. Restart the server.

The `make_certificate` script is deliberately a no-op if `cert.pem`
already exists — that's why step 2 is required. Configuration (CN, SANs)
is in
[`src/server/localhost.conf`](https://github.com/markburgess/SSTorytime/blob/main/src/server/localhost.conf);
edit it before regenerating if you need additional DNS names or IPs.

For any deployment outside `localhost`, replace the self-signed pair with
a real certificate from your CA — or, preferably, terminate TLS at a
reverse proxy (see "Production hardening" below).

## Logging

The server logs to standard output:

- A banner of the resource root on startup
  ([`http_server.go:115-118`](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go#L115-L118)).
- One line per incoming search with the decoded command
  ([`http_server.go:229`](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go#L229)).
- One line per redirect from `:8080` to `:8443`
  ([`http_server.go:127`](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go#L127)).
- Upload successes and failures with the resulting file path.
- A "Shutting down servers…" line plus a "Server exited properly" line on clean exit.

Pipe stdout into your logging pipeline (journald, Loki, Cloud Logging)
with the usual one-liner.

## Graceful shutdown

On `SIGINT` or `SIGTERM`, both listeners are given a 10-second context to
drain in-flight requests before the process exits
([`http_server.go:136-152`](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go#L136-L152)).
Long-running `/searchN4L` queries that exceed this window are cut off; set
your client timeouts below 10 seconds to fail fast.

The persistent PostgreSQL connection opened by
[`SST.Open`](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go#L89)
is closed implicitly when the process exits; there is no explicit `Close`
in the signal handler.

## Production hardening

The default binary is suitable for local development and trusted networks.
For anything beyond that:

- **Put a real reverse proxy in front.** Terminate TLS with a CA-issued
  certificate at nginx, Caddy, or Traefik, and have SSTorytime listen on
  a loopback-only HTTP port behind it. The self-signed cert and the
  permissive CORS policy then only affect the hop inside the trusted
  network.
- **Add authentication at the proxy.** The server itself has no notion of
  users or API keys. HTTP Basic, mTLS, or an OIDC side-car are all
  straightforward.
- **Set body and rate limits.** `/Upload` is capped at 32 MB by the
  multipart parser, but there is no rate limit on `/searchN4L`.
- **Isolate the `-resources` tree.** The file server is recursive —
  whatever you point it at is readable. A dedicated directory, not
  `/` or `/home`, is the safe default.

## See also

- [Web API](WebAPI.md) — the JSON protocol spoken by every endpoint.
- [MCP-SST](http-api/mcp-sst.md) — how LLM clients talk to this server.
