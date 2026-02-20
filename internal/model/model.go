// Package model defines the core data types shared across all packages.
// These structs represent feeds, articles, and bookmarks — the nouns of the
// application. They are plain data holders with no business logic.
package model

import "time"

// Feed represents a single feed source from the config file.
// It can be an RSS/Atom feed or a GitHub release tracker.
type Feed struct {
	Name          string `toml:"name" json:"name"`
	URL           string `toml:"url" json:"url"`
	Tag           string `toml:"tag,omitempty" json:"tag,omitempty"`
	RetentionDays *int   `toml:"retention_days,omitempty" json:"retention_days,omitempty"`
}

// IsGitHub reports whether this feed tracks GitHub releases
// rather than an RSS/Atom feed.
func (f Feed) IsGitHub() bool {
	// WHY: We use a prefix convention ("github:owner/repo") instead of a
	// separate field because it keeps the config simpler — one URL field
	// handles all feed types.
	return len(f.URL) > 7 && f.URL[:7] == "github:"
}

// GitHubRepo extracts the "owner/repo" portion from a GitHub feed URL.
// Returns an empty string if this is not a GitHub feed.
func (f Feed) GitHubRepo() string {
	if !f.IsGitHub() {
		return ""
	}
	return f.URL[7:]
}

// Article represents a single entry from a feed (RSS item, Atom entry,
// or GitHub release). This is the primary unit of content in the app.
type Article struct {
	GUID            string    `json:"guid"`
	FeedURL         string    `json:"feed_url"`
	Title           string    `json:"title"`
	URL             string    `json:"url,omitempty"`
	Author          string    `json:"author,omitempty"`
	Summary         string    `json:"summary,omitempty"`
	Content         string    `json:"content,omitempty"`
	PublishedAt     time.Time `json:"published_at"`
	FetchedAt       time.Time `json:"fetched_at"`
	NormalizedTitle string    `json:"normalized_title,omitempty"`
}

// Bookmark represents a saved article with optional user notes.
type Bookmark struct {
	FeedName  string    `json:"feed_name"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Date      time.Time `json:"date"`
	Notes     string    `json:"notes,omitempty"`
	SavedAt   time.Time `json:"saved_at"`
}
