# TLS for the web server

## Default: no cert needed

```bash
./bin/sstorytime serve
# HTTP :8080
```

## Local HTTPS (`-tls`)

Missing `cert.pem` / `key.pem` are **auto-generated** (stdlib crypto; no openssl):

```bash
./bin/sstorytime serve -tls
# HTTPS :8443
```

Optional flags: `-cert`, `-key`, `-https-addr`, and `-http-addr` for an opt-in HTTP→HTTPS **307** redirect.

Optional openssl helper:

```bash
./internal/app/httpserver/make_certificate
```
