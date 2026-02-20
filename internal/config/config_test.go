package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[settings]
refresh_interval_minutes = 15
retention_days = 14

[[feeds]]
name = "Go Blog"
url = "https://go.dev/blog/feed.atom"
tag = "go"

[[feeds]]
name = "Tokio"
url = "github:tokio-rs/tokio"
retention_days = 30
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Settings.RefreshIntervalMinutes != 15 {
		t.Errorf("refresh interval = %d, want 15", cfg.Settings.RefreshIntervalMinutes)
	}
	if cfg.Settings.RetentionDays != 14 {
		t.Errorf("retention days = %d, want 14", cfg.Settings.RetentionDays)
	}
	if len(cfg.Feeds) != 2 {
		t.Fatalf("feeds count = %d, want 2", len(cfg.Feeds))
	}
	if cfg.Feeds[0].Name != "Go Blog" {
		t.Errorf("feed[0].Name = %q, want %q", cfg.Feeds[0].Name, "Go Blog")
	}
	if cfg.Feeds[1].IsGitHub() != true {
		t.Error("feed[1] should be detected as GitHub")
	}
	if cfg.Feeds[1].GitHubRepo() != "tokio-rs/tokio" {
		t.Errorf("feed[1].GitHubRepo() = %q, want %q", cfg.Feeds[1].GitHubRepo(), "tokio-rs/tokio")
	}
}

func TestLoad_Defaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	// Minimal config — only required fields, settings use defaults
	content := `
[[feeds]]
name = "Test Feed"
url = "https://example.com/feed.xml"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Settings.RefreshIntervalMinutes != 30 {
		t.Errorf("default refresh interval = %d, want 30", cfg.Settings.RefreshIntervalMinutes)
	}
	if cfg.Settings.RetentionDays != 7 {
		t.Errorf("default retention days = %d, want 7", cfg.Settings.RetentionDays)
	}
	if cfg.Settings.CacheFile == "" {
		t.Error("cache file path should have a default")
	}
	if cfg.Settings.StateFile == "" {
		t.Error("state file path should have a default")
	}
	if cfg.Settings.BookmarkFile == "" {
		t.Error("bookmark file path should have a default")
	}
}

func TestLoad_NoFeeds(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[settings]
refresh_interval_minutes = 30
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for config with no feeds")
	}
}

func TestLoad_MissingFeedName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[[feeds]]
url = "https://example.com/feed.xml"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for feed with no name")
	}
}

func TestLoad_MissingFeedURL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[[feeds]]
name = "No URL"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for feed with no url")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.toml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_ExpandHome(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[settings]
bookmark_file = "~/my-bookmarks.md"
state_file = "~/my-state.json"
cache_file = "~/my-cache.json"

[[feeds]]
name = "Test"
url = "https://example.com/feed.xml"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	home, _ := os.UserHomeDir()
	if cfg.Settings.BookmarkFile != filepath.Join(home, "my-bookmarks.md") {
		t.Errorf("bookmark file = %q, want ~ expanded", cfg.Settings.BookmarkFile)
	}
}

func TestRetentionDays_FeedOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[settings]
retention_days = 7

[[feeds]]
name = "Default"
url = "https://example.com/feed.xml"

[[feeds]]
name = "Custom"
url = "https://example.com/other.xml"
retention_days = 30
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if days := cfg.RetentionDays(cfg.Feeds[0]); days != 7 {
		t.Errorf("default feed retention = %d, want 7", days)
	}
	if days := cfg.RetentionDays(cfg.Feeds[1]); days != 30 {
		t.Errorf("custom feed retention = %d, want 30", days)
	}
}
