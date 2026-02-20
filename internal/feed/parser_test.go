package feed

import (
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

func TestMapItem_FullItem(t *testing.T) {
	pub := time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC)
	item := &gofeed.Item{
		GUID:            "unique-123",
		Title:           "Go 1.24 Released",
		Link:            "https://go.dev/blog/go1.24",
		Description:     "The Go team announces Go 1.24.",
		Content:         "<p>Full article content here.</p>",
		Author:          &gofeed.Person{Name: "Go Team"},
		PublishedParsed: &pub,
	}

	a := mapItem("https://go.dev/blog/feed.atom", item)

	if a.GUID != "unique-123" {
		t.Errorf("GUID = %q, want %q", a.GUID, "unique-123")
	}
	if a.Title != "Go 1.24 Released" {
		t.Errorf("Title = %q", a.Title)
	}
	if a.URL != "https://go.dev/blog/go1.24" {
		t.Errorf("URL = %q", a.URL)
	}
	if a.Author != "Go Team" {
		t.Errorf("Author = %q", a.Author)
	}
	if a.Summary != "The Go team announces Go 1.24." {
		t.Errorf("Summary = %q", a.Summary)
	}
	if a.Content != "<p>Full article content here.</p>" {
		t.Errorf("Content = %q", a.Content)
	}
	if !a.PublishedAt.Equal(pub) {
		t.Errorf("PublishedAt = %v, want %v", a.PublishedAt, pub)
	}
	if a.NormalizedTitle != "go 124 released" {
		t.Errorf("NormalizedTitle = %q", a.NormalizedTitle)
	}
	if a.FeedURL != "https://go.dev/blog/feed.atom" {
		t.Errorf("FeedURL = %q", a.FeedURL)
	}
}

func TestMapItem_NoGUID_FallbackToLink(t *testing.T) {
	item := &gofeed.Item{
		Title: "Test",
		Link:  "https://example.com/post",
	}

	a := mapItem("https://example.com/feed", item)
	if a.GUID != "https://example.com/post" {
		t.Errorf("GUID should fall back to link, got %q", a.GUID)
	}
}

func TestMapItem_NoGUIDNoLink_GeneratesHash(t *testing.T) {
	item := &gofeed.Item{
		Title: "Orphan Article",
	}

	a := mapItem("https://example.com/feed", item)
	if a.GUID == "" {
		t.Error("GUID should be generated when no GUID or link")
	}
	if a.GUID[:7] != "sha256:" {
		t.Errorf("generated GUID should start with sha256:, got %q", a.GUID)
	}

	// Same input should produce same hash.
	a2 := mapItem("https://example.com/feed", item)
	if a.GUID != a2.GUID {
		t.Error("generated GUID should be deterministic")
	}
}

func TestMapItem_NoPublishedDate_FallbackToUpdated(t *testing.T) {
	updated := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	item := &gofeed.Item{
		Title:         "Updated Only",
		UpdatedParsed: &updated,
	}

	a := mapItem("https://example.com/feed", item)
	if !a.PublishedAt.Equal(updated) {
		t.Errorf("PublishedAt should fall back to UpdatedParsed, got %v", a.PublishedAt)
	}
}

func TestMapItem_NoAuthor(t *testing.T) {
	item := &gofeed.Item{
		Title: "No Author",
	}

	a := mapItem("https://example.com/feed", item)
	if a.Author != "" {
		t.Errorf("Author should be empty, got %q", a.Author)
	}
}

func TestMapItem_DescriptionPreferredOverContent(t *testing.T) {
	item := &gofeed.Item{
		Title:       "Test",
		Description: "Short summary",
		Content:     "Full long content",
	}

	a := mapItem("https://example.com/feed", item)
	if a.Summary != "Short summary" {
		t.Errorf("Summary should prefer description, got %q", a.Summary)
	}
}
