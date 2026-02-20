// Package config handles loading and validating the TOML configuration file.
// The config file is the source of truth for the feed list and application
// settings. It lives at $XDG_CONFIG_HOME/feeder/config.toml by default.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"
	"github.com/mayknxyz/my-feeder/internal/model"
)

// defaultSettings provides sensible defaults for all settings.
// These are used when a setting is omitted from the config file.
var defaultSettings = Settings{
	RefreshIntervalMinutes: 30,
	RetentionDays:          7,
}

// Config is the top-level configuration loaded from config.toml.
type Config struct {
	Settings Settings     `toml:"settings"`
	Feeds    []model.Feed `toml:"feeds"`
}

// Settings holds global application preferences.
type Settings struct {
	RefreshIntervalMinutes int    `toml:"refresh_interval_minutes"`
	RetentionDays          int    `toml:"retention_days"`
	BookmarkFile           string `toml:"bookmark_file"`
	StateFile              string `toml:"state_file"`
	CacheFile              string `toml:"cache_file"`
	GitHubToken            string `toml:"github_token,omitempty"`
}

// DefaultConfigPath returns the default config file location following
// the XDG Base Directory Specification.
func DefaultConfigPath() string {
	// LEARN: xdg.ConfigHome resolves to $XDG_CONFIG_HOME if set,
	// otherwise falls back to ~/.config. This is the standard on Linux
	// for user-specific config files.
	return filepath.Join(xdg.ConfigHome, "feeder", "config.toml")
}

// Load reads and parses the config file at the given path.
// If path is empty, it uses the default XDG config path.
func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}

	cfg := &Config{
		Settings: defaultSettings,
	}

	// LEARN: toml.Decode merges into the existing struct, so defaults
	// set above are preserved when the TOML file omits those fields.
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	cfg.resolveDefaults()

	return cfg, nil
}

// validate checks that the config has the minimum required fields.
func (c *Config) validate() error {
	if len(c.Feeds) == 0 {
		return fmt.Errorf("no feeds configured — add at least one [[feeds]] entry")
	}
	for i, f := range c.Feeds {
		if f.Name == "" {
			return fmt.Errorf("feed #%d: missing name", i+1)
		}
		if f.URL == "" {
			return fmt.Errorf("feed %q: missing url", f.Name)
		}
	}
	if c.Settings.RefreshIntervalMinutes < 1 {
		return fmt.Errorf("refresh_interval_minutes must be >= 1")
	}
	if c.Settings.RetentionDays < 1 {
		return fmt.Errorf("retention_days must be >= 1")
	}
	return nil
}

// resolveDefaults fills in any settings that weren't specified in the
// config file with XDG-compliant default paths.
func (c *Config) resolveDefaults() {
	if c.Settings.BookmarkFile == "" {
		c.Settings.BookmarkFile = filepath.Join(xdg.DataHome, "feeder", "bookmarks.md")
	} else {
		c.Settings.BookmarkFile = expandHome(c.Settings.BookmarkFile)
	}

	if c.Settings.StateFile == "" {
		c.Settings.StateFile = filepath.Join(xdg.DataHome, "feeder", "state.json")
	} else {
		c.Settings.StateFile = expandHome(c.Settings.StateFile)
	}

	if c.Settings.CacheFile == "" {
		// WHY: Cache goes under XDG cache dir, not data dir, because it's
		// ephemeral and can be rebuilt by re-fetching feeds.
		c.Settings.CacheFile = filepath.Join(xdg.CacheHome, "feeder", "cache.json")
	} else {
		c.Settings.CacheFile = expandHome(c.Settings.CacheFile)
	}
}

// RetentionDays returns the effective retention for a feed, falling back
// to the global setting if the feed doesn't specify one.
func (c *Config) RetentionDays(feed model.Feed) int {
	if feed.RetentionDays != nil {
		return *feed.RetentionDays
	}
	return c.Settings.RetentionDays
}

// expandHome replaces a leading "~/" with the user's home directory.
func expandHome(path string) string {
	if len(path) < 2 || path[:2] != "~/" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[2:])
}
