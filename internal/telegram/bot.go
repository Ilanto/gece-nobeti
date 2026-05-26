package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/burak/linux-dashboard/internal/anomaly"
	"github.com/burak/linux-dashboard/internal/collector"
	"github.com/burak/linux-dashboard/internal/config"
	"github.com/burak/linux-dashboard/internal/event"
	"github.com/burak/linux-dashboard/internal/storage"
)

// Bot handles Telegram bot interactions.
type Bot struct {
	mu         sync.RWMutex
	cfg        *config.Config
	store      *storage.Store
	alerts     *anomaly.AlertStore
	httpClient *http.Client

	offset    int64
	lastToken string
	rootCtx   context.Context
}

type tgUpdate struct {
	UpdateID int64      `json:"update_id"`
	Message  *tgMessage `json:"message,omitempty"`
}

type tgMessage struct {
	MessageID int64  `json:"message_id"`
	Text      string `json:"text"`
	Chat      tgChat `json:"chat"`
}

type tgChat struct {
	ID int64 `json:"id"`
}

type updateResp struct {
	OK          bool       `json:"ok"`
	Result      []tgUpdate `json:"result"`
	Description string     `json:"description"`
	ErrorCode   int        `json:"error_code"`
}

const maxTelegramResponseBytes = 1 << 20

// New creates a new Telegram bot.
func New(cfg *config.Config, store *storage.Store, alerts *anomaly.AlertStore, emitter *event.Emitter) *Bot {
	b := &Bot{
		cfg:        cfg,
		store:      store,
		alerts:     alerts,
		httpClient: &http.Client{Timeout: 40 * time.Second},
		rootCtx:    context.Background(),
	}
	if emitter != nil {
		emitter.On(anomaly.EventAnomalyDetected, b.handleAnomalyDetected)
	}
	return b
}

// SetConfig updates the bot configuration.
func (b *Bot) SetConfig(cfg *config.Config) {
	b.mu.Lock()
	b.cfg = cfg
	b.mu.Unlock()
}

// Start begins the bot's update loop.
func (b *Bot) Start(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	b.mu.Lock()
	b.rootCtx = ctx
	b.mu.Unlock()
	go b.loop(ctx)
}

func (b *Bot) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		cfg := b.currentConfig()
		if cfg == nil || !cfg.Telegram.Enabled || cfg.Telegram.BotToken == "" || len(cfg.Telegram.ChatIDs) == 0 {
			if !sleepContext(ctx, 5*time.Second) {
				return
			}
			continue
		}

		token := cfg.Telegram.BotToken
		if token != b.lastToken {
			b.offset = 0
			b.lastToken = token
		}

		updates, err := b.getUpdates(ctx, cfg)
		if err != nil {
			log.Printf("telegram: getUpdates: %v", err)
			if !sleepContext(ctx, 3*time.Second) {
				return
			}
			continue
		}
		for _, upd := range updates {
			b.offset = upd.UpdateID + 1
			if upd.Message == nil || strings.TrimSpace(upd.Message.Text) == "" {
				continue
			}
			if !isAllowedChat(cfg.Telegram.ChatIDs, upd.Message.Chat.ID) {
				continue
			}
			b.handleMessage(ctx, cfg, upd.Message)
		}
	}
}

func (b *Bot) currentConfig() *config.Config {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.cfg
}

func sleepContext(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func isAllowedChat(chatIDs []int64, id int64) bool {
	for _, allowed := range chatIDs {
		if allowed == id {
			return true
		}
	}
	return false
}

func (b *Bot) getUpdates(ctx context.Context, cfg *config.Config) ([]tgUpdate, error) {
	timeoutSec := 25
	body := map[string]any{
		"offset":          b.offset,
		"timeout":         timeoutSec,
		"allowed_updates": []string{"message"},
	}
	var resp updateResp
	if err := b.apiCall(ctx, cfg, "getUpdates", body, &resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("telegram %d: %s", resp.ErrorCode, resp.Description)
	}
	return resp.Result, nil
}

func (b *Bot) apiCall(ctx context.Context, cfg *config.Config, method string, body any, dst any) error {
	baseURL := "https://api.telegram.org"
	url := fmt.Sprintf("%s/bot%s/%s", baseURL, cfg.Telegram.BotToken, method)

	var rdr io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxTelegramResponseBytes+1))
	if err != nil {
		return err
	}
	if len(raw) > maxTelegramResponseBytes {
		return fmt.Errorf("telegram response exceeds %d bytes", maxTelegramResponseBytes)
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("decode telegram response: %w", err)
	}
	return nil
}

