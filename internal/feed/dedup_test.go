package feed

import (
	"testing"
	"time"

	"github.com/mayknxyz/my-feeder/internal/model"
)

func TestNormalizeTitle(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Go 1.24 Released!", "go 124 released"},
		{"  Multiple   Spaces  ", "multiple spaces"},
		{"Rust's New Feature (v2.0)", "rusts new feature v20"},
		{"ALL CAPS TITLE", "all caps title"},
		{"", ""},
		{"simple", "simple"},
		{"Hello, World! — 2025", "hello world 2025"},
	}

	for _, tt := range tests {
		got := NormalizeTitle(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeTitle(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsDuplicate_GUIDMatch(t *testing.T) {
	existing := []model.Article{
		{GUID: "abc-123", Title: "Old Article"},
	}
	article := model.Article{GUID: "abc-123", Title: "Totally Different Title"}

	if !IsDuplicate(article, existing, 7) {
		t.Error("should detect GUID duplicate")
	}
}

func TestIsDuplicate_URLMatch(t *testing.T) {
	existing := []model.Article{
		{GUID: "different-guid", URL: "https://example.com/post/1", Title: "Old"},
	}
	article := model.Article{GUID: "new-guid", URL: "https://example.com/post/1", Title: "New"}

	if !IsDuplicate(article, existing, 7) {
		t.Error("should detect URL duplicate")
	}
}

func TestIsDuplicate_FuzzyTitleMatch(t *testing.T) {
	now := time.Now()
	existing := []model.Article{
		{
			GUID:            "old",
			URL:             "https://example.com/old",
			Title:           "Go 1.24 Released",
			NormalizedTitle: NormalizeTitle("Go 1.24 Released"),
			PublishedAt:     now.AddDate(0, 0, -3),
		},
	}
	article := model.Article{
		GUID:            "new",
		URL:             "https://other.com/new",
		Title:           "Go 1.24.0 Released",
		NormalizedTitle: NormalizeTitle("Go 1.24.0 Released"),
	}

	if !IsDuplicate(article, existing, 7) {
		t.Error("should detect fuzzy title duplicate")
	}
}

func TestIsDuplicate_FuzzyTitleNoMatch(t *testing.T) {
	now := time.Now()
	existing := []model.Article{
		{
			GUID:            "old",
			URL:             "https://example.com/old",
			Title:           "Go 1.24 Released",
			NormalizedTitle: NormalizeTitle("Go 1.24 Released"),
			PublishedAt:     now.AddDate(0, 0, -3),
		},
	}
	article := model.Article{
		GUID:            "new",
		URL:             "https://other.com/new",
		Title:           "Rust 2.0 Announced",
		NormalizedTitle: NormalizeTitle("Rust 2.0 Announced"),
	}

	if IsDuplicate(article, existing, 7) {
		t.Error("should NOT detect completely different title as duplicate")
	}
}

func TestIsDuplicate_FuzzyTitleOutsideWindow(t *testing.T) {
	existing := []model.Article{
		{
			GUID:            "old",
			URL:             "https://example.com/old",
			Title:           "Go 1.24 Released",
			NormalizedTitle: NormalizeTitle("Go 1.24 Released"),
			PublishedAt:     time.Now().AddDate(0, 0, -30),
		},
	}
	article := model.Article{
		GUID:            "new",
		URL:             "https://other.com/new",
		Title:           "Go 1.24.0 Released",
		NormalizedTitle: NormalizeTitle("Go 1.24.0 Released"),
	}

	if IsDuplicate(article, existing, 7) {
		t.Error("should NOT match fuzzy title outside dedup window")
	}
}

func TestIsDuplicate_NoDuplicate(t *testing.T) {
	existing := []model.Article{
		{GUID: "existing", URL: "https://example.com/1", Title: "Existing"},
	}
	article := model.Article{
		GUID: "brand-new",
		URL:  "https://example.com/2",
	}

	if IsDuplicate(article, existing, 7) {
		t.Error("should NOT detect as duplicate")
	}
}

func TestIsDuplicate_EmptyExisting(t *testing.T) {
	article := model.Article{GUID: "new", Title: "New Article"}
	if IsDuplicate(article, nil, 7) {
		t.Error("should not be duplicate against empty list")
	}
}
