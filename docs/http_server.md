
# `http_server` and web

The http server provided is a generic browsing interface. It isn't meant to be the last
word on browsing the graph. In principle, every application might have its own custom
interface. This web page illustrates the Web API and is used to develop our thinking around
graphs.

Build from the project root (`make build` or `go build -o bin/http_server ./cmd/http_server`).

## TLS certificates

For local HTTPS, create a self-signed cert (writes `cert.pem` and `key.pem` in the **current working directory**):

```bash
./cmd/http_server/make_certificate
```

The script always loads `localhost.conf` from its own directory, so you can run it from anywhere.
Pass the PEMs to the server:

```bash
./bin/http_server -cert cert.pem -key key.pem
# defaults are already cert.pem / key.pem in CWD
./bin/http_server
```

## Resources directory

The web server accepts a resources root for file paths referenced in URLs (images, documents, etc.):

```bash
./bin/http_server -resources /data/directory
```

For example, if we share a folder called `/mnt/Recordings`:

```bash
./bin/http_server -resources /mnt/Recordings
```

then a disk file

```text
/mnt/Recordings/Rush/Presto/Folder.jpg
```

maps an image reference

```text
/Resources/Rush/Presto/Folder.jpg
```

to the URL

```text
http://localhost:8080/Resources/Rush/Presto/Folder.jpg
```

* HTTP on port **8080**; HTTPS on **8443** (with `-cert` / `-key`).

## Four search formats

The web server renders four different kinds of page.

* Ad hoc topic view, showing the orbits of random search sets (e.g. `brain&!whale)
* Page notes (N4L view, e.g. `\notes chinese`)
* Story/Sequence view (`\seq astronomy` or `\story (huli)`)
* Path solutions (`\from` a set of nodes `\to` a set of nodes).
