# journal-view

A Go CLI for fetching and rendering daily journal entries from Dropbox, displayed in a terminal TUI with Markdown rendering.

## Features

- Fetches journal entries from Dropbox by date
- Beautiful terminal UI powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Markdown rendering with [Glamour](https://github.com/charmbracelet/glamour)
- OAuth2 authentication with automatic token refresh
- Raw output mode for piping to other tools
- Multiple date input formats

## Installation

```bash
go install
```

This places the binary as `journal-view` in your `$GOPATH/bin` (or `$HOME/go/bin`).

## Usage

```bash
# View today's entry (TUI)
journal-view

# View a specific date
journal-view 2025-06-05
journal-view 06/05/2025
journal-view "Jun 5, 2025"

# Dump raw Markdown to stdout
journal-view --raw
journal-view --raw 2025-06-05
```

### Supported Date Formats

| Format          | Example         |
|-----------------|-----------------|
| `YYYY-MM-DD`    | `2025-06-05`    |
| `YYYYMMDD`      | `20250605`      |
| `MM/DD/YYYY`    | `06/05/2025`    |
| `YYYY/MM/DD`    | `2025/06/05`    |
| `Mon D, YYYY`   | `Jun 5, 2025`   |
| `Month D, YYYY` | `June 5, 2025`  |
| `YYYY-M-D`      | `2025-6-5`      |

## Authentication

`journal-view` shares credentials with [dropbox-appender](https://github.com/tgruben/dropbox-appender). It resolves an OAuth2 access token via the following priority chain:

1. **`DROPBOX_TOKEN`** env var — a direct, short-lived access token
2. **Refresh token flow** — uses `DROPBOX_REFRESH_TOKEN`, `DROPBOX_APP_KEY`, and `DROPBOX_APP_SECRET` to obtain a fresh access token automatically
3. **Config file** — reads from `~/.config/dropbox-appender/config.json` (same file as dropbox-appender)

### Config File Format

```json
{
  "app_key": "your-app-key",
  "app_secret": "your-app-secret",
  "refresh_token": "your-refresh-token"
}
```

Env vars override config file values.

## Journal Path Pattern

Entries are fetched from Dropbox using the path pattern:

```
/Notes/Journal/YYYY/MM/NoteYYYYMMDD.md
```

For example, June 5, 2025 resolves to:

```
/Notes/Journal/2025/06/Note20250605.md
```

## Keyboard Controls

| Key | Action |
|---|---|
| `q` / `Esc` / `Ctrl+C` | Quit |
| `←` / `h` | Navigate to previous day |
| `→` / `l` | Navigate to next day |
| `↑` / `↓` | Scroll up/down |
| `j` / `k` | Scroll down/up (vim-style) |
| `PgUp` / `PgDn` | Page up/down |
| `g` / `G` | Jump to top/bottom |
| Mouse wheel | Scroll up/down |

## Project Structure

```
├── main.go           # CLI entry: date parsing, config/auth, launches TUI or raw dump
├── config.go         # Reads ~/.config/dropbox-appender/config.json + env overrides
├── auth.go           # OAuth2 token resolution (direct token → refresh flow → error)
├── dropbox.go        # Dropbox content API client (/2/files/download)
├── journal.go        # Date → /Notes/Journal/YYYY/MM/NoteYYYYMMDD.md path
├── ui/
│   └── view.go       # Bubble Tea TUI with spinner + Glamour markdown rendering
├── go.mod
└── go.sum
```

## Dependencies

- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [charmbracelet/bubbles](https://github.com/charmbracelet/bubbles) — Spinner + Viewport components
- [charmbracelet/glamour](https://github.com/charmbracelet/glamour) — Markdown rendering
- [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
