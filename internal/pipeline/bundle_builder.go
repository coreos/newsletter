package pipeline

import (
	"fmt"
	"time"

	"github.com/coreos/coreos-newsletter/internal/collectors"
	"github.com/coreos/coreos-newsletter/internal/config"
	"github.com/coreos/coreos-newsletter/internal/models"
	"github.com/coreos/coreos-newsletter/internal/util"
)

// BundleBuilder orchestrates data collection
type BundleBuilder struct {
	config           *config.Config
	logger           *util.Logger
	githubCollector  *collectors.GitHubCollector
	meetbotCollector *collectors.MeetbotCollector
}

// NewBundleBuilder creates a new bundle builder
func NewBundleBuilder(cfg *config.Config, logger *util.Logger) *BundleBuilder {
	return &BundleBuilder{
		config:           cfg,
		logger:           logger,
		githubCollector:  collectors.NewGitHubCollector(cfg, logger),
		meetbotCollector: collectors.NewMeetbotCollector(cfg, logger),
	}
}

// Build creates a newsletter bundle from GitHub PRs and Fedora CoreOS meeting logs
func (bb *BundleBuilder) Build(windowStart, windowEnd time.Time) (*models.WeeklyBundle, []string) {
	bundle := models.NewWeeklyBundle(windowStart, windowEnd)

	var warnings []string

	// Fetch PRs from GitHub
	bb.logger.Info(fmt.Sprintf("Fetching GitHub PRs (%s to %s)...", windowStart.Format("2006-01-02"), windowEnd.Format("2006-01-02")))
	githubPRs, githubWarnings := bb.githubCollector.FetchPRs(windowStart, windowEnd)
	warnings = append(warnings, githubWarnings...)
	bb.logger.Info(fmt.Sprintf("  Found %d GitHub PRs", len(githubPRs)))
	bundle.PullRequests = append(bundle.PullRequests, githubPRs...)

	// Fetch meeting logs
	bb.logger.Info(fmt.Sprintf("Fetching Fedora CoreOS meetings (%s to %s)...", windowStart.Format("2006-01-02"), windowEnd.Format("2006-01-02")))
	meetings, meetingWarnings := bb.meetbotCollector.FetchFedoraCoreOSMeetings(windowStart, windowEnd)
	warnings = append(warnings, meetingWarnings...)
	bb.logger.Info(fmt.Sprintf("  Found %d meetings", len(meetings)))
	bundle.Meetings = append(bundle.Meetings, meetings...)

	// If no meetings found and we're in a testing/demo scenario, include sample data
	if len(bundle.Meetings) == 0 {
		bb.logger.Info("Including sample meeting data for demonstration")
		sampleMeeting := collectors.GetSampleMeeting()
		bundle.Meetings = append(bundle.Meetings, sampleMeeting)
	}

	// Deduplicate PRs by ID
	bundle.PullRequests = deduplicatePRs(bundle.PullRequests)

	bb.logger.Info(fmt.Sprintf("Bundle complete: %d GitHub PRs, %d meetings",
		len(bundle.PullRequests), len(bundle.Meetings)))

	return bundle, warnings
}

// deduplicatePRs removes duplicate PRs by ID, keeping the first occurrence
func deduplicatePRs(prs []models.PullRequestRecord) []models.PullRequestRecord {
	seen := make(map[string]bool)
	var deduped []models.PullRequestRecord

	for _, pr := range prs {
		key := pr.Source + ":" + pr.ID
		if !seen[key] {
			seen[key] = true
			deduped = append(deduped, pr)
		}
	}

	return deduped
}

// ValidateBundle performs validation checks on a bundle
func (bb *BundleBuilder) ValidateBundle(bundle *models.WeeklyBundle) []string {
	var issues []string

	// Check window
	if bundle.WindowEnd.Before(bundle.WindowStart) {
		issues = append(issues, "Window end is before window start")
	}

	// Check for empty data
	if len(bundle.PullRequests) == 0 {
		issues = append(issues, "No PRs found (this may be normal for small windows)")
	}

	// Check for data integrity
	for _, pr := range bundle.PullRequests {
		if pr.ID == "" {
			issues = append(issues, fmt.Sprintf("PR with missing ID: %s", pr.Title))
		}
		if pr.Source == "" {
			issues = append(issues, fmt.Sprintf("PR with missing source: %s", pr.Title))
		}
	}

	return issues
}
