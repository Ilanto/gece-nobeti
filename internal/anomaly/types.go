package anomaly

import (
	"sync"
	"time"
)

// Anomaly represents a detected system anomaly.
type Anomaly struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // "cpu", "memory", "disk", "network", "process"
	Severity  string    `json:"severity"` // "warning", "critical"
	Message   string    `json:"message"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Timestamp time.Time `json:"timestamp"`
}

// AlertStore manages active anomaly alerts.
type AlertStore struct {
	mu      sync.RWMutex
	alerts  []*Anomaly
}

// NewAlertStore creates a new AlertStore.
func NewAlertStore() *AlertStore {
	return &AlertStore{}
}

// Store adds a new anomaly alert to the store.
func (s *AlertStore) Store(a *Anomaly) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alerts = append(s.alerts, a)
}

// GetAll returns all stored alerts.
func (s *AlertStore) GetAll() []*Anomaly {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Anomaly, len(s.alerts))
	copy(result, s.alerts)
	return result
}

// Clear removes all alerts.
func (s *AlertStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alerts = nil
}