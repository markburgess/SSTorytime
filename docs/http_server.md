
# `http_server` and web

The http server provided is a generic browsing interface. It isn't meant to be the last
word on browsing the graph. In principle, every application might have its own custom
interface. This web page illustrates the Web API and is used to develop our thinking around
graphs.

The web server has a single argument:
<pre>
./bin/http_server -resources /data/directory
</pre>
This is a directory path which serves as a root for any file paths referenced in URLs, e.g.
where images of documents may be cached in order to be accessible from links rendered in the
browser. It may include any kind of MIME type, such as music files, images, documents etc.

For example, if we share a folder called `/mnt/Recordings`, then start the server
<pre>
./bin/http_server -resources /mnt/Recordings
</pre>
which leads to a disk file
<pre>
/mnt/Recordings/Rush/Presto/Folder.jpg
</pre>
which maps an image reference
<pre>
/Resources/Rush/Presto/Folder.jpg
</pre>
to the URL
<pre>
https://localhost:8443/Resources/Rush/Presto/Folder.jpg
</pre>

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

For production deployments, terminate TLS at an external reverse proxy
(nginx, Caddy, Traefik) and have the SSTorytime server listen on a
loopback-only HTTP port behind it.

## Four search formats

The web server renders four different kinds of page.

* Ad hoc topic view, showing the orbits of random search sets (e.g. `brain&!whale)
* Page notes (N4L view, e.g. `\notes chinese`)
* Story/Sequence view (`\seq astronomy` or `\story (huli)`)
* Path solutions (`\from` a set of nodes `\to` a set of nodes).

