package collector

import (
	"sync"
	"time"

	"github.com/burak/linux-dashboard/internal/linuxproc"
)

type NetworkCollector struct {
	prevStats map[string]linuxproc.InterfaceStats
	prevTime  time.Time
	mu        sync.Mutex
}

func NewNetworkCollector() *NetworkCollector {
	return &NetworkCollector{
		prevStats: make(map[string]linuxproc.InterfaceStats),
		prevTime:  time.Now(),
	}
}

func (n *NetworkCollector) Collect() NetworkMetrics {
	current := linuxproc.CollectNetwork()
	now := time.Now()

	n.mu.Lock()
	defer n.mu.Unlock()

	var totalUp, totalDown uint64
	interfaces := make([]InterfaceInfo, 0, len(current))

	for _, s := range current {
		var inBPS, outBPS uint64
		var inPPS, outPPS uint64

		if prev, ok := n.prevStats[s.Name]; ok {
			dt := now.Sub(n.prevTime).Seconds()
			if dt > 0 {
				if s.InBytes >= prev.InBytes {
					inBPS = uint64(float64(s.InBytes-prev.InBytes) / dt)
				}
				if s.OutBytes >= prev.OutBytes {
					outBPS = uint64(float64(s.OutBytes-prev.OutBytes) / dt)
				}
				if s.InPackets >= prev.InPackets {
					inPPS = uint64(float64(s.InPackets-prev.InPackets) / dt)
				}
				if s.OutPackets >= prev.OutPackets {
					outPPS = uint64(float64(s.OutPackets-prev.OutPackets) / dt)
				}
			}
		}

		if s.Status == "up" && s.Name != "lo" {
			totalDown += inBPS
			totalUp += outBPS
		}

		interfaces = append(interfaces, InterfaceInfo{
			Name:      s.Name,
			Type:      s.Type,
			Status:    s.Status,
			SpeedMbps: s.SpeedMbps,
			InBPS:     inBPS,
			OutBPS:    outBPS,
			InPPS:     inPPS,
			OutPPS:    outPPS,
			InErrors:  s.InErrors,
			OutErrors: s.OutErrors,
			Address:   s.Address,
		})

		n.prevStats[s.Name] = s
	}

	n.prevTime = now

	return NetworkMetrics{
		Interfaces:   interfaces,
		TotalUpBPS:   totalUp,
		TotalDownBPS: totalDown,
	}
}