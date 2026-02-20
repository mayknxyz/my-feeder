// Package feed handles fetching, parsing, and deduplicating content
// from RSS/Atom feeds and GitHub releases.
package feed

import (
	"strings"
	"time"
	"unicode"

	"github.com/mayknxyz/my-feeder/internal/model"
	"github.com/xrash/smetrics"
)

// dedupThreshold is the minimum Jaro-Winkler similarity score for two
// normalized titles to be considered duplicates.
const dedupThreshold = 0.85

// NormalizeTitle prepares a title for dedup comparison: lowercase,
// strip punctuation, collapse whitespace.
func NormalizeTitle(title string) string {
	title = strings.ToLower(title)

	// LEARN: strings.Map applies a function to every rune. Returning -1
	// drops the rune. We keep letters, digits, and spaces.
	title = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			return r
		}
		return -1
	}, title)

	// Collapse multiple spaces into one and trim.
	fields := strings.Fields(title)
	return strings.Join(fields, " ")
}

// IsDuplicate checks whether an article is a duplicate of any existing
// article using the 3-tier strategy: GUID → URL → fuzzy title.
// retentionDays controls the window for fuzzy title matching.
func IsDuplicate(article model.Article, existing []model.Article, retentionDays int) bool {
	// Tier 1: exact GUID match.
	if article.GUID != "" {
		for _, e := range existing {
			if e.GUID == article.GUID {
				return true
			}
		}
	}

	// Tier 2: exact URL match.
	if article.URL != "" {
		for _, e := range existing {
			if e.URL == article.URL {
				return true
			}
		}
	}

	// Tier 3: fuzzy title match within the dedup window.
	if article.NormalizedTitle == "" {
		return false
	}

	// WHY: The dedup window uses max(retentionDays, 14) to avoid a gap
	// where articles expire from cache but could still be re-fetched
	// and re-inserted as "new".
	window := retentionDays
	if window < 14 {
		window = 14
	}
	cutoff := time.Now().AddDate(0, 0, -window)

	for _, e := range existing {
		if e.PublishedAt.Before(cutoff) {
			continue
		}
		if e.NormalizedTitle == "" {
			continue
		}

		// WHY: Jaro-Winkler penalizes early-character mismatches more
		// than late ones, which suits article titles where the meaningful
		// words tend to come first ("Go 1.24 Released" vs "Go 1.24.0 Released").
		score := smetrics.JaroWinkler(article.NormalizedTitle, e.NormalizedTitle, 0.7, 4)
		if score >= dedupThreshold {
			return true
		}
	}

	return false
}
