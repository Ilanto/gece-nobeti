// Package linuxproc provides low-level access to Linux's /proc filesystem.
package linuxproc

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// DriveInfo holds per-disk/partition information.
type DriveInfo struct {
	Letter     string // e.g. "sda1" or "nvme0n1p1"
	Label      string
	FSType     string
	TotalBytes uint64
	FreeBytes  uint64
	UsedBytes  uint64
	UsedPct    float64
	ReadBPS    uint64
	WriteBPS   uint64
}

// CollectDisk returns disk usage for all mounted filesystems.
func CollectDisk() []DriveInfo {
	var drives []DriveInfo

	// Use /proc/mounts to get mounted filesystems
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return drives
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		device := fields[0]
		mountPoint := fields[1]
		fsType := fields[2]

		// Skip虚拟文件系统
		if fsType == "proc" || fsType == "sysfs" || fsType == "devtmpfs" ||
			fsType == "devpts" || fsType == "tmpfs" || fsType == "securityfs" ||
			fsType == "cgroup" || fsType == "cgroup2" || fsType == "pstore" ||
			fsType == "debugfs" || fsType == "tracefs" || fsType == "autofs" ||
			fsType == "mqueue" || fsType == "hugetlbfs" || fsType == "configfs" ||
			fsType == "fusectl" || fsType == "binfmt_misc" || fsType == "overlay" {
			continue
		}

		letter := mountPoint
		if mountPoint == "/" {
			letter = "/"
		} else {
			// 取最后一个路径组件
			parts := strings.Split(mountPoint, "/")
			letter = parts[len(parts)-1]
			if letter == "" {
				letter = device
			}
		}

		// Read df -B1 output
		info := DriveInfo{
			Letter: letter,
			FSType: fsType,
		}

		// Get block device info from /sys/block/<dev>/stat
		// We'll skip detailed I/O stats for now and focus on capacity
		if device != "" && !strings.HasPrefix(device, "/dev/loop") {
			// Try to get size from /sys/class/block/<dev>/size
			devName := strings.TrimPrefix(device, "/dev/")
			sizeFile := "/sys/class/block/" + devName + "/size"
			if data := ReadFile(sizeFile); data != "" {
				if sectors, err := strconv.ParseUint(strings.TrimSpace(data), 10, 64); err == nil {
					info.TotalBytes = sectors * 512
				}
			}
		}

		drives = append(drives, info)
	}

	return drives
}

// CollectDiskUsage is a simpler wrapper that uses ReadFile/fs usage.
func CollectDiskUsage() []DriveInfo {
	return CollectDisk()
}

// CollectDiskStats reads /proc/diskstats for I/O counters.
func CollectDiskStats() map[string]DiskIO {
	stats := make(map[string]DiskIO)

	// Read /proc/diskstats
	f, err := os.Open("/proc/diskstats")
	if err != nil {
		return stats
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		fields := strings.Fields(line)
		if len(fields) < 14 {
			continue
		}

		// Field layout (12 fields after major/minor):
		//  3 - device name
		//  4 - reads completed
		//  5 - reads merged
		//  6 - sectors read
		//  7 - read time (ms)
		//  8 - writes completed
		//  9 - writes merged
		// 10 - sectors written
		// 11 - write time (ms)
		// 12 - I/O in progress
		// 13 - I/O time (ms)
		// 14 - weighted I/O time (ms)

		name := fields[2]
		reads, _ := strconv.ParseUint(fields[4], 10, 64)
		rbytes, _ := strconv.ParseUint(fields[5], 10, 64)
		writes, _ := strconv.ParseUint(fields[8], 10, 64)
		wbytes, _ := strconv.ParseUint(fields[9], 10, 64)

		stats[name] = DiskIO{
			Name:       name,
			ReadOps:    reads,
			WriteOps:   writes,
			ReadBytes:  rbytes * 512,
			WriteBytes: wbytes * 512,
		}
	}

	return stats
}

// DiskIO holds I/O statistics for a disk device.
type DiskIO struct {
	Name       string
	ReadOps    uint64
	WriteOps   uint64
	ReadBytes  uint64
	WriteBytes uint64
}