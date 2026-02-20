// Package main is the entry point for feeder — a TUI feed reader.
// In Phase 1, it loads the config, initializes storage files, and
// prints a summary. The TUI will be added in Phase 3.
package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/mayknxyz/my-feeder/internal/config"
	"github.com/mayknxyz/my-feeder/internal/store"
)

func main() {
	// TODO: Replace with cobra CLI arg parsing in Phase 6.
	configPath := ""
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatal("Failed to load config", "error", err)
	}

	log.Info("Config loaded", "feeds", len(cfg.Feeds))

	// Load or initialize storage files.
	cache, err := store.LoadCache(cfg.Settings.CacheFile)
	if err != nil {
		log.Fatal("Failed to load cache", "error", err)
	}

	state, err := store.LoadState(cfg.Settings.StateFile)
	if err != nil {
		log.Fatal("Failed to load state", "error", err)
	}

	// Print summary.
	fmt.Println("Feeder")
	fmt.Println("======")
	fmt.Printf("Feeds configured: %d\n", len(cfg.Feeds))
	fmt.Printf("Cached articles:  %d\n", cache.ArticleCount())
	fmt.Printf("Articles read:    %d\n", len(state.Read))
	fmt.Println()

	for _, feed := range cfg.Feeds {
		feedType := "rss"
		if feed.IsGitHub() {
			feedType = "github"
		}
		tag := feed.Tag
		if tag == "" {
			tag = "-"
		}
		fmt.Printf("  [%s] [%s] %s\n", tag, feedType, feed.Name)
	}

	fmt.Println()
	fmt.Printf("Config:    %s\n", config.DefaultConfigPath())
	fmt.Printf("Cache:     %s\n", cfg.Settings.CacheFile)
	fmt.Printf("State:     %s\n", cfg.Settings.StateFile)
	fmt.Printf("Bookmarks: %s\n", cfg.Settings.BookmarkFile)
}
