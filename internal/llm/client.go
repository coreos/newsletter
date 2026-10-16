package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"github.com/coreos/coreos-newsletter/internal/config"
	"github.com/coreos/coreos-newsletter/internal/util"
)

// Client interface for LLM providers
type Client interface {
	GenerateNewsletter(ctx context.Context, prompt string) (string, error)
	Close() error
}

// NewClient creates an LLM client based on configuration
func NewClient(cfg *config.Config, logger *util.Logger) (Client, error) {
	provider := strings.ToLower(cfg.LLMProvider)

	if provider == "claude" {
		logger.Debug("Using Anthropic Claude provider")
		// Check if using Vertex AI (Google Cloud)
		if cfg.VertexProjectID != "" {
			logger.Debug("Using Claude via Vertex AI")
			return NewVertexClaudeClient(cfg, logger)
		}
		// Use direct Anthropic API
		return NewClaudeClient(cfg, logger)
	}

	if provider == "gemini" {
		logger.Debug("Using Google Gemini provider")
		// Check if using Vertex AI (Google Cloud)
		if cfg.VertexProjectID != "" {
			logger.Debug("Using Gemini via Vertex AI")
			return NewVertexGeminiClient(cfg, logger)
		}
		// Use direct Gemini API
		return NewGeminiClient(cfg, logger)
	}

	// Default to mock if no provider specified
	logger.Warn("No LLM provider specified, using mock client")
	return NewMockClient(cfg, logger), nil
}

// ============================================================================
// CLAUDE CLIENT
// ============================================================================

// ClaudeClient uses Anthropic's Claude API via HTTP
// Reads ANTHROPIC_API_KEY from environment variable automatically
type ClaudeClient struct {
	config *config.Config
	logger *util.Logger
	client *http.Client
}

// NewClaudeClient creates a Claude client
func NewClaudeClient(cfg *config.Config, logger *util.Logger) (*ClaudeClient, error) {
	return &ClaudeClient{
		config: cfg,
		logger: logger,
		client: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// ClaudeRequest represents a request to Claude API
type ClaudeRequest struct {
	Model     string          `json:"model"`
	Messages  []ClaudeMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens"`
}

// ClaudeMessage represents a message for Claude
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeResponse represents the response from Claude API
type ClaudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// GenerateNewsletter generates a newsletter using Claude
func (cc *ClaudeClient) GenerateNewsletter(ctx context.Context, prompt string) (string, error) {
	cc.logger.Debug("Generating newsletter with Anthropic Claude")

	// Read API key from environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		// Try from config as fallback
		apiKey = cc.config.AnthropicAPIKey
	}

	if apiKey == "" {
		cc.logger.Error("ANTHROPIC_API_KEY not found in environment or config")
		return "", fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	// Log model being used
	cc.logger.Debug(fmt.Sprintf("Using Claude model: %s", cc.config.ClaudeModel))

	// Prepare request
	req := ClaudeRequest{
		Model: cc.config.ClaudeModel,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: 2000,
	}

	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API call
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("x-api-key", apiKey)

	resp, err := cc.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call Claude API: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Claude API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var respData ClaudeResponse
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if respData.Error != nil {
		return "", fmt.Errorf("Claude API error: %s - %s", respData.Error.Type, respData.Error.Message)
	}

	// Extract text from response
	if len(respData.Content) > 0 && respData.Content[0].Type == "text" {
		return respData.Content[0].Text, nil
	}

	return "", fmt.Errorf("no text content in Claude response")
}

// Close closes the Claude client
func (cc *ClaudeClient) Close() error {
	return nil
}

// ============================================================================
// VERTEX AI CLAUDE CLIENT
// ============================================================================

// VertexClaudeClient uses Claude API via Google Cloud Vertex AI
type VertexClaudeClient struct {
	config *config.Config
	logger *util.Logger
	client *genai.Client
}

// NewVertexClaudeClient creates a Claude client via Vertex AI
func NewVertexClaudeClient(cfg *config.Config, logger *util.Logger) (*VertexClaudeClient, error) {
	if cfg.VertexProjectID == "" {
		return nil, fmt.Errorf("ANTHROPIC_VERTEX_PROJECT_ID not set")
	}

	// Create Vertex AI client (uses Application Default Credentials from: gcloud auth application-default login)
	ctx := context.Background()
	client, err := genai.NewClient(ctx, cfg.VertexProjectID, cfg.VertexRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vertex AI client: %w", err)
	}

	return &VertexClaudeClient{
		config: cfg,
		logger: logger,
		client: client,
	}, nil
}

// GenerateNewsletter generates a newsletter using Claude via Vertex AI
func (vc *VertexClaudeClient) GenerateNewsletter(ctx context.Context, prompt string) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("context is required for GenerateNewsletter")
	}
	vc.logger.Debug("Generating newsletter with Claude via Vertex AI")

	// Get model name
	modelName := vc.config.ClaudeModel
	if modelName == "" {
		modelName = config.DefaultClaudeModel
	}

	// Log model and project being used
	vc.logger.Debug(fmt.Sprintf("Using Claude model: %s in project: %s region: %s", 
		modelName, vc.config.VertexProjectID, vc.config.VertexRegion))

	// Create model (Vertex AI models are referenced with "publishers/google/models/..." prefix)
	model := vc.client.GenerativeModel(modelName)
	model.SetTemperature(0.7)
	model.SetTopP(0.9)
	model.SetMaxOutputTokens(2000)

	// Generate content
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	// Extract text from response
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in response")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	// Extract text
	if len(candidate.Content.Parts) > 0 {
		if text, ok := candidate.Content.Parts[0].(genai.Text); ok {
			return string(text), nil
		}
	}
	return "", fmt.Errorf("no text content in response")
}

