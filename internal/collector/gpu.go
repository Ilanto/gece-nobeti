package collector

import (
	"os/exec"
	"strconv"
	"strings"
)

type GPUCollector struct{}

func NewGPUCollector() *GPUCollector {
	return &GPUCollector{}
}

func (g *GPUCollector) Collect() GPUMetrics {
	// Try nvidia-smi
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,utilization.gpu,memory.used,memory.total,temperature.gpu", "--format=csv,noheader,nounits")
	out, err := cmd.Output()
	if err != nil {
		return GPUMetrics{Available: false}
	}

	// Parse CSV: "Name, Util%, MemoryUsed MB, MemoryTotal MB, Temp C"
	// Example: "NVIDIA GeForce RTX 5080, 15, 1234, 16384, 42"
	line := strings.TrimSpace(string(out))
	fields := strings.Split(line, ",")
	if len(fields) < 5 {
		return GPUMetrics{Available: false}
	}

	name := strings.TrimSpace(fields[0])
	util, _ := strconv.ParseFloat(strings.TrimSpace(fields[1]), 64)
	memUsed, _ := strconv.ParseUint(strings.TrimSpace(fields[2]), 10, 64)
	memTotal, _ := strconv.ParseUint(strings.TrimSpace(fields[3]), 10, 64)
	temp, _ := strconv.Atoi(strings.TrimSpace(fields[4]))

	return GPUMetrics{
		Name:        name,
		Utilization: util,
		VRAMUsed:    memUsed * 1024 * 1024,
		VRAMTotal:   memTotal * 1024 * 1024,
		Temperature: temp,
		Available:   true,
	}
}