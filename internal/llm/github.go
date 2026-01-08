package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	githubModelsEndpoint = "https://models.github.ai/inference/chat/completions"
	defaultModel         = "openai/gpt-4o"
	defaultTimeout       = 30 * time.Second
)

// Config holds the LLM summarization configuration
type Config struct {
	Model   string
	Lang    string
	Prompt  string // User's custom prompt injection
	Timeout time.Duration
}

// SummaryInput contains the data to summarize
type SummaryInput struct {
	DateLabel     string
	IssueGroups   []IssueEntry
	StandalonePRs []PREntry
	Reviews       []ReviewEntry
}

// IssueEntry represents an issue with its linked PRs
type IssueEntry struct {
	Title string
	URL   string
	Body  string
	PRs   []PREntry
}

// PREntry represents a pull request
type PREntry struct {
	Title string
	URL   string
	Body  string
}

// ReviewEntry represents reviews submitted on a PR
type ReviewEntry struct {
	PRTitle string
	PRURL   string
	States  []string
}

// chatRequest is the request body for the chat completions API
type chatRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse is the response from the chat completions API
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Summarize generates a summary using GitHub Models API
func Summarize(ctx context.Context, cfg Config, input SummaryInput) (string, error) {
	// Apply defaults
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.Lang == "" {
		cfg.Lang = "en"
	}

	// Get GitHub token
	token, err := getGHToken()
	if err != nil {
		return "", fmt.Errorf("failed to get GitHub token: %w", err)
	}

	// Build messages
	messages := buildMessages(cfg, input)

	// Create request
	reqBody := chatRequest{
		Model:    cfg.Model,
		Messages: messages,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with context timeout
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", githubModelsEndpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var chatResp chatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from model")
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}

// getGHToken retrieves the GitHub token from environment or gh CLI
func getGHToken() (string, error) {
	// Try GH_TOKEN first
	if token := os.Getenv("GH_TOKEN"); token != "" {
		return token, nil
	}

	// Try GITHUB_TOKEN
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}

	// Fall back to gh auth token
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get token from gh CLI: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
