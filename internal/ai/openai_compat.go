package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAICompatProvider implements AIProvider for OpenAI-compatible endpoints.
// Used for: openai, openrouter, minimax, deepseek, groq.
type OpenAICompatProvider struct {
	APIKey      string
	Endpoint    string // "https://api.openai.com/v1/chat/completions" or custom
	Model       string
	MaxTokens   int
	Temperature float64
}

type openaiRequest struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Chat sends a chat request to an OpenAI-compatible endpoint.
func (p *OpenAICompatProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	if p.Endpoint == "" {
		return "", fmt.Errorf("endpoint not configured")
	}

	systemContent := ""
	var userContent string

	for _, m := range messages {
		if m.Role == "system" {
			systemContent = m.Content
		} else if m.Role == "user" {
			userContent = m.Content
		}
	}

	maxTokens := p.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1024
	}

	temp := p.Temperature
	if temp == 0 {
		temp = 0.7
	}

	reqMessages := []openaiMessage{}
	if systemContent != "" {
		reqMessages = append(reqMessages, openaiMessage{Role: "system", Content: systemContent})
	}
	reqMessages = append(reqMessages, openaiMessage{Role: "user", Content: userContent})

	reqBody := openaiRequest{
		Model:       p.Model,
		Messages:    reqMessages,
		MaxTokens:   maxTokens,
		Temperature: temp,
	}

	buf, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.Endpoint, bytes.NewReader(buf))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if len(body) > maxResponseBytes {
		return "", fmt.Errorf("response exceeds %d bytes", maxResponseBytes)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("%s %d: %s", p.Endpoint, resp.StatusCode, truncate(body))
	}

	var parsed openaiResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("%s: %s", parsed.Error.Type, parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}