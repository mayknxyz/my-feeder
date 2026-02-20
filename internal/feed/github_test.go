package feed

import (
	"testing"
	"time"

	"github.com/google/go-github/v68/github"
)

func TestMapRelease(t *testing.T) {
	id := int64(12345)
	name := "v1.0.0 — Initial Release"
	tag := "v1.0.0"
	htmlURL := "https://github.com/owner/repo/releases/tag/v1.0.0"
	body := "First stable release with all the features."
	login := "contributor"
	pub := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	ts := github.Timestamp{Time: pub}

	rel := &github.RepositoryRelease{
		ID:          &id,
		Name:        &name,
		TagName:     &tag,
		HTMLURL:     &htmlURL,
		Body:        &body,
		Author:      &github.User{Login: &login},
		PublishedAt: &ts,
	}

	a := mapRelease("owner/repo", rel)

	if a.GUID != "github:owner/repo:12345" {
		t.Errorf("GUID = %q", a.GUID)
	}
	if a.FeedURL != "github:owner/repo" {
		t.Errorf("FeedURL = %q", a.FeedURL)
	}
	if a.Title != "v1.0.0 — Initial Release" {
		t.Errorf("Title = %q", a.Title)
	}
	if a.URL != htmlURL {
		t.Errorf("URL = %q", a.URL)
	}
	if a.Author != "contributor" {
		t.Errorf("Author = %q", a.Author)
	}
	if a.Content != body {
		t.Errorf("Content = %q", a.Content)
	}
	if !a.PublishedAt.Equal(pub) {
		t.Errorf("PublishedAt = %v, want %v", a.PublishedAt, pub)
	}
}

func TestMapRelease_NoName_FallsBackToTag(t *testing.T) {
	id := int64(1)
	tag := "v2.0.0"
	rel := &github.RepositoryRelease{
		ID:      &id,
		TagName: &tag,
	}

	a := mapRelease("tokio-rs/tokio", rel)
	if a.Title != "tokio-rs/tokio v2.0.0" {
		t.Errorf("Title = %q, want %q", a.Title, "tokio-rs/tokio v2.0.0")
	}
}

func TestMapRelease_NoNameNoTag(t *testing.T) {
	id := int64(1)
	rel := &github.RepositoryRelease{
		ID: &id,
	}

	a := mapRelease("owner/repo", rel)
	if a.Title != "owner/repo release" {
		t.Errorf("Title = %q, want %q", a.Title, "owner/repo release")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly ten", 11, "exactly ten"},
		{"this is a long string that should be truncated", 20, "this is a long st..."},
		{"", 10, ""},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}
