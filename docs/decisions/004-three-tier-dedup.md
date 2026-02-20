# 004 — Three-tier deduplication

## Status

Accepted

## Context

RSS feeds frequently produce duplicates:
- Same article appears in multiple feeds (cross-posted, syndicated)
- Feed regeneration assigns new GUIDs to existing articles
- Titles change slightly between updates ("v1.0 Released" → "v1.0.0 Released")

Without dedup, the article list fills with noise. Over-aggressive dedup hides legitimate distinct articles.

## Decision

Three-tier dedup, checked in order at insert time (before writing to cache):

1. **Exact GUID match** — if the article has a GUID and it matches an existing article, skip. Fastest check, covers most RSS/Atom feeds that assign stable IDs.

2. **Exact URL match** — if the article URL matches an existing article's URL, skip. Catches cross-feed duplicates where different feeds link to the same page.

3. **Fuzzy title match** — normalize both titles (lowercase, strip punctuation, collapse whitespace), compute Jaro-Winkler similarity. If >= 0.85 against any article from the last `max(retention_days, 14)` days, skip.

The 0.85 threshold was chosen because:
- 1.0 = exact match only (too strict — misses "Go 1.24" vs "Go 1.24.0")
- 0.80 = catches too many false positives ("Rust 1.84" matches "Rust 1.83")
- 0.85 = sweet spot in testing with real RSS feeds

The dedup window uses `max(retention_days, 14)` to avoid a gap where articles expire from cache but are still within re-fetch range.

## Consequences

**Easier:**
- Clean article lists without manual curation
- Works across feeds (URL match catches syndication)
- Normalized titles stored in cache for fast comparison

**Harder:**
- Fuzzy matching is O(n) per new article against the dedup window — acceptable for thousands of articles, but the window must be bounded
- Threshold tuning may need adjustment for non-English feeds
- False positives are possible — a legitimately new article with a very similar title could be skipped

**Giving up:**
- Content-based dedup (comparing article bodies) — too expensive and unreliable with varying extraction quality
- User-configurable thresholds — keeping it simple for v1, can add later if needed
