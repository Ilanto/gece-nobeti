package collector

import (
	"sync"
	"time"
)

// SnapshotStore holds the latest system snapshot and recent history.
type SnapshotStore struct {
	mu      sync.RWMutex
	latest  *SystemSnapshot
}

func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{}
}

func (s *SnapshotStore) SetLatest(snap *SystemSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latest = snap
}

func (s *SnapshotStore) Latest() *SystemSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latest
}

// TimestampedSnapshot pairs a snapshot with its collection time.
type TimestampedSnapshot struct {
	Time    time.Time
	CPU     CPUMetrics
	Memory  MemoryMetrics
	GPU     GPUMetrics
	Network NetworkMetrics
	Disk    DiskMetrics
}