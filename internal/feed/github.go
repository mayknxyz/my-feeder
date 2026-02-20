package feed

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v68/github"
	"github.com/mayknxyz/my-feeder/internal/model"
)

// FetchGitHubReleases fetches releases from a GitHub repository and
// maps them to Article structs. The repo string should be "owner/repo".
// If token is empty, unauthenticated requests are used (lower rate limit).
func FetchGitHubReleases(ctx context.Context, repo string, token string) ([]model.Article, error) {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid GitHub repo format %q, expected owner/repo", repo)
	}
	owner, repoName := parts[0], parts[1]

	client := newGitHubClient(token)

	// LEARN: ListReleases returns paginated results. For a personal
	// reader we only need the first page (most recent releases).
	releases, _, err := client.Repositories.ListReleases(ctx, owner, repoName, &github.ListOptions{
		PerPage: 25,
	})
	if err != nil {
		return nil, fmt.Errorf("fetching releases for %s: %w", repo, err)
	}

	articles := make([]model.Article, 0, len(releases))
	for _, rel := range releases {
		if rel.Draft != nil && *rel.Draft {
			continue
		}
		articles = append(articles, mapRelease(repo, rel))
	}
	return articles, nil
}

// newGitHubClient creates a GitHub API client, optionally authenticated.
func newGitHubClient(token string) *github.Client {
	if token != "" {
		// LEARN: go-github uses a static token for authentication.
		// This increases the rate limit from 60 to 5000 requests/hour.
		return github.NewClient(nil).WithAuthToken(token)
	}
	return github.NewClient(nil)
}

// mapRelease converts a GitHub release to an Article.
func mapRelease(repo string, rel *github.RepositoryRelease) model.Article {
	a := model.Article{
		GUID:    fmt.Sprintf("github:%s:%d", repo, rel.GetID()),
		FeedURL: "github:" + repo,
		Title:   releaseTitle(repo, rel),
		URL:     rel.GetHTMLURL(),
		Author:  releaseAuthor(rel),
		Summary: truncate(rel.GetBody(), 200),
		Content: rel.GetBody(),
	}

	if rel.PublishedAt != nil {
		a.PublishedAt = rel.PublishedAt.Time
	} else if rel.CreatedAt != nil {
		a.PublishedAt = rel.CreatedAt.Time
	} else {
		a.PublishedAt = time.Now()
	}

	a.FetchedAt = time.Now()
	a.NormalizedTitle = NormalizeTitle(a.Title)

	return a
}

// releaseTitle builds a human-readable title from the release.
func releaseTitle(repo string, rel *github.RepositoryRelease) string {
	name := rel.GetName()
	tag := rel.GetTagName()

	if name != "" {
		return name
	}
	if tag != "" {
		// WHY: Many repos only set the tag, not the release name.
		// "owner/repo v1.2.3" is more readable than just "v1.2.3".
		return repo + " " + tag
	}
	return repo + " release"
}

// releaseAuthor extracts the author login from the release.
func releaseAuthor(rel *github.RepositoryRelease) string {
	if rel.Author != nil {
		return rel.Author.GetLogin()
	}
	return ""
}

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
