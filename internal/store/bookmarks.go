package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mayknxyz/my-feeder/internal/model"
)

// AppendBookmark adds a bookmark entry to the markdown file. The file is
// append-only — entries are never removed, only added. This makes it
// safe for git sync and human editing.
func AppendBookmark(path string, bm model.Bookmark) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating directory for %s: %w", path, err)
	}

	// LEARN: os.OpenFile with O_APPEND|O_CREATE|O_WRONLY opens the file
	// for appending, creating it if it doesn't exist. This is safer than
	// reading + writing the whole file for append-only operations.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening bookmark file %s: %w", path, err)
	}
	defer f.Close()

	entry := formatBookmark(bm)
	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("writing bookmark: %w", err)
	}
	return nil
}

// formatBookmark renders a bookmark as a markdown entry.
func formatBookmark(bm model.Bookmark) string {
	s := fmt.Sprintf("## %s\n\n", bm.Title)
	s += fmt.Sprintf("- **Source**: %s\n", bm.FeedName)
	s += fmt.Sprintf("- **URL**: %s\n", bm.URL)
	s += fmt.Sprintf("- **Date**: %s\n", bm.Date.Format("2006-01-02"))
	if bm.Notes != "" {
		s += fmt.Sprintf("- **Notes**: %s\n", bm.Notes)
	}
	s += "\n---\n\n"
	return s
}