// Close closes the Vertex AI client
func (vc *VertexClaudeClient) Close() error {
	if vc.client != nil {
		return vc.client.Close()
	}
	return nil
}

// ============================================================================
// VERTEX AI GEMINI CLIENT
// ============================================================================

// VertexGeminiClient uses Gemini API via Google Cloud Vertex AI
type VertexGeminiClient struct {
	config *config.Config
	logger *util.Logger
	client *genai.Client
}

// NewVertexGeminiClient creates a Gemini client via Vertex AI
func NewVertexGeminiClient(cfg *config.Config, logger *util.Logger) (*VertexGeminiClient, error) {
	if cfg.VertexProjectID == "" {
		return nil, fmt.Errorf("ANTHROPIC_VERTEX_PROJECT_ID not set")
	}

	// Create Vertex AI client (uses Application Default Credentials from: gcloud auth application-default login)
	ctx := context.Background()
	client, err := genai.NewClient(ctx, cfg.VertexProjectID, cfg.VertexRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vertex AI client: %w", err)
	}

	return &VertexGeminiClient{
		config: cfg,
		logger: logger,
		client: client,
	}, nil
}

// GenerateNewsletter generates a newsletter using Gemini via Vertex AI
func (vg *VertexGeminiClient) GenerateNewsletter(ctx context.Context, prompt string) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("context is required for GenerateNewsletter")
	}
	vg.logger.Debug("Generating newsletter with Gemini via Vertex AI")

	// Get model name
	modelName := vg.config.GeminiModel
	if modelName == "" {
		modelName = config.DefaultGeminiModel
	}

	// Log model and project being used
	vg.logger.Debug(fmt.Sprintf("Using Gemini model: %s in project: %s region: %s", 
		modelName, vg.config.VertexProjectID, vg.config.VertexRegion))

	// Create model (Vertex AI models are referenced by name)
	model := vg.client.GenerativeModel(modelName)
	model.SetTemperature(0.7)
	model.SetTopP(0.9)
	model.SetMaxOutputTokens(2000)

	// Generate content
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	// Extract text from response
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in response")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	// Extract text
	if len(candidate.Content.Parts) > 0 {
		if text, ok := candidate.Content.Parts[0].(genai.Text); ok {
			return string(text), nil
		}
	}
	return "", fmt.Errorf("no text content in response")
}

