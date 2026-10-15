package collectors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/coreos-newsletter/internal/config"
	"github.com/coreos/coreos-newsletter/internal/models"
	"github.com/coreos/coreos-newsletter/internal/util"
)

const (
	meetbotTimeout = 30 * time.Second
	meetbotBase    = "https://meetbot.fedoraproject.org"
	meetbotCalAPI  = meetbotBase + "/cal/events"
	meetbotFragAPI = meetbotBase + "/fragedpt/"
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

	// Try to fetch meetings via the calendar API endpoint
	// This is the official Mote API for fetching events by period
	calMeetings, err := mc.fetchViaCalendarAPI(windowStart, windowEnd)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Calendar API failed: %v", err))
		// Fall back to trying the fragedpt API
		fragMeetings, fragWarnings := mc.fetchViaFragedptAPI(windowStart, windowEnd)
		warnings = append(warnings, fragWarnings...)
		return fragMeetings, warnings
	}

	meetings = append(meetings, calMeetings...)

	if len(meetings) == 0 {
		warnings = append(warnings, "No meetings found in time window")
	}

	return meetings, warnings
}

// fetchViaCalendarAPI fetches meetings using the /cal/events endpoint
// This is the official calendar API that returns JSON
func (mc *MeetbotCollector) fetchViaCalendarAPI(windowStart, windowEnd time.Time) ([]models.MeetingRecord, error) {
	var meetings []models.MeetingRecord

	// Format dates for the API (ISO 8601 format)
	startStr := windowStart.Format("2006-01-02T15:04:05")
	endStr := windowEnd.Format("2006-01-02T15:04:05")

	// Build URL
	url := fmt.Sprintf("%s?start=%s&end=%s", meetbotCalAPI, startStr, endStr)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Fedora CoreOS Newsletter Generator")

	resp, err := mc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Parse JSON response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var events []map[string]interface{}
	err = json.Unmarshal(body, &events)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert events to MeetingRecords, filtering for fedora-coreos
	for _, event := range events {
		meeting := mc.parseCalendarEvent(event)
		if strings.Contains(strings.ToLower(meeting.Title), "fedora") && 
		   strings.Contains(strings.ToLower(meeting.Title), "coreos") {
			meetings = append(meetings, meeting)
		}
	}

	return meetings, nil
}

// parseCalendarEvent converts a calendar event JSON to MeetingRecord
func (mc *MeetbotCollector) parseCalendarEvent(event map[string]interface{}) models.MeetingRecord {
	meeting := models.MeetingRecord{
		Title: "Fedora CoreOS Meeting",
		Date:  time.Now(),
		URL:   meetbotBase,
	}

	// Parse title
	if title, ok := event["title"].(string); ok {
		meeting.Title = title
	}

	// Parse start date
	if start, ok := event["start"].(string); ok {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			meeting.Date = t
		}
	}

	meeting.Raw = map[string]interface{}{
		"url":     meetbotBase,
		"fetched": time.Now(),
		"event":   event,
	}

	return meeting
}

// fetchViaFragedptAPI fetches meetings using the /fragedpt/ endpoint
// This is a fallback API for fetching meetings by channel and date
func (mc *MeetbotCollector) fetchViaFragedptAPI(windowStart, windowEnd time.Time) ([]models.MeetingRecord, []string) {
	var meetings []models.MeetingRecord
	var warnings []string

	// Try to list meetings for fedora-coreos channel
	current := windowStart
	for current.Before(windowEnd) || current.Equal(windowEnd) {
		dateStr := current.Format("2006-01-02")

		// Fetch meetings for this date
		url := fmt.Sprintf("%s?rqstdata=listmeet&channame=fedora-coreos&datename=%s", meetbotFragAPI, dateStr)

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("User-Agent", "Fedora CoreOS Newsletter Generator")

		resp, err := mc.client.Do(req)
		if err != nil {
			current = current.AddDate(0, 0, 1)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			current = current.AddDate(0, 0, 1)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			current = current.AddDate(0, 0, 1)
			continue
		}

		// Parse meetings list
		var meetingsList []map[string]interface{}
		err = json.Unmarshal(body, &meetingsList)
		if err != nil {
			current = current.AddDate(0, 0, 1)
			continue
		}

		// Convert each meeting
		for _, m := range meetingsList {
			meeting := mc.parseFragedptMeeting(m, dateStr)
			meetings = append(meetings, meeting)
		}

		current = current.AddDate(0, 0, 1)
	}

	return meetings, warnings
}

// parseFragedptMeeting converts a fragedpt meeting JSON to MeetingRecord
func (mc *MeetbotCollector) parseFragedptMeeting(meeting map[string]interface{}, dateStr string) models.MeetingRecord {
	record := models.MeetingRecord{
		Title: "Fedora CoreOS Meeting",
		Date:  time.Now(),
		URL:   meetbotBase,
	}

	// Parse date
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		record.Date = t
	}

	// Parse topic
	if topic, ok := meeting["topic"].(string); ok {
		record.Title = topic
	}

	// Parse URL
	if urls, ok := meeting["url"].(map[string]interface{}); ok {
		if summary, ok := urls["summary"].(string); ok {
			record.URL = meetbotBase + summary
		}
	}

	record.Raw = map[string]interface{}{
		"url":     meetbotBase,
		"fetched": time.Now(),
		"meeting": meeting,
	}

	return record
}
