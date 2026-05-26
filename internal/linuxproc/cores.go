package linuxproc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CoreInfo holds per-CPU-core metadata.
type CoreInfo struct {
	ID        int
	FreqKHz   uint64
	Governor  string
	NumaNode  int
	Microcode string
}

// CollectCores reads per-core info from /sys/devices/system/cpu/cpu*.
func CollectCores() []CoreInfo {
	var cores []CoreInfo

	for i := 0; ; i++ {
		base := fmt.Sprintf("/sys/devices/system/cpu/cpu%d", i)
		if _, err := os.Stat(base); err != nil {
			break
		}
		core := CoreInfo{ID: i}

		if b, err := os.ReadFile(base + "/cpufreq/scaling_cur_freq"); err == nil {
			if v, err := strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64); err == nil {
				core.FreqKHz = v
			}
		}
		if core.FreqKHz == 0 {
			if b, err := os.ReadFile(base + "/cpufreq/cpuinfo_cur_freq"); err == nil {
				if v, err := strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64); err == nil {
					core.FreqKHz = v
				}
			}
		}

		if b, err := os.ReadFile(base + "/cpufreq/scaling_governor"); err == nil {
			core.Governor = strings.TrimSpace(string(b))
		}

		if entries, err := os.ReadDir(base); err == nil {
			for _, e := range entries {
				if strings.HasPrefix(e.Name(), "node") {
					if v, err := strconv.Atoi(e.Name()[4:]); err == nil {
						core.NumaNode = v
					}
				}
			}
		}

		if b, err := os.ReadFile("/sys/devices/system/cpu/cpu0/microcode/version"); err == nil {
			core.Microcode = strings.TrimSpace(string(b))
		}

		cores = append(cores, core)
	}

	return cores
}
