# Connect an LLM via MCP-SST

Model Context Protocol (MCP) is an open standard for letting LLMs reach out to external tools over a well-defined RPC. **MCP-SST** is a separate proxy project that translates MCP tool calls into SSTorytime HTTP API requests, so an LLM client (Claude Desktop, for example) can search your personal knowledge graph, browse notes, and walk paths.

This cookbook walks the full path: start the SSTorytime server, install MCP-SST, wire up a client, and run a sample conversation.

!!! info "Where is MCP-SST?"
    MCP-SST lives in a separate repository from SSTorytime itself. Check the
    [SSTorytime README](https://github.com/markburgess/SSTorytime) and the
    [LinkedIn SSTorytime group](https://www.linkedin.com/groups/15875004/)
    for the current canonical URL. At the time of writing the proxy is still evolving
    and not all details below are guaranteed; assumptions made without direct code
    access are flagged **[inferred]**.

## 1. Start the SSTorytime HTTPS server

From the repo root:

```bash
cd src
make                            # builds src/bin/* including http_server
./bin/http_server -resources ../resources
```

The server listens on **HTTPS :8443** (primary) with a TLS redirect on :8080. Certificates are generated on first build by `src/server/make_certificate` and stored in `src/server/cert.pem` / `src/server/key.pem` (365-day, self-signed RSA-4096). For details see [http_server.md](../http_server.md) and [WebAPI.md](../WebAPI.md).

Verify the endpoints are up:

```bash
curl -k https://localhost:8443/searchN4L -d '{"Query":"start"}' -H "Content-Type: application/json"
```

The `-k` flag accepts the self-signed cert. You should get back a JSON envelope with `Content`, `Time`, `Intent`, and `Ambient` fields.

!!! warning "Production TLS"
    Self-signed certificates are fine for localhost development. For anything exposed
    beyond your own machine, terminate TLS at a reverse proxy (nginx, Caddy, Traefik)
    with a real certificate, and point the proxy at `http://localhost:8080` (which SST
    silently serves when it sees `X-Forwarded-Proto: https`). Never expose :8443 with
    the default cert to the public internet.

## 2. Install MCP-SST

**[inferred — verify against the actual repo]** The proxy is typically distributed as a small Go or Node.js binary. Assuming the Go flavor:

```bash
git clone https://github.com/<org>/mcp-sst.git
cd mcp-sst
go build -o mcp-sst .
```

Configuration is typically a small JSON or YAML file describing:

- **SSTorytime endpoint URL** — `https://localhost:8443` for the setup above.
- **TLS trust policy** — either `insecure: true` for self-signed certs in development, or a `ca_cert_path` pointing to your PEM bundle.
- **Tool allowlist** — which MCP tool calls the proxy should expose. A reasonable default set: `search`, `notes`, `paths`, `arrows`, `chapters`.

Example minimal config (`mcp-sst.yaml`):

```yaml
sstorytime:
  url: https://localhost:8443
  insecure: true

tools:
  - search
  - notes
  - paths
  - arrows
```

## 3. Configure an LLM client

Claude Desktop is the most common target. It reads MCP server definitions from `~/Library/Application Support/Claude/claude_desktop_config.json` on macOS or `%APPDATA%\Claude\claude_desktop_config.json` on Windows. The Linux location is typically `~/.config/Claude/claude_desktop_config.json` **[inferred]**.

Add an entry for the MCP-SST proxy:

```json
{
  "mcpServers": {
    "sstorytime": {
      "command": "/path/to/mcp-sst",
      "args": ["--config", "/path/to/mcp-sst.yaml"]
    }
  }
}
```

Restart Claude Desktop. In the **Tools** panel you should see `sstorytime` with the configured tool set.

For other clients (Cursor, Continue.dev, custom MCP hosts), the exact config differs but the shape is identical: point the client at the `mcp-sst` binary and pass its config file.

## 4. Sample conversation

Once wired up, you can ask the LLM to reach into the graph. Example dialogue:

> **User:** What do I have in my knowledge graph about the Chinese word for "food"?

> **Claude** (invokes `sstorytime.search` with `{"query": "food"}`): I found several related entries in your `chinese` chapter. The hanzi 食物 maps to pinyin `shíwù` and English `food`. It appears in orbit with 菜 (vegetable, `cài`) and 肉 (meat, `ròu`). Want me to pull the full notes page for the `chinese` chapter?

> **User:** Yes, and also find a path from 菜 to 肉.

> **Claude** (invokes `sstorytime.notes` with `{"chapter": "chinese"}`, then `sstorytime.paths` with `{"from": "菜", "to": "肉"}`): Here are the two relevant pages. For the path: 菜 → (english has hanzi) → vegetable → (is a kind of) → food → (is a kind of) → meat → (english has hanzi) → 肉.

Under the hood, each tool call becomes an HTTPS request to the SSTorytime server. The LLM shapes the result into natural language. The MCP proxy is stateless — each turn is an independent RPC.

## 5. Trust model

MCP-SST sits between your LLM client and your database. That makes the proxy a genuine security boundary:

- **The LLM sees whatever the proxy exposes.** If you configure `tools: [search, notes]` but omit `write`, the LLM literally cannot modify the graph.
- **Credentials stay on the server side.** The SSTorytime DSN is not visible to the LLM or to the cloud hosting the model. Only the HTTPS endpoint of SSTorytime is reached over the network, and that is ideally on localhost.
- **Self-signed certificates** mean the proxy must trust them explicitly (`insecure: true` or CA pinning). Do not copy `insecure: true` into production.
- **Rate limiting and logging** are the proxy's responsibility. The SSTorytime server does not throttle by itself. A misbehaving LLM can easily run thousands of queries; configure per-minute caps in the proxy.

## Troubleshooting

- **LLM says "no tool called sstorytime.search"** — the MCP client did not load the config. Restart the client; check its logs for JSON parse errors.
- **Proxy cannot connect to SSTorytime** — start the server (step 1) and verify with `curl -k https://localhost:8443/...`.
- **TLS error in the proxy** — either set `insecure: true` or fix the cert trust chain (add SSTorytime's cert to the system trust store or point `ca_cert_path` at it).
- **Empty results for plausible queries** — verify the graph actually contains the data: run the same query with `../src/bin/searchN4L` locally before blaming the LLM path.

## Next steps

- If no MCP-SST exists for your platform, the [WebAPI.md](../WebAPI.md) endpoints are simple enough that rolling a minimal MCP wrapper is an afternoon's work.
- The response envelope (`Content`, `Time`, `Intent`, `Ambient`) is designed to be LLM-friendly — `Ambient` is the chapter/context metadata, `Intent` is the parsed query object, and `Content` is a tagged union of result types. Use the `Intent` echo to verify the proxy is translating MCP parameters correctly.
