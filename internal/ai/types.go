package ai

import "context"

// Config holds AI provider settings.
type Config struct {
	Provider    string  // "anthropic", "openai", "openrouter", "minimax", "deepseek", "groq"
	APIKey      string
	Model       string
	Endpoint    string // optional override
	MaxTokens   int
	Temperature float64
}

// Advisor wraps an AI provider and provides anomaly analysis.
type Advisor struct {
	cfg      Config
	provider AIProvider
}

// AIProvider is the interface for AI backends.
type AIProvider interface {
	Chat(ctx context.Context, messages []Message) (string, error)
}

// Message represents a single chat message.
type Message struct {
	Role    string // "system", "user", "assistant"
	Content string
}

// AnomalyContext holds context about a detected anomaly for AI analysis.
type AnomalyContext struct {
	Type        string  // "runaway_cpu", "memory_leak", "orphan", "port_conflict", "spawn_storm"
	ProcessName string
	PID         uint32
	Metric      string
	Value       float64
	Threshold   float64
	Duration    string
	Details     string
}