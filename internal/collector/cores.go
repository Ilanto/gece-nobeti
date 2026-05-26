package collector

import "github.com/burak/linux-dashboard/internal/linuxproc"

type CoresCollector struct{}

func NewCoresCollector() *CoresCollector { return &CoresCollector{} }

func (c *CoresCollector) Collect() []CoreMetrics {
	raw := linuxproc.CollectCores()
	out := make([]CoreMetrics, len(raw))
	for i, r := range raw {
		out[i] = CoreMetrics{
			ID:        r.ID,
			FreqMHz:   float64(r.FreqKHz) / 1000.0,
			Governor:  r.Governor,
			NumaNode:  r.NumaNode,
			Microcode: r.Microcode,
		}
	}
	return out
}
