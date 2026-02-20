# 002 — Flat files over SQLite

## Status

Accepted

## Context

The original plan used SQLite (via rusqlite) for storing feeds, articles, read state, and bookmarks. This gave indexed queries, relational integrity, and transactional writes.

However, the primary storage for this project is a GitHub repo — config, read state, and bookmarks all sync across machines via git. This changes the requirements:

- Persistent state must be in git-friendly formats (text, not binary)
- SQLite's `.db` file is binary — bad for git diffs, merges, and conflict resolution
- The data volume is small (dozens of feeds, thousands of articles at most)
- Articles are ephemeral and re-fetchable — they don't need durable storage

## Decision

Use flat files with Go's stdlib:

| File | Format | Synced? | Purpose |
|------|--------|---------|---------|
| `config.toml` | TOML | User's choice | Feed list, settings |
| `cache.json` | JSON | No | Fetched articles (ephemeral) |
| `state.json` | JSON | Yes | Read article GUIDs |
| `bookmarks.md` | Markdown | Yes | Saved articles |

Key design choices:
- **Cache is ephemeral** — delete it and articles re-fetch on next refresh. No migration headaches.
- **State is minimal** — just a list of read GUIDs. Small, mergeable, diffable.
- **Bookmarks are human-readable** — markdown, not JSON. Can be read/edited outside the app.

## Consequences

**Easier:**
- No C dependency (rusqlite bundles SQLite from C source)
- Git-friendly — text files diff and merge cleanly
- No schema migrations ever
- Cache can be blown away without data loss
- Human-readable state files — debuggable with `cat`

**Harder:**
- No indexed queries — dedup scans are O(n) over the cache. Fine for thousands of articles, wouldn't scale to millions.
- No relational integrity — app code must enforce consistency between cache, state, and config
- Concurrent writes need care — file locking or write-to-temp-then-rename pattern

**Giving up:**
- Complex queries (aggregations, joins) — not needed for this use case
- Transactional writes — acceptable risk given the data is ephemeral or rebuildable
