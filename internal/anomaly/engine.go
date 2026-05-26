package anomaly

import (
	"context"
	"sync"

	"github.com/burak/linux-dashboard/internal/collector"
	"github.com/burak/linux-dashboard/internal/event"
)

// Event names emitted by the engine.
const (
	EventAnomalyDetected = "anomaly.detected"
)

// Engine runs anomaly detection on system snapshots.
type Engine struct {
	emitter   *event.Emitter
	alerts    *AlertStore
	snapshots *snapshotHistory

	mu sync.RWMutex
}

type snapshotHistory struct {
	mu     sync.RWMutex
	snaps  []collector.SystemSnapshot
	maxLen int
}

// NewEngine creates a new anomaly detection engine.
func NewEngine(emitter *event.Emitter, alerts *AlertStore) *Engine {
	e := &Engine{
		emitter:   emitter,
		alerts:    alerts,
		snapshots: &snapshotHistory{maxLen: 300}, // 5 min history at 1s interval
	}
	return e
}

// Start begins listening for metric snapshots.
func (e *Engine) Start(ctx context.Context) {
	if e.emitter != nil {
		e.emitter.On("metrics.snapshot", e.handleSnapshot)
	}
}

func (e *Engine) handleSnapshot(data any) {
	snap, ok := data.(*collector.SystemSnapshot)
	if !ok {
		return
	}

	e.snapshots.add(*snap)

	anomalies := e.detectAnomalies(snap)
	for _, a := range anomalies {
		e.alerts.Store(a)
		if e.emitter != nil {
			e.emitter.Emit(EventAnomalyDetected, a)
		}
	}
}

func (e *Engine) detectAnomalies(snap *collector.SystemSnapshot) []*Anomaly {
	var anomalies []*Anomaly

	if a := DetectCPUAnomaly(snap); a != nil {
		anomalies = append(anomalies, a)
	}
	if a := DetectMemoryAnomaly(snap); a != nil {
		anomalies = append(anomalies, a)
	}
	if a := DetectDiskAnomaly(snap); a != nil {
		anomalies = append(anomalies, a)
	}
	if a := DetectNetworkAnomaly(snap); a != nil {
		anomalies = append(anomalies, a)
	}
	if a := DetectProcessAnomaly(snap); a != nil {
		anomalies = append(anomalies, a)
	}

	return anomalies
}

// snapshotHistory maintains a ring buffer of recent snapshots.
func (sh *snapshotHistory) add(snap collector.SystemSnapshot) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	sh.snaps = append(sh.snaps, snap)
	if len(sh.snaps) > sh.maxLen {
		sh.snaps = sh.snaps[1:]
	}
}

func (sh *snapshotHistory) getLastN(n int) []collector.SystemSnapshot {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	if len(sh.snaps) < n {
		return sh.snaps
	}
	return sh.snaps[len(sh.snaps)-n:]
}

// GetHistory returns the last N snapshots for analysis.
func (e *Engine) GetHistory(n int) []collector.SystemSnapshot {
	return e.snapshots.getLastN(n)
}