
# `http_server` / `sstorytime serve` and web

The http server provided is a generic browsing interface. It isn't meant to be the last
word on browsing the graph. In principle, every application might have its own custom
interface. This web page illustrates the Web API and is used to develop our thinking around
graphs.

Build from the project root with `make build` (or `go build -o bin/sstorytime ./cmd/sstorytime`).

## Listen modes

| Mode | Command | Ports |
|------|---------|--------|
| Plain HTTP (default) | `./bin/http_server` or `./bin/sstorytime serve` | `-addr` (default `:8080`) only |
| HTTPS | `… serve -tls` | `-https-addr` (default `:8443`) only; self-signed cert if missing |
| HTTPS + redirect | `… serve -tls -http-addr :8080` | HTTPS + optional HTTP **307** temporary redirect |

Default plain HTTP is what you want behind a reverse proxy (proxy terminates TLS). The app does not speak ACME.

## TLS certificates (`-tls` only)

If `-cert` / `-key` files are missing, the server **generates** a self-signed localhost certificate
with Go `crypto/x509` (no openssl). Paths default to `cert.pem` / `key.pem` in the process CWD.

```bash
./bin/sstorytime serve -tls
# or with explicit paths
./bin/sstorytime serve -tls -cert /path/cert.pem -key /path/key.pem
```

Optional openssl helper (still available):

```bash
./internal/app/httpserver/make_certificate
```

Browsers will warn on self-signed certificates; that is expected for local development.

## Resources directory

```bash
./bin/http_server -resources /data/directory
```

For example, `/mnt/Recordings/Rush/Presto/Folder.jpg` maps to
`http://localhost:8080/Resources/Rush/Presto/Folder.jpg` when started with
`-resources /mnt/Recordings`.

## Four search formats

The web server renders four different kinds of page.

* Ad hoc topic view, showing the orbits of random search sets (e.g. `brain&!whale)
* Page notes (N4L view, e.g. `\notes chinese`)
* Story/Sequence view (`\seq astronomy` or `\story (huli)`)
* Path solutions (`\from` a set of nodes `\to` a set of nodes).
