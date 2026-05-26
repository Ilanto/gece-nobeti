package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfigPath returns the effective config file path used.
func (c *Config) ConfigPath() string {
	return ""
}

func Load(path string) (*Config, error) {
	if path == "" {
		home := os.Getenv("HOME")
		if home == "" {
			home = "."
		}
		path = filepath.Join(home, ".config", "linux-dashboard", "config.yaml")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Apply defaults for any unset fields
	if cfg.Monitoring.Interval == 0 {
		cfg.Monitoring.Interval = 1 * time.Second
	}
	if cfg.Monitoring.ProcessTreeInterval == 0 {
		cfg.Monitoring.ProcessTreeInterval = 2 * time.Second
	}
	if cfg.Monitoring.PortScanInterval == 0 {
		cfg.Monitoring.PortScanInterval = 3 * time.Second
	}
	if cfg.Monitoring.MaxProcesses == 0 {
		cfg.Monitoring.MaxProcesses = 2000
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.WellKnownPorts == nil {
		cfg.WellKnownPorts = map[uint16]string{}
	}

	return &cfg, nil
}

type Config struct {
	SchemaVersion  int
	Server         ServerConfig
	Monitoring     MonitoringConfig
	Controller     ControllerConfig
	Anomaly        AnomalyConfig
	Notifications  NotificationsConfig
	WellKnownPorts map[uint16]string
	AI             AIConfig
	Telegram       TelegramConfig
	UI             UIConfig
	Rules          []Rule
}

type ServerConfig struct {
	Host        string
	Port        int
	OpenBrowser bool
}

type MonitoringConfig struct {
	Interval            time.Duration
	ProcessTreeInterval time.Duration
	PortScanInterval    time.Duration
	GPUInterval         time.Duration
	HistoryDuration     time.Duration
	MaxProcesses        int
}

type ControllerConfig struct {
	ProtectedProcesses []string
	ConfirmKillSystem  bool
}

type AnomalyConfig struct{}

type NotificationsConfig struct{}

type AIConfig struct {
	Enabled         bool
	Provider        string // "anthropic", "openai", "openrouter", "minimax", "deepseek", "groq"
	APIKey          string
	Model           string
	Endpoint        string
	MaxTokens       int
	Temperature     float64
	Language        string
	MaxRequestsPerMinute int
}

type TelegramConfig struct {
	Enabled        bool
	BotToken       string
	ChatIDs        []int64
	AllowedChatIDs []int64 // alias for ChatIDs
	RequireConfirm bool
	ConfirmTTL     time.Duration
	PollTimeout    time.Duration
	APIBaseURL     string
}

type UIConfig struct {}

type Rule struct {}

func DefaultConfig() *Config {
	return &Config{
		Monitoring: MonitoringConfig{
			Interval:            1 * time.Second,
			ProcessTreeInterval: 2 * time.Second,
			PortScanInterval:    3 * time.Second,
			MaxProcesses:        2000,
		},
		WellKnownPorts: map[uint16]string{},
	}
}
