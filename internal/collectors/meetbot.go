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
	meetbotBase    = "https://meetbot.fedoraproject.org/meeting-1_matrix_fedoraproject-org"
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

	// Generate meeting URLs for the time window
	urls := mc.generateMeetingURLs(windowStart, windowEnd)

	// Try to fetch each URL
	for _, url := range urls {
		meeting, err := mc.fetchAndParseMeeting(url)
		if err != nil {
			// 404s are normal (no meeting on that day/time)
			if strings.Contains(err.Error(), "404") {
				continue
			}
			warnings = append(warnings, fmt.Sprintf("Meetbot fetch failed: %v", err))
			continue
		}

		meetings = append(meetings, meeting)
	}

	return meetings, warnings
}

// generateMeetingURLs generates potential meeting URLs for a date range
func (mc *MeetbotCollector) generateMeetingURLs(windowStart, windowEnd time.Time) []string {
	var urls []string

	// Common meeting times (UTC)
	times := []string{"14:00", "15:30", "16:00"}

	// Iterate through each day in the window
	current := windowStart

	for current.Before(windowEnd) || current.Equal(windowEnd) {
		// Only check weekdays (Monday=1, Friday=5)
		if current.Weekday() >= 1 && current.Weekday() <= 5 {
			// Format date as YYYY-MM-DD
			dateStr := current.Format("2006-01-02")

			// Generate URL for each common meeting time
			for _, timeStr := range times {
				// Replace colons with dots for URL
				timeURL := strings.ReplaceAll(timeStr, ":", ".")
				url := fmt.Sprintf("%s/%s/fedora-coreos-meeting.%s-%s.log.html",
					meetbotBase, dateStr, dateStr, timeURL)
				urls = append(urls, url)
			}
		}

		current = current.AddDate(0, 0, 1)
	}

	return urls
}

// fetchAndParseMeeting fetches and parses a single meeting log
func (mc *MeetbotCollector) fetchAndParseMeeting(url string) (models.MeetingRecord, error) {
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := mc.client.Do(req)
	if err != nil {
		return models.MeetingRecord{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.MeetingRecord{}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Read HTML content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.MeetingRecord{}, err
	}

	// Parse the HTML
	meeting, err := mc.parseMeetingLog(string(body), url)
	if err != nil {
		return models.MeetingRecord{}, err
	}

	return meeting, nil
}

// parseMeetingLog parses the HTML meeting log
func (mc *MeetbotCollector) parseMeetingLog(html string, url string) (models.MeetingRecord, error) {
	meeting := models.MeetingRecord{
		Title: "Fedora CoreOS Meeting",
		URL:   url,
	}

	// Extract date from URL (format: YYYY-MM-DD)
	datePattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2})/fedora-coreos`)
	if matches := datePattern.FindStringSubmatch(url); len(matches) > 1 {
		if parsedTime, err := time.Parse("2006-01-02", matches[1]); err == nil {
			meeting.Date = parsedTime
		}
	}

	// Extract attendees
	// Look for lines like "Present: <name1> <name2> ..."
	attendeesPattern := regexp.MustCompile(`<li>.*?[Pp]resent:?\s+(.+?)<`)
	if matches := attendeesPattern.FindStringSubmatch(html); len(matches) > 1 {
		attendeeStr := matches[1]
		// Split by common delimiters
		names := strings.FieldsFunc(attendeeStr, func(r rune) bool {
			return r == ',' || r == ' '
		})

		for _, name := range names {
			name = strings.TrimSpace(name)
			if name != "" {
				meeting.Attendees = append(meeting.Attendees, name)
			}
		}
	}

	// Extract topics discussed
	// Look for sections like "TOPIC: ..."
	topicPattern := regexp.MustCompile(`<li><strong>TOPIC</strong>.*?:\s*(.+?)<`)
	for _, matches := range topicPattern.FindAllStringSubmatch(html, -1) {
		if len(matches) > 1 {
			topic := strings.TrimSpace(matches[1])
			if topic != "" {
				meeting.Topics = append(meeting.Topics, topic)
			}
		}
	}

	// Extract decisions
	// Look for sections like "AGREED: ..."
	decisionPattern := regexp.MustCompile(`<li><strong>AGREED</strong>.*?:\s*(.+?)<`)
	for _, matches := range decisionPattern.FindAllStringSubmatch(html, -1) {
		if len(matches) > 1 {
			decision := strings.TrimSpace(matches[1])
			if decision != "" {
				meeting.Decisions = append(meeting.Decisions, decision)
			}
		}
	}

	// Extract action items
	// Look for sections like "ACTION: ..."
	actionPattern := regexp.MustCompile(`<li><strong>ACTION</strong>.*?:\s*(.+?)<`)
	for _, matches := range actionPattern.FindAllStringSubmatch(html, -1) {
		if len(matches) > 1 {
			action := strings.TrimSpace(matches[1])
			if action != "" {
				meeting.ActionItems = append(meeting.ActionItems, action)
			}
		}
	}

	// Store raw response
	meeting.Raw = map[string]interface{}{
		"url":     url,
		"fetched": time.Now(),
	}

	return meeting, nil
}
