package executor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	anthropicAPIURL = "https://api.anthropic.com/v1/messages"
	defaultModel    = "claude-sonnet-4-5-20250929"
	apiVersion      = "2023-06-01"
)

// ClaudeClient is a client for the Claude API
type ClaudeClient struct {
	apiKey     string
	httpClient *http.Client
	model      string
}

// NewClaudeClient creates a new Claude API client
func NewClaudeClient() (*ClaudeClient, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("CLAUDE_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("API key not found. Set ANTHROPIC_API_KEY or CLAUDE_API_KEY environment variable")
	}

	return &ClaudeClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Minute, // Long timeout for implementation tasks
		},
		model: defaultModel,
	}, nil
}

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// APIRequest represents a request to the Claude API
type APIRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

// APIResponse represents a response from the Claude API
type APIResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// ErrorResponse represents an error from the Claude API
type ErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Implement executes an implementation task using Claude API
func (c *ClaudeClient) Implement(prompt string) (string, error) {
	reqBody := APIRequest{
		Model:     c.model,
		MaxTokens: 8000, // Sufficient for implementation responses
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", anthropicAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", apiVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return "", fmt.Errorf("API error: %s - %s", errResp.Error.Type, errResp.Error.Message)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	return apiResp.Content[0].Text, nil
}
