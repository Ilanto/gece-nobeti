// Package linuxproc provides low-level access to Linux's /proc filesystem.
package linuxproc

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// MemoryInfo holds memory statistics.
type MemoryInfo struct {
	Total       uint64
	Available   uint64
	Used        uint64
	UsedPercent float64
	Buffers     uint64
	Cached      uint64
	SwapTotal   uint64
	SwapFree    uint64
}

// Collect returns current memory usage snapshot.
func CollectMemory() MemoryInfo {
	info := MemoryInfo{}

	// Read /proc/meminfo
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return info
	}
	defer f.Close()

	m := make(map[string]uint64)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.TrimSuffix(val, " kB")
		val = strings.TrimSuffix(val, " k")
		val = strings.TrimSuffix(val, "KB")
		v, _ := strconv.ParseUint(val, 10, 64)
		m[key] = v * 1024 // kB to bytes
	}

	info.Total = m["MemTotal"]
	info.Buffers = m["Buffers"]
	info.Cached = m["Cached"]
	info.SwapTotal = m["SwapTotal"]
	info.SwapFree = m["SwapFree"]

	// Available = MemAvailable (kernel 3.14+) or MemFree - LowFree - Hidden
	if avail, ok := m["MemAvailable"]; ok {
		info.Available = avail
	} else {
		// Fallback
		info.Available = m["MemFree"]
	}

	info.Used = info.Total - info.Available
	if info.Total > 0 {
		info.UsedPercent = float64(info.Used) / float64(info.Total) * 100
	}

	return info
}