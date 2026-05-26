// Package linuxproc provides low-level access to Linux's /proc filesystem.
// All reads here are directly from /proc, /sys and syscalls — no external deps.
package linuxproc

import (
	"bufio"
	"os"
	"path"
	"strconv"
	"strings"
)

// ReadFile reads entire file and returns content. Returns "" if file doesn't exist.
func ReadFile(file string) string {
	data, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	return string(data)
}

// ReadDir returns filenames in a directory.
func ReadDir(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// ParseKeyValue reads a file with "key value" lines and returns map.
// Example: /proc/meminfo, /proc/stat
func ParseKeyValue(file string) map[string]uint64 {
	m := make(map[string]uint64)
	data := ReadFile(file)
	if data == "" {
		return m
	}
	sc := bufio.NewScanner(strings.NewReader(data))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Remove "kB" suffix and convert
		val = strings.TrimSuffix(val, " kB")
		val = strings.TrimSuffix(val, "KB")
		val = strings.TrimSuffix(val, " k")
		val = strings.TrimSuffix(val, "B")
		if v, err := strconv.ParseUint(val, 10, 64); err == nil {
			m[key] = v * 1024 // everything in meminfo is in kB
		}
	}
	return m
}

// ParseInt parses an int64 from a string, returns def if fail.
func ParseInt(s string, def int64) int64 {
	if v, err := strconv.ParseInt(s, 10, 64); err == nil {
		return v
	}
	return def
}

// ParseUint parses a uint64 from a string, returns def if fail.
func ParseUint(s string, def uint64) uint64 {
	if v, err := strconv.ParseUint(s, 10, 64); err == nil {
		return v
	}
	return def
}

// ParseFloat parses a float64 from a string, returns def if fail.
func ParseFloat(s string, def float64) float64 {
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v
	}
	return def
}

// ExePath returns symlink target of /proc/<pid>/exe.
func ExePath(pid int) string {
	return path.Join("/proc", strconv.Itoa(pid), "exe")
}

// ReadLink reads symlink and returns target string.
func ReadLink(file string) string {
	if target, err := os.Readlink(file); err == nil {
		return target
	}
	return ""
}