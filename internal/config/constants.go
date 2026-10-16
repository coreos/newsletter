package config

import "time"

// HTTP timeouts
const (
	GitHubAPITimeout  = 30 * time.Second
	MeetbotAPITimeout = 30 * time.Second
	LLMAPITimeout     = 60 * time.Second
)

// API pagination and limits
const (
	GitHubAPIPerPage = 100
	MaxPRsToShow     = 20
	MaxShoutouts     = 10
	MaxTopPRs        = 5
)

// LLM generation settings
const (
	LLMMaxTokens = 2000
)

// Default values
const (
	DefaultGeminiModel        = "gemini-2.5-flash"
	DefaultClaudeModel        = "claude-sonnet-4-20250514"
	DefaultCloudMLRegion      = "us-central1"
	DefaultGoogleCloudLocation = "us-central1"
	DefaultLLMProvider        = "gemini"
)
