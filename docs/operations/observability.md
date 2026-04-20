# Observability

SSTorytime's observability surface is deliberately minimal. This page
documents what you have, not what you wish you had. If your deployment
requires Prometheus scrapes, distributed traces, or structured JSON
logs, you will need to add them — they are not in the box.

## 1. What is *not* there

!!! warning "No `/health`, no `/metrics`, no structured logging"
    The HTTPS server at
    [`src/server/http_server.go`](https://github.com/markburgess/SSTorytime/blob/main/src/server/http_server.go)
    exposes the HTML UI and a handful of JSON query endpoints — see
    [HTTP API walkthrough](../http-api/mcp-sst.md) — and nothing else.
    There is specifically:

    - **No `/health` or `/ready` endpoint.** The server does not publish
      its own liveness. Use the transport-level signals under
      [Liveness today](#liveness-today) instead.
    - **No `/metrics` endpoint.** There is no Prometheus exposition
      anywhere in the tree. Adding one means wiring `prometheus/client_golang`
      into the HTTP handler chain, which has not been done.
    - **No structured logs.** Handler code uses `fmt.Println` and
      `log.Println` — free-form lines to stdout. There is no JSON field
      structure, no request ID, no level discrimination beyond the
      choice of function call.
    - **No trace propagation.** W3C `traceparent` headers are not
      forwarded to or from the database; the cone-search PL/pgSQL walkers
      do not emit span boundaries.

### Liveness today

"Is SSTorytime up?" is answered by two external checks:

```bash
# Can we reach the database?
psql -h localhost -U sstoryline -d sstoryline -c 'SELECT 1'

# Does the HTTPS server serve its landing HTML?
curl -sk https://localhost:8443/ | head -1
```

If both succeed, the system is alive. If either fails, the system is
not alive — there is no finer-grained signal available without adding
one. Operators wanting a probe for Kubernetes/systemd health checks
typically wrap these two commands in a small shell script and point
their orchestrator at it.

## 2. `LastSeen` as the one activity surface

The one observability primitive SSTorytime does ship is the
[`LastSeen` table](../Database/Schema.md#lastseen). It is **not** an
HTTP metric; it is a database-side activity log updated by two PL/pgSQL
functions every time a [node](../concepts/glossary.md#node) or section
is touched through the library API.

| Column | Meaning |
|---|---|
| `Section` / `NPtr` | What was observed. |
| `First` | Timestamp of the first observation. |
| `Last`  | Timestamp of the most recent observation. |
| `Delta` | Exponentially-weighted moving average of inter-observation gap (seconds). |
| `Freq`  | Monotonic hit counter. |

Both update functions apply a **60-second sampling threshold** before
counting a new observation — see
[Performance → `LastSeen` 60-second sampling threshold](../Database/Performance.md#lastseen-60-second-sampling-threshold)
for the rationale. The effect is that high-frequency repeat queries
(double clicks, MCP retry storms) collapse into a single counted event,
and `Delta` represents genuine user-visible sessions rather than
machine-gun traffic.

The Go-side accessor is `ReadLastSeenDrift`, which returns the
rolling averages over the table. Use it to answer:

- *Which nodes are hot this week?* — `ORDER BY Last DESC LIMIT 50`.
- *Which sections are trending up?* — compare `Freq` across snapshots.
- *Where are the idle parts of the graph?* — `WHERE Last < NOW() - interval '30 days'`.

!!! info "`LastSeen` is `LOGGED` — analytics survive a crash"
    Unlike the bulk-load tables, `LastSeen` is WAL-protected. An
    unclean Postgres shutdown during `N4L -u` truncates `Node` and
    `PageMap` but **not** `LastSeen`. Activity history survives across
    graph rebuilds.

## 3. Monitoring strategies that work today

Given the gaps above, the realistic production pattern is to **put
observability outside the SSTorytime binary**.

### Reverse proxy in front of :8443

A reverse proxy (Caddy, nginx, Traefik) in front of the SSTorytime
HTTPS server solves several problems at once:

- **Access logs** — structured, filterable by path, status, latency.
- **Healthcheck proxy** — the proxy's own liveness endpoint gives you
  a `/health` surface even though SSTorytime does not.
- **TLS termination with a real cert** — see the
  [Production deployments section of TLS certificates](../tls-certificates.md#production-deployments).
  The self-signed dev cert then lives only on the loopback hop
  between proxy and backend.
- **Rate limiting and IP allowlists** — a natural defence for a system
  without authentication built in.

A minimal Caddy config that gives you `access.log` in JSON plus
HTTPS cert management:

```caddy
sst.example.com {
    reverse_proxy localhost:8443 {
        transport http {
            tls_insecure_skip_verify
        }
    }
    log {
        output file /var/log/caddy/sst-access.log
        format json
    }
}
```

Ship those logs to whatever aggregator you already run (Loki, Splunk,
ELK, CloudWatch). You now have per-request structured telemetry without
touching the Go code.

### Postgres-side telemetry

SSTorytime pushes most of its work into PL/pgSQL, so the database
itself is the right place to instrument query-level behaviour.

- **`log_min_duration_statement`** — set this in `postgresql.conf` (or
  via `ALTER SYSTEM SET`) to log every query slower than, say, 500 ms.
  The cone-search walkers and search GIN lookups are the natural
  suspects when latency spikes; this surfaces them.
- **`pg_stat_statements`** — add `shared_preload_libraries =
  'pg_stat_statements'` and `CREATE EXTENSION pg_stat_statements`.
  `SELECT * FROM pg_stat_statements ORDER BY total_exec_time DESC
  LIMIT 20` gives you a ranked list of the most expensive queries over
  the server uptime window. This is the single highest-leverage thing
  you can enable on the database side.

Both combine well with the `LastSeen` surface above: `pg_stat_statements`
tells you which queries are slow, `LastSeen` tells you which nodes
drive the traffic behind them.

## 4. Gaps worth naming

SSTorytime's current observability model has real limits. Name them
rather than pretending they are features:

- **No Prometheus endpoint.** Rate-of-change metrics — QPS, p99
  latency, error rate — do not exist. The reverse proxy's access log
  is the closest substitute and it is lower fidelity.
- **No request IDs or trace propagation.** You cannot correlate a slow
  HTTP response with its PL/pgSQL walker on the DB side. Debugging a
  latency anomaly means eyeballing `log_min_duration_statement` output
  and matching timestamps.
- **No graceful-degradation signals.** The server does not publish
  "degraded" or "read-only" states. If the database becomes
  unreachable, the server simply starts returning errors from query
  paths — callers have to infer.
- **No alerting primitives.** There are no counters you can threshold
  on, no SLI dictionary, no burn-rate calculations built in. Any
  alerting strategy layers on the reverse proxy access log and the
  external `psql`/`curl` liveness checks.

The [versioning page's "No database migration framework" section](../versioning.md#known-limitations)
mirrors this honesty about the current scope. These are known gaps,
not bugs. They are reasonable for the project's current stage; they
will not be reasonable for a multi-tenant production deployment. When
that deployment profile matters to you, plan to add observability
rather than discovering its absence at 3 a.m.

## See also

- [TLS certificates → Production deployments](../tls-certificates.md#production-deployments)
  — the reverse-proxy pattern this page leans on.
- [Database → Performance → `LastSeen`](../Database/Performance.md#lastseen-60-second-sampling-threshold)
  — threshold behaviour behind the activity table.
- [Database → Schema → `LastSeen`](../Database/Schema.md#lastseen)
  — the table definition.
- [Backup, restore, delete](../cookbooks/backup-restore-delete.md)
  — related operational hygiene.
