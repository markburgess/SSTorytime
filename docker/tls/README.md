# Trusted localhost certificate (optional, mkcert)

Without files here, the environment generates a self-signed certificate and
browsers warn about it. To get **https://localhost:8443** (or your
`SST_WEB_PORT`) **without a warning**,
put a certificate from [mkcert](https://github.com/FiloSottile/mkcert) here.
mkcert creates a local CA, and your browser trusts it once it is installed.

Run mkcert **where your browser runs**. On Windows that means PowerShell, even
if the checkout lives in WSL: `mkcert -install` inside WSL would only trust it
for Linux programs.

```powershell
# one time: install mkcert (e.g. winget install FiloSottile.mkcert), then
mkcert -install

# go to this directory (docker\tls) of your checkout, e.g.
#   cd C:\Users\<you>\SSTorytime\docker\tls                      (checkout on C:)
#   cd \\wsl.localhost\<distro>\home\<you>\SSTorytime\docker\tls   (checkout in WSL)
mkcert -cert-file cert.pem -key-file key.pem localhost 127.0.0.1 ::1
copy "$(mkcert -CAROOT)\rootCA.pem" .
```

If PowerShell says `mkcert` is not recognized right after installing it, open a
new PowerShell window: the open one doesn't know the new program yet.

On Linux or macOS the commands are the same, with
`cp "$(mkcert -CAROOT)/rootCA.pem" .` as the last step. On Linux, `mkcert -install`
needs `certutil` (package `libnss3-tools` on Debian/Ubuntu) so that Firefox and
Chrome trust the CA; see the mkcert documentation for other distributions.

Then run `docker compose -f docker/compose.yaml run --rm start`. It checks that
`cert.pem` is unexpired, matches `key.pem`, covers `localhost` and is signed by
`rootCA.pem`. Only then does it copy them into the environment's work volume
and switch the web server over, with no data loss and no restart by hand. Until
the next start, the server keeps the last validated certificate, so
half-regenerated files are never loaded. To go back to self-signed, delete the
three files and run start again.

| File | What it is | Secret? |
|---|---|---|
| `cert.pem` | server certificate for localhost | no |
| `key.pem` | server private key | yes, local only (gitignored) |
| `rootCA.pem` | mkcert's CA **certificate**, used by the verify step to trust the server | no |

**Never** copy `rootCA-key.pem`, the CA's private key, here. Anyone with it can
create certificates your browser trusts for any site. start refuses to run if
it finds that file here. mkcert certificates last about two years; regenerate
them when start reports them as expired.

Firefox on Windows may keep its own certificate store and still warn. In that
case enable `security.enterprise_roots.enabled` in `about:config`.
