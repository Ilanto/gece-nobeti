// Package linuxproc provides low-level access to Linux's /proc filesystem.
package linuxproc

import (
	"bufio"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// InterfaceStats holds per-interface network statistics.
type InterfaceStats struct {
	Name        string
	Type        string
	Status      string
	SpeedMbps   uint64
	InBytes     uint64
	OutBytes    uint64
	InPackets   uint64
	OutPackets  uint64
	InErrors    uint64
	OutErrors   uint64
	Address     string // "192.168.1.6/24" from `ip -4 addr show`
}

// CollectNetwork returns all network interface statistics.
func CollectNetwork() []InterfaceStats {
	var stats []InterfaceStats

	// Read /proc/net/dev for per-interface byte/packet/error counts
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return stats
	}
	defer f.Close()

	interfaces := make(map[string]InterfaceStats)

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		// Format: "  eth0:1234 5678 ..."
		line = strings.TrimSpace(line)
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		if name == "" || name == "face" || name == "lo" { // skip header and loopback for now
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 10 {
			continue
		}

		inBytes, _ := strconv.ParseUint(fields[0], 10, 64)
		inPackets, _ := strconv.ParseUint(fields[1], 10, 64)
		// fields[2] = in_errors
		// fields[3] = in_dropped
		outBytes, _ := strconv.ParseUint(fields[8], 10, 64)
		outPackets, _ := strconv.ParseUint(fields[9], 10, 64)

		speed := uint64(0)
		speedFile := "/sys/class/net/" + name + "/speed"
		if data := ReadFile(speedFile); data != "" {
			if v, err := strconv.ParseUint(strings.TrimSpace(data), 10, 64); err == nil {
				speed = v
			}
		}

		status := "up"
		operstateFile := "/sys/class/net/" + name + "/operstate"
		if data := ReadFile(operstateFile); data != "" {
			if strings.TrimSpace(data) == "down" {
				status = "down"
			}
		}

		ifaceType := "ethernet"
		if strings.HasPrefix(name, "wl") || strings.HasPrefix(name, "wifi") {
			ifaceType = "wifi"
		} else if strings.HasPrefix(name, "docker") || strings.HasPrefix(name, "br-") || strings.HasPrefix(name, "veth") {
			ifaceType = "virtual"
		} else if strings.HasPrefix(name, "lo") {
			ifaceType = "loopback"
		}

		interfaces[name] = InterfaceStats{
			Name:       name,
			Type:       ifaceType,
			Status:     status,
			SpeedMbps:  speed,
			InBytes:    inBytes,
			OutBytes:   outBytes,
			InPackets:  inPackets,
			OutPackets: outPackets,
			InErrors:   ParseUint(fields[2], 0),
			OutErrors:  ParseUint(fields[10], 0),
		}
	}

	// Add loopback
	if _, ok := interfaces["lo"]; !ok {
		interfaces["lo"] = InterfaceStats{Name: "lo", Type: "loopback", Status: "up", SpeedMbps: 0}
	}

	// IP addresses via `ip -4 addr show`
	if out, err := exec.Command("ip", "-4", "-o", "addr", "show").Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			parts := strings.Fields(line)
			if len(parts) >= 4 && parts[2] == "inet" {
				ifName := parts[1]
				if s, ok := interfaces[ifName]; ok {
					s.Address = parts[3]
					interfaces[ifName] = s
				}
			}
		}
	}

	for _, s := range interfaces {
		stats = append(stats, s)
	}
	return stats
}

// CollectNetworkTotals returns aggregated total bps.
func CollectNetworkTotals() (upBPS, downBPS uint64) {
	stats := CollectNetwork()
	for _, s := range stats {
		if s.Status == "down" {
			continue
		}
		// These are cumulative counters; delta is computed in the collector layer.
		// Here we just return raw values; caller does the delta.
		downBPS += s.InBytes
		upBPS += s.OutBytes
	}
	return
}