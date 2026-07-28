package githubapi

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
)

// NewClient returns a GitHub API client. If GITHUB_TOKEN is set it's used
// for authenticated requests (needed for private repos, and to get GitHub's
// much higher authenticated rate limit); otherwise an unauthenticated
// client is returned, which still works for public repos.
func NewClient(ctx context.Context) *github.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return github.NewClient(nil)
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return github.NewClient(oauth2.NewClient(ctx, ts))
}

// ParseOwnerRepo extracts "owner" and "repo" from a GitHub URL such as
// "https://github.com/owner/repo" or "https://github.com/owner/repo.git".
func ParseOwnerRepo(githubURL string) (owner, repo string, err error) {
	u, err := url.Parse(strings.TrimSpace(githubURL))
	if err != nil {
		return "", "", fmt.Errorf("invalid GitHub URL %q: %w", githubURL, err)
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("could not find owner/repo in GitHub URL %q", githubURL)
	}
	return parts[0], strings.TrimSuffix(parts[1], ".git"), nil
}

type IssueSummary struct {
	Number int
	Title  string
	State  string
	URL    string
}

type PRSummary struct {
	Number int
	Title  string
	State  string
	URL    string
}

type CommitSummary struct {
	SHA     string
	Message string
	Author  string
	URL     string
}

type Activity struct {
	Issues  []IssueSummary
	PRs     []PRSummary
	Commits []CommitSummary
}

// Fetcher loads the activity summary for a project's GitHub URL. It's a
// seam so handlers can be tested without hitting the real GitHub API.
type Fetcher func(ctx context.Context, githubURL string) (*Activity, error)

// DefaultFetcher is the real implementation: parse owner/repo out of the
// URL, build a client (authenticated if GITHUB_TOKEN is set), and fetch.
func DefaultFetcher(ctx context.Context, githubURL string) (*Activity, error) {
	owner, repo, err := ParseOwnerRepo(githubURL)
	if err != nil {
		return nil, err
	}
	client := NewClient(ctx)
	return FetchActivity(ctx, client, owner, repo)
}

// FetchActivity reads the most recent issues, pull requests, and commits
// for owner/repo from the GitHub API.
func FetchActivity(ctx context.Context, client *github.Client, owner, repo string) (*Activity, error) {
	issues, _, err := client.Issues.ListByRepo(ctx, owner, repo, &github.IssueListByRepoOptions{
		State:       "all",
		ListOptions: github.ListOptions{PerPage: 10},
	})
	if err != nil {
		return nil, fmt.Errorf("list issues: %w", err)
	}

	prs, _, err := client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
		State:       "all",
		ListOptions: github.ListOptions{PerPage: 10},
	})
	if err != nil {
		return nil, fmt.Errorf("list pull requests: %w", err)
	}

	commits, _, err := client.Repositories.ListCommits(ctx, owner, repo, &github.CommitsListOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	})
	if err != nil {
		return nil, fmt.Errorf("list commits: %w", err)
	}

	activity := &Activity{}
	for _, issue := range issues {
		if issue.IsPullRequest() {
			continue // GitHub's issues endpoint also returns PRs; we list those separately above
		}
		activity.Issues = append(activity.Issues, IssueSummary{
			Number: issue.GetNumber(),
			Title:  issue.GetTitle(),
			State:  issue.GetState(),
			URL:    issue.GetHTMLURL(),
		})
	}
	for _, pr := range prs {
		activity.PRs = append(activity.PRs, PRSummary{
			Number: pr.GetNumber(),
			Title:  pr.GetTitle(),
			State:  pr.GetState(),
			URL:    pr.GetHTMLURL(),
		})
	}
	for _, commit := range commits {
		message := commit.GetCommit().GetMessage()
		if idx := strings.IndexByte(message, '\n'); idx >= 0 {
			message = message[:idx]
		}
		activity.Commits = append(activity.Commits, CommitSummary{
			SHA:     commit.GetSHA(),
			Message: message,
			Author:  commit.GetCommit().GetAuthor().GetName(),
			URL:     commit.GetHTMLURL(),
		})
	}
	return activity, nil
}
