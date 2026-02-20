# Feeder — Personal TUI Feed Reader

## Context

Build an open-source TUI feed reader in Go. Motivated by wanting a modern alternative to newsboat that's lightweight, configurable, and supports more than just RSS. State syncs via a GitHub repo — config, read state, and bookmarks all live as flat files in the repo.

## v1 Feature Set

- **Feed sources**: RSS/Atom + GitHub releases
- **Retention**: configurable per feed, default 7 days, auto-expiry
- **Deduplication**: 3-tier (GUID → URL → fuzzy title via Jaro-Winkler at 0.85 threshold)
- **Sorting**: chronological, newest first
- **Full article extraction**: on-demand readability extraction for summary-only feeds
- **Dashboard**: single screen with unread counts per feed/tag
- **Feed list + reader**: standard pane-based TUI with vim-style keys
- **Bookmarks**: quick save to local markdown file (bookmarked articles survive expiry)
- **Images**: displayed as copyable URLs, `o` to open in browser
- **Config**: TOML file at `~/.config/feeder/config.toml`
- **State sync**: read state + bookmarks stored as flat files, synced via GitHub repo
- **Single binary**: no runtime dependencies

## Dependencies

| Purpose | Package |
|---------|---------|
| TUI framework | `charmbracelet/bubbletea` |
| TUI styling | `charmbracelet/lipgloss` |
| Markdown rendering | `charmbracelet/glamour` |
| HTTP client | `net/http` (stdlib) |
| Feed parsing | `mmcdole/gofeed` |
| GitHub API | `google/go-github` |
| Readability | `go-shiori/go-readability` |
| Fuzzy matching | `xrash/smetrics` (Jaro-Winkler) |
| Config | `BurntSushi/toml` |
| XDG paths | `adrg/xdg` |
| Logging | `charmbracelet/log` |
| CLI args | `spf13/cobra` |
| Open URLs | `pkg/browser` |

## Project Structure

```
feeder/
├── go.mod
├── go.sum
├── config.example.toml
├── main.go                # Entry: parse args, load config, run TUI
├── internal/
│   ├── app/
│   │   └── app.go         # Bubbletea model, Update/View dispatch
│   ├── config/
│   │   └── config.go      # TOML parsing, defaults, validation
│   ├── model/
│   │   └── model.go       # Feed, Article, Bookmark structs
│   ├── store/
│   │   ├── store.go       # Read/write state files (JSON)
│   │   ├── cache.go       # Article cache (ephemeral local JSON)
│   │   └── bookmarks.go   # Bookmark markdown file operations
│   ├── feed/
│   │   ├── fetcher.go     # HTTP fetch, concurrent with errgroup
│   │   ├── parser.go      # gofeed → Article mapping
│   │   ├── github.go      # GitHub releases via go-github
│   │   ├── extractor.go   # Readability full-text extraction
│   │   └── dedup.go       # Title normalization, similarity check
│   └── ui/
│       ├── dashboard.go   # Feed overview with unread counts
│       ├── feedlist.go    # Article list for feed/tag
│       ├── reader.go      # Full article with glamour rendering
│       └── help.go        # Keybinding overlay
```

## Architecture

**Pattern**: Elm-style via bubbletea — single `Model` struct owns all state. `Update()` handles messages, `View()` renders. Each screen is a sub-model with its own `Update`/`View`.

**Message loop**: Bubbletea manages the event loop. Background tasks (feed refresh, article extraction) use `tea.Cmd` to send messages back:

- `tea.KeyMsg` — key press
- `tea.WindowSizeMsg` — terminal resize
- `FeedUpdatedMsg` — background fetch complete
- `FeedErrorMsg` — fetch failed
- `ArticleExtractedMsg` — readability done
- `TickMsg` — periodic refresh timer

**Storage**: Flat files, no database.

- **Config** (`config.toml`) — TOML, source of truth for feed list. Lives in `~/.config/feeder/` or the synced repo.
- **Article cache** (`cache.json`) — local ephemeral JSON file. Stores fetched articles. Can be rebuilt from scratch by re-fetching. Not synced.
- **Read state** (`state.json`) — JSON file tracking read article IDs and per-feed metadata. Synced via GitHub repo.
- **Bookmarks** (`bookmarks.md`) — markdown file. Synced via GitHub repo.

**Feed refresh**: goroutine on a ticker, concurrent fetches with `errgroup` + semaphore. Results sent back as bubbletea messages.

**Article extraction**: Lazy — only fetches full text when user opens an article with short/missing content.

**GitHub releases**: `github:owner/repo` URL prefix in config triggers go-github instead of RSS fetch.

## Storage Formats

### Article Cache (`cache.json`)

Local only, ephemeral — can be deleted and rebuilt by re-fetching.

```json
{
  "version": 1,
  "articles": {
    "feed-url-hash": [
      {
        "guid": "unique-id",
        "title": "Article Title",
        "url": "https://example.com/post",
        "author": "Author Name",
        "summary": "Short summary...",
        "content": "Full extracted content if available",
        "published_at": "2025-01-09T00:00:00Z",
        "fetched_at": "2025-01-09T12:00:00Z",
        "normalized_title": "article title"
      }
    ]
  },
  "last_fetched": {
    "feed-url-hash": "2025-01-09T12:00:00Z"
  }
}
```

### Read State (`state.json`)

Synced via GitHub repo. Tracks which articles have been read.

```json
{
  "version": 1,
  "read": ["guid-1", "guid-2", "guid-3"],
  "last_synced": "2025-01-09T12:00:00Z"
}
```

### Bookmarks (`bookmarks.md`)

Synced via GitHub repo. Append-only markdown.

```markdown
## Article Title

- **Source**: Feed Name
- **URL**: https://example.com/post
- **Date**: 2025-01-09
- **Notes**: user notes if any

---
```

## Config Format

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

## TUI Screens

**Dashboard** → Feed table with unread counts, tags summary, refresh status

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

**Feed List** → Article table for selected feed/tag, `*` for unread

```
┌─ Go Blog ───────────────────────────────────────┐
│                                                  │
│  > * Go 1.24 Released              2025-01-09    │
│      Go Survey Results 2024        2025-01-05    │
│    * Structured Logging in Go      2025-01-02    │
│      Go 1.23 Released              2024-12-28    │
│                                                  │
│  4 articles (2 unread)                           │
│  [enter]open  [b]ookmark  [m]ark read  [d]ash   │
└──────────────────────────────────────────────────┘
```

**Reader** → Glamour-rendered article, images as `[Image: url]`, scrollable

```
┌─ Go 1.24 Released ─────────────────────────────┐
│  Source: go.dev/blog  |  2025-01-09             │
│  ────────────────────────────────────────────── │
│                                                  │
│  The Go team is happy to announce version        │
│  1.24 of Go. This release includes...           │
│                                                  │
│  [Image: https://go.dev/blog/img/...]            │
│                                                  │
│  ## What's New                                   │
│  ...                                             │
│                                                  │
│  [j/k]scroll  [o]pen in browser  [b]ookmark     │
│  [y]ank URL  [d]ash  [esc]back                  │
└──────────────────────────────────────────────────┘
```

**Navigation**: `Enter` to drill in, `Esc`/`d` to go back, `q` to quit, `r` to refresh, `b` to bookmark, `o` to open in browser, `?` for help.

## Deduplication Strategy

Runs at **insert time** — after parsing a feed, before writing to cache.

1. **Exact GUID match** — fastest path, covers most cases
2. **Exact URL match** — catches same article from different feeds
3. **Fuzzy title match** — Jaro-Winkler similarity at 0.85 threshold against articles from last 14 days

Title normalization: lowercase, strip punctuation, collapse whitespace. Stored as `normalized_title` in cache for fast lookups.

Note: the dedup window (14 days) should always be `max(retention_days, 14)` to avoid re-inserting articles that were expired but not yet outside the dedup window.

## Build-Up Path (6 phases)

### Phase 1: Skeleton + Config + Storage

- `go mod init`, config structs, TOML loading
- Store package: read/write JSON cache and state files
- Bookmark markdown append
- Tests for config parsing and store operations
- **Milestone**: binary reads config, creates storage files, prints feed count

### Phase 2: Feed Fetching + Parsing

- HTTP fetch for RSS/Atom via gofeed
- GitHub releases via go-github
- Dedup logic (normalize + GUID/URL/fuzzy check)
- Article cache insert/query
- Tests for parser, dedup, GitHub mapping
- **Milestone**: binary fetches real feeds, caches articles, prints counts

### Phase 3: Basic TUI Shell

- Bubbletea model with screen enum
- Dashboard, feed list, reader views (plain text, no glamour yet)
- Navigation between screens
- **Milestone**: working TUI that displays articles from cache

### Phase 4: Background Refresh + Polish

- Background fetch via `tea.Cmd` on ticker
- Status bar with refresh status/errors
- Auto-expiry of old articles from cache
- `r` key for manual refresh
- **Milestone**: feeds refresh automatically while TUI runs

### Phase 5: Article Extraction + Markdown

- go-readability extraction on demand
- glamour markdown rendering in reader
- Image URLs as copyable text, `o` to open browser
- **Milestone**: rich article reading experience

### Phase 6: Bookmarks + Final Polish

- Bookmark to markdown file, `b` key
- Help overlay with `?`
- Yank URL with `y`
- Toggle read/unread with `m`
- Graceful error display in status bar
- cobra CLI args (e.g. `--config`, `--refresh`)
- **Milestone**: feature-complete v1

## Verification

After each phase:

1. `go build ./...` compiles clean
2. `go test ./...` passes
3. `go vet ./...` clean
4. Manual test with real feeds (Go Blog, HN, a GitHub repo)

End-to-end v1 test:

1. Create config with 3+ feeds (mix of RSS and GitHub)
2. Run `feeder` — dashboard shows feeds with unread counts
3. Enter a feed → articles listed newest first
4. Open article → full text rendered, images as URLs
5. Bookmark an article → check markdown file created
6. Wait for auto-refresh → new articles appear
7. Check articles older than retention are purged (but bookmarks survive)
8. Delete cache.json, re-run → articles re-fetched, read state preserved from state.json
