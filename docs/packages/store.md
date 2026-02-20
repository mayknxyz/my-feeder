# Package: `internal/store`

> Fill in as Phase 1 is built.

## Purpose

Reads and writes flat-file storage: article cache, read state, and bookmarks.

## Key Files

| File | Responsibility |
|------|---------------|
| `store.go` | Shared types, file path resolution |
| `cache.go` | Article cache — read/write `cache.json` |
| `bookmarks.go` | Bookmark markdown file — append entries |

## Concepts to Document

- [ ] Write-to-temp-then-rename pattern for safe file writes
- [ ] JSON marshaling/unmarshaling with `encoding/json`
- [ ] How cache invalidation and expiry work
- [ ] File path expansion (`~` → home directory)
- [ ] State file format and merge-friendliness for git
