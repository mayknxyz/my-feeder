# Feeder

A TUI feed reader built in Go. Supports RSS, Atom, and GitHub releases. Single binary, no runtime dependencies.

State (read status, bookmarks) syncs via a GitHub repo as flat files — no database.

```
┌─ Feeder ────────────────────────────────────────┐
│                                                  │
│  Feeds                              Unread       │
│  ─────                              ──────       │
│  > [go]    Go Blog                     3         │
│    [rust]  This Week in Rust           1         │
│    [gh]    tokio releases              2         │
│    [news]  Hacker News                12         │
│                                                  │
│  Tags: go(3)  rust(3)  news(12)  all(18)         │
│                                                  │
│  Last refresh: 2 min ago    Next: in 28 min      │
│  [r]efresh  [a]ll  [q]uit  [?]help               │
└──────────────────────────────────────────────────┘
```

## Features

- **RSS/Atom + GitHub releases** — follow any feed or `github:owner/repo` for release tracking
- **Three-tier dedup** — GUID, URL, and fuzzy title matching to keep your list clean
- **On-demand article extraction** — full-text readability for summary-only feeds
- **Bookmarks** — save articles to a markdown file, survives article expiry
- **Configurable retention** — per-feed or global, auto-expiry of old articles
- **Vim-style navigation** — `j`/`k` movement, `Enter` to drill in, `Esc` to go back
- **Sync via GitHub** — read state and bookmarks are flat files, sync them however you like

## Install

```bash
go install github.com/mayknxyz/my-feeder@latest
```

Or build from source:

```bash
git clone https://github.com/mayknxyz/my-feeder.git
cd my-feeder
go build -o feeder .
```

## Config

Create `~/.config/feeder/config.toml`:

```toml
[settings]
refresh_interval_minutes = 30
retention_days = 7
bookmark_file = "~/Documents/feeder-bookmarks.md"
state_file = "~/Documents/feeder-state.json"
cache_file = "~/.cache/feeder/cache.json"
# github_token = "ghp_..."

[[feeds]]
name = "Go Blog"
url = "https://go.dev/blog/feed.atom"
tag = "go"

[[feeds]]
name = "Tokio Releases"
url = "github:tokio-rs/tokio"
tag = "rust"
retention_days = 30
```

## Keybindings

| Key | Action |
|-----|--------|
| `j` / `k` | Move down / up |
| `Enter` | Open feed / article |
| `Esc` | Go back |
| `d` | Go to dashboard |
| `r` | Refresh feeds |
| `a` | Show all articles |
| `b` | Bookmark article |
| `m` | Toggle read / unread |
| `o` | Open in browser |
| `y` | Yank (copy) URL |
| `?` | Help overlay |
| `q` | Quit |

## Storage

| File | Format | Synced? | Purpose |
|------|--------|---------|---------|
| `config.toml` | TOML | Your choice | Feed list and settings |
| `cache.json` | JSON | No | Fetched articles (ephemeral, rebuildable) |
| `state.json` | JSON | Yes | Read article IDs |
| `bookmarks.md` | Markdown | Yes | Saved articles |

The cache can be deleted at any time — articles re-fetch on the next refresh. Read state and bookmarks are designed to be synced via a git repo.

## Docs

- [Architecture](docs/architecture.md) — data flow, screen flow, concurrency model
- [Commenting guide](docs/commenting.md) — code commenting conventions
- [Architecture decisions](docs/decisions/) — ADRs explaining key technical choices

## License

MIT
