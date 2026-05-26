// Package linuxproc provides low-level access to Linux's /proc filesystem.
package linuxproc

import (
	"bufio"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

// ProcessInfo holds raw data for a single process.
type ProcessInfo struct {
	PID        uint32
	ParentPID  uint32
	Name       string
	State      string
	PPID       uint32
	Threads    uint32
	StartTime  int64
	UTime      uint64
	STime      uint64
	VMSize     uint64 // VM size in KB
	VMRSS      uint64 // resident set size in KB
	ExePath    string
	CWD        string
	Command    string
	UID        uint32 // real UID from /proc/<pid>/status
	CoreID     uint32 // last CPU core from /proc/<pid>/stat field 39
}

// AllProcesses returns a slice of all running processes.
func AllProcesses() ([]ProcessInfo, error) {
	const procDir = "/proc"

	entries, err := os.ReadDir(procDir)
	if err != nil {
		return nil, err
	}

	var procs []ProcessInfo
	var mu sync.Mutex

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if name[0] < '0' || name[0] > '9' {
			continue // not a PID
		}
		pid, err := strconv.ParseUint(name, 10, 64)
		if err != nil {
			continue
		}

		info := readProcessInfo(uint32(pid))
		if info.PID == 0 {
			continue
		}
		mu.Lock()
		procs = append(procs, info)
		mu.Unlock()
	}

	return procs, nil
}

func readProcessInfo(pid uint32) ProcessInfo {
	procDir := "/proc/" + strconv.FormatUint(uint64(pid), 10)

	// Read /proc/<pid>/stat
	statData := ReadFile(path.Join(procDir, "stat"))

	var info ProcessInfo
	info.PID = pid

	if statData == "" {
		return info
	}

	// Format: pid (comm) state ppid pgrp session tty_nr tpgid flags minflt cminflt majflt cmajflt utime stime
	// We need to extract fields carefully — comm can contain spaces and parens
	// Find the last ')' to split
	closeIdx := strings.LastIndex(statData, ")")
	if closeIdx == -1 {
		return info
	}

	after := strings.TrimSpace(statData[closeIdx+1:])
	fields := strings.Fields(after)
	if len(fields) < 17 {
		return info
	}

	// ppid is field 1 after ')'
	info.ParentPID = uint32(parseUintSafe(fields[1], 0))
	// state is field 2
	info.State = fields[2]
	// utime field 11 (index 10), stime field 12 (index 11)
	info.UTime = parseUintSafe(fields[10], 0)
	info.STime = parseUintSafe(fields[11], 0)
	// vsize field 19 (index 18)
	info.VMSize = parseUintSafe(fields[18], 0)
	// rss field 21 (index 20)
	info.VMRSS = parseUintSafe(fields[20], 0)
	// start time field 20 (index 19)
	info.StartTime = parseIntSafe(fields[19], 0)

	// Extract comm (between '(' and ')')
	openIdx := strings.LastIndex(statData, "(")
	if openIdx > 0 && closeIdx > openIdx {
		info.Name = statData[openIdx+1 : closeIdx]
	}

	// /proc/<pid>/comm
	if comm := ReadFile(path.Join(procDir, "comm")); comm != "" {
		info.Name = strings.TrimSpace(comm)
	}

	// exe symlink
	info.ExePath = ReadLink(path.Join(procDir, "exe"))
	// cwd symlink
	info.CWD = ReadLink(path.Join(procDir, "cwd"))

	// cmdline
	if cmdData := ReadFile(path.Join(procDir, "cmdline")); cmdData != "" {
		// Replace null bytes with spaces
		cmdData = strings.ReplaceAll(cmdData, "\x00", " ")
		info.Command = strings.TrimSpace(cmdData)
	}

	// Number of threads and UID from /proc/<pid>/status
	if statusData := ReadFile(path.Join(procDir, "status")); statusData != "" {
		sc := bufio.NewScanner(strings.NewReader(statusData))
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "Threads:") {
				f := strings.Fields(line)
				if len(f) >= 2 {
					t, _ := strconv.ParseUint(f[1], 10, 64)
					info.Threads = uint32(t)
				}
			}
			if strings.HasPrefix(line, "Uid:") {
				f := strings.Fields(line)
				if len(f) >= 2 {
					if v, err := strconv.ParseUint(f[1], 10, 32); err == nil {
						info.UID = uint32(v)
					}
				}
			}
		}
	}

	// CoreID: /proc/<pid>/stat field index 38 (0-based after ')') = processor
	if closeIdx >= 0 {
		after2 := strings.TrimSpace(statData[closeIdx+1:])
		fields2 := strings.Fields(after2)
		if len(fields2) >= 37 {
			if v, err := strconv.ParseUint(fields2[36], 10, 32); err == nil {
				info.CoreID = uint32(v)
			}
		}
	}

	return info
}

// PIDsWithName returns PIDs that match the given name.
func PIDsWithName(name string) []uint32 {
	procs, _ := AllProcesses()
	var pids []uint32
	for _, p := range procs {
		if p.Name == name {
			pids = append(pids, p.PID)
		}
	}
	return pids
}

func parseUintSafe(s string, def uint64) uint64 {
	if v, err := strconv.ParseUint(s, 10, 64); err == nil {
		return v
	}
	return def
}

func parseIntSafe(s string, def int64) int64 {
	if v, err := strconv.ParseInt(s, 10, 64); err == nil {
		return v
	}
	return def
}