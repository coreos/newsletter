package models

import (
	"time"
)

// PullRequestRecord represents a pull request from GitHub or other sources
type PullRequestRecord struct {
	Source       string            `json:"source"` // "github" or "meetbot"
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	URL          string            `json:"url"`
	Author       *string           `json:"author,omitempty"`
	State        string            `json:"state"`
	UpdatedAt    time.Time         `json:"updated_at"`
	MergedAt     *time.Time        `json:"merged_at,omitempty"`
	Additions    *int              `json:"additions,omitempty"`
	Deletions    *int              `json:"deletions,omitempty"`
	ChangedFiles *int              `json:"changed_files,omitempty"`
	Labels       []string          `json:"labels"`
	Raw          map[string]interface{} `json:"raw"`
}

// MeetingRecord represents a meeting from meetbot logs
type MeetingRecord struct {
	Title       string                 `json:"title"`
	Date        time.Time              `json:"date"`
	URL         string                 `json:"url"`
	Attendees   []string               `json:"attendees"`
	Topics      []string               `json:"topics"`
	Decisions   []string               `json:"decisions"`
	ActionItems []string               `json:"action_items"`
	Raw         map[string]interface{} `json:"raw"`
}

// WeeklyBundle contains all collected data for a week
type WeeklyBundle struct {
	WindowStart    time.Time              `json:"window_start"`
	WindowEnd      time.Time              `json:"window_end"`
	PullRequests   []PullRequestRecord    `json:"pull_requests"`
	Meetings       []MeetingRecord        `json:"meetings"`
	Meta           map[string]interface{} `json:"meta"`
}

// NewWeeklyBundle creates a new bundle with initialized slices and meta map
func NewWeeklyBundle(start, end time.Time) *WeeklyBundle {
	return &WeeklyBundle{
		WindowStart:  start,
		WindowEnd:    end,
		PullRequests: make([]PullRequestRecord, 0),
		Meetings:     make([]MeetingRecord, 0),
		Meta:         make(map[string]interface{}),
	}
}
