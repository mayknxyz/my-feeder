# Package: `internal/feed`

> Fill in as Phase 2 is built.

## Purpose

Handles fetching, parsing, and deduplicating feed content from RSS/Atom sources and GitHub releases.

## Key Files

| File | Responsibility |
|------|---------------|
| `fetcher.go` | HTTP fetch orchestration, concurrent requests via errgroup |
| `parser.go` | gofeed → Article struct mapping |
| `github.go` | GitHub releases via go-github |
| `extractor.go` | On-demand readability extraction |
| `dedup.go` | Title normalization, 3-tier similarity check |

## Concepts to Document

- [ ] How gofeed unifies RSS and Atom into one struct
- [ ] errgroup + semaphore pattern for bounded concurrency
- [ ] Jaro-Winkler scoring and threshold tuning
- [ ] go-readability extraction: what works, what doesn't
- [ ] GitHub API rate limiting and token auth
