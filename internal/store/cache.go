package store

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mayknxyz/my-feeder/internal/model"
)

// Cache holds fetched articles grouped by feed URL. This file is local
// and ephemeral — it can be deleted and rebuilt by re-fetching feeds.
type Cache struct {
	Version     int                    `json:"version"`
	Articles    map[string][]model.Article `json:"articles"`
	LastFetched map[string]string      `json:"last_fetched"`
}

// LoadCache reads the cache file from disk. If the file doesn't exist,
// returns an empty cache — this is expected on first run or after
// clearing the cache.
func LoadCache(path string) (*Cache, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return newCache(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading cache file %s: %w", path, err)
	}

	var c Cache
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parsing cache file %s: %w", path, err)
	}

	// Ensure maps are initialized even if the JSON had null values.
	if c.Articles == nil {
		c.Articles = make(map[string][]model.Article)
	}
	if c.LastFetched == nil {
		c.LastFetched = make(map[string]string)
	}
	return &c, nil
}

// SaveCache writes the cache to disk atomically.
func SaveCache(path string, cache *Cache) error {
	return writeJSON(path, cache)
}

// ArticlesForFeed returns all cached articles for a given feed URL.
// Returns an empty slice if the feed has no cached articles.
func (c *Cache) ArticlesForFeed(feedURL string) []model.Article {
	articles, ok := c.Articles[feedURL]
	if !ok {
		return []model.Article{}
	}
	return articles
}

// AllArticles returns every cached article across all feeds.
func (c *Cache) AllArticles() []model.Article {
	var all []model.Article
	for _, articles := range c.Articles {
		all = append(all, articles...)
	}
	return all
}

// SetArticles replaces all cached articles for a feed URL.
func (c *Cache) SetArticles(feedURL string, articles []model.Article) {
	c.Articles[feedURL] = articles
}

// ArticleCount returns the total number of cached articles across all feeds.
func (c *Cache) ArticleCount() int {
	count := 0
	for _, articles := range c.Articles {
		count += len(articles)
	}
	return count
}

// newCache creates an empty cache with initialized maps.
func newCache() *Cache {
	return &Cache{
		Version:     1,
		Articles:    make(map[string][]model.Article),
		LastFetched: make(map[string]string),
	}
}
