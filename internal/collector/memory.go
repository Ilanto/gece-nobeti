package collector

import "github.com/burak/linux-dashboard/internal/linuxproc"

type MemoryCollector struct{}

func NewMemoryCollector() *MemoryCollector {
	return &MemoryCollector{}
}

func (m *MemoryCollector) Collect() MemoryMetrics {
	info := linuxproc.CollectMemory()
	freePhys := info.Total - info.Used - info.Buffers - info.Cached
	if freePhys > info.Total {
		freePhys = 0
	}
	swapUsed := uint64(0)
	if info.SwapTotal > info.SwapFree {
		swapUsed = info.SwapTotal - info.SwapFree
	}
	return MemoryMetrics{
		TotalPhys:     info.Total,
		AvailPhys:     info.Available,
		UsedPhys:      info.Used,
		FreePhys:      freePhys,
		Buffers:       info.Buffers,
		Cached:        info.Cached,
		UsedPercent:   info.UsedPercent,
		TotalPageFile: info.Total + info.SwapTotal,
		AvailPageFile: info.Available + info.SwapFree,
		CommitCharge:  0,
		SwapUsed:      swapUsed,
	}
}