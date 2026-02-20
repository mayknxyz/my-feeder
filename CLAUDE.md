# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Feeder — a TUI RSS/Atom feed reader in Go. Also supports GitHub releases via `github:owner/repo` URL syntax. State (read status, bookmarks) syncs via a GitHub repo as flat files. Single binary, no runtime dependencies.

## Build / Test / Lint

```bash
go build ./...           # compile
go run .                 # run the TUI
go test ./...            # all tests
go test ./internal/feed  # single package tests
go vet ./...             # lint (must pass clean)
```

Every change should pass `go build ./...` (no warnings), `go test ./...`, and `go vet ./...` clean.

## Architecture

**Elm-style state machine via bubbletea**: A single `Model` struct owns all state. Each `ui/` module provides its own bubbletea sub-model with `Update()` and `View()`. No shared mutable state.

**Message loop**: Bubbletea manages the event loop. Background tasks (feed refresh, article extraction) use `tea.Cmd` to send messages (`FeedUpdatedMsg`, `FeedErrorMsg`, `ArticleExtractedMsg`, etc.) back to the main model.

**Storage**: Flat files, no database.
- `config.toml` — TOML, source of truth for feed list. Lives in `~/.config/feeder/`.
- `cache.json` — local ephemeral article cache. Can be deleted and rebuilt by re-fetching. Not synced.
- `state.json` — read article IDs and per-feed metadata. Synced via GitHub repo.
- `bookmarks.md` — append-only markdown. Synced via GitHub repo.

**Feed refresh**: Goroutine on a ticker with `errgroup` + semaphore for concurrent fetches. GitHub releases use go-github; RSS/Atom uses gofeed.

**Article extraction**: Lazy — go-readability extraction only fires when a user opens an article with short/missing content.

**Dedup** (at insert time, before cache write):
1. Exact GUID match
2. Exact URL match
3. Fuzzy title match — Jaro-Winkler >= 0.85 against articles from last `max(retention_days, 14)` days (titles normalized: lowercase, strip punctuation, collapse whitespace, stored as `normalized_title`)

## Key Dependencies

bubbletea + lipgloss (TUI), glamour (markdown rendering), gofeed (feed parsing), go-github (GitHub API), go-readability (extraction), smetrics (Jaro-Winkler dedup), BurntSushi/toml (config), adrg/xdg (XDG base directories), charmbracelet/log (structured logging), cobra (CLI), pkg/browser (open URLs).

## Config

TOML at `~/.config/feeder/config.toml`. Feeds defined as `[[feeds]]` entries with `name`, `url`, optional `tag` and `retention_days`. Global settings under `[settings]` (refresh interval, default retention, bookmark/state/cache file paths, optional GitHub token).

## Commenting Convention

This is a learning project — code should be well-commented. Follow `docs/commenting.md` for the full guide. Summary:

1. **Package comments** — every package gets a godoc comment at the top
2. **Exported godoc** — every exported function, type, and method gets a doc comment starting with the symbol name
3. **Inline comments** — two prefixed categories:
   - `// WHY:` — explains non-obvious design decisions (why this approach, not another)
   - `// LEARN:` — explains Go idioms, library behavior, or patterns that are educational (can be stripped later)

Do not comment obvious code. Do not write changelog comments. If the code is self-explanatory, skip the comment.

## TUI Screens

Three screens: **Dashboard** (feed table with unread counts) → **Feed List** (articles for a feed/tag) → **Reader** (glamour-rendered article). Navigation: `Enter` drills in, `Esc`/`d` goes back, vim-style `j`/`k` for movement.
