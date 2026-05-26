package collector

import (
	"sync"
	"time"

	"github.com/burak/linux-dashboard/internal/linuxproc"
)

type ProcessCollector struct {
	maxProcs int
	prevProc map[uint32]procSample
	mu       sync.Mutex
}

type procSample struct {
	utime    uint64
	stime    uint64
	lastTime time.Time
}

func NewProcessCollector(maxProcs int) *ProcessCollector {
	if maxProcs <= 0 {
		maxProcs = 2000
	}
	return &ProcessCollector{
		maxProcs: maxProcs,
		prevProc: make(map[uint32]procSample),
	}
}

func (p *ProcessCollector) Collect() []ProcessInfo {
	procs, _ := linuxproc.AllProcesses()

	p.mu.Lock()
	defer p.mu.Unlock()

	result := make([]ProcessInfo, 0, len(procs))
	now := time.Now()

	clockTicks := float64(100) // usually 100 Hz, but could be 250 or 1000

	for _, lp := range procs {
		sample := procSample{
			utime:    lp.UTime,
			stime:    lp.STime,
			lastTime: now,
		}

		var cpuPercent float64
		if prev, ok := p.prevProc[lp.PID]; ok {
			dt := now.Sub(prev.lastTime).Seconds()
			if dt > 0 {
				deltaU := float64(lp.UTime - prev.utime)
				deltaS := float64(lp.STime - prev.stime)
				totalDelta := deltaU + deltaS
				cpuPercent = (totalDelta / dt) / clockTicks * 100
				if cpuPercent > 100 {
					cpuPercent = 100
				}
			}
		}

		p.prevProc[lp.PID] = sample

		status := processState(lp.State)

		result = append(result, ProcessInfo{
			PID:         lp.PID,
			ParentPID:   lp.ParentPID,
			Name:        lp.Name,
			ExePath:     lp.ExePath,
			CPUPercent:  cpuPercent,
			WorkingSet:  lp.VMRSS * 1024,
			PrivateBytes: lp.VMSize * 1024,
			PageFaults:   0,
			IOReadBytes:  0,
			IOWriteBytes: 0,
			IOReadOps:    0,
			IOWriteOps:   0,
			ThreadCount:  lp.Threads,
			CreateTime:   lp.StartTime,
			IsCritical:   false,
			Status:       status,
			Connections:  0,
			PriorityClass: 0,
			UID:          lp.UID,
			CoreID:       lp.CoreID,
		})
	}

	// Prune old entries
	for pid := range p.prevProc {
		if !findPID(procs, pid) {
			delete(p.prevProc, pid)
		}
	}

	// Limit result
	if len(result) > p.maxProcs {
		result = result[:p.maxProcs]
	}

	return result
}

func findPID(procs []linuxproc.ProcessInfo, pid uint32) bool {
	for _, p := range procs {
		if p.PID == pid {
			return true
		}
	}
	return false
}

func processState(s string) string {
	switch s {
	case "R":
		return "running"
	case "S":
		return "sleeping"
	case "D":
		return "disk_sleep"
	case "Z":
		return "zombie"
	case "T":
		return "stopped"
	case "t":
		return "tracing"
	case "X", "x":
		return "dead"
	case "I":
		return "idle"
	default:
		return s
	}
}