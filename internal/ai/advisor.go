package ai

import (
	"context"
	"fmt"
	"strings"
)

// NewAdvisor creates a new AI advisor with the given configuration.
func NewAdvisor(cfg Config) *Advisor {
	a := &Advisor{cfg: cfg}
	a.cfg.Provider = strings.ToLower(strings.TrimSpace(a.cfg.Provider))
	switch a.cfg.Provider {
	case "anthropic":
		a.provider = &AnthropicProvider{
			APIKey: cfg.APIKey,
			Model:  cfg.Model,
		}
	case "openai", "openrouter", "minimax", "deepseek", "groq":
		endpoint := cfg.Endpoint
		if endpoint == "" {
			switch cfg.Provider {
			case "openai":
				endpoint = "https://api.openai.com/v1/chat/completions"
			case "openrouter":
				endpoint = "https://openrouter.ai/api/v1/chat/completions"
			case "minimax":
				endpoint = "https://api.minimax.chat/v1/text/chatcompletion_v2"
			case "deepseek":
				endpoint = "https://api.deepseek.com/v1/chat/completions"
			case "groq":
				endpoint = "https://api.groq.com/openai/v1/chat/completions"
			default:
				endpoint = "https://api.openai.com/v1/chat/completions"
			}
		}
		a.provider = &OpenAICompatProvider{
			APIKey:      cfg.APIKey,
			Endpoint:    endpoint,
			Model:       cfg.Model,
			MaxTokens:   cfg.MaxTokens,
			Temperature: cfg.Temperature,
		}
	}
	return a
}

// Analyze takes an AnomalyContext and returns an AI-generated analysis.
func (a *Advisor) Analyze(ctx context.Context, ac AnomalyContext) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("AI provider not configured")
	}

	systemPrompt := "You are a Linux system administrator assistant. Analyze anomalies and provide actionable advice."
	userPrompt := buildAnomalyPrompt(ac)

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	return a.provider.Chat(ctx, messages)
}

// IsEnabled returns true if the advisor has a valid provider configured.
func (a *Advisor) IsEnabled() bool {
	return a.provider != nil && a.cfg.APIKey != ""
}

// Configure updates the advisor configuration and recreates the provider.
func (a *Advisor) Configure(cfg Config) {
	*a = *NewAdvisor(cfg)
}

// ProviderName returns the current provider name.
func (a *Advisor) ProviderName() string {
	return a.cfg.Provider
}

// buildAnomalyPrompt creates a user prompt from anomaly context.
func buildAnomalyPrompt(ac AnomalyContext) string {
	return fmt.Sprintf(
		"Anomaly Detected:\n"+
			"Type: %s\n"+
			"Process: %s (PID: %d)\n"+
			"Metric: %s\n"+
			"Value: %.2f (Threshold: %.2f)\n"+
			"Duration: %s\n"+
			"Details: %s\n\n"+
			"Provide analysis and recommended actions:",
		ac.Type, ac.ProcessName, ac.PID, ac.Metric, ac.Value, ac.Threshold, ac.Duration, ac.Details,
	)
}