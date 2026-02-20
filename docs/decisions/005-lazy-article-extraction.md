# 005 — Lazy article extraction

## Status

Accepted

## Context

Many RSS feeds only include a summary or the first paragraph of an article. To show full content in the reader, we need to fetch the original web page and extract the article text (stripping navigation, ads, sidebars).

This extraction could happen:
- **Eagerly** — during feed refresh, for every new article
- **Lazily** — only when the user actually opens an article

## Decision

Lazy extraction. When the user opens an article in the reader:

1. If `content` is already populated and substantial, display it immediately
2. If `content` is empty or very short (< 200 chars), show the summary and kick off a background `tea.Cmd` that:
   - Fetches the article's URL with `net/http`
   - Extracts readable content with go-readability
   - Sends an `ArticleExtractedMsg` back to update the reader view
   - Caches the extracted content for future views

The user sees the summary instantly, then the full content replaces it when extraction completes (typically < 1 second).

## Consequences

**Easier:**
- Feed refresh is fast — no extra HTTP requests per article
- Respects source servers — only fetches pages the user actually reads
- Most articles from full-content feeds never trigger extraction at all
- Extraction failures are visible immediately (user is looking at the article) rather than silently logged

**Harder:**
- Brief flash of summary → full content when extraction completes
- User must be online to extract (offline reading limited to what's in the cache)
- Extraction quality varies by site — some pages resist readability parsing

**Giving up:**
- Offline-first reading of full articles (would require eager extraction)
- Pre-cached full content for instant display (acceptable tradeoff — summary is shown instantly as a fallback)
