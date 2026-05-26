package ai

import "github.com/burak/linux-dashboard/internal/config"

// loadAIConfig maps config.AI settings to ai.Config.
func loadAIConfig(cfg *config.Config) Config {
	if cfg == nil {
		return Config{}
	}
	return Config{
		Provider:    cfg.AI.Provider,
		APIKey:      cfg.AI.APIKey,
		Model:       cfg.AI.Model,
		Endpoint:    cfg.AI.Endpoint,
		MaxTokens:   cfg.AI.MaxTokens,
		Temperature: cfg.AI.Temperature,
	}
}