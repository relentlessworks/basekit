# basekit

Agentic-first base encoding and decoding service. Encode and decode data using Base64, Base64URL, Base32, Base32Hex, Base58 (Bitcoin), Base85 (ASCII85), Hex, and URL encoding.

The agent IS the interface. No UI, no SDK. Plain text API, agent-driven, single Go binary. No database needed — pure stateless computation.

## Quick Start

```bash
# Build
make build

# Run
./basekit
# basekit listening on :7700

# Encode
curl -X POST 'http://localhost:7700/encode?encoding=base64&data=hello'
# encoding=base64 input=hello output=aGVsbG8=

# Decode
curl -X POST 'http://localhost:7700/decode?encoding=base64&data=aGVsbG8='
# encoding=base64 input=aGVsbG8= output=hello

# Convert between encodings
curl -X POST 'http://localhost:7700/convert?from=hex&to=base64&data=68656c6c6f'
# from=hex to=base64 input=68656c6c6f output=aGVsbG8=

# Detect encoding
curl -X POST 'http://localhost:7700/detect?data=aGVsbG8='
# encoding=base64

# List supported encodings
curl http://localhost:7700/list
# encoding=base64
# encoding=base64url
# encoding=base32
# encoding=base32hex
# encoding=base58
# encoding=base85
# encoding=hex
# encoding=url

# Get help
curl http://localhost:7700/help

# JSON output
curl -H 'Accept: application/json' -X POST 'http://localhost:7700/encode?encoding=base64&data=hello'
```

## Supported Encodings

| Encoding    | Description                          |
|-------------|--------------------------------------|
| base64      | Standard Base64                      |
| base64url   | URL-safe Base64 (- and _)            |
| base32      | Standard Base32 (RFC 4648)          |
| base32hex   | Hex Base32 (0-9A-V)                 |
| base58      | Bitcoin Base58 (no 0, O, I, l)      |
| base85      | ASCII85 (Base85)                     |
| hex         | Hexadecimal encoding                 |
| url         | URL percent-encoding                 |

## API Reference

### `GET /`
Service info.

### `GET /help` or `GET /.well-known/agent.md`
Operating manual for agents.

### `GET /list`
List all supported encodings.

### `POST /encode`
Encode data.
- `encoding` — encoding to use
- `data` — text to encode (or send as raw body)

### `POST /decode`
Decode data.
- `encoding` — encoding to use
- `data` — encoded text (or send as raw body)

### `POST /convert`
Convert between encodings.
- `from` — source encoding
- `to` — target encoding
- `data` — data to convert

### `POST /detect`
Detect possible encodings of a string.
- `data` — text to analyze

### `POST /mcp`
MCP (Model Context Protocol) JSON-RPC 2.0 endpoint for chat client integrations.

## Configuration

| Flag      | Env          | Default | Description       |
|-----------|--------------|---------|-------------------|
| `-addr`   | `BASEKIT_ADDR` | `:7700` | Listen address    |

Config priority: defaults < env < flags.

## Build

```bash
make build    # CGO_ENABLED=0 go build -trimpath
make test     # go test -race
make vet      # go vet
make run      # go run
```

## License

MIT
