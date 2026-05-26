package linuxproc

import (
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
)

// SyslogEntry represents one journalctl log line.
type SyslogEntry struct {
	Timestamp int64
	Facility  string
	Severity  string
	Message   string
	Source    string
}

var facilityNames = map[int]string{
	0: "kern", 1: "user", 2: "mail", 3: "daemon",
	4: "auth", 5: "syslog", 6: "lpr", 7: "news",
	8: "uucp", 9: "cron", 10: "authpriv",
}

var severityNames = map[int]string{
	0: "crit", 1: "crit", 2: "crit", 3: "crit",
	4: "warn", 5: "warn",
	6: "info", 7: "info",
}

// CollectSyslog runs journalctl and returns the last n entries (newest first).
func CollectSyslog(n int) []SyslogEntry {
	if n <= 0 {
		n = 200
	}
	cmd := exec.Command("journalctl", "-n", strconv.Itoa(n), "-o", "json", "--no-pager", "-q")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var entries []SyslogEntry
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			continue
		}

		entry := SyslogEntry{}

		if v, ok := m["__REALTIME_TIMESTAMP"]; ok {
			var s string
			if err := json.Unmarshal(v, &s); err == nil {
				if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
					entry.Timestamp = ts / 1000 // µs → ms
				}
			}
		}

		if v, ok := m["MESSAGE"]; ok {
			var s string
			if err := json.Unmarshal(v, &s); err == nil {
				entry.Message = s
			}
		}
		if entry.Message == "" {
			continue
		}

		prio := 6
		if v, ok := m["PRIORITY"]; ok {
			var s string
			if err := json.Unmarshal(v, &s); err == nil {
				if p, err := strconv.Atoi(s); err == nil {
					prio = p
				}
			}
		}
		entry.Severity = severityNames[prio]
		if entry.Severity == "" {
			entry.Severity = "info"
		}

		fac := 3
		if v, ok := m["SYSLOG_FACILITY"]; ok {
			var s string
			if err := json.Unmarshal(v, &s); err == nil {
				if f, err := strconv.Atoi(s); err == nil {
					fac = f
				}
			}
		}
		if name, ok := facilityNames[fac]; ok {
			entry.Facility = name
		} else {
			entry.Facility = "daemon"
		}

		if v, ok := m["SYSLOG_IDENTIFIER"]; ok {
			var s string
			if err := json.Unmarshal(v, &s); err == nil {
				entry.Source = s
			}
		}

		entries = append(entries, entry)
	}

	// Reverse so newest is first
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	return entries
}
