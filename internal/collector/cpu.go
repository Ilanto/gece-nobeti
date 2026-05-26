package collector

import "github.com/burak/linux-dashboard/internal/linuxproc"

type CPUCollector struct{}

func NewCPUCollector(name string, freqMHz uint32) *CPUCollector {
	return &CPUCollector{}
}

func (c *CPUCollector) Collect() CPUMetrics {
	snap := linuxproc.GlobalCPU().Collect()
	return CPUMetrics{
		TotalPercent: snap.TotalPercent,
		PerCore:      snap.PerCore,
		NumLogical:   snap.NumLogical,
		Name:         snap.Name,
		FreqMHz:      snap.FreqMHz,
	}
}