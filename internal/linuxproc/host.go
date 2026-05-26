package linuxproc

import (
	"bufio"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// HostInfo holds basic system identification data.
type HostInfo struct {
	Hostname      string
	KernelVersion string
	Arch          string
	OS            string
	UptimeSeconds float64
	User          string
	UID           uint32
	Shell         string
	TTY           string
}

// CollectHost reads system identity from /proc and /etc.
func CollectHost() HostInfo {
	info := HostInfo{}

	if b, err := os.ReadFile("/proc/sys/kernel/hostname"); err == nil {
		info.Hostname = strings.TrimSpace(string(b))
	}

	if b, err := os.ReadFile("/proc/version"); err == nil {
		parts := strings.Fields(string(b))
		if len(parts) >= 3 {
			info.KernelVersion = parts[2]
		}
	}

	if out, err := exec.Command("uname", "-m").Output(); err == nil {
		info.Arch = strings.TrimSpace(string(out))
	} else {
		info.Arch = "x86_64"
	}

	if b, err := os.ReadFile("/proc/uptime"); err == nil {
		parts := strings.Fields(string(b))
		if len(parts) >= 1 {
			if v, err := strconv.ParseFloat(parts[0], 64); err == nil {
				info.UptimeSeconds = v
			}
		}
	}

	if f, err := os.Open("/etc/os-release"); err == nil {
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				info.OS = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), `"`)
				break
			}
		}
	}

	if b, err := os.ReadFile("/proc/self/status"); err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			if strings.HasPrefix(line, "Uid:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					if v, err := strconv.ParseUint(parts[1], 10, 32); err == nil {
						info.UID = uint32(v)
					}
				}
			}
		}
	}

	if u := os.Getenv("USER"); u != "" {
		info.User = u
	} else if u := os.Getenv("LOGNAME"); u != "" {
		info.User = u
	} else {
		info.User = "user"
	}

	if s := os.Getenv("SHELL"); s != "" {
		info.Shell = s
	} else {
		info.Shell = "/bin/sh"
	}

	if t := os.Getenv("SSH_TTY"); t != "" {
		info.TTY = t
	} else if t := os.Getenv("TTY"); t != "" {
		info.TTY = t
	} else {
		info.TTY = "pts/0"
	}

	return info
}
