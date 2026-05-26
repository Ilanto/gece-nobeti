package storage

import (
	"sync"
	"time"

	"github.com/burak/linux-dashboard/internal/collector"
	"github.com/burak/linux-dashboard/internal/config"
)

// Store provides access to system snapshots and configuration.
type Store struct {
	mu      sync.RWMutex
	latest  *collector.SystemSnapshot
	config  *config.Config
	alerts  []string // active anomaly alerts
}

func NewStore(historyCap, procHistoryCap int) *Store {
	return &Store{
		config: config.DefaultConfig(),
		alerts: []string{},
	}
}

func (s *Store) SetLatest(snap *collector.SystemSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latest = snap
}

func (s *Store) Latest() *collector.SystemSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latest
}

func (s *Store) UpdateLatest(update func(*collector.SystemSnapshot)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.latest == nil {
		return false
	}
	update(s.latest)
	return true
}

func (s *Store) PruneStaleProcesses(cutoff time.Time) int {
	return 0
}

// ActiveAlerts returns the list of active anomaly alerts.
func (s *Store) ActiveAlerts() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]string, len(s.alerts))
	copy(result, s.alerts)
	return result
}

// GetConfig returns the current configuration.
func (s *Store) GetConfig() *config.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg := *s.config
	return &cfg
}

// UpdateConfig applies a partial config update.
func (s *Store) UpdateConfig(updates map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Apply updates to s.config
	// For now, just return nil (full implementation would merge updates)
	return nil
}