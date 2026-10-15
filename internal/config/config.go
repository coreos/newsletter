package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration from environment variables
type Config struct {
	// GitHub
	GitHubToken string
	GitHubOwner string
	GitHubRepo  string

	// LLM Provider
	LLMProvider string

	// Google Gemini
	GoogleAPIKey        string
	GoogleCloudProject  string
	GoogleCloudLocation string
	GeminiModel         string

	// Claude (Anthropic)
	AnthropicAPIKey string
	ClaudeModel     string

	// Claude via Vertex AI (Google Cloud)
	VertexProjectID string
	VertexRegion    string
}

// Load reads configuration from .env file and environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (non-fatal if missing)
	_ = godotenv.Load()

	cfg := &Config{
		// GitHub
		GitHubToken: os.Getenv("GITHUB_TOKEN"),
		GitHubOwner: os.Getenv("GITHUB_OWNER"),
		GitHubRepo:  os.Getenv("GITHUB_REPO"),

		// LLM
		LLMProvider: strings.ToLower(getEnvOrDefault("LLM_PROVIDER", "gemini")),

		// Gemini
		GoogleAPIKey:        os.Getenv("GOOGLE_API_KEY"),
		GoogleCloudProject:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
		GoogleCloudLocation: getEnvOrDefault("GOOGLE_CLOUD_LOCATION", "us-central1"),
		GeminiModel:         getEnvOrDefault("GEMINI_MODEL", "gemini-2.5-flash"),

		// Claude
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		ClaudeModel:     getEnvOrDefault("CLAUDE_MODEL", "claude-sonnet-4-20250514"),

		// Claude via Vertex AI
		VertexProjectID: os.Getenv("ANTHROPIC_VERTEX_PROJECT_ID"),
		VertexRegion:    getEnvOrDefault("CLOUD_ML_REGION", "us-central1"),
	}

	return cfg, nil
}

// GitHubRepoList parses comma-separated GitHub repos
func (c *Config) GitHubRepoList() []string {
	if c.GitHubRepo == "" {
		return []string{}
	}
	var repos []string
	for _, repo := range strings.Split(c.GitHubRepo, ",") {
		if trimmed := strings.TrimSpace(repo); trimmed != "" {
			repos = append(repos, trimmed)
		}
	}
	return repos
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
