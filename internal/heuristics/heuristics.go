package heuristics

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/coreos/coreos-newsletter/internal/models"
)

// PRScore contains ranking and categorization of a PR
type PRScore struct {
	PR              models.PullRequestRecord
	Score           float64
	Category        string // "feature", "bugfix", "refactor", "docs"
	Significance    string // "critical", "high", "medium", "low"
	IsHumanAuthor   bool
}

// RankPRs ranks pull requests by significance and category
func RankPRs(prs []models.PullRequestRecord) []PRScore {
	var scores []PRScore

	for _, pr := range prs {
		score := scorePR(pr)
		scores = append(scores, score)
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].Score != scores[j].Score {
			return scores[i].Score > scores[j].Score
		}
		// Tie-breaker: by date (newer first)
		return scores[i].PR.UpdatedAt.After(scores[j].PR.UpdatedAt)
	})

	return scores
}

// scorePR calculates a score for a PR based on various heuristics
func scorePR(pr models.PullRequestRecord) PRScore {
	score := PRScore{
		PR:            pr,
		IsHumanAuthor: isHumanAuthor(pr),
	}

	// Start with base score
	baseScore := 0.0

	// Score by state
	if pr.State == "merged" {
		baseScore += 100.0
	} else if pr.State == "closed" {
		baseScore += 10.0
	} else if pr.State == "open" {
		baseScore += 50.0
	}

	// Score by complexity (lines changed)
	if pr.Additions != nil && pr.Deletions != nil {
		totalLines := *pr.Additions + *pr.Deletions
		if totalLines > 1000 {
			baseScore += 50.0
			score.Significance = "critical"
		} else if totalLines > 500 {
			baseScore += 30.0
			score.Significance = "high"
		} else if totalLines > 100 {
			baseScore += 15.0
			score.Significance = "medium"
		} else {
			score.Significance = "low"
		}
	}

	// Score by recency
	age := time.Since(pr.UpdatedAt)
	if age < 24*time.Hour {
		baseScore += 30.0
	} else if age < 7*24*time.Hour {
		baseScore += 15.0
	}

	// Categorize by labels and title
	category, categoryScore := categorizePR(pr)
	score.Category = category
	baseScore += categoryScore

	// Penalize if automation source
	if !score.IsHumanAuthor {
		baseScore *= 0.1
	}

	score.Score = baseScore
	return score
}

// isHumanAuthor checks if the PR author is a human (not a bot)
func isHumanAuthor(pr models.PullRequestRecord) bool {
	if pr.Author == nil {
		return false
	}

	authorLower := strings.ToLower(*pr.Author)
	// Filter out known bot accounts
	botPatterns := []string{
		"bot",
		"dependabot",
		"coreosbot",
		"ci-",
		"automation",
		"github-actions",
	}

	for _, pattern := range botPatterns {
		if strings.Contains(authorLower, pattern) {
			return false
		}
	}

	return true
}

// categorizePR categorizes a PR based on labels and title
func categorizePR(pr models.PullRequestRecord) (string, float64) {
	labelStr := strings.ToLower(strings.Join(pr.Labels, " "))
	titleStr := strings.ToLower(pr.Title)
	combined := labelStr + " " + titleStr

	// Check for category indicators
	if strings.Contains(combined, "feature") || strings.Contains(combined, "enhancement") {
		return "feature", 40.0
	}
	if strings.Contains(combined, "fix") || strings.Contains(combined, "bugfix") || strings.Contains(combined, "bug") {
		return "bugfix", 50.0
	}
	if strings.Contains(combined, "refactor") {
		return "refactor", 25.0
	}
	if strings.Contains(combined, "doc") || strings.Contains(combined, "documentation") {
		return "docs", 10.0
	}

	return "other", 0.0
}

// FilterPRsForShoutouts filters and ranks PRs suitable for contributor recognition
func FilterPRsForShoutouts(prs []models.PullRequestRecord, limit int) []models.PullRequestRecord {
	// Filter: only merged PRs from humans
	var filtered []models.PullRequestRecord
	for _, pr := range prs {
		if pr.State != "merged" {
			continue
		}
		if !isHumanAuthor(pr) {
			continue
		}
		filtered = append(filtered, pr)
	}

	// Rank
	scores := RankPRs(filtered)

	// Extract top N
	var result []models.PullRequestRecord
	for i, score := range scores {
		if i >= limit {
			break
		}
		result = append(result, score.PR)
	}

	return result
}

// ContributorStats tracks statistics about contributors
type ContributorStats struct {
	Author           string
	PRCount          int
	MergedCount      int
	TotalAdditions   int
	TotalDeletions   int
	Categories       map[string]int
	MostRecentUpdate time.Time
}

