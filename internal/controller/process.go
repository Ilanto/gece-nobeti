//go:build linux

package controller

import (
	"os"
	"strconv"
	"strings"
)

// GetProcessName returns the process name for a given PID.
// It first tries /proc/<pid>/comm, then falls back to /proc/<pid>/cmdline.
func GetProcessName(pid uint32) string {
	// Try /proc/<pid>/comm first (most reliable for process name)
	data, err := os.ReadFile("/proc/" + strconv.FormatUint(uint64(pid), 10) + "/comm")
	if err == nil && len(data) > 0 {
		return strings.TrimSpace(string(data))
	}

	// Fallback: read /proc/<pid>/cmdline
	data, err = os.ReadFile("/proc/" + strconv.FormatUint(uint64(pid), 10) + "/cmdline")
	if err != nil || len(data) == 0 {
		return ""
	}

	// cmdline is null-separated, extract first entry
	parts := strings.Split(strings.TrimSpace(string(data)), "\x00")
	if len(parts) > 0 && parts[0] != "" {
		// Extract basename from path
		if idx := strings.LastIndex(parts[0], "/"); idx >= 0 {
			return parts[0][idx+1:]
		}
		return parts[0]
	}

	return ""
}

// GetProcessComm returns just the comm name from /proc/<pid>/comm.
func GetProcessComm(pid uint32) string {
	data, err := os.ReadFile("/proc/" + strconv.FormatUint(uint64(pid), 10) + "/comm")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// GetProcessCmdline returns the full command line from /proc/<pid>/cmdline.
func GetProcessCmdline(pid uint32) string {
	data, err := os.ReadFile("/proc/" + strconv.FormatUint(uint64(pid), 10) + "/cmdline")
	if err != nil {
		return ""
	}
	// cmdline is null-separated, replace nulls with spaces
	return strings.Join(strings.Split(strings.TrimSpace(string(data)), "\x00"), " ")
}

// IsKernelThread checks if a process is a kernel thread by checking
// if /proc/<pid>/comm is empty or matches known kernel thread patterns.
// Kernel threads typically have comm names like: migration/, ksoftirqd/, watchdog/, etc.
func IsKernelThread(pid uint32) bool {
	comm := GetProcessComm(pid)
	if comm == "" {
		return false
	}

	// Kernel threads often have names ending with /
	// e.g., "migration/0", "ksoftirqd/0", "watchdog/1"
	if strings.HasSuffix(comm, "/") {
		return true
	}

	// Known kernel thread patterns
	kernelThreads := []string{
		"migration/", "ksoftirqd/", "watchdog/", "events/", "khelper/",
		"kthreadd", "sync", "kpsmoused", "deferwq", "charger_manager",
		"led-manager", "suspend", "kworker/",
	}

	for _, pattern := range kernelThreads {
		if strings.HasPrefix(comm, pattern) {
			return true
		}
	}

	return false
}