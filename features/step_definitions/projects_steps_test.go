package steps

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"goapp/internal/githubapi"
)

// fakeGithubFetcher stands in for githubapi.DefaultFetcher in tests, so the
// Hebrew Cucumber suite never depends on network access or GitHub's rate
// limits. It returns a small, fixed activity summary for any githubURL.
func fakeGithubFetcher(_ context.Context, githubURL string) (*githubapi.Activity, error) {
	return &githubapi.Activity{
		Issues: []githubapi.IssueSummary{
			{Number: 1, Title: "Example issue", State: "open", URL: githubURL + "/issues/1"},
		},
		PRs: []githubapi.PRSummary{
			{Number: 2, Title: "Example pull request", State: "closed", URL: githubURL + "/pull/2"},
		},
		Commits: []githubapi.CommitSummary{
			{SHA: "abc1234", Message: "Example commit", Author: "Test Author", URL: githubURL + "/commit/abc1234"},
		},
	}, nil
}

func (ts *testState) addProject(name, githubURL string) error {
	resp, err := ts.client.PostForm(ts.server.URL+"/projects", url.Values{
		"project_name": {name},
		"github_url":   {githubURL},
	})
	if err != nil {
		return err
	}
	return ts.captureResponse(resp)
}

func (ts *testState) projectExists(name, githubURL string) error {
	_, err := ts.conn.Exec(`INSERT INTO projects (project_name, github_url) VALUES (?, ?)`, name, githubURL)
	return err
}

func (ts *testState) projectAppearsInList(name string) error {
	resp, err := ts.client.Get(ts.server.URL + "/projects")
	if err != nil {
		return err
	}
	if err := ts.captureResponse(resp); err != nil {
		return err
	}
	return ts.bodyContains(name)
}

func (ts *testState) openProjectScreen(name string) error {
	var id int64
	if err := ts.conn.QueryRow(`SELECT id FROM projects WHERE project_name = ?`, name).Scan(&id); err != nil {
		return fmt.Errorf("project %q not found: %w", name, err)
	}
	resp, err := ts.client.Get(ts.server.URL + "/projects/" + strconv.FormatInt(id, 10))
	if err != nil {
		return err
	}
	return ts.captureResponse(resp)
}

func (ts *testState) githubActivityDisplayed() error {
	for _, needle := range []string{"Example issue", "Example pull request", "Example commit"} {
		if err := ts.bodyContains(needle); err != nil {
			return err
		}
	}
	return nil
}