// Close closes the Vertex AI Gemini client
func (vg *VertexGeminiClient) Close() error {
	if vg.client != nil {
		return vg.client.Close()
	}
	return nil
}

// ============================================================================
// GEMINI CLIENT
// ============================================================================

// GeminiClient uses Google's Gemini API via HTTP
type GeminiClient struct {
	config *config.Config
	logger *util.Logger
	client *http.Client
}

// NewGeminiClient creates a Gemini client
func NewGeminiClient(cfg *config.Config, logger *util.Logger) (*GeminiClient, error) {
	// Read API key from environment
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		// Try from config as fallback
		apiKey = cfg.GoogleAPIKey
	}

	if apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY not set in environment or config")
	}

	return &GeminiClient{
		config: cfg,
		logger: logger,
		client: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// GeminiRequest represents a request to Gemini API
type GeminiRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
}

// GeminiResponse represents the response from Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// GenerateNewsletter generates a newsletter using Gemini
func (gc *GeminiClient) GenerateNewsletter(ctx context.Context, prompt string) (string, error) {
	gc.logger.Debug("Generating newsletter with Google Gemini")

	// Read API key from environment
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		// Try from config as fallback
		apiKey = gc.config.GoogleAPIKey
	}

	if apiKey == "" {
		return "", fmt.Errorf("GOOGLE_API_KEY not set")
	}

	// Get model name (default to gemini-2.5-flash if not specified)
	modelName := gc.config.GeminiModel
	if modelName == "" {
		modelName = config.DefaultGeminiModel
	}

	// Log model being used
	gc.logger.Debug(fmt.Sprintf("Using Gemini model: %s", modelName))

	// Prepare request
	req := GeminiRequest{
		Contents: []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			{
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: prompt},
				},
			},
		},
	}

	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL with API key
	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		modelName, url.QueryEscape(apiKey))

	// Make API call
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := gc.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call Gemini API: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Gemini API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var respData GeminiResponse
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if respData.Error != nil {
		return "", fmt.Errorf("Gemini API error: %d - %s", respData.Error.Code, respData.Error.Message)
	}

	// Extract text from response
	if len(respData.Candidates) > 0 && len(respData.Candidates[0].Content.Parts) > 0 {
		return respData.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("no text content in Gemini response")
}

// Close closes the Gemini client
func (gc *GeminiClient) Close() error {
	return nil
}

// ============================================================================
// MOCK CLIENT (for testing without API keys)
// ============================================================================

// MockClient for testing
type MockClient struct {
	config *config.Config
	logger *util.Logger
}

// NewMockClient creates a mock LLM client for testing
func NewMockClient(cfg *config.Config, logger *util.Logger) *MockClient {
	return &MockClient{
		config: cfg,
		logger: logger,
	}
}

// GenerateNewsletter generates a mock newsletter
func (mc *MockClient) GenerateNewsletter(ctx context.Context, prompt string) (string, error) {
	mc.logger.Debug("Generating mock newsletter")

	newsletter := `# Fedora CoreOS Newsletter

## This Week's Highlights

The Fedora CoreOS project moves forward with valuable contributions and active participation.

## Notable Contributions

- Pull requests improving core functionality have been merged
- Meeting discussions focused on feature planning and roadmap alignment
- Documentation improvements make it easier for new contributors to get started

## This Week's Activities

This week featured active participation in Fedora CoreOS meetings with discussions on:
- Project priorities and direction
- Technical decisions and implementations
- Feedback and requirements

## Thanks to All Contributors

We appreciate the ongoing support and contributions from all who help drive Fedora CoreOS forward!

---
*Generated by Fedora CoreOS Newsletter Generator*
`
	return newsletter, nil
}

// Close closes the mock client
func (mc *MockClient) Close() error {
	return nil
}
