package collector

import (
	"sync"
	"time"

	"github.com/burak/linux-dashboard/internal/event"
	"github.com/burak/linux-dashboard/internal/linuxproc"
)

// Manager coordinates all collectors and provides a unified snapshot API.
type Manager struct {
	cpu       *CPUCollector
	memory    *MemoryCollector
	disk      *DiskCollector
	network   *NetworkCollector
	gpu       *GPUCollector
	processes *ProcessCollector
	ports     *PortCollector
	host      *HostCollector
	cores     *CoresCollector
	sensors   *SensorsCollector

	store    *SnapshotStore
	emitter  *event.Emitter
	mu       sync.RWMutex
	stopCh   chan struct{}
}

func NewManager(emit *event.Emitter) *Manager {
	m := &Manager{
		cpu:       NewCPUCollector("", 0),
		memory:    NewMemoryCollector(),
		disk:      NewDiskCollector(),
		network:   NewNetworkCollector(),
		gpu:       NewGPUCollector(),
		processes: NewProcessCollector(2000),
		ports:     NewPortCollector(nil),
		host:      NewHostCollector(),
		cores:     NewCoresCollector(),
		sensors:   NewSensorsCollector(),
		store:     NewSnapshotStore(),
		emitter:   emit,
		stopCh:    make(chan struct{}),
	}
	return m
}

// Start begins periodic collection in the background.
func (m *Manager) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				m.Collect()
			case <-m.stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop halts background collection.
func (m *Manager) Stop() {
	close(m.stopCh)
}

// Collect gathers all metrics and stores the snapshot.
func (m *Manager) Collect() {
	m.mu.Lock()
	defer m.mu.Unlock()

	procs := m.processes.Collect()
	tree := BuildProcessTree(procs)
	ports := m.ports.Collect(nil)

	snap := &SystemSnapshot{
		Timestamp:    time.Now().Unix(),
		CPU:          m.cpu.Collect(),
		Memory:       m.memory.Collect(),
		Disk:         m.disk.Collect(),
		Network:      m.network.Collect(),
		GPU:          m.gpu.Collect(),
		Processes:    procs,
		ProcessTree:  tree,
		PortBindings: ports,
	}

	m.store.SetLatest(snap)

	if m.emitter != nil {
		m.emitter.Emit("metrics.snapshot", snap)
	}
}

// LatestSnapshot returns the most recent system snapshot.
func (m *Manager) LatestSnapshot() *SystemSnapshot {
	return m.store.Latest()
}

// LatestProcessTree returns the most recent process tree.
func (m *Manager) LatestProcessTree() []*ProcessNode {
	snap := m.store.Latest()
	if snap == nil {
		return nil
	}
	return snap.ProcessTree
}

// LatestPortBindings returns the most recent port bindings.
func (m *Manager) LatestPortBindings() []PortBinding {
	snap := m.store.Latest()
	if snap == nil {
		return nil
	}
	return snap.PortBindings
}

// LatestHost returns current system identity info.
func (m *Manager) LatestHost() HostMetrics {
	return m.host.Collect()
}

// LatestCores returns per-CPU-core info.
func (m *Manager) LatestCores() []CoreMetrics {
	return m.cores.Collect()
}

// LatestSensors returns hardware sensor readings.
func (m *Manager) LatestSensors() []SensorMetric {
	return m.sensors.Collect()
}

// FetchSyslog runs journalctl and returns the last n log entries.
func (m *Manager) FetchSyslog(n int) []SyslogMetric {
	raw := linuxproc.CollectSyslog(n)
	out := make([]SyslogMetric, len(raw))
	for i, r := range raw {
		out[i] = SyslogMetric{
			Timestamp: r.Timestamp,
			Facility:  r.Facility,
			Severity:  r.Severity,
			Message:   r.Message,
			Source:    r.Source,
		}
	}
	return out
}

// FetchConnections reads all TCP/UDP connections from /proc/net.
func (m *Manager) FetchConnections() []ConnectionMetric {
	raw := linuxproc.CollectConnections()
	out := make([]ConnectionMetric, len(raw))
	for i, r := range raw {
		out[i] = ConnectionMetric{
			Protocol:   r.Protocol,
			LocalAddr:  r.LocalAddr,
			LocalPort:  r.LocalPort,
			RemoteAddr: r.RemoteAddr,
			RemotePort: r.RemotePort,
			State:      r.State,
			PID:        r.PID,
			Process:    r.Process,
		}
	}
	return out
}