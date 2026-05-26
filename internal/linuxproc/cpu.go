// Package linuxproc provides low-level access to Linux's /proc filesystem.
package linuxproc

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CPUInfo holds raw per-CPU stats from /proc/stat.
type CPUInfo struct {
	User    uint64
	Nice    uint64
	System  uint64
	Idle    uint64
	Iowait  uint64
	Irq     uint64
	Softirq uint64
	Steal   uint64
	Guest   uint64
	Total   uint64
}

// CPUSnapshot holds total and per-core CPU usage percentages.
type CPUSnapshot struct {
	TotalPercent float64
	PerCore      []float64
	NumLogical   int
	Name         string
	FreqMHz      uint32
}

// cpuCollector holds state between samples.
type cpuCollector struct {
	mu      sync.RWMutex
	prev    map[string]CPUInfo // keyed by "cpu0", "cpu1", ... or "cpu"
	prevTs  time.Time
	name    string
	freqMHz uint32
}

var (
	globalCPUCollector *cpuCollector
	onceCPU            sync.Once
)

// Global returns the singleton CPU collector.
func GlobalCPU() *cpuCollector {
	onceCPU.Do(func() {
		globalCPUCollector = &cpuCollector{
			prev: make(map[string]CPUInfo),
			name: CPUName(),
		}
		globalCPUCollector.freqMHz = CPUFreqMHz()
	})
	return globalCPUCollector
}

// Collect returns current CPU usage snapshot.
func (c *cpuCollector) Collect() CPUSnapshot {
	return c.collectWithClock(time.Now())
}

func (c *cpuCollector) collectWithClock(now time.Time) CPUSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	snapshot := CPUSnapshot{
		Name:         c.name,
		FreqMHz:      c.freqMHz,
		NumLogical:   runtime.NumCPU(),
		PerCore:      []float64{},
		TotalPercent: 0,
	}

	// Read all lines from /proc/stat
	f, err := os.Open("/proc/stat")
	if err != nil {
		return snapshot
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	var cpuLine string // reserved for future total line parsing
	_ = cpuLine
	perCore := []float64{}

	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "cpu") {
			fields := strings.Fields(line)
			if len(fields) < 8 {
				continue
			}

			ci := CPUInfo{
				User:    ParseUint(fields[1], 0),
				Nice:    ParseUint(fields[2], 0),
				System:  ParseUint(fields[3], 0),
				Idle:    ParseUint(fields[4], 0),
				Iowait:  ParseUint(fields[5], 0),
				Irq:     ParseUint(fields[6], 0),
				Softirq: ParseUint(fields[7], 0),
			}
			if len(fields) >= 9 {
				ci.Steal = ParseUint(fields[8], 0)
			}
			if len(fields) >= 10 {
				ci.Guest = ParseUint(fields[9], 0)
			}
			ci.Total = ci.User + ci.Nice + ci.System + ci.Idle + ci.Iowait + ci.Irq + ci.Softirq + ci.Steal + ci.Guest

			key := fields[0]

			if key == "cpu" {
				cpuLine = line
				prev, ok := c.prev[key]
				if ok {
					deltaTotal := ci.Total - prev.Total
					deltaIdle := ci.Idle - prev.Idle
					if deltaTotal > 0 {
						snapshot.TotalPercent = float64(deltaTotal-deltaIdle) / float64(deltaTotal) * 100
					}
				}
				c.prev[key] = ci
			} else {
				// Per-core
				prev, ok := c.prev[key]
				var pct float64
				if ok {
					deltaTotal := ci.Total - prev.Total
					deltaIdle := ci.Idle - prev.Idle
					if deltaTotal > 0 {
						pct = float64(deltaTotal-deltaIdle) / float64(deltaTotal) * 100
					}
				}
				perCore = append(perCore, pct)
				c.prev[key] = ci
			}
		}
	}

	snapshot.PerCore = perCore
	return snapshot
}

// CPUName returns processor name from /proc/cpuinfo.
func CPUName() string {
	data := ReadFile("/proc/cpuinfo")
	if data == "" {
		return "unknown"
	}
	sc := bufio.NewScanner(strings.NewReader(data))
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "model name") || strings.HasPrefix(line, "Hardware") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return "unknown"
}

// CPUFreqMHz returns CPU frequency in MHz from /sys.
func CPUFreqMHz() uint32 {
	data := ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq")
	if data == "" {
		return 0
	}
	v, _ := strconv.ParseUint(strings.TrimSpace(data), 10, 64)
	return uint32(v / 1000) // kHz to MHz
}

// CollectCPU is the top-level function used by the collector package.
func CollectCPU() (total float64, perCore []float64, numLogical int, name string, freqMHz uint32) {
	snap := GlobalCPU().Collect()
	return snap.TotalPercent, snap.PerCore, snap.NumLogical, snap.Name, snap.FreqMHz
}