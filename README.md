# claude-desktop-config

Manage multiple Claude Desktop MCP server configurations with per-server enable/disable flags.

Claude Desktop's `claude_desktop_config.json` is plain JSON — you can't comment things out. This tool lets you maintain a **source file** with as many server entries as you like, each with an `"enabled"` flag, and generates the real config from it.

## Install

```bash
go install github.com/damianoneill/claude-desktop-config@latest
```

Or download a binary from the [releases page](https://github.com/damianoneill/claude-desktop-config/releases).

> **macOS note:** Binaries downloaded from the releases page are not code-signed or notarized. macOS Gatekeeper will block the binary on first run. To remove the quarantine attribute:
>
> ```bash
> xattr -d com.apple.quarantine ./claude-desktop-config
> ```

## Quick start

```bash
# 1. Create source file from the committed example
claude-desktop-config init

# 2. Edit it — add your real credentials, toggle enabled/disabled as needed
$EDITOR claude_desktop_config.source.json

# 3. Preview what will be generated (nothing is written)
claude-desktop-config dry-run

# 4. Apply — writes to Claude Desktop config and backs up the previous one
claude-desktop-config apply

# 5. Restart Claude Desktop
```

## Usage

```
claude-desktop-config [--source PATH] <command>
```

| Command              | Description                                                      |
| -------------------- | ---------------------------------------------------------------- |
| `init`               | Create source file from `.example` (skips if already exists)    |
| `apply`              | Generate and write Claude Desktop config from source file        |
| `dry-run`            | Preview generated config without writing to disk                 |
| `list`               | List all servers with `[on]`/`[off]` status and URL             |
| `enable <name>`      | Enable a server by name                                          |
| `disable <name>`     | Disable a server by name                                         |
| `version`            | Print version                                                    |

Global flags:

```
--source PATH   Path to source JSON file (default: ./claude_desktop_config.source.json)
```

## Source file format

`claude_desktop_config.source.json` mirrors the real Claude Desktop config structure,
with two extra fields per server entry:

- `"enabled"` — `true` to include, `false` to exclude from the generated config
- `"_comment"` — optional freeform note, stripped from output

```json
{
  "mcpServers": {
    "local-server-a": {
      "enabled": true,
      "_comment": "Local dev instance — enabled for day-to-day testing",
      "command": "npx",
      "args": [
        "mcp-remote",
        "http://localhost:8080/mcp/server-a",
        "--header",
        "Authorization:Bearer YOUR_API_TOKEN"
      ],
      "env": {
        "NODE_OPTIONS": "--use-system-ca",
        "NODE_TLS_REJECT_UNAUTHORIZED": "0"
      }
    },
    "prod-server-a": {
      "enabled": false,
      "_comment": "Production — keep disabled unless explicitly testing prod",
      "command": "npx",
      "args": [
        "mcp-remote",
        "https://api.example.com/mcp/server-a",
        "--header",
        "Authorization:Bearer YOUR_PROD_API_TOKEN"
      ],
      "env": {
        "NODE_OPTIONS": "--use-system-ca"
      }
    }
  }
}
```

`claude_desktop_config.source.json` is **gitignored** — it should never be committed since it contains real API tokens. The committed `.example` file uses placeholder values and is safe to share.

## Config file locations

The correct path is auto-detected by OS:

| OS      | Path                                                              |
| ------- | ----------------------------------------------------------------- |
| macOS   | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Linux   | `~/.config/Claude/claude_desktop_config.json`                     |
| Windows | `%APPDATA%/Claude/claude_desktop_config.json`                     |

## How apply works

1. Reads the source file
2. Filters to only `"enabled": true` servers
3. Strips `"enabled"` and `"_comment"` fields
4. Backs up the existing Claude Desktop config with a timestamp suffix (`.YYYYMMDD-HHMMSS.bak`)
5. Merges the new `mcpServers` block into the existing config, preserving any other top-level keys Claude Desktop may have added
6. Writes the result

## Development

```bash
make build          # build ./claude-desktop-config binary
make test           # run all tests
make lint           # run golangci-lint
make release-dry    # local goreleaser snapshot
```
