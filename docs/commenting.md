# Commenting Guide

Three tiers of comments, each serving a different purpose.

## 1. Package Comments

Go convention — one per package, at the top of the main file. Explains what the package does.

```go
// Package feed handles fetching, parsing, and deduplicating
// content from RSS/Atom feeds and GitHub releases.
package feed
```

Every package must have one. Keep it to 1-3 sentences.

## 2. Function/Type Godoc

On every exported function, type, and method. Explains **what** it does and **when** to use it. Starts with the symbol name per Go convention.

```go
// Normalize strips a title for dedup comparison:
// lowercase, remove punctuation, collapse whitespace.
func Normalize(title string) string {

// Article represents a single feed entry, regardless of source
// (RSS, Atom, or GitHub release).
type Article struct {

// Refresh fetches all configured feeds concurrently and returns
// the merged results. Respects the semaphore limit for max
// concurrent HTTP requests. Called on startup and by the ticker.
func (f *Fetcher) Refresh(ctx context.Context) ([]Article, error) {
```

Unexported helpers get a comment if the purpose isn't obvious from the name and signature. Skip comments on trivial helpers.

## 3. Inline Comments

Two prefixed categories for easy scanning:

### WHY — Design rationale

Explains a non-obvious choice. The code shows *what*, the comment explains *why this way and not another*.

```go
// WHY: Jaro-Winkler penalizes early-character mismatches more
// than late ones, which suits article titles where the meaningful
// words tend to come first ("Go 1.24 Released" vs "Go 1.24.0 Released").
score := smetrics.JaroWinkler(a, b, 0.7, 4)

// WHY: write to temp file then rename — prevents partial reads
// if the process crashes mid-write.
tmpPath := path + ".tmp"
```

### LEARN — Go idioms and library behavior

Marks patterns that are new or non-obvious. These are meant to be educational — they can be stripped once the pattern is second nature.

```go
// LEARN: errgroup cancels all goroutines if any one returns an error.
// The semaphore (buffered channel) limits concurrent HTTP requests.
g, ctx := errgroup.WithContext(ctx)
sem := make(chan struct{}, maxConcurrent)

// LEARN: type switch is Go's way of handling sum types / tagged unions.
// Each message type gets its own case.
switch msg := msg.(type) {
case tea.KeyMsg:
case FeedUpdatedMsg:
}

// LEARN: defer runs when the surrounding function returns, not when
// the block scope ends. This closes the file after the function exits.
defer f.Close()
```

## What NOT to Comment

```go
// Bad: restates the code
// Increment i by 1
i++

// Bad: obvious from the name and type
// GetTitle returns the title
func (a *Article) GetTitle() string {

// Bad: changelog comment
// Added 2025-01-09: new field for tags
Tag string
```

## When to Add Comments

- Writing a new function or type → add godoc
- Making a non-obvious choice → add WHY
- Using a Go pattern for the first time → add LEARN
- If you had to think about it for more than a moment → it deserves a comment