func (b *Bot) handleMessage(ctx context.Context, cfg *config.Config, msg *tgMessage) {
	cmd, _ := parseCommand(msg.Text)
	if cmd == "" {
		return
	}

	var reply string
	switch cmd {
	case "start", "help":
		reply = helpText()
	case "status":
		reply = b.statusText()
	case "alerts":
		reply = b.alertsText()
	default:
		reply = "Unknown command. Send /help."
	}

	if err := b.sendMessage(ctx, cfg, msg.Chat.ID, reply); err != nil {
		log.Printf("telegram: send reply: %v", err)
	}
}

func parseCommand(text string) (string, []string) {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 {
		return "", nil
	}
	cmd := strings.TrimPrefix(fields[0], "/")
	if idx := strings.IndexByte(cmd, '@'); idx >= 0 {
		cmd = cmd[:idx]
	}
	return strings.ToLower(cmd), fields[1:]
}

func helpText() string {
	return strings.Join([]string{
		"Linux Dashboard Bot Commands:",
		"/start - Show welcome message",
		"/status - Show CPU, memory, and top processes",
		"/alerts - Show active anomaly alerts",
		"/help - Show this help message",
	}, "\n")
}

func (b *Bot) statusText() string {
	snap := b.store.Latest()
	if snap == nil {
		return "No snapshot yet."
	}
	top := topProcessesByCPU(snap.Processes, 3)
	lines := []string{
		fmt.Sprintf("CPU %.1f%% (%d cores)", snap.CPU.TotalPercent, snap.CPU.NumLogical),
		fmt.Sprintf("Memory %.1f%%", snap.Memory.UsedPercent),
		fmt.Sprintf("Network ↓ %s/s ↑ %s/s", formatBytes(int64(snap.Network.TotalDownBPS)), formatBytes(int64(snap.Network.TotalUpBPS))),
		"Top CPU:",
	}
	for _, p := range top {
		lines = append(lines, fmt.Sprintf("- %s PID %d CPU %.1f%% MEM %s", p.Name, p.PID, p.CPUPercent, formatBytes(int64(p.WorkingSet))))
	}
	return strings.Join(lines, "\n")
}

func topProcessesByCPU(procs []collector.ProcessInfo, n int) []collector.ProcessInfo {
	if len(procs) == 0 {
		return nil
	}
	sorted := make([]collector.ProcessInfo, len(procs))
	copy(sorted, procs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CPUPercent > sorted[j].CPUPercent
	})
	if len(sorted) > n {
		return sorted[:n]
	}
	return sorted
}

func (b *Bot) alertsText() string {
	items := b.alerts.GetAll()
	if len(items) == 0 {
		return "No active alerts."
	}
	if len(items) > 8 {
		items = items[:8]
	}
	lines := []string{"Active alerts:"}
	for _, a := range items {
		lines = append(lines, fmt.Sprintf("- [%s] %s: %s", strings.ToUpper(a.Severity), a.Type, a.Message))
	}
	return strings.Join(lines, "\n")
}

func (b *Bot) handleAnomalyDetected(data any) {
	a, ok := data.(*anomaly.Anomaly)
	if !ok {
		return
	}
	b.mu.RLock()
	cfg := b.cfg
	b.mu.RUnlock()

	if cfg == nil || !cfg.Telegram.Enabled || len(cfg.Telegram.ChatIDs) == 0 {
		return
	}

	ctx := context.Background()
	severity := "⚠️"
	if a.Severity == "critical" {
		severity = "🚨"
	}
	msg := fmt.Sprintf("%s Anomaly Detected\nType: %s\nMessage: %s\nValue: %.1f%%", severity, a.Type, a.Message, a.Value)
	for _, chatID := range cfg.Telegram.ChatIDs {
		if err := b.sendMessage(ctx, cfg, chatID, msg); err != nil {
			log.Printf("telegram: failed to send anomaly alert to %d: %v", chatID, err)
		}
	}
}

func (b *Bot) sendMessage(ctx context.Context, cfg *config.Config, chatID int64, text string) error {
	body := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}
	var resp map[string]any
	return b.apiCall(ctx, cfg, "sendMessage", body, &resp)
}

func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return strconv.FormatInt(n, 10) + " B"
	}
	exp := 0
	for n >= unit && exp < 4 {
		n /= unit
		exp++
	}
	return strconv.FormatInt(n, 10) + " " + string("BKMGT"[exp]) + "B"
}