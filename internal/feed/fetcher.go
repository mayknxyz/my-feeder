package feed

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/mayknxyz/my-feeder/internal/model"
	"github.com/mayknxyz/my-feeder/internal/store"
)

// maxConcurrent limits the number of simultaneous HTTP requests during
// a feed refresh cycle.
const maxConcurrent = 5

// FetchResult holds the outcome of fetching a single feed.
type FetchResult struct {
	Feed     model.Feed
	Articles []model.Article
	Err      error
}

// Fetcher coordinates concurrent feed fetching and deduplication.
type Fetcher struct {
	Feeds       []model.Feed
	Cache       *store.Cache
	GitHubToken string
	RetentionFn func(model.Feed) int
}

// RefreshAll fetches all configured feeds concurrently, deduplicates
// new articles against the cache, and returns results per feed.
func (f *Fetcher) RefreshAll(ctx context.Context) []FetchResult {
	results := make([]FetchResult, len(f.Feeds))

	// LEARN: A buffered channel acts as a counting semaphore. Each
	// goroutine sends a value before starting work and receives after
	// finishing, limiting concurrency to the channel's buffer size.
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i, feed := range f.Feeds {
		wg.Add(1)
		go func(idx int, fd model.Feed) {
			defer wg.Done()

			sem <- struct{}{}        // acquire semaphore slot
			defer func() { <-sem }() // release semaphore slot

			articles, err := f.fetchOne(ctx, fd)
			results[idx] = FetchResult{Feed: fd, Articles: articles, Err: err}
		}(i, feed)
	}

	wg.Wait()
	return results
}

// fetchOne fetches a single feed, routes to the right parser, and
// deduplicates against the existing cache.
func (f *Fetcher) fetchOne(ctx context.Context, feed model.Feed) ([]model.Article, error) {
	var raw []model.Article
	var err error

	if feed.IsGitHub() {
		raw, err = FetchGitHubReleases(ctx, feed.GitHubRepo(), f.GitHubToken)
	} else {
		raw, err = ParseRSS(feed.URL)
	}
	if err != nil {
		return nil, err
	}

	// Deduplicate against existing cached articles for this feed.
	existing := f.Cache.ArticlesForFeed(feed.URL)
	retDays := 7
	if f.RetentionFn != nil {
		retDays = f.RetentionFn(feed)
	}

	var fresh []model.Article
	for _, a := range raw {
		if !IsDuplicate(a, existing, retDays) {
			fresh = append(fresh, a)
		}
	}

	log.Info("Feed fetched",
		"feed", feed.Name,
		"total", len(raw),
		"new", len(fresh),
		"dupes", len(raw)-len(fresh),
	)

	// Merge: new articles first, then existing (newest first).
	merged := append(fresh, existing...)
	f.Cache.SetArticles(feed.URL, merged)
	f.Cache.LastFetched[feed.URL] = time.Now().UTC().Format(time.RFC3339)

	return fresh, nil
}

// ExpireOld removes articles older than their feed's retention period
// from the cache. Bookmarked articles are never expired (handled by
// the caller checking state before expiry).
func (f *Fetcher) ExpireOld() int {
	expired := 0
	for _, feed := range f.Feeds {
		retDays := 7
		if f.RetentionFn != nil {
			retDays = f.RetentionFn(feed)
		}
		cutoff := time.Now().AddDate(0, 0, -retDays)

		articles := f.Cache.ArticlesForFeed(feed.URL)
		var kept []model.Article
		for _, a := range articles {
			if a.PublishedAt.Before(cutoff) {
				expired++
				continue
			}
			kept = append(kept, a)
		}
		if len(kept) != len(articles) {
			f.Cache.SetArticles(feed.URL, kept)
		}
	}
	if expired > 0 {
		log.Info("Expired old articles", "count", expired)
	}
	return expired
}

// FormatDuration returns a human-readable duration string like
// "2 min ago" or "1 hour ago". Used for displaying last refresh time.
func FormatDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d min ago", mins)
	default:
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
}
