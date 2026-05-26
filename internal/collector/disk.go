package collector

import (
	"bufio"
	"os"
	"strings"
	"syscall"

	"github.com/burak/linux-dashboard/internal/linuxproc"
)

type DiskCollector struct{}

func NewDiskCollector() *DiskCollector {
	return &DiskCollector{}
}

func (d *DiskCollector) Collect() DiskMetrics {
	drives := linuxproc.CollectDisk()
	ioStats := linuxproc.CollectDiskStats()

	result := make([]DriveInfo, 0, len(drives))
	for _, drv := range drives {
		var readBPS, writeBPS, readIOPS, writeIOPS uint64
		for name, io := range ioStats {
			if strings.Contains(drv.Letter, name) || name == drv.Letter {
				readBPS = io.ReadBytes
				writeBPS = io.WriteBytes
				readIOPS = io.ReadOps
				writeIOPS = io.WriteOps
				break
			}
		}

		info := DriveInfo{
			Letter:     drv.Letter,
			Label:      drv.Label,
			FSType:     drv.FSType,
			TotalBytes: drv.TotalBytes,
			FreeBytes:  drv.FreeBytes,
			UsedBytes:  drv.UsedBytes,
			UsedPct:    drv.UsedPct,
			ReadBPS:    readBPS,
			WriteBPS:   writeBPS,
			ReadIOPS:   readIOPS,
			WriteIOPS:  writeIOPS,
		}

		result = append(result, info)
	}

	if len(result) == 0 {
		result = collectDiskFromMounts()
	}

	return DiskMetrics{Drives: result}
}

func collectDiskFromMounts() []DriveInfo {
	var drives []DriveInfo

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

		if fsType == "proc" || fsType == "sysfs" || fsType == "devtmpfs" ||
			fsType == "devpts" || fsType == "tmpfs" || fsType == "securityfs" ||
			fsType == "cgroup" || fsType == "cgroup2" || fsType == "pstore" ||
			fsType == "debugfs" || fsType == "tracefs" || fsType == "autofs" ||
			fsType == "mqueue" || fsType == "hugetlbfs" || fsType == "configfs" ||
			fsType == "fusectl" || fsType == "binfmt_misc" || fsType == "overlay" ||
			fsType == "nsfs" {
			continue
		}

		letter := mountPoint
		if mountPoint == "/" {
			letter = "/"
		} else {
			parts := strings.Split(mountPoint, "/")
			letter = parts[len(parts)-1]
			if letter == "" {
				letter = device
			}
		}

		total, free, used := statfsCapacity(mountPoint)

		var usedPct float64
		if total > 0 {
			usedPct = float64(used) / float64(total) * 100
		}

		drives = append(drives, DriveInfo{
			Letter:     letter,
			Label:      "",
			FSType:     fsType,
			TotalBytes: total,
			FreeBytes:  free,
			UsedBytes:  used,
			UsedPct:    usedPct,
		})
	}

	return drives
}

func statfsCapacity(path string) (total, free, used uint64) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0, 0, 0
	}
	total = stat.Blocks * uint64(stat.Bsize)
	free = stat.Bfree * uint64(stat.Bsize)
	used = total - free
	return
}