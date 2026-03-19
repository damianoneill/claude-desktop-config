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

# 2. Launch the interactive TUI — toggle servers, then press 'a' to apply
claude-desktop-config

# Or use the CLI directly:

# 3. Preview what will be generated (nothing is written)
claude-desktop-config dry-run

# 4. Apply — writes to Claude Desktop config, backs up the previous one
claude-desktop-config apply

# 5. Restart Claude Desktop
```

## Usage

Run without arguments to launch the interactive TUI:

```bash
claude-desktop-config
```

Or use a subcommand directly:

```
claude-desktop-config [--source PATH] [--keep-backups N] <command>
```

| Command   | Description                                                  |
| --------- | ------------------------------------------------------------ |
| `tui`     | Launch the interactive TUI (default when no command given)   |
| `init`    | Create source file from `.example` (skips if already exists) |
| `apply`   | Generate and write Claude Desktop config from source file    |
| `dry-run` | Preview generated config without writing to disk             |
| `version` | Print version                                                |

Global flags:

```
--source PATH       Path to source JSON file (default: ./claude_desktop_config.source.json)
--keep-backups N    Number of backup files to retain (default: 3)
```

## Interactive TUI

Running `claude-desktop-config` (or `claude-desktop-config tui`) opens a full-screen terminal UI:

```
Claude Desktop MCP Servers  3 enabled / 15 total

────────────────────────────────────────────────────────────────────────
       NAME                                       URL
  ●  local-server-a                               http://localhost:8080/mcp/server-a
  ○  local-server-b                               http://localhost:8080/mcp/server-b
  ○  prod-server-a                                https://api.example.com/mcp/server-a
  ●~ staging-server-a                             https://staging.example.com/mcp/server-a
────────────────────────────────────────────────────────────────────────
↑↓/jk navigate  space toggle  s save  a apply  d dry-run  q quit
```

| Symbol | Meaning                     |
| ------ | --------------------------- |
| `●`    | Enabled                     |
| `○`    | Disabled                    |
| `●~`   | Staged to enable (unsaved)  |
| `○~`   | Staged to disable (unsaved) |

Changes are **staged** until explicitly committed:

- **`space`** — toggle the selected server (staged, not yet written)
- **`s`** — save staged changes to the source file
- **`a`** — save staged changes and apply to Claude Desktop config
- **`d`** — dry-run preview of currently enabled servers (shown in status bar)
- **`q`** — quit (discards any unsaved staged changes)

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
3. Strips `"enabled"` and `"_comment"` fields from the output
4. Backs up the existing Claude Desktop config with a timestamp suffix (`.YYYYMMDD-HHMMSS.bak`)
5. Prunes old backups, keeping the most recent N (default 3, configurable via `--keep-backups`)
6. Merges the new `mcpServers` block into the existing config, preserving any other top-level keys Claude Desktop may have written
7. Writes the result

## Development

```bash
make setup          # install pre-commit hooks
make build          # build ./claude-desktop-config binary
make test           # run all tests
make lint           # run golangci-lint
make fmt            # format source code
make release-dry    # local goreleaser snapshot
```