// GatherContributorStats creates statistics about contributors
func GatherContributorStats(prs []models.PullRequestRecord) map[string]*ContributorStats {
	stats := make(map[string]*ContributorStats)

	for _, pr := range prs {
		if pr.Author == nil {
			continue
		}

		author := *pr.Author

		if stats[author] == nil {
			stats[author] = &ContributorStats{
				Author:     author,
				Categories: make(map[string]int),
			}
		}

		s := stats[author]
		s.PRCount++

		if pr.State == "merged" {
			s.MergedCount++
		}

		if pr.Additions != nil {
			s.TotalAdditions += *pr.Additions
		}
		if pr.Deletions != nil {
			s.TotalDeletions += *pr.Deletions
		}

		_, category := categorizePR(pr)
		_ = category // category is scoring, not the name
		cat := categorizeSimple(pr)
		s.Categories[cat]++

		if pr.UpdatedAt.After(s.MostRecentUpdate) {
			s.MostRecentUpdate = pr.UpdatedAt
		}
	}

	return stats
}

// categorizeSimple returns just the category name
func categorizeSimple(pr models.PullRequestRecord) string {
	cat, _ := categorizePR(pr)
	return cat
}

// HumanContributors returns list of unique human contributors from PRs
func HumanContributors(prs []models.PullRequestRecord) []string {
	seen := make(map[string]bool)
	var contributors []string

	for _, pr := range prs {
		if pr.Author == nil {
			continue
		}

		if !isHumanAuthor(pr) {
			continue
		}

		author := *pr.Author
		if !seen[author] {
			seen[author] = true
			contributors = append(contributors, author)
		}
	}

	// Sort for consistency
	sort.Strings(contributors)
	return contributors
}

// BuildLLMPayload builds a payload for LLM processing
func BuildLLMPayload(bundle *models.WeeklyBundle) map[string]interface{} {
	payload := make(map[string]interface{})

	// Window information
	payload["window_start"] = bundle.WindowStart.Format("2006-01-02")
	payload["window_end"] = bundle.WindowEnd.Format("2006-01-02")

	// Rank PRs
	rankedPRs := RankPRs(bundle.PullRequests)

	// Take top PRs for summary
	topCount := 20
	var topPRs []map[string]interface{}
	for i, score := range rankedPRs {
		if i >= topCount {
			break
		}

		prMap := map[string]interface{}{
			"title":        score.PR.Title,
			"author":       score.PR.Author,
			"source":       score.PR.Source,
			"state":        score.PR.State,
			"url":          score.PR.URL,
			"category":     score.Category,
			"significance": score.Significance,
			"score":        score.Score,
		}
		topPRs = append(topPRs, prMap)
	}
	payload["top_prs"] = topPRs
	payload["total_prs"] = len(bundle.PullRequests)

	// Shoutouts (top merged PRs from humans)
	shoutoutPRs := FilterPRsForShoutouts(bundle.PullRequests, 10)
	var shoutoutList []map[string]interface{}
	for _, pr := range shoutoutPRs {
		shoutMap := map[string]interface{}{
			"author": pr.Author,
			"title":  pr.Title,
			"url":    pr.URL,
		}
		shoutoutList = append(shoutoutList, shoutMap)
	}
	payload["shoutouts"] = shoutoutList
	payload["human_contributors"] = HumanContributors(bundle.PullRequests)


	// Meetings
	var meetingsList []map[string]interface{}
	for _, meeting := range bundle.Meetings {
		meetingMap := map[string]interface{}{
			"title":      meeting.Title,
			"date":       meeting.Date.Format("2006-01-02"),
			"url":        meeting.URL,
			"attendees":  meeting.Attendees,
			"topics":     meeting.Topics,
			"decisions":  meeting.Decisions,
			"actions":    meeting.ActionItems,
		}
		meetingsList = append(meetingsList, meetingMap)
	}
	payload["meetings"] = meetingsList
	payload["total_meetings"] = len(bundle.Meetings)

	// Metadata
	payload["generated_at"] = time.Now().Format(time.RFC3339)

	return payload
}

// SummarizePRs creates a text summary of the top PRs
func SummarizePRs(prs []models.PullRequestRecord) string {
	scores := RankPRs(prs)

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("## Pull Request Summary\n\n"))
	summary.WriteString(fmt.Sprintf("Total PRs/MRs: %d\n\n", len(prs)))

	// Count by source
	sourceCount := make(map[string]int)
	for _, pr := range prs {
		sourceCount[pr.Source]++
	}
	summary.WriteString("### By Source\n")
	for source, count := range sourceCount {
		summary.WriteString(fmt.Sprintf("- %s: %d\n", source, count))
	}

	summary.WriteString("\n### Top PRs\n\n")
	for i, score := range scores {
		if i >= 10 {
			break
		}
		summary.WriteString(fmt.Sprintf("%d. [%s] %s (%s, %s)\n",
			i+1, score.PR.Source, score.PR.Title, score.Category, score.Significance))
		summary.WriteString(fmt.Sprintf("   Author: %v | URL: %s\n\n", score.PR.Author, score.PR.URL))
	}

	return summary.String()
}
