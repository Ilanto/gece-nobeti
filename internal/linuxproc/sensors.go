package linuxproc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// SensorReading holds one temperature sensor reading.
type SensorReading struct {
	Name     string
	Value    float64
	Unit     string
	Critical float64
}

// CollectSensors reads temperature sensors from /sys/class/hwmon.
func CollectSensors() []SensorReading {
	var readings []SensorReading

	for i := 0; ; i++ {
		base := fmt.Sprintf("/sys/class/hwmon/hwmon%d", i)
		if _, err := os.Stat(base); err != nil {
			break
		}

		chipName := "unknown"
		if b, err := os.ReadFile(base + "/name"); err == nil {
			chipName = strings.TrimSpace(string(b))
		}

		for j := 1; j <= 10; j++ {
			inputPath := fmt.Sprintf("%s/temp%d_input", base, j)
			b, err := os.ReadFile(inputPath)
			if err != nil {
				continue
			}
			raw, err := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
			if err != nil {
				continue
			}
			tempC := float64(raw) / 1000.0

			label := fmt.Sprintf("%s_temp%d", chipName, j)
			if lb, err := os.ReadFile(fmt.Sprintf("%s/temp%d_label", base, j)); err == nil {
				label = fmt.Sprintf("%s_%s", chipName, strings.TrimSpace(string(lb)))
			}

			var critC float64
			if cb, err := os.ReadFile(fmt.Sprintf("%s/temp%d_crit", base, j)); err == nil {
				if v, err := strconv.ParseInt(strings.TrimSpace(string(cb)), 10, 64); err == nil {
					critC = float64(v) / 1000.0
				}
			}
			if critC == 0 {
				if cb, err := os.ReadFile(fmt.Sprintf("%s/temp%d_max", base, j)); err == nil {
					if v, err := strconv.ParseInt(strings.TrimSpace(string(cb)), 10, 64); err == nil {
						critC = float64(v) / 1000.0
					}
				}
			}

			readings = append(readings, SensorReading{
				Name:     label,
				Value:    tempC,
				Unit:     "°C",
				Critical: critC,
			})
		}
	}

	return readings
}
