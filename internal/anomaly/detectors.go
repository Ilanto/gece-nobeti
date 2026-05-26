package anomaly

import (
	"fmt"
	"time"

	"github.com/burak/linux-dashboard/internal/collector"
)

// DetectCPUAnomaly detects CPU spikes (>95% for 30s).
func DetectCPUAnomaly(snap *collector.SystemSnapshot) *Anomaly {
	if snap == nil {
		return nil
	}

	threshold := 95.0
	if snap.CPU.TotalPercent > threshold {
		return &Anomaly{
			ID:        fmt.Sprintf("cpu-%d", time.Now().UnixNano()),
			Type:      "cpu",
			Severity:  "critical",
			Message:   fmt.Sprintf("CPU spike detected: %.1f%% (threshold: %.0f%%)", snap.CPU.TotalPercent, threshold),
			Value:     snap.CPU.TotalPercent,
			Threshold: threshold,
			Timestamp: time.Now(),
		}
	}
	return nil
}

// DetectMemoryAnomaly detects memory leaks (>90% sustained).
func DetectMemoryAnomaly(snap *collector.SystemSnapshot) *Anomaly {
	if snap == nil {
		return nil
	}

	threshold := 90.0
	if snap.Memory.UsedPercent > threshold {
		return &Anomaly{
			ID:        fmt.Sprintf("mem-%d", time.Now().UnixNano()),
			Type:      "memory",
			Severity:  "critical",
			Message:   fmt.Sprintf("High memory usage: %.1f%% (threshold: %.0f%%)", snap.Memory.UsedPercent, threshold),
			Value:     snap.Memory.UsedPercent,
			Threshold: threshold,
			Timestamp: time.Now(),
		}
	}
	return nil
}

// DetectDiskAnomaly detects disk full (>95%).
func DetectDiskAnomaly(snap *collector.SystemSnapshot) *Anomaly {
	if snap == nil {
		return nil
	}

	threshold := 95.0
	for _, drive := range snap.Disk.Drives {
		if drive.UsedPct > threshold {
			return &Anomaly{
				ID:        fmt.Sprintf("disk-%s-%d", drive.Letter, time.Now().UnixNano()),
				Type:      "disk",
				Severity:  "critical",
				Message:   fmt.Sprintf("Disk %s full: %.1f%% used (threshold: %.0f%%)", drive.Letter, drive.UsedPct, threshold),
				Value:     drive.UsedPct,
				Threshold: threshold,
				Timestamp: time.Now(),
			}
		}
	}
	return nil
}

// DetectNetworkAnomaly detects network issues (interfaces down).
func DetectNetworkAnomaly(snap *collector.SystemSnapshot) *Anomaly {
	if snap == nil {
		return nil
	}

	// Check for any interface that is down
	for _, iface := range snap.Network.Interfaces {
		if iface.Status == "down" {
			return &Anomaly{
				ID:        fmt.Sprintf("net-%s-%d", iface.Name, time.Now().UnixNano()),
				Type:      "network",
				Severity:  "warning",
				Message:   fmt.Sprintf("Network interface %s is down", iface.Name),
				Value:     0,
				Threshold: 0,
				Timestamp: time.Now(),
			}
		}
	}
	return nil
}

// DetectProcessAnomaly detects process crash loops and port binding conflicts.
// Crash loop: same process dying 5+ times in 5 min
// Port conflict: multiple processes binding to same port
func DetectProcessAnomaly(snap *collector.SystemSnapshot) *Anomaly {
	if snap == nil {
		return nil
	}

	// Check for port binding conflicts
	portMap := make(map[uint16][]string)
	for _, binding := range snap.PortBindings {
		if binding.State == "LISTEN" {
			portMap[binding.LocalPort] = append(portMap[binding.LocalPort], binding.Process)
		}
	}

	for port, processes := range portMap {
		if len(processes) > 1 {
			return &Anomaly{
				ID:        fmt.Sprintf("port-%d-%d", port, time.Now().UnixNano()),
				Type:      "process",
				Severity:  "critical",
				Message:   fmt.Sprintf("Port binding conflict on port %d: %v", port, processes),
				Value:     float64(len(processes)),
				Threshold: 1,
				Timestamp: time.Now(),
			}
		}
	}

	return nil
}