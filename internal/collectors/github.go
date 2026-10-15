package collectors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/coreos/coreos-newsletter/internal/config"
	"github.com/coreos/coreos-newsletter/internal/models"
	"github.com/coreos/coreos-newsletter/internal/util"
)

const (
	githubAPIBase = "https://api.github.com"
	githubTimeout = 30 * time.Second
)

// GitHubCollector fetches PRs from GitHub repositories
type GitHubCollector struct {
	config *config.Config
	logger *util.Logger
	client *http.Client
}

// NewGitHubCollector creates a new GitHub collector
func NewGitHubCollector(cfg *config.Config, logger *util.Logger) *GitHubCollector {
	return &GitHubCollector{
		config: cfg,
		logger: logger,
		client: &http.Client{Timeout: githubTimeout},
	}
}

// FetchPRs fetches pull requests from configured GitHub repositories
func (gc *GitHubCollector) FetchPRs(windowStart, windowEnd time.Time) ([]models.PullRequestRecord, []string) {
	var prs []models.PullRequestRecord
	var warnings []string

	// Check if GitHub is configured
	if gc.config.GitHubToken == "" || gc.config.GitHubOwner == "" {
		warnings = append(warnings, "GitHub skipped: set GITHUB_TOKEN and GITHUB_OWNER")
		return prs, warnings
	}

	repos := gc.config.GitHubRepoList()
	if len(repos) == 0 {
		warnings = append(warnings, "GitHub skipped: set GITHUB_REPO")
		return prs, warnings
	}

	// Fetch PRs for each repository
	for _, repo := range repos {
		repoPRs, repoWarnings := gc.fetchOneRepo(gc.config.GitHubOwner, repo, windowStart, windowEnd)
		prs = append(prs, repoPRs...)
		warnings = append(warnings, repoWarnings...)
	}

	return prs, warnings
}

// fetchOneRepo fetches PRs from a single repository with owner
func (gc *GitHubCollector) fetchOneRepo(owner, repo string, windowStart, windowEnd time.Time) ([]models.PullRequestRecord, []string) {
	var prs []models.PullRequestRecord
	var warnings []string

	url := fmt.Sprintf("%s/repos/%s/%s/pulls", githubAPIBase, owner, repo)

	// Fetch PR list
	req, _ := http.NewRequest("GET", url, nil)
	gc.addHeaders(req)
	q := req.URL.Query()
	q.Add("state", "all")
	q.Add("sort", "updated")
	q.Add("direction", "desc")
	q.Add("per_page", "100")
	req.URL.RawQuery = q.Encode()

	resp, err := gc.client.Do(req)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("GitHub repo %s/%s: %v", owner, repo, err))
		return prs, warnings
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		warnings = append(warnings, fmt.Sprintf("GitHub repo %s/%s: HTTP %d", owner, repo, resp.StatusCode))
		return prs, warnings
	}

	var items []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		warnings = append(warnings, fmt.Sprintf("GitHub repo %s/%s: decode error", owner, repo))
		return prs, warnings
	}

	// Process each PR
	for _, raw := range items {
		pr := gc.processPR(raw, owner, repo, windowStart, windowEnd)
		if pr != nil {
			prs = append(prs, *pr)
		}
	}

	return prs, warnings
}

// processPR processes a single PR from the list response
func (gc *GitHubCollector) processPR(raw map[string]interface{}, owner, repo string, windowStart, windowEnd time.Time) *models.PullRequestRecord {
	// Parse updated_at
	updatedAtStr, ok := raw["updated_at"].(string)
	if !ok {
		return nil
	}

	updatedAt, err := util.ParseDate(updatedAtStr)
	if err != nil || updatedAt == nil {
		return nil
	}

	// Check time window
	if updatedAt.Before(windowStart.UTC()) {
		return nil
	}
	if updatedAt.After(windowEnd.UTC()) {
		return nil
	}

	// Use list data by default, only fetch details if needed
	d := raw
	if d["merged_at"] == nil && d["state"] == "closed" {
		// Fetch detail for closed PRs without merged_at
		if num, ok := raw["number"].(float64); ok {
			detailPR := gc.fetchPRDetail(owner, repo, int64(num))
			if detailPR != nil {
				d = detailPR
			}
		}
	}

	// Extract labels
	var labels []string
	if labelsList, ok := d["labels"].([]interface{}); ok {
		for _, l := range labelsList {
			if labelMap, ok := l.(map[string]interface{}); ok {
				if name, ok := labelMap["name"].(string); ok {
					labels = append(labels, name)
				}
			}
		}
	}

	// Extract author
	var author *string
	if user, ok := d["user"].(map[string]interface{}); ok {
		if login, ok := user["login"].(string); ok {
			author = util.StringPtr(login)
		}
	}

	// Extract merged_at
	var mergedAt *time.Time
	if mergedAtStr, ok := d["merged_at"].(string); ok && mergedAtStr != "" {
		if mt, err := util.ParseDate(mergedAtStr); err == nil {
			mergedAt = mt
		}
	}

	// Extract counts
	var additions, deletions, changedFiles *int
	if add, ok := d["additions"].(float64); ok {
		additions = util.IntPtr(int(add))
	}
	if del, ok := d["deletions"].(float64); ok {
		deletions = util.IntPtr(int(del))
	}
	if files, ok := d["changed_files"].(float64); ok {
		changedFiles = util.IntPtr(int(files))
	}

	// Build raw data
	raw2 := make(map[string]interface{})
	raw2["repo"] = fmt.Sprintf("%s/%s", owner, repo)
	if num, ok := d["number"].(float64); ok {
		raw2["number"] = int64(num)
	}
	if base, ok := d["base"].(map[string]interface{}); ok {
		if ref, ok := base["ref"].(string); ok {
			raw2["base"] = ref
		}
	}
	if head, ok := d["head"].(map[string]interface{}); ok {
		if ref, ok := head["ref"].(string); ok {
			raw2["head"] = ref
		}
	}

	return &models.PullRequestRecord{
		Source:       "github",
		ID:           fmt.Sprintf("%.0f", d["id"]),
		Title:        getStringField(d, "title"),
		URL:          getStringField(d, "html_url"),
		Author:       author,
		State:        getStringField(d, "state"),
		UpdatedAt:    *updatedAt,
		MergedAt:     mergedAt,
		Additions:    additions,
		Deletions:    deletions,
		ChangedFiles: changedFiles,
		Labels:       labels,
		Raw:          raw2,
	}
}

// fetchPRDetail fetches detailed PR information
func (gc *GitHubCollector) fetchPRDetail(owner, repo string, number int64) map[string]interface{} {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", githubAPIBase, owner, repo, number)

	req, _ := http.NewRequest("GET", url, nil)
	gc.addHeaders(req)

	resp, err := gc.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var detail map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return nil
	}

	return detail
}

// addHeaders adds authentication and API version headers
func (gc *GitHubCollector) addHeaders(req *http.Request) {
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", gc.config.GitHubToken))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")
}

// Helper function to get string field from map
func getStringField(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

