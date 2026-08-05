# Setting up TLS for the web server

From the project root (or any directory; PEMs are written to the **current working directory**):

```bash
./internal/app/httpserver/make_certificate
```

This runs `openssl` with `internal/app/httpserver/localhost.conf` (localhost + 127.0.0.1 SANs).
Equivalent manual command:

```bash
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem \
  -days 365 -nodes -config internal/app/httpserver/localhost.conf
```

Then (after `make build`):

```bash
./bin/http_server -cert cert.pem -key key.pem
# or
./bin/sstorytime serve -cert cert.pem -key key.pem
```

Browsers will warn on self-signed certificates; that is expected for local development.
