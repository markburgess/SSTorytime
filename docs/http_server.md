
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

**Default: one plain HTTP port** on `:8080` (reverse-proxy / ACME on the proxy;
the app never does ACME and never opens a second port).

| Mode | Command | Listeners |
|------|---------|-----------|
| Proxy / normal | `sstorytime serve` | HTTP `:8080` only |
| Local HTTPS | `sstorytime serve --tls` | HTTPS `:8443` only |
| Dual + redirect | `sstorytime serve --tls --http-addr :8080` | HTTPS + HTTP 301 (only if 8443 is open end-to-end) |

If HTTP redirects to HTTPS but the firewall drops the HTTPS port, browsers
**hang** after the 301. Prefer single-port modes.

### Certificates (`--tls` only)

Missing `cert.pem` / `key.pem` → self-signed localhost via Go `crypto/x509`.

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

