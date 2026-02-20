// Package main is the entry point for feeder — a TUI feed reader.
// In Phase 2, it loads config, fetches all feeds, deduplicates, caches
// articles, and prints a summary. The TUI will be added in Phase 3.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/mayknxyz/my-feeder/internal/config"
	"github.com/mayknxyz/my-feeder/internal/feed"
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

	// Fetch all feeds concurrently.
	fetcher := &feed.Fetcher{
		Feeds:       cfg.Feeds,
		Cache:       cache,
		GitHubToken: cfg.Settings.GitHubToken,
		RetentionFn: cfg.RetentionDays,
	}

	fmt.Println("Fetching feeds...")
	results := fetcher.RefreshAll(context.Background())

	// Expire old articles.
	fetcher.ExpireOld()

	// Save updated cache.
	if err := store.SaveCache(cfg.Settings.CacheFile, cache); err != nil {
		log.Fatal("Failed to save cache", "error", err)
	}

	// Print summary.
	fmt.Println()
	fmt.Println("Feeder")
	fmt.Println("======")
	fmt.Printf("Feeds configured: %d\n", len(cfg.Feeds))
	fmt.Printf("Cached articles:  %d\n", cache.ArticleCount())
	fmt.Printf("Articles read:    %d\n", len(state.Read))
	fmt.Println()

	for _, r := range results {
		status := "ok"
		if r.Err != nil {
			status = fmt.Sprintf("error: %v", r.Err)
		}

		feedType := "rss"
		if r.Feed.IsGitHub() {
			feedType = "github"
		}
		tag := r.Feed.Tag
		if tag == "" {
			tag = "-"
		}

		count := len(cache.ArticlesForFeed(r.Feed.URL))
		newCount := len(r.Articles)

		fmt.Printf("  [%s] [%s] %-25s %3d articles (%d new)  %s\n",
			tag, feedType, r.Feed.Name, count, newCount, status)
	}

	fmt.Println()
	fmt.Printf("Config:    %s\n", config.DefaultConfigPath())
	fmt.Printf("Cache:     %s\n", cfg.Settings.CacheFile)
	fmt.Printf("State:     %s\n", cfg.Settings.StateFile)
	fmt.Printf("Bookmarks: %s\n", cfg.Settings.BookmarkFile)
}
