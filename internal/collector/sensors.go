package collector

import "github.com/burak/linux-dashboard/internal/linuxproc"

type SensorsCollector struct{}

func NewSensorsCollector() *SensorsCollector { return &SensorsCollector{} }

func (s *SensorsCollector) Collect() []SensorMetric {
	raw := linuxproc.CollectSensors()
	out := make([]SensorMetric, len(raw))
	for i, r := range raw {
		out[i] = SensorMetric{
			Name:     r.Name,
			Value:    r.Value,
			Unit:     r.Unit,
			Critical: r.Critical,
		}
	}
	return out
}
