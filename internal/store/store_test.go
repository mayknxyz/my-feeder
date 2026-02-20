package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mayknxyz/my-feeder/internal/model"
)

// --- State tests ---

func TestLoadState_FileNotExist(t *testing.T) {
	s, err := LoadState("/nonexistent/state.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Version != 1 {
		t.Errorf("version = %d, want 1", s.Version)
	}
	if len(s.Read) != 0 {
		t.Errorf("read list should be empty, got %d", len(s.Read))
	}
}

func TestState_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	state := &State{
		Version: 1,
		Read:    []string{"guid-1", "guid-2"},
	}

	if err := SaveState(path, state); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadState(path)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}

	if len(loaded.Read) != 2 {
		t.Fatalf("read count = %d, want 2", len(loaded.Read))
	}
	if loaded.Read[0] != "guid-1" || loaded.Read[1] != "guid-2" {
		t.Errorf("read = %v, want [guid-1 guid-2]", loaded.Read)
	}
}

func TestState_MarkRead(t *testing.T) {
	s := &State{Version: 1, Read: []string{}}

	s.MarkRead("guid-1")
	if !s.IsRead("guid-1") {
		t.Error("guid-1 should be read")
	}

	// Marking the same GUID again should not duplicate it.
	s.MarkRead("guid-1")
	if len(s.Read) != 1 {
		t.Errorf("duplicate mark: read count = %d, want 1", len(s.Read))
	}
}

func TestState_MarkUnread(t *testing.T) {
	s := &State{Version: 1, Read: []string{"guid-1", "guid-2", "guid-3"}}

	s.MarkUnread("guid-2")
	if s.IsRead("guid-2") {
		t.Error("guid-2 should be unread after MarkUnread")
	}
	if len(s.Read) != 2 {
		t.Errorf("read count = %d, want 2", len(s.Read))
	}

	// Removing a non-existent GUID should be a no-op.
	s.MarkUnread("guid-999")
	if len(s.Read) != 2 {
		t.Errorf("read count = %d, want 2 after no-op unread", len(s.Read))
	}
}

// --- Cache tests ---

func TestLoadCache_FileNotExist(t *testing.T) {
	c, err := LoadCache("/nonexistent/cache.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Version != 1 {
		t.Errorf("version = %d, want 1", c.Version)
	}
	if c.ArticleCount() != 0 {
		t.Errorf("article count = %d, want 0", c.ArticleCount())
	}
}

func TestCache_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	cache := newCache()
	now := time.Now()
	cache.SetArticles("https://example.com/feed.xml", []model.Article{
		{
			GUID:        "article-1",
			FeedURL:     "https://example.com/feed.xml",
			Title:       "Test Article",
			URL:         "https://example.com/post/1",
			PublishedAt: now,
			FetchedAt:   now,
		},
	})

	if err := SaveCache(path, cache); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadCache(path)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}

	if loaded.ArticleCount() != 1 {
		t.Fatalf("article count = %d, want 1", loaded.ArticleCount())
	}

	articles := loaded.ArticlesForFeed("https://example.com/feed.xml")
	if len(articles) != 1 {
		t.Fatalf("feed articles = %d, want 1", len(articles))
	}
	if articles[0].Title != "Test Article" {
		t.Errorf("title = %q, want %q", articles[0].Title, "Test Article")
	}
}

func TestCache_ArticlesForFeed_Empty(t *testing.T) {
	c := newCache()
	articles := c.ArticlesForFeed("https://nonexistent.com/feed")
	if len(articles) != 0 {
		t.Errorf("expected empty slice, got %d articles", len(articles))
	}
}

func TestCache_AllArticles(t *testing.T) {
	c := newCache()
	now := time.Now()

	c.SetArticles("feed-1", []model.Article{
		{GUID: "a1", Title: "Article 1", FetchedAt: now},
		{GUID: "a2", Title: "Article 2", FetchedAt: now},
	})
	c.SetArticles("feed-2", []model.Article{
		{GUID: "a3", Title: "Article 3", FetchedAt: now},
	})

	all := c.AllArticles()
	if len(all) != 3 {
		t.Errorf("all articles = %d, want 3", len(all))
	}
}

// --- Bookmark tests ---

func TestAppendBookmark(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bookmarks.md")

	bm := model.Bookmark{
		FeedName: "Go Blog",
		Title:    "Go 1.24 Released",
		URL:      "https://go.dev/blog/go1.24",
		Date:     time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC),
		Notes:    "Great release",
	}

	if err := AppendBookmark(path, bm); err != nil {
		t.Fatalf("append error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "## Go 1.24 Released") {
		t.Error("missing title heading")
	}
	if !strings.Contains(content, "**Source**: Go Blog") {
		t.Error("missing source")
	}
	if !strings.Contains(content, "**URL**: https://go.dev/blog/go1.24") {
		t.Error("missing URL")
	}
	if !strings.Contains(content, "**Date**: 2025-01-09") {
		t.Error("missing date")
	}
	if !strings.Contains(content, "**Notes**: Great release") {
		t.Error("missing notes")
	}
}

func TestAppendBookmark_Multiple(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bookmarks.md")

	now := time.Now()
	for i := range 3 {
		bm := model.Bookmark{
			FeedName: "Test",
			Title:    fmt.Sprintf("Article %d", i+1),
			URL:      fmt.Sprintf("https://example.com/%d", i+1),
			Date:     now,
		}
		if err := AppendBookmark(path, bm); err != nil {
			t.Fatalf("append %d error: %v", i, err)
		}
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	// Should have 3 separator lines (one per bookmark).
	if strings.Count(content, "---") != 3 {
		t.Errorf("expected 3 separators, got %d", strings.Count(content, "---"))
	}
}

func TestAppendBookmark_NoNotes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bookmarks.md")

	bm := model.Bookmark{
		FeedName: "Test",
		Title:    "No Notes Article",
		URL:      "https://example.com",
		Date:     time.Now(),
	}

	if err := AppendBookmark(path, bm); err != nil {
		t.Fatalf("append error: %v", err)
	}

	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "**Notes**") {
		t.Error("notes field should be omitted when empty")
	}
}
