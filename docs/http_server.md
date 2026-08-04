
# Web server (`sstorytime serve` / `http_server`)

The server is a generic browsing interface and JSON Web API. It is not meant to
be the last word on browsing the graph; applications may grow custom UIs. This
document describes how to run it and how paths map to files.

## Run

```text
go build -o bin/sstorytime ./cmd/sstorytime
./bin/sstorytime serve
# multi-call: ln -s sstorytime http_server && ./http_server
```

**Default: plain HTTP** on `:8080` (reverse-proxy / ACME-friendly; the app never
does ACME itself).

| Mode | Command | Listeners |
|------|---------|-----------|
| Production / proxy | `sstorytime serve` | HTTP `:8080` (app) |
| Local TLS (upstream-like) | `sstorytime serve --tls` | HTTPS `:8443` (app) + HTTP `:8080` → redirect |

### Certificates (`--tls` only)

If `cert.pem` / `key.pem` are missing, a **self-signed localhost** cert is written
with Go `crypto/x509` (no openssl). Browsers will warn — expected for dev.

For real deployments: terminate TLS on the reverse proxy and keep the default
HTTP backend, or pass your own PEMs with `--tls --cert … --key …`.

```text
sstorytime serve --addr :8080
sstorytime serve --tls --https-addr :8443
```

### Resources directory

```text
sstorytime serve --resources /data/directory
```

That path is the root for `/Resources/…` URLs (images, audio, PDFs, etc.).

Example: share `/mnt/Recordings`, start with `--resources /mnt/Recordings`, then
disk file `/mnt/Recordings/Rush/Presto/Folder.jpg` is served as:

```text
https://localhost:8443/Resources/Rush/Presto/Folder.jpg
```

(HTTP on :8080 redirects to the HTTPS URL.)

## Four search formats

The web server renders four different kinds of page.

* Ad hoc topic view, showing the orbits of random search sets (e.g. `brain&!whale)
* Page notes (N4L view, e.g. `\notes chinese`)
* Story/Sequence view (`\seq astronomy` or `\story (huli)`)
* Path solutions (`\from` a set of nodes `\to` a set of nodes).

