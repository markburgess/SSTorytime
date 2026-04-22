# TLS certificates

The SSTorytime HTTPS server terminates TLS locally using a self-signed
certificate it generates for itself on first build. This keeps the
developer path friction-free — no cert-authority account, no ACME
setup, no manual `openssl` invocation — at the cost of a browser warning
the first time you hit the server. This page documents the full
lifecycle: when certs are generated, what the script does, where they
live on disk, how long they're valid, and what to do for a real
deployment.

!!! warning "Self-signed is for local development only"
    The cert the bootstrap script produces is not trusted by any
    browser. It's fine for `localhost` testing; it's **not** fit for
    production. See [Production deployments](#production-deployments)
    below for the recommended reverse-proxy pattern.

## When certs are generated

The first time you run `make` in `src/server/` (or `make all` from the
repo root, which descends into `src/server/`), the server's Makefile
invokes
[`src/server/make_certificate`](https://github.com/markburgess/SSTorytime/blob/main/src/server/make_certificate)
before `go build`. Subsequent builds skip the generation step —
`make_certificate` exits early if `cert.pem` already exists:

```sh
if [ -f "cert.pem" ]; then
    echo The certificate files exist
    exit 0
fi
```

So the workflow is: first `make` creates the cert, every `make` after
that reuses it.

## What the script does

The script is six lines of shell. It invokes `openssl req` with the
X.509 flow:

```sh
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem \
            -days 365 -nodes -config localhost.conf
```

Breaking that down:

| Flag | Meaning |
|---|---|
| `-x509` | Emit a self-signed certificate, not a CSR. |
| `-newkey rsa:4096` | Generate a fresh 4096-bit RSA key pair. |
| `-keyout key.pem` | Write the private key here. |
| `-out cert.pem` | Write the certificate here. |
| `-days 365` | Validity period — one year. |
| `-nodes` | No DES — the key is stored unencrypted (so the server can read it without a passphrase). |
| `-config localhost.conf` | Use the shipped OpenSSL config for the subject + SANs. |

The config file is
[`src/server/localhost.conf`](https://github.com/markburgess/SSTorytime/blob/main/src/server/localhost.conf).
It sets a fixed subject (`C=NO`, `ST=Oslo`, `O=ChiTek-i`, `CN=localhost`)
and lists the Subject Alternative Names the cert is valid for:

```ini
[alt_names]
DNS.1 = localhost
IP.1  = 127.0.0.1
IP.2  = ::1
```

That SAN list is why the cert works for `https://localhost:8443`,
`https://127.0.0.1:8443`, and `https://[::1]:8443` out of the box.

!!! code-ref "See in code"
    Bootstrap script:
    [`src/server/make_certificate`](https://github.com/markburgess/SSTorytime/blob/main/src/server/make_certificate).
    OpenSSL config:
    [`src/server/localhost.conf`](https://github.com/markburgess/SSTorytime/blob/main/src/server/localhost.conf).

## Where certs live

After a successful build you have two files in `src/server/`:

```
src/server/
├── cert.pem        ← public certificate
└── key.pem         ← private key (unencrypted)
```

Both are gitignored — they never get committed. Each developer's build
produces their own pair.

The server process (built to `src/bin/http_server`) reads them via a
path relative to its own binary — `../server/cert.pem` and
`../server/key.pem` — so the conventional workflow (`cd src && ./bin/http_server`)
"just works" without any path configuration.

!!! warning "`key.pem` is unencrypted — set `chmod 600` immediately"
    Because the server has to start without a human typing a passphrase,
    the private key is stored in cleartext. Treat `src/server/key.pem`
    the same way you'd treat any other secret on disk — don't copy it to
    a shared directory, don't include it in a backup that leaves the
    host, and rotate it if a box is compromised.

    The `make_certificate` script does not set restrictive permissions
    on its output, so after first build run:

    ```bash
    chmod 600 src/server/key.pem
    chmod 644 src/server/cert.pem   # cert is public, key is not
    ```

    Confirm with `ls -l src/server/{cert,key}.pem` — `key.pem` should
    read `-rw-------`. A world-readable `key.pem` on a shared host means
    any local user can impersonate the server.

## Validity

- **365 days from the moment the cert was generated.** Once it expires,
  the server will still start, but every browser and HTTPS client will
  reject the connection.
- **No auto-renewal.** Nothing watches the expiry date. If you keep a
  long-lived dev box, set a reminder.

To check the current validity window:

```bash
openssl x509 -in src/server/cert.pem -noout -dates
```

Output looks like:

```
notBefore=Apr 20 10:00:00 2026 GMT
notAfter =Apr 20 10:00:00 2027 GMT
```

## Regenerating / rotating

### The simple way

Delete both files and rebuild:

```bash
rm src/server/cert.pem src/server/key.pem
(cd src/server && make)
```

The next `make` will re-run `make_certificate` and produce a fresh pair
good for another 365 days.

### Without a rebuild

You can regenerate the cert without re-compiling the Go server — just
re-run the OpenSSL command by hand from inside `src/server/`:

```bash
cd src/server
rm cert.pem key.pem
./make_certificate
```

Then restart the `http_server` process so it picks up the new files.
This is useful if you've tweaked `localhost.conf` (for example, adding
a new SAN for a custom hostname) and want the change to take effect
without touching Go.

### Custom key material

If you already have a cert/key pair you'd rather use (say, a cert
signed by your organisation's internal CA), simply drop them in as
`src/server/cert.pem` and `src/server/key.pem`. The server doesn't
verify where they came from — it just reads PEM. Make sure the key
permissions are strict (`chmod 600 src/server/key.pem`).

## Browser warnings

Every browser will show a "Not secure" / "Your connection is not
private" interstitial the first time you visit
`https://localhost:8443`. This is correct behaviour — the cert isn't
signed by a trusted authority, and the browser has no way to
distinguish a legitimate self-signed dev cert from a hostile one.

### Accepting for local development

- **Chrome / Edge / Brave:** click *Advanced* → *Proceed to localhost
  (unsafe)*.
- **Firefox:** click *Advanced* → *Accept the Risk and Continue*.
- **Safari:** click *Show Details* → *visit this website*. macOS may
  additionally prompt for your login password to install the cert in
  the system keychain.

Once accepted, the browser remembers the exception until the cert
changes (e.g. after regeneration). After rotation you'll get the prompt
once more.

### Trusting the cert system-wide (optional)

If you test the local server a lot and want to stop dismissing
warnings, you can install `cert.pem` into your OS / browser trust
store. Instructions vary by platform — search for "trust self-signed
cert on \[your OS\]" — but the mechanic is always the same: copy
`cert.pem` into a trusted-roots directory and refresh the store.

!!! tip "Mark the cert as localhost-only in your trust store"
    If your OS supports per-usage trust flags, limit the cert to
    `localhost` / TLS-server only. This avoids trusting the cert for
    arbitrary sites and mitigates the damage if the key file leaks.

## Production deployments

Self-signed certs are the wrong tool for production. A few reasons:

- No user will see anything but a scary warning.
- The cert isn't revocable through any public CA mechanism.
- The 365-day rotation has to be done by hand.
- Browsers and clients correctly refuse to submit credentials through it.

The recommended pattern is to put a proper reverse proxy in front of
`http_server` and let the proxy handle TLS:

| Proxy | Notes |
|---|---|
| [nginx](https://nginx.org/) | Mature, ubiquitous. Pair with [Certbot](https://certbot.eff.org/) for automatic Let's Encrypt rotation. |
| [Caddy](https://caddyserver.com/) | Zero-config HTTPS — Caddy asks Let's Encrypt for a cert on startup and auto-renews. Single binary. |
| [Traefik](https://traefik.io/) | Dynamic reverse proxy that integrates with Docker / Kubernetes and does ACME natively. |

All three integrate with [ACME](https://datatracker.ietf.org/doc/html/rfc8555)
(Let's Encrypt) so the cert renews itself every ~60 days without your
intervention.

A minimal Caddy config:

```caddy
sst.example.com {
    reverse_proxy localhost:8443 {
        transport http {
            tls_insecure_skip_verify
        }
    }
}
```

(The `tls_insecure_skip_verify` is on the proxy-to-backend hop, not the
public side — Caddy still enforces a real cert to the outside world.)

With this in place, disable the `:8443` bind address on the public
interface and let the proxy front everything. The self-signed
`cert.pem` inside `src/server/` is then only used on the loopback
interface between the proxy and the Go server.
