package feed

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/mayknxyz/my-feeder/internal/model"
)

// ParseRSS fetches and parses an RSS or Atom feed URL, returning
// articles mapped to the common Article model.
func ParseRSS(feedURL string) ([]model.Article, error) {
	// LEARN: gofeed.NewParser().ParseURL handles both RSS and Atom
	// transparently — it detects the format and returns a unified Feed struct.
	fp := gofeed.NewParser()
	parsed, err := fp.ParseURL(feedURL)
	if err != nil {
		return nil, fmt.Errorf("parsing feed %s: %w", feedURL, err)
	}

	articles := make([]model.Article, 0, len(parsed.Items))
	for _, item := range parsed.Items {
		articles = append(articles, mapItem(feedURL, item))
	}
	return articles, nil
}

// mapItem converts a gofeed.Item into our Article model.
func mapItem(feedURL string, item *gofeed.Item) model.Article {
	a := model.Article{
		GUID:    itemGUID(feedURL, item),
		FeedURL: feedURL,
		Title:   item.Title,
		URL:     item.Link,
		Summary: itemSummary(item),
		Content: item.Content,
	}

	if item.Author != nil {
		a.Author = item.Author.Name
	}

	// WHY: gofeed provides PublishedParsed and UpdatedParsed as *time.Time.
	// We prefer published date, falling back to updated, then current time.
	// This handles feeds that only set one or the other.
	if item.PublishedParsed != nil {
		a.PublishedAt = *item.PublishedParsed
	} else if item.UpdatedParsed != nil {
		a.PublishedAt = *item.UpdatedParsed
	} else {
		a.PublishedAt = time.Now()
	}

	a.FetchedAt = time.Now()
	a.NormalizedTitle = NormalizeTitle(a.Title)

	return a
}

// itemGUID returns a stable identifier for the item. Falls back to
// a hash of feed URL + title if no GUID or link is available.
func itemGUID(feedURL string, item *gofeed.Item) string {
	if item.GUID != "" {
		return item.GUID
	}
	if item.Link != "" {
		return item.Link
	}

	// WHY: Some feeds have neither GUID nor link. We generate a
	// deterministic ID from the feed URL + title so the same article
	// always gets the same GUID across fetches.
	h := sha256.Sum256([]byte(feedURL + "|" + item.Title))
	return fmt.Sprintf("sha256:%x", h[:8])
}

// itemSummary extracts a summary from the item, preferring the
// description field over content for brevity.
func itemSummary(item *gofeed.Item) string {
	if item.Description != "" {
		return strings.TrimSpace(item.Description)
	}
	return ""
}
