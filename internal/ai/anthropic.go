package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AnthropicProvider implements AIProvider for Anthropic Claude.
type AnthropicProvider struct {
	APIKey string
	Model  string
}

type anthropicRequest struct {
	Model     string              `json:"model"`
	System    string              `json:"system"`
	MaxTokens int                 `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage *struct {
		InputTokens  int `json:"input_tokens,omitempty"`
		OutputTokens int `json:"output_tokens,omitempty"`
	} `json:"usage,omitempty"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

const maxResponseBytes = 2 << 20 // 2MB

// Chat sends a chat request to Anthropic's Messages API.
func (p *AnthropicProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages provided")
	}

	systemContent := ""
	userContent := ""
	for _, m := range messages {
		if m.Role == "system" {
			systemContent = m.Content
		} else if m.Role == "user" {
			userContent = m.Content
		}
	}

	maxTokens := 1024
	if systemContent != "" {
		maxTokens = 2048
	}

	reqBody := anthropicRequest{
		Model:     p.Model,
		System:    systemContent,
		MaxTokens: maxTokens,
		Messages: []anthropicMessage{
			{Role: "user", Content: userContent},
		},
	}

	buf, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	endpoint := "https://api.anthropic.com/v1/messages"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(buf))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if len(body) > maxResponseBytes {
		return "", fmt.Errorf("response exceeds %d bytes", maxResponseBytes)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		time.Sleep(5 * time.Second)
		return p.Chat(ctx, messages)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("anthropic %d: %s", resp.StatusCode, truncate(body))
	}

	var parsed anthropicResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("anthropic %s: %s", parsed.Error.Type, parsed.Error.Message)
	}

	for _, c := range parsed.Content {
		if c.Type == "text" && c.Text != "" {
			return c.Text, nil
		}
	}
	for _, c := range parsed.Content {
		if c.Text != "" {
			return c.Text, nil
		}
	}
	return "", fmt.Errorf("no text content in response")
}

func truncate(b []byte) string {
	const max = 512
	s := string(b)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}