
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

Default (same idea as classic upstream):

| Listener | Role |
|----------|------|
| **https://localhost:8443** | UI + API (TLS) |
| **http://localhost:8080** | 301 redirect to HTTPS |

### Certificates

If `cert.pem` and `key.pem` are not present in the working directory, the server
**writes a self-signed localhost certificate** using the Go standard library
(`crypto/x509` + ECDSA P-256). You do not need `openssl` or the old
`make_certificate` script.

Browsers will warn about the self-signed cert — accept the exception for local
development. For production, put real PEMs in place (or use `--cert` / `--key`)
or put a reverse proxy in front and run with `--http-only`.

```text
sstorytime serve --cert /etc/ssl/sst.pem --key /etc/ssl/sst-key.pem
sstorytime serve --http-only --http-addr :8080   # plain HTTP only
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

