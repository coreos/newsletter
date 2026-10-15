package collectors

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/coreos/coreos-newsletter/internal/config"
	"github.com/coreos/coreos-newsletter/internal/models"
	"github.com/coreos/coreos-newsletter/internal/util"
)

const (
	meetbotTimeout = 30 * time.Second
	meetbotBase    = "https://meetbot.fedoraproject.org"
)

// MeetbotCollector fetches Fedora CoreOS meeting logs
type MeetbotCollector struct {
	config *config.Config
	logger *util.Logger
	client *http.Client
}

// NewMeetbotCollector creates a new Meetbot collector
func NewMeetbotCollector(cfg *config.Config, logger *util.Logger) *MeetbotCollector {
	return &MeetbotCollector{
		config: cfg,
		logger: logger,
		client: &http.Client{Timeout: meetbotTimeout},
	}
}

// FetchFedoraCoreOSMeetings fetches Fedora CoreOS meeting logs from meetbot
func (mc *MeetbotCollector) FetchFedoraCoreOSMeetings(windowStart, windowEnd time.Time) ([]models.MeetingRecord, []string) {
	var meetings []models.MeetingRecord
	var warnings []string

	// Fetch the main meetbot page which loads meetings dynamically
	// We'll extract meeting info from the HTML content
	html, err := mc.fetchMeetbotPage()
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Meetbot fetch failed: %v", err))
		return meetings, warnings
	}

	// Parse the Fedora CoreOS meeting info from the page
	// The page contains AI-generated summaries with meeting details
	meetings, parseWarnings := mc.extractMeetingsFromPage(html, windowStart, windowEnd)
	warnings = append(warnings, parseWarnings...)

	return meetings, warnings
}

// fetchMeetbotPage fetches the main meetbot page
func (mc *MeetbotCollector) fetchMeetbotPage() (string, error) {
	req, _ := http.NewRequest("GET", meetbotBase, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")
	
	resp, err := mc.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// extractMeetingsFromPage extracts meeting info from the meetbot HTML page
func (mc *MeetbotCollector) extractMeetingsFromPage(html string, windowStart, windowEnd time.Time) ([]models.MeetingRecord, []string) {
	var meetings []models.MeetingRecord
	var warnings []string

	// Look for fedora coreos meeting sections in the HTML
	// Pattern: fedora coreos meeting 11 people | 5 topics
	pattern := regexp.MustCompile(`fedora\s+coreos\s+meeting\s+(\d+)\s+people\s*\|\s*(\d+)\s+topics`)
	
	matchesFound := pattern.FindAllStringSubmatch(html, -1)
	if len(matchesFound) > 0 {
		// Found at least one meeting mention
		// Extract summary sections that follow
		summaryPattern := regexp.MustCompile(`Key Discussion Points.*?(?:Action Items|$)`)
		
		if summaryMatch := summaryPattern.FindString(html); summaryMatch != "" {
			// Parse the summary to extract key points
			meeting := mc.parseMeetingSummary(summaryMatch, html)
			
			// Only include if it's in our time window
			if meeting.Date.After(windowStart.Add(-24*time.Hour)) && meeting.Date.Before(windowEnd.Add(24*time.Hour)) {
				meetings = append(meetings, meeting)
			}
		}
	}

	// If no meetings found, check for the "No meetings" indication in our time window
	if len(meetings) == 0 && !strings.Contains(html, "no community meetings") {
		// Meetings might exist but we couldn't parse them
		warnings = append(warnings, "Could not extract meeting details from meetbot page")
	}

	return meetings, warnings
}

// parseMeetingSummary parses a meeting summary from the HTML
func (mc *MeetbotCollector) parseMeetingSummary(summary string, fullHTML string) models.MeetingRecord {
	meeting := models.MeetingRecord{
		Title: "Fedora CoreOS Meeting",
		Date:  time.Now(),
		URL:   meetbotBase,
	}

	// Extract topics from "roll call", "Action items from last meeting", "Review", "Migrate", "Open Floor" etc
	topicsPattern := regexp.MustCompile(`<li[^>]*>([^<]+)</li>`)
	for _, match := range topicsPattern.FindAllStringSubmatch(summary, -1) {
		topic := strings.TrimSpace(match[1])
		if topic != "" && len(topic) > 2 && !strings.HasPrefix(topic, "role:") {
			meeting.Topics = append(meeting.Topics, topic)
		}
	}

	// Extract key discussion points
	discussionPattern := regexp.MustCompile(`<li[^>]*>([^<]+?The team[^<]*)</li>`)
	for _, match := range discussionPattern.FindAllStringSubmatch(summary, -1) {
		point := strings.TrimSpace(match[1])
		if point != "" {
			meeting.Topics = append(meeting.Topics, point)
		}
	}

	// Extract attendees from "People present" section
	peopleSectionPattern := regexp.MustCompile(`People present.*?(?:Topics covered|$)`)
	if peopleMatch := peopleSectionPattern.FindString(summary); peopleMatch != "" {
		attendeePattern := regexp.MustCompile(`@([^:]+):`)
		for _, attendeeMatch := range attendeePattern.FindAllStringSubmatch(peopleMatch, -1) {
			attendee := strings.TrimSpace(attendeeMatch[1])
			if attendee != "" {
				meeting.Attendees = append(meeting.Attendees, attendee)
			}
		}
	}

	// Extract action items
	actionSectionPattern := regexp.MustCompile(`Action Items.*?(?:Lookup discussions|$)`)
	if actionMatch := actionSectionPattern.FindString(summary); actionMatch != "" {
		actionPattern := regexp.MustCompile(`<li[^>]*>([^<]+?:.*?)</li>`)
		for _, actionMMatch := range actionPattern.FindAllStringSubmatch(actionMatch, -1) {
			action := strings.TrimSpace(actionMMatch[1])
			if action != "" && !strings.HasPrefix(action, "role:") {
				meeting.ActionItems = append(meeting.ActionItems, action)
			}
		}
	}

	meeting.Raw = map[string]interface{}{
		"url":     meetbotBase,
		"fetched": time.Now(),
	}

	return meeting
}
