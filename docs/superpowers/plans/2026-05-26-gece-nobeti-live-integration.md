# Gece Nöbeti Canlı Entegrasyon Uygulama Planı

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** linux-dashboard Go binary'sine eksik endpoint'leri ekle, Gece Nöbeti dashboard'unu binary içine göm — tek `./ldm` komutuyla gerçek sistem verisi gösteren dashboard çalışsın.

**Architecture:** Go backend `/proc`/`/sys`/`journalctl`'dan veri okuyup JSON döner. Frontend aynı origin'den REST çağrısı yapar, adapter fonksiyonları Go JSON'u React state'e dönüştürür. `go:embed` zaten mevcut — sadece `web/` klasörünü güncellemek yeterli.

**Tech Stack:** Go 1.21+, chi router, React 18 (CDN+Babel), JetBrains Mono/Barlow Condensed/Newsreader fonts

---

## Dosya Haritası

### Oluşturulacak
- `internal/linuxproc/host.go` — hostname, kernel, uptime, os, arch, user, tty
- `internal/linuxproc/cores.go` — per-core freq, governor, numa node
- `internal/linuxproc/sensors.go` — hwmon sıcaklıkları
- `internal/linuxproc/syslog.go` — journalctl JSON parser
- `internal/linuxproc/connections.go` — /proc/net/tcp parser + inode→pid
- `internal/collector/host.go` — HostInfo collector
- `internal/collector/cores.go` — CoreInfo collector
- `internal/collector/sensors.go` — SensorInfo collector
- `internal/server/web/live.js` — data.js yerine geçen adapter + fetch katmanı

### Değiştirilecek
- `internal/linuxproc/process.go` — UID ve CPU core ID alanları eklenir
- `internal/linuxproc/memory.go` — Cached + FreeMem alanları eklenir
- `internal/collector/types.go` — yeni struct'lar eklenir
- `internal/collector/memory.go` — Buffers/Cached/SwapUsed alanları doldurulur
- `internal/collector/process.go` — UID + CoreID alanları doldurulur
- `internal/collector/manager.go` — yeni collector'lar eklenir
- `internal/server/handlers.go` — yeni handler fonksiyonları eklenir
- `internal/server/router.go` — yeni rotalar eklenir
- `internal/server/web/index.html` — data.js → live.js, API base URL eklenir

### Taşınacak (~/Masaüstü/43/ → internal/server/web/)
- `app.jsx`, `panels.jsx`, `cores.jsx`, `surveillance.jsx`, `palette.jsx`, `tweaks-panel.jsx`

---

## Task 1: Memory — Buffers/Cached/SwapUsed alanları

`linuxproc.CollectMemory()` zaten bu değerleri okuyor ama collector bunları döndürmüyor.

**Files:**
- Modify: `internal/collector/types.go`
- Modify: `internal/collector/memory.go`

- [ ] **Adım 1: types.go — MemoryMetrics struct'ına alanlar ekle**

`internal/collector/types.go` dosyasında `MemoryMetrics` struct'ını bul ve şununla değiştir:

```go
type MemoryMetrics struct {
	TotalPhys     uint64  `json:"total_phys"`
	AvailPhys     uint64  `json:"avail_phys"`
	UsedPhys      uint64  `json:"used_phys"`
	FreePhys      uint64  `json:"free_phys"`
	Buffers       uint64  `json:"buffers"`
	Cached        uint64  `json:"cached"`
	UsedPercent   float64 `json:"used_percent"`
	TotalPageFile uint64  `json:"total_page_file"`
	AvailPageFile uint64  `json:"avail_page_file"`
	CommitCharge  uint64  `json:"commit_charge"`
	SwapUsed      uint64  `json:"swap_used"`
}
```

- [ ] **Adım 2: memory.go — collector yeni alanları doldursun**

`internal/collector/memory.go` dosyasındaki `Collect()` metodunu şununla değiştir:

```go
func (m *MemoryCollector) Collect() MemoryMetrics {
	info := linuxproc.CollectMemory()
	freePhys := info.Total - info.Used - info.Buffers - info.Cached
	if freePhys > info.Total {
		freePhys = 0
	}
	swapUsed := uint64(0)
	if info.SwapTotal > info.SwapFree {
		swapUsed = info.SwapTotal - info.SwapFree
	}
	return MemoryMetrics{
		TotalPhys:     info.Total,
		AvailPhys:     info.Available,
		UsedPhys:      info.Used,
		FreePhys:      freePhys,
		Buffers:       info.Buffers,
		Cached:        info.Cached,
		UsedPercent:   info.UsedPercent,
		TotalPageFile: info.Total + info.SwapTotal,
		AvailPageFile: info.Available + info.SwapFree,
		CommitCharge:  0,
		SwapUsed:      swapUsed,
	}
}
```

- [ ] **Adım 3: Derleme testi**

```bash
cd ~/linux-dashboard && go build ./...
```

Beklenen: hata yok

- [ ] **Adım 4: Commit**

```bash
cd ~/linux-dashboard
git add internal/collector/types.go internal/collector/memory.go
git commit -m "feat: memory metrics — add buffers/cached/swap_used fields"
```

---

## Task 2: Process — UID ve CPU Core ID eklentisi

`/proc/<pid>/status`'tan UID, `/proc/<pid>/stat` alanı 39'dan (processor) CPU core ID okunur.

**Files:**
- Modify: `internal/linuxproc/process.go`
- Modify: `internal/collector/types.go`
- Modify: `internal/collector/process.go`

- [ ] **Adım 1: linuxproc/process.go — ProcessInfo struct'ına UID ve CoreID ekle**

`internal/linuxproc/process.go` dosyasında `type ProcessInfo struct` bloğuna şu iki alanı ekle:

```go
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
	VMSize     uint64
	VMRSS      uint64
	ExePath    string
	CWD        string
	Command    string
	UID        uint32   // ← YENİ: /proc/<pid>/status'tan
	CoreID     uint32   // ← YENİ: /proc/<pid>/stat alanı 39
}
```

- [ ] **Adım 2: linuxproc/process.go — readProcessInfo'ya UID ve CoreID okuma ekle**

`readProcessInfo` fonksiyonunun `info.Name = ...` satırından sonrasına şunu ekle:

```go
	// UID: /proc/<pid>/status → "Uid:\t<real> ..."
	if statusData := ReadFile(path.Join(procDir, "status")); statusData != "" {
		for _, line := range strings.Split(statusData, "\n") {
			if strings.HasPrefix(line, "Uid:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					if v, err := strconv.ParseUint(parts[1], 10, 32); err == nil {
						info.UID = uint32(v)
					}
				}
				break
			}
		}
	}

	// CoreID: /proc/<pid>/stat field index 38 (0-based) = processor
	// statData already parsed above; re-parse fields after ')'
	if closeIdx >= 0 {
		after2 := strings.TrimSpace(statData[closeIdx+1:])
		fields2 := strings.Fields(after2)
		if len(fields2) >= 37 {
			if v, err := strconv.ParseUint(fields2[36], 10, 32); err == nil {
				info.CoreID = uint32(v)
			}
		}
	}
```

- [ ] **Adım 3: collector/types.go — ProcessInfo struct'ına UID ve CoreID ekle**

`internal/collector/types.go` dosyasında `type ProcessInfo struct` bloğunu bul ve şu iki JSON alanını ekle:

```go
type ProcessInfo struct {
	PID           uint32  `json:"pid"`
	ParentPID     uint32  `json:"parent_pid"`
	Name          string  `json:"name"`
	ExePath       string  `json:"exe_path"`
	CPUPercent    float64 `json:"cpu_percent"`
	WorkingSet    uint64  `json:"working_set"`
	PrivateBytes  uint64  `json:"private_bytes"`
	PageFaults    uint32  `json:"page_faults"`
	IOReadBytes   uint64  `json:"io_read_bytes"`
	IOWriteBytes  uint64  `json:"io_write_bytes"`
	IOReadOps     uint64  `json:"io_read_ops"`
	IOWriteOps    uint64  `json:"io_write_ops"`
	ThreadCount   uint32  `json:"thread_count"`
	CreateTime    int64   `json:"create_time"`
	IsCritical    bool    `json:"is_critical"`
	Status        string  `json:"status"`
	Connections   int     `json:"connections"`
	PriorityClass uint32  `json:"priority_class"`
	UID           uint32  `json:"uid"`      // ← YENİ
	CoreID        uint32  `json:"core_id"`  // ← YENİ
}
```

- [ ] **Adım 4: collector/process.go — UID ve CoreID alanlarını doldur**

`result = append(result, ProcessInfo{...})` çağrısına şu iki satırı ekle:

```go
		result = append(result, ProcessInfo{
			PID:         lp.PID,
			ParentPID:   lp.ParentPID,
			Name:        lp.Name,
			ExePath:     lp.ExePath,
			CPUPercent:  cpuPercent,
			WorkingSet:  lp.VMRSS * 1024,
			PrivateBytes: lp.VMSize * 1024,
			PageFaults:   0,
			IOReadBytes:  0,
			IOWriteBytes: 0,
			IOReadOps:    0,
			IOWriteOps:   0,
			ThreadCount:  lp.Threads,
			CreateTime:   lp.StartTime,
			IsCritical:   false,
			Status:       status,
			Connections:  0,
			PriorityClass: 0,
			UID:          lp.UID,     // ← YENİ
			CoreID:       lp.CoreID,  // ← YENİ
		})
```

- [ ] **Adım 5: Derleme testi**

```bash
cd ~/linux-dashboard && go build ./...
```

Beklenen: hata yok

- [ ] **Adım 6: Commit**

```bash
cd ~/linux-dashboard
git add internal/linuxproc/process.go internal/collector/types.go internal/collector/process.go
git commit -m "feat: process — add uid and core_id fields from /proc"
```

---

## Task 3: HostInfo — /proc + /etc/os-release okuyucu

**Files:**
- Create: `internal/linuxproc/host.go`
- Create: `internal/collector/host.go`
- Modify: `internal/collector/types.go`
- Modify: `internal/collector/manager.go`
- Modify: `internal/server/handlers.go`
- Modify: `internal/server/router.go`

- [ ] **Adım 1: internal/linuxproc/host.go oluştur**

```go
package linuxproc

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// HostInfo holds basic system identification data.
type HostInfo struct {
	Hostname       string
	KernelVersion  string
	Arch           string
	OS             string
	UptimeSeconds  float64
	User           string
	UID            uint32
	Shell          string
	TTY            string
}

// CollectHost reads system identity from /proc and /etc.
func CollectHost() HostInfo {
	info := HostInfo{}

	// hostname
	if b, err := os.ReadFile("/proc/sys/kernel/hostname"); err == nil {
		info.Hostname = strings.TrimSpace(string(b))
	}

	// kernel version
	if b, err := os.ReadFile("/proc/version"); err == nil {
		parts := strings.Fields(string(b))
		if len(parts) >= 3 {
			info.KernelVersion = parts[2]
		}
	}

	// arch
	if b, err := os.ReadFile("/proc/sys/kernel/arch"); err == nil {
		info.Arch = strings.TrimSpace(string(b))
	} else {
		info.Arch = "x86_64" // fallback
	}

	// uptime
	if b, err := os.ReadFile("/proc/uptime"); err == nil {
		parts := strings.Fields(string(b))
		if len(parts) >= 1 {
			if v, err := strconv.ParseFloat(parts[0], 64); err == nil {
				info.UptimeSeconds = v
			}
		}
	}

	// OS from /etc/os-release
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

	// current user from /proc/self/status
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
			if strings.HasPrefix(line, "Name:") {
				// /proc/self is the current process (ldm), not the user
				// use environment variable instead
			}
		}
	}

	// user from environment
	if u := os.Getenv("USER"); u != "" {
		info.User = u
	} else if u := os.Getenv("LOGNAME"); u != "" {
		info.User = u
	} else {
		info.User = "user"
	}

	// shell from environment
	if s := os.Getenv("SHELL"); s != "" {
		info.Shell = s
	} else {
		info.Shell = "/bin/sh"
	}

	// TTY from environment
	if t := os.Getenv("SSH_TTY"); t != "" {
		info.TTY = t
	} else if t := os.Getenv("TTY"); t != "" {
		info.TTY = t
	} else {
		info.TTY = "pts/0"
	}

	return info
}
```

- [ ] **Adım 2: collector/types.go — HostMetrics struct ekle**

`internal/collector/types.go` dosyasına `SystemSnapshot` struct'ından önce şunu ekle:

```go
// HostMetrics holds basic system identification.
type HostMetrics struct {
	Hostname      string  `json:"hostname"`
	KernelVersion string  `json:"kernel_version"`
	Arch          string  `json:"arch"`
	OS            string  `json:"os"`
	UptimeSeconds float64 `json:"uptime_seconds"`
	User          string  `json:"user"`
	UID           uint32  `json:"uid"`
	Shell         string  `json:"shell"`
	TTY           string  `json:"tty"`
}
```

- [ ] **Adım 3: internal/collector/host.go oluştur**

```go
package collector

import "github.com/burak/linux-dashboard/internal/linuxproc"

// HostCollector reads static system identity info.
type HostCollector struct{}

func NewHostCollector() *HostCollector { return &HostCollector{} }

func (h *HostCollector) Collect() HostMetrics {
	info := linuxproc.CollectHost()
	return HostMetrics{
		Hostname:      info.Hostname,
		KernelVersion: info.KernelVersion,
		Arch:          info.Arch,
		OS:            info.OS,
		UptimeSeconds: info.UptimeSeconds,
		User:          info.User,
		UID:           info.UID,
		Shell:         info.Shell,
		TTY:           info.TTY,
	}
}
```

- [ ] **Adım 4: server/handlers.go — handleHost ekle**

`internal/server/handlers.go` dosyasına `handleSystem` fonksiyonundan önce şunu ekle:

```go
func handleHost(col *collector.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, col.LatestHost())
	}
}
```

- [ ] **Adım 5: collector/manager.go — host collector ekle**

`Manager` struct'ına `host *HostCollector` alanı ekle:

```go
type Manager struct {
	cpu       *CPUCollector
	memory    *MemoryCollector
	disk      *DiskCollector
	network   *NetworkCollector
	gpu       *GPUCollector
	processes *ProcessCollector
	ports     *PortCollector
	host      *HostCollector  // ← YENİ
	// ...
}
```

`NewManager` fonksiyonuna `host: NewHostCollector(),` satırını ekle.

`Manager`'a yeni metod ekle:

```go
func (m *Manager) LatestHost() HostMetrics {
	return m.host.Collect()
}
```

- [ ] **Adım 6: server/router.go — /api/v1/host rotası ekle**

`r.Get("/api/v1/system", ...)` satırından hemen sonrasına ekle:

```go
r.Get("/api/v1/host", handleHost(col))
```

- [ ] **Adım 7: Derleme testi**

```bash
cd ~/linux-dashboard && go build ./...
```

Beklenen: hata yok

- [ ] **Adım 8: Manuel test**

```bash
cd ~/linux-dashboard && ./ldm &
sleep 2 && curl -s http://localhost:19876/api/v1/host | python3 -m json.tool
kill %1
```

Beklenen çıktı (örnek):
```json
{
  "hostname": "MayAta",
  "kernel_version": "7.0.9-1-cachyos",
  "arch": "x86_64",
  "os": "CachyOS",
  "uptime_seconds": 23000.0,
  "user": "mayata",
  "uid": 1000,
  "shell": "/usr/bin/zsh",
  "tty": "pts/0"
}
```

- [ ] **Adım 9: Commit**

```bash
cd ~/linux-dashboard
git add internal/linuxproc/host.go internal/collector/host.go \
        internal/collector/types.go internal/collector/manager.go \
        internal/server/handlers.go internal/server/router.go
git commit -m "feat: /api/v1/host — system identity from /proc and /etc"
```

---

## Task 4: CoreInfo — /sys/devices/system/cpu/ okuyucu

**Files:**
- Create: `internal/linuxproc/cores.go`
- Create: `internal/collector/cores.go`
- Modify: `internal/collector/types.go`
- Modify: `internal/collector/manager.go`
- Modify: `internal/server/handlers.go`
- Modify: `internal/server/router.go`

- [ ] **Adım 1: internal/linuxproc/cores.go oluştur**

```go
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

		// Frequency in kHz
		if b, err := os.ReadFile(base + "/cpufreq/scaling_cur_freq"); err == nil {
			if v, err := strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64); err == nil {
				core.FreqKHz = v
			}
		}
		// Fallback: cpuinfo_cur_freq
		if core.FreqKHz == 0 {
			if b, err := os.ReadFile(base + "/cpufreq/cpuinfo_cur_freq"); err == nil {
				if v, err := strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64); err == nil {
					core.FreqKHz = v
				}
			}
		}

		// Governor
		if b, err := os.ReadFile(base + "/cpufreq/scaling_governor"); err == nil {
			core.Governor = strings.TrimSpace(string(b))
		}

		// NUMA node
		numaGlob := base + "/node*"
		if entries, err := os.ReadDir(base); err == nil {
			for _, e := range entries {
				if strings.HasPrefix(e.Name(), "node") {
					if v, err := strconv.Atoi(e.Name()[4:]); err == nil {
						core.NumaNode = v
					}
				}
			}
		}
		_ = numaGlob

		// Microcode
		if b, err := os.ReadFile("/sys/devices/system/cpu/cpu0/microcode/version"); err == nil {
			core.Microcode = strings.TrimSpace(string(b))
		}

		cores = append(cores, core)
	}

	return cores
}
```

- [ ] **Adım 2: collector/types.go — CoreMetrics struct ekle**

```go
// CoreMetrics holds per-CPU-core info.
type CoreMetrics struct {
	ID       int    `json:"id"`
	FreqMHz  float64 `json:"freq_mhz"`
	Governor string  `json:"governor"`
	NumaNode int    `json:"numa_node"`
	Microcode string `json:"microcode"`
}
```

- [ ] **Adım 3: internal/collector/cores.go oluştur**

```go
package collector

import "github.com/burak/linux-dashboard/internal/linuxproc"

type CoresCollector struct{}

func NewCoresCollector() *CoresCollector { return &CoresCollector{} }

func (c *CoresCollector) Collect() []CoreMetrics {
	raw := linuxproc.CollectCores()
	out := make([]CoreMetrics, len(raw))
	for i, r := range raw {
		out[i] = CoreMetrics{
			ID:        r.ID,
			FreqMHz:   float64(r.FreqKHz) / 1000.0,
			Governor:  r.Governor,
			NumaNode:  r.NumaNode,
			Microcode: r.Microcode,
		}
	}
	return out
}
```

- [ ] **Adım 4: server/handlers.go — handleCores ekle**

```go
func handleCores(col *collector.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, col.LatestCores())
	}
}
```

- [ ] **Adım 5: collector/manager.go — cores collector ekle**

`Manager` struct'ına `cores *CoresCollector` ekle.  
`NewManager`'a `cores: NewCoresCollector()` ekle.  
`Manager`'a metod ekle:

```go
func (m *Manager) LatestCores() []CoreMetrics {
	return m.cores.Collect()
}
```

- [ ] **Adım 6: router.go — /api/v1/cores rotası ekle**

```go
r.Get("/api/v1/cores", handleCores(col))
```

- [ ] **Adım 7: Test + commit**

```bash
cd ~/linux-dashboard && go build ./... && ./ldm &
sleep 2 && curl -s http://localhost:19876/api/v1/cores | python3 -m json.tool | head -30
kill %1
```

Beklenen: JSON dizi, her eleman `{"id":0,"freq_mhz":3738.9,"governor":"powersave",...}` formatında

```bash
git add internal/linuxproc/cores.go internal/collector/cores.go \
        internal/collector/types.go internal/collector/manager.go \
        internal/server/handlers.go internal/server/router.go
git commit -m "feat: /api/v1/cores — per-core freq and governor from /sys"
```

---

## Task 5: SensorInfo — /sys/class/hwmon okuyucu

**Files:**
- Create: `internal/linuxproc/sensors.go`
- Create: `internal/collector/sensors.go`
- Modify: `internal/collector/types.go`
- Modify: `internal/collector/manager.go`
- Modify: `internal/server/handlers.go`
- Modify: `internal/server/router.go`

- [ ] **Adım 1: internal/linuxproc/sensors.go oluştur**

```go
package linuxproc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// SensorReading holds one temperature or fan sensor reading.
type SensorReading struct {
	Name     string
	Value    float64 // °C for temp, rpm for fan
	Unit     string
	Critical float64 // 0 if not applicable
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

		// Read temp1..temp10
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

			// Label
			label := fmt.Sprintf("%s_temp%d", chipName, j)
			if lb, err := os.ReadFile(fmt.Sprintf("%s/temp%d_label", base, j)); err == nil {
				label = fmt.Sprintf("%s_%s", chipName, strings.TrimSpace(string(lb)))
			}

			// Critical threshold
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
```

- [ ] **Adım 2: collector/types.go — SensorMetric struct ekle**

```go
// SensorMetric holds one sensor reading.
type SensorMetric struct {
	Name     string  `json:"name"`
	Value    float64 `json:"value"`
	Unit     string  `json:"unit"`
	Critical float64 `json:"critical"`
}
```

- [ ] **Adım 3: internal/collector/sensors.go oluştur**

```go
package collector

import "github.com/burak/linux-dashboard/internal/linuxproc"

type SensorsCollector struct{}

func NewSensorsCollector() *SensorsCollector { return &SensorsCollector{} }

func (s *SensorsCollector) Collect() []SensorMetric {
	raw := linuxproc.CollectSensors()
	out := make([]SensorMetric, len(raw))
	for i, r := range raw {
		out[i] = SensorMetric{
			Name:     r.Name,
			Value:    r.Value,
			Unit:     r.Unit,
			Critical: r.Critical,
		}
	}
	return out
}
```

- [ ] **Adım 4: server/handlers.go — handleSensors ekle**

```go
func handleSensors(col *collector.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, col.LatestSensors())
	}
}
```

- [ ] **Adım 5: collector/manager.go — sensors collector ekle**

`Manager` struct'ına `sensors *SensorsCollector` ekle.  
`NewManager`'a `sensors: NewSensorsCollector()` ekle.  
Metod ekle:

```go
func (m *Manager) LatestSensors() []SensorMetric {
	return m.sensors.Collect()
}
```

- [ ] **Adım 6: router.go — /api/v1/sensors rotası ekle**

```go
r.Get("/api/v1/sensors", handleSensors(col))
```

- [ ] **Adım 7: Test + commit**

```bash
cd ~/linux-dashboard && go build ./... && ./ldm &
sleep 2 && curl -s http://localhost:19876/api/v1/sensors | python3 -m json.tool
kill %1
```

Beklenen: `[{"name":"nvme_Composite","value":44.85,"unit":"°C","critical":84.85}, ...]`

```bash
git add internal/linuxproc/sensors.go internal/collector/sensors.go \
        internal/collector/types.go internal/collector/manager.go \
        internal/server/handlers.go internal/server/router.go
git commit -m "feat: /api/v1/sensors — hwmon temperatures from /sys"
```

---

## Task 6: SyslogEntry — journalctl okuyucu

**Files:**
- Create: `internal/linuxproc/syslog.go`
- Modify: `internal/collector/types.go`
- Modify: `internal/server/handlers.go`
- Modify: `internal/server/router.go`

- [ ] **Adım 1: internal/linuxproc/syslog.go oluştur**

```go
package linuxproc

import (
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
)

// SyslogEntry represents one journalctl log line.
type SyslogEntry struct {
	Timestamp int64  // Unix miliseconds
	Facility  string
	Severity  string // "info" | "warn" | "crit"
	Message   string
	Source    string // SYSLOG_IDENTIFIER
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

// CollectSyslog runs journalctl and returns the last n entries.
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

		// Timestamp: __REALTIME_TIMESTAMP in microseconds
		if v, ok := m["__REALTIME_TIMESTAMP"]; ok {
			var s string
			if err := json.Unmarshal(v, &s); err == nil {
				if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
					entry.Timestamp = ts / 1000 // µs → ms
				}
			}
		}

		// Message
		if v, ok := m["MESSAGE"]; ok {
			var s string
			if err := json.Unmarshal(v, &s); err == nil {
				entry.Message = s
			}
		}
		if entry.Message == "" {
			continue
		}

		// Priority → severity
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

		// Facility
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

		// Source identifier
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
```

- [ ] **Adım 2: collector/types.go — SyslogMetric struct ekle**

```go
// SyslogMetric represents one log entry for the API.
type SyslogMetric struct {
	Timestamp int64  `json:"timestamp"` // Unix ms
	Facility  string `json:"facility"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Source    string `json:"source"`
}
```

- [ ] **Adım 3: server/handlers.go — handleSyslog ekle**

```go
func handleSyslog(col *collector.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// journalctl is called directly — no collector caching needed
		raw := col.FetchSyslog(200)
		writeJSON(w, http.StatusOK, map[string]any{"entries": raw})
	}
}
```

- [ ] **Adım 4: collector/manager.go — FetchSyslog metodu ekle**

```go
func (m *Manager) FetchSyslog(n int) []SyslogMetric {
	raw := linuxproc.CollectSyslog(n)
	out := make([]SyslogMetric, len(raw))
	for i, r := range raw {
		out[i] = SyslogMetric{
			Timestamp: r.Timestamp,
			Facility:  r.Facility,
			Severity:  r.Severity,
			Message:   r.Message,
			Source:    r.Source,
		}
	}
	return out
}
```

`manager.go` başına `linuxproc` importunu ekle:
```go
import (
    "sync"
    "time"
    "github.com/burak/linux-dashboard/internal/event"
    "github.com/burak/linux-dashboard/internal/linuxproc"
)
```

- [ ] **Adım 5: router.go — /api/v1/syslog rotası ekle**

```go
r.Get("/api/v1/syslog", handleSyslog(col))
```

- [ ] **Adım 6: Test + commit**

```bash
cd ~/linux-dashboard && go build ./... && ./ldm &
sleep 2 && curl -s http://localhost:19876/api/v1/syslog | python3 -m json.tool | head -40
kill %1
```

Beklenen: `{"entries":[{"timestamp":1779823..., "facility":"kern","severity":"warn","message":"..."},...]}` 

```bash
git add internal/linuxproc/syslog.go internal/collector/types.go \
        internal/collector/manager.go internal/server/handlers.go \
        internal/server/router.go
git commit -m "feat: /api/v1/syslog — journalctl JSON reader"
```

---

## Task 7: ConnectionEntry — /proc/net/tcp* okuyucu

**Files:**
- Create: `internal/linuxproc/connections.go`
- Modify: `internal/collector/types.go`
- Modify: `internal/server/handlers.go`
- Modify: `internal/server/router.go`
- Modify: `internal/collector/manager.go`

- [ ] **Adım 1: internal/linuxproc/connections.go oluştur**

```go
package linuxproc

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// ConnectionEntry represents one TCP/UDP connection.
type ConnectionEntry struct {
	Protocol   string
	LocalAddr  string
	LocalPort  uint16
	RemoteAddr string
	RemotePort uint16
	State      string
	PID        uint32
	Process    string
}

var tcpStates = map[string]string{
	"01": "ESTABLISHED", "02": "SYN_SENT", "03": "SYN_RECV",
	"04": "FIN_WAIT1", "05": "FIN_WAIT2", "06": "TIME_WAIT",
	"07": "CLOSE", "08": "CLOSE_WAIT", "09": "LAST_ACK",
	"0A": "LISTEN", "0B": "CLOSING",
}

func hexToIPv4(h string) string {
	b, err := hex.DecodeString(h)
	if err != nil || len(b) != 4 {
		return "0.0.0.0"
	}
	// /proc/net/tcp stores little-endian
	ip := binary.LittleEndian.Uint32(b)
	return fmt.Sprintf("%d.%d.%d.%d", ip&0xff, (ip>>8)&0xff, (ip>>16)&0xff, (ip>>24)&0xff)
}

func hexToPort(h string) uint16 {
	v, _ := strconv.ParseUint(h, 16, 16)
	return uint16(v)
}

// buildInodeMap maps socket inodes to PIDs and process names.
func buildInodeMap() map[string][2]string {
	m := make(map[string][2]string)
	procDir, _ := filepath.Glob("/proc/[0-9]*/fd/*")
	for _, fdPath := range procDir {
		link, err := os.Readlink(fdPath)
		if err != nil {
			continue
		}
		if !strings.HasPrefix(link, "socket:[") {
			continue
		}
		inode := link[8 : len(link)-1]
		// extract PID from /proc/<pid>/fd/<n>
		parts := strings.Split(fdPath, "/")
		if len(parts) < 3 {
			continue
		}
		pid := parts[2]
		name := ""
		if b, err := os.ReadFile("/proc/" + pid + "/comm"); err == nil {
			name = strings.TrimSpace(string(b))
		}
		m[inode] = [2]string{pid, name}
	}
	return m
}

func parseNetFile(path, proto string, inodeMap map[string][2]string) []ConnectionEntry {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var entries []ConnectionEntry
	sc := bufio.NewScanner(f)
	sc.Scan() // skip header

	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 10 {
			continue
		}

		localParts := strings.Split(fields[1], ":")
		remoteParts := strings.Split(fields[2], ":")
		if len(localParts) != 2 || len(remoteParts) != 2 {
			continue
		}

		stateHex := strings.ToUpper(fields[3])
		state := tcpStates[stateHex]
		if state == "" {
			state = stateHex
		}

		inode := fields[9]
		pid := uint32(0)
		procName := ""
		if info, ok := inodeMap[inode]; ok {
			if v, err := strconv.ParseUint(info[0], 10, 32); err == nil {
				pid = uint32(v)
			}
			procName = info[1]
		}

		localIP := hexToIPv4(localParts[0])
		remoteIP := hexToIPv4(remoteParts[0])

		// For ipv6 files the hex is 32 chars; use net package
		if len(localParts[0]) == 32 {
			b, _ := hex.DecodeString(localParts[0])
			localIP = net.IP(b).String()
			b2, _ := hex.DecodeString(remoteParts[0])
			remoteIP = net.IP(b2).String()
		}

		entry := ConnectionEntry{
			Protocol:   proto,
			LocalAddr:  localIP,
			LocalPort:  hexToPort(localParts[1]),
			RemoteAddr: remoteIP,
			RemotePort: hexToPort(remoteParts[1]),
			State:      state,
			PID:        pid,
			Process:    procName,
		}
		entries = append(entries, entry)
	}
	return entries
}

// CollectConnections reads all TCP/UDP connections from /proc/net.
func CollectConnections() []ConnectionEntry {
	inodeMap := buildInodeMap()
	var all []ConnectionEntry
	all = append(all, parseNetFile("/proc/net/tcp",  "tcp",  inodeMap)...)
	all = append(all, parseNetFile("/proc/net/tcp6", "tcp6", inodeMap)...)
	all = append(all, parseNetFile("/proc/net/udp",  "udp",  inodeMap)...)
	// Suppress unused import if net is not used elsewhere
	_ = syscall.AF_INET
	return all
}
```

Not: `syscall` importunu kullanmıyorsak kaldır. `net` import için `net.IP` kullanımı yeterli.

Düzeltilmiş import bloğu (`syscall` olmadan):

```go
import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)
```

- [ ] **Adım 2: collector/types.go — ConnectionMetric struct ekle**

```go
// ConnectionMetric represents one network connection.
type ConnectionMetric struct {
	Protocol   string `json:"protocol"`
	LocalAddr  string `json:"local_addr"`
	LocalPort  uint16 `json:"local_port"`
	RemoteAddr string `json:"remote_addr"`
	RemotePort uint16 `json:"remote_port"`
	State      string `json:"state"`
	PID        uint32 `json:"pid"`
	Process    string `json:"process"`
}
```

- [ ] **Adım 3: server/handlers.go — handleConnections ekle**

```go
func handleConnections(col *collector.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, col.FetchConnections())
	}
}
```

- [ ] **Adım 4: collector/manager.go — FetchConnections metodu ekle**

```go
func (m *Manager) FetchConnections() []ConnectionMetric {
	raw := linuxproc.CollectConnections()
	out := make([]ConnectionMetric, len(raw))
	for i, r := range raw {
		out[i] = ConnectionMetric{
			Protocol:   r.Protocol,
			LocalAddr:  r.LocalAddr,
			LocalPort:  r.LocalPort,
			RemoteAddr: r.RemoteAddr,
			RemotePort: r.RemotePort,
			State:      r.State,
			PID:        r.PID,
			Process:    r.Process,
		}
	}
	return out
}
```

- [ ] **Adım 5: router.go — /api/v1/connections rotası ekle**

```go
r.Get("/api/v1/connections", handleConnections(col))
```

- [ ] **Adım 6: Test + commit**

```bash
cd ~/linux-dashboard && go build ./... && ./ldm &
sleep 2 && curl -s http://localhost:19876/api/v1/connections | python3 -m json.tool | head -40
kill %1
```

Beklenen: `[{"protocol":"tcp","local_addr":"127.0.0.1","local_port":11434,"state":"LISTEN","process":"ollama",...},...]`

```bash
git add internal/linuxproc/connections.go internal/collector/types.go \
        internal/collector/manager.go internal/server/handlers.go \
        internal/server/router.go
git commit -m "feat: /api/v1/connections — /proc/net/tcp parser with inode→pid"
```

---

## Task 8: Network arayüzlerine IP adresi ekle

`/api/v1/network` şu an IP adresi döndürmüyor. Dashboard için gerekli.

**Files:**
- Modify: `internal/linuxproc/network.go`
- Modify: `internal/collector/types.go`
- Modify: `internal/collector/network.go`

- [ ] **Adım 1: linuxproc/network.go — InterfaceStats struct'ına Address alanı ekle**

`InterfaceStats` struct'ına şu alanı ekle:

```go
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
	Address     string  // ← YENİ: "192.168.1.6/24" formatında
}
```

- [ ] **Adım 2: linuxproc/network.go — IP adresini `ip` komutuyla oku**

`CollectNetwork()` fonksiyonunun return'ünden önce şunu ekle:

```go
	// IP addresses via `ip -4 addr show`
	cmd := exec.Command("ip", "-4", "-o", "addr", "show")
	if out, err := cmd.Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			parts := strings.Fields(line)
			if len(parts) < 4 {
				continue
			}
			ifName := parts[1]
			if parts[2] == "inet" && len(parts) >= 4 {
				if s, ok := interfaces[ifName]; ok {
					s.Address = parts[3]
					interfaces[ifName] = s
				}
			}
		}
	}
```

`import "os/exec"` ekle (veya mevcut import bloğuna dahil et).

- [ ] **Adım 3: collector/types.go — InterfaceInfo'ya Address ekle**

```go
type InterfaceInfo struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	SpeedMbps uint64 `json:"speed_mbps"`
	InBPS     uint64 `json:"in_bps"`
	OutBPS    uint64 `json:"out_bps"`
	InPPS     uint64 `json:"in_pps"`
	OutPPS    uint64 `json:"out_pps"`
	InErrors  uint64 `json:"in_errors"`
	OutErrors uint64 `json:"out_errors"`
	Address   string `json:"address"`  // ← YENİ
}
```

- [ ] **Adım 4: collector/network.go — Address alanını kopyala**

Network collector'ın `InterfaceInfo` oluşturduğu yerde `Address: i.Address,` satırını ekle.

- [ ] **Adım 5: Test + commit**

```bash
cd ~/linux-dashboard && go build ./... && ./ldm &
sleep 2 && curl -s http://localhost:19876/api/v1/network | python3 -m json.tool | grep -A2 "address"
kill %1
```

Beklenen: `"address": "192.168.1.6/24"`

```bash
git add internal/linuxproc/network.go internal/collector/types.go internal/collector/network.go
git commit -m "feat: network — add IP address to interface info"
```

---

## Task 9: Dashboard dosyalarını web/ klasörüne taşı

Mevcut `internal/server/web/` klasörü değiştirilir; Gece Nöbeti dosyaları buraya kopyalanır.

**Files:**
- Delete: `internal/server/web/index.html` (eski React frontend)
- Delete: `internal/server/web/assets/` (eski assets)
- Copy: `~/Masaüstü/43/*.jsx` ve `~/Masaüstü/43/*.js` → `internal/server/web/`

- [ ] **Adım 1: Eski frontend'i kaldır, yeni dosyaları kopyala**

```bash
cd ~/linux-dashboard

# Eski frontend'i yedekle
mkdir -p internal/server/web/_old
mv internal/server/web/assets internal/server/web/_old/ 2>/dev/null || true
mv internal/server/web/index.html internal/server/web/_old/ 2>/dev/null || true
mv internal/server/web/favicon.svg internal/server/web/_old/ 2>/dev/null || true

# Yeni dashboard dosyalarını kopyala
cp ~/Masaüstü/43/app.jsx \
   ~/Masaüstü/43/panels.jsx \
   ~/Masaüstü/43/cores.jsx \
   ~/Masaüstü/43/surveillance.jsx \
   ~/Masaüstü/43/palette.jsx \
   ~/Masaüstü/43/tweaks-panel.jsx \
   ~/Masaüstü/43/index.html \
   internal/server/web/
```

- [ ] **Adım 2: index.html'de data.js referansını kaldır, live.js ekle**

`internal/server/web/index.html` dosyasında şu satırı:

```html
<script type="text/babel" src="data.js"></script>
```

Şununla değiştir:

```html
<script src="live.js"></script>
```

(`type="text/babel"` olmamalı — live.js sade JavaScript olacak, JSX değil)

- [ ] **Adım 3: Derleme testi (web/ embed kontrol)**

```bash
cd ~/linux-dashboard && go build ./...
```

Beklenen: hata yok (web/ klasöründe dosyalar var, embed çalışır)

- [ ] **Adım 4: Commit**

```bash
cd ~/linux-dashboard
git add internal/server/web/
git commit -m "feat: replace embedded frontend with Gece Nöbeti dashboard files"
```

---

## Task 10: live.js — API adapter katmanı

Bu dosya `data.js`'in yerini alır. Tüm `window.NWL` nesnesini gerçek API verisiyle doldurur, ardından React uygulamasını başlatır.

**Files:**
- Create: `internal/server/web/live.js`

- [ ] **Adım 1: internal/server/web/live.js oluştur**

```javascript
/* live.js — Gece Nöbeti canlı veri katmanı */
/* data.js'in yerini alır — window.NWL'yi gerçek API'den doldurur */

(function () {
  const API = window.location.origin;

  /* ── Statik UI metinleri (data.js'den kopyalandı) ── */
  const SYSLOG_NOTES = {
    "Out of memory: Killed process": "Çekirdek bellek baskısı altında. OOM-killer en obur süreci seçti.",
    "thermal: cpu_thermal trip point": "İşlemcinin termal eşiği aşıldı.",
    "sshd: invalid user root": "Tanımadığımız bir IP root denemesi yaptı.",
    "sshd: maximum authentication attempts exceeded": "Birisi denemekten yoruldu. fail2ban'i kontrol et.",
    "GPU HANG detected": "GPU bir kareyi çevirmekte tereddüt etti. Genelde toparlanır.",
    "TCP: out of memory": "Soket tamponları doluyor. /proc/sys/net/ipv4/tcp_mem.",
    "nf_conntrack: table full": "Bağlantı izleme tablosu doldu. nf_conntrack_max'i büyüt.",
    "EXT4-fs: warning: maximal mount count": "Dosya sistemi bakım istiyor. fsck düşün.",
    "segfault at": "Bir süreç olmayan bir yere uzandı. Çekirdek sonlandırdı.",
    "NetworkManager: connection lost": "Kablosuz bağlantı kesildi. Otomatik geri dönüş aktif.",
    "systemd-resolved: DNSSEC validation failed": "İmza doğrulaması başarısız.",
  };
  const SYSLOG_NOTE_GENERIC = "Kayda alındı. Çekirdek hatırlar. Sen unutsan da o unutmaz.";

  const WHISPERS = [
    "takas alanı bu gece sığ",
    "92 dosya tanımlayıcı açık. hangileri olduğunu biliyorsun.",
    "çekirdek uyanık. her zaman.",
    "sen nefes aldın. inotify de aldı.",
    "TCP tekrar göndermeleri tırmanıyor",
    "önbellek sıcak",
    "epoll boşuna bekledi",
    "PID 1 boot'tan beri gözünü kırpmadı",
    "sen kıpırdamadın. ben de.",
    "yük ritmine kavuşuyor",
  ];

  const COMMANDS = [
    { cmd: ":focus core <n>",   tr: "Çekirdek <n>'e odaklan",        example: ":focus core 2" },
    { cmd: ":focus proc <pid>", tr: "PID'e göre süreç dosyasını aç", example: ":focus proc 1" },
    { cmd: ":ack all",          tr: "Tüm kritik alarmları onayla",    example: ":ack all" },
    { cmd: ":replay <s>",       tr: "<s> saniye geri sar",            example: ":replay 30" },
    { cmd: ":live",             tr: "Canlıya dön",                    example: ":live" },
    { cmd: ":filter <sev>",     tr: "Olay akışını seviyeye göre süz", example: ":filter crit" },
    { cmd: ":clear",            tr: "Alarm sırasını temizle",         example: ":clear" },
    { cmd: ":help",             tr: "Yardımı aç",                     example: ":help" },
  ];

  const KEYS = [
    { k: ":",      tr: "Komut paletini aç" },
    { k: "?",      tr: "Yardımı göster" },
    { k: "j / k",  tr: "Olay akışında ↓ / ↑" },
    { k: "g / G",  tr: "En başa / en sona" },
    { k: "Enter",  tr: "Seçili olayı aç" },
    { k: "/",      tr: "Olay süzgecini değiştir" },
    { k: "1-8",    tr: "Çekirdek 0-7'ye odaklan" },
    { k: "0",      tr: "Odağı bırak" },
    { k: "Esc",    tr: "Kapat · canlıya dön" },
  ];

  const L = {
    brand: "GECE NÖBETİ", subbrand: "ana sistem", kernel: "çekirdek",
    uptime: "çalışma", sessions: "oturum", tty: "tty", observed: "GÖZLEM AKTİF",
    live: "CANLI", recording: "▮ KAYITTA", session: "OTURUM", host: "MAKİNE",
    uid: "UID", shell: "kabuk", ssh: "ssh", idle: "boşta",
    loadavg: "YÜK ORTALAMASI · 1DK", mem: "BELLEK", swap: "TAKAS",
    used: "KULLANIM", buff: "TAMPON", cache: "ÖNBELLEK", free: "BOŞ",
    disk: "DİSK", mounts: "bağlama", read: "OKUMA", write: "YAZMA",
    gpu: "GPU", driver: "sürücü", gpu_util: "SM YÜK", vram: "VRAM",
    power: "GÜÇ", temp: "SICAKLIK", sensors: "ALGILAYICILAR",
    cores: "İŞLEMCİ ÇEKİRDEKLERİ", topProcs: "EN AKTİF SÜREÇLER · CPU%",
    load60: "YÜK · SON 60 SANİYE", syslog: "OLAY AKIŞI",
    syslogSub: "journalctl · dmesg · auth", lines: "satır",
    all: "HEPSİ", info: "BİLGİ", warn: "UYARI", crit: "KRİTİK",
    network: "AĞ · ARAYÜZLER VE BAĞLANTILAR", conn: "BAĞLANTI",
    alerts: "ALARM SIRASI", quiet: "— sessiz —",
    noAlerts: "kritik olay yok. çekirdek eşit ritimle nefes alıyor.",
    pending: "bekleyen", process: "SÜREÇ", dossier: "DOSYA",
    syslogEntry: "OLAY · KAYIT", notes: "NOTLAR · NÖBET",
    whisper: "▮ SİSTEM FISILTISI", cur: "İML", bpm: "BPM", close: "KAPAT",
    ack: "ONAY", acknowledge: "ONAYLA", dismiss: "AT", drop: "DÜŞÜR",
    cpu: "CPU", memTab: "BELLEK", threads: "İŞ PARÇACIĞI", pid: "PID",
    state: "DURUM", selectCore: "bir çekirdek seç · yukarıdaki herhangi bir rayı tıkla",
    detail: "AYRINTI", tweakTitle: "Ayarlar", paletteSec: "Palet", toneLbl: "Ton",
    motionSec: "Hareket", tempoLbl: "Tempo", atmosSec: "Atmosfer",
    scanlines: "Tarama çizgileri", grain: "Film tanesi", vignette: "Vinyet",
    flicker: "Titreme", surveillanceSec: "Gözetim", eyeLbl: "Göz + imleç kaydı",
    whispersLbl: "Sistem fısıltıları", motion_still: "Hareketsiz",
    motion_calm: "Sakin", motion_living: "Canlı",
    pal_noir: "Noir", pal_crt: "CRT", pal_blue: "Mavi Saat", pal_amber: "Amber",
    replay: "GERİ SARMA", replayHint: "← / → · 1sn ileri-geri  ·  Esc · canlıya dön",
    live2: "canlı", secAgo: "sn önce", you: "sen", cmdTitle: "KOMUT PALETİ",
    cmdHint: "komutu yaz · Enter çalıştır · Esc kapat · ↑↓ önceki",
    helpTitle: "KISAYOLLAR",
    ticker: [
      "systemctl status nginx · etkin (çalışıyor)",
      "journalctl -f · akıyor",
      "uptime · yük ortalaması okunuyor",
      "free -h · bellek durumu",
      "ss -tunap · bağlantılar izleniyor",
      "iostat -x 1 · disk metrikleri",
    ],
    nightNotes: "Gerçek veri. Gerçek sistem. Çekirdek kendi kendine çalışıyor.",
  };

  /* ── Yardımcı fonksiyonlar ── */
  async function fetchJSON(path) {
    const res = await fetch(API + path);
    if (!res.ok) throw new Error(path + ' → ' + res.status);
    return res.json();
  }

  function classifyProcess(p) {
    if (p.uid === 0 && !p.exe_path) return 'kernel';
    if (p.uid === 0) return 'system';
    return 'user';
  }

  function adaptProcesses(apiProcs, shellName) {
    return apiProcs.map(p => ({
      uid: 'P-' + p.pid,
      name: p.name,
      kind: classifyProcess(p),
      cpu: p.cpu_percent,
      mem: Math.round(p.working_set / (1024 * 1024)),
      pid: p.pid,
      threads: p.thread_count,
      isOperator: p.name === shellName,
      core: p.core_id || 0,
      cpuLive: p.cpu_percent,
      cpuTarget: p.cpu_percent,
      state: p.status || 'R',
      container: null,
      lastActive: Date.now(),
    }));
  }

  function adaptMemory(apiMem) {
    const MB = 1024 * 1024;
    return {
      total: Math.round(apiMem.total_phys / MB),
      used: Math.round(apiMem.used_phys / MB),
      buff: Math.round(apiMem.buffers / MB),
      cache: Math.round(apiMem.cached / MB),
      free: Math.round(apiMem.free_phys / MB),
      swap_total: Math.round((apiMem.total_page_file - apiMem.total_phys) / MB) || 8192,
      swap_used: Math.round(apiMem.swap_used / MB),
    };
  }

  function adaptGpu(apiGpu) {
    if (!apiGpu || !apiGpu.available) {
      return { name: '—', driver: '—', util: 0, mem_used: 0, mem_total: 1,
               temp: 0, power: 0, power_max: 1, fan: 0, procs: [] };
    }
    const GB = 1024 * 1024 * 1024;
    return {
      name: apiGpu.name,
      driver: apiGpu.driver || '—',
      util: apiGpu.utilization,
      mem_used: apiGpu.vram_used / GB,
      mem_total: apiGpu.vram_total / GB,
      temp: apiGpu.temperature,
      power: 0,
      power_max: 300,
      fan: 0,
      procs: [],
    };
  }

  function adaptSensors(apiSensors) {
    return (apiSensors || []).map(s => ({
      name: s.name,
      tr: s.name,
      val: Math.round(s.value * 10) / 10,
      unit: s.unit,
      crit: Math.round(s.critical),
    }));
  }

  function adaptNetworkIfs(apiNetwork) {
    return (apiNetwork.interfaces || []).map(i => ({
      name: i.name,
      ip: i.address || '—',
      state: i.status,
      rxBase: i.in_bps / (1024 * 1024),
      txBase: i.out_bps / (1024 * 1024),
    }));
  }

  function adaptConnections(apiConns) {
    return (apiConns || []).slice(0, 20).map(c => ({
      proto: c.protocol,
      local: c.local_addr + ':' + c.local_port,
      remote: c.remote_port > 0 ? c.remote_addr + ':' + c.remote_port : '*',
      state: c.state,
      proc: c.process || '—',
    }));
  }

  function adaptMounts(apiDisk) {
    const GB = 1024 * 1024 * 1024;
    return (apiDisk.drives || []).map(d => ({
      mp: d.letter,
      fs: d.fs_type,
      used: d.used_bytes / GB,
      total: d.total_bytes / GB,
      hot: d.used_pct > 80,
    }));
  }

  function adaptCores(apiCores) {
    return (apiCores || []).map((c, i) => ({
      id: 'CPU' + i,
      freq: c.freq_mhz / 1000,
      governor: c.governor || 'unknown',
      microcode: c.microcode || '—',
      numa: 'node' + (c.numa_node || 0),
    }));
  }

  /* ── Syslog ── */
  function adaptSyslogEntry(e, idx) {
    return {
      id: 'S-live-' + (e.timestamp || idx),
      ts: e.timestamp,
      fac: e.facility,
      sev: e.severity,
      text: e.message,
      ack: false,
    };
  }

  /* ── Başlangıç verisi yükle, sonra uygulamayı başlat ── */
  async function init() {
    let HOST, CORES, SENSORS, GPU, MOUNTS, CONNECTIONS, NETWORK_IFS, initialSyslog;

    try {
      const [hostData, coresData, sensorsData, gpuData, diskData, netData, connsData, syslogData] =
        await Promise.allSettled([
          fetchJSON('/api/v1/host'),
          fetchJSON('/api/v1/cores'),
          fetchJSON('/api/v1/sensors'),
          fetchJSON('/api/v1/gpu'),
          fetchJSON('/api/v1/disk'),
          fetchJSON('/api/v1/network'),
          fetchJSON('/api/v1/connections'),
          fetchJSON('/api/v1/syslog'),
        ]);

      const h = hostData.status === 'fulfilled' ? hostData.value : {};
      HOST = {
        hostname: h.hostname || 'localhost',
        user:     h.user    || 'user',
        kernel:   h.kernel_version || '—',
        arch:     h.arch    || 'x86_64',
        os:       h.os      || 'Linux',
        tty:      h.tty     || 'pts/0',
        shell:    h.shell   || '/bin/sh',
        uid:      String(h.uid || '1000'),
        ssh:      '—',
        sessions: 1,
        boot: Date.now() - ((h.uptime_seconds || 3600) * 1000),
      };

      const rawCores = coresData.status === 'fulfilled' ? coresData.value : [];
      CORES = adaptCores(Array.isArray(rawCores) ? rawCores : []);
      if (CORES.length === 0) {
        const n = navigator.hardwareConcurrency || 4;
        for (let i = 0; i < n; i++)
          CORES.push({ id: 'CPU' + i, freq: 2.4, governor: 'unknown', microcode: '—', numa: 'node0' });
      }

      SENSORS = adaptSensors(sensorsData.status === 'fulfilled' ? sensorsData.value : []);
      GPU = adaptGpu(gpuData.status === 'fulfilled' ? gpuData.value : null);
      MOUNTS = adaptMounts(diskData.status === 'fulfilled' ? diskData.value : { drives: [] });
      NETWORK_IFS = adaptNetworkIfs(netData.status === 'fulfilled' ? netData.value : { interfaces: [] });
      CONNECTIONS = adaptConnections(connsData.status === 'fulfilled' ? connsData.value : []);

      const sl = syslogData.status === 'fulfilled' ? (syslogData.value.entries || []) : [];
      initialSyslog = sl.map(adaptSyslogEntry);

    } catch (err) {
      console.error('live.js init error:', err);
      HOST = { hostname: 'localhost', user: 'user', kernel: '—', arch: 'x86_64',
               os: 'Linux', tty: 'pts/0', shell: '/bin/sh', uid: '1000',
               ssh: '—', sessions: 1, boot: Date.now() - 3600000 };
      CORES = [{ id: 'CPU0', freq: 2.4, governor: 'unknown', microcode: '—', numa: 'node0' }];
      SENSORS = []; GPU = { name: '—', driver: '—', util: 0, mem_used: 0, mem_total: 1, temp: 0, power: 0, power_max: 1, fan: 0, procs: [] };
      MOUNTS = []; NETWORK_IFS = []; CONNECTIONS = []; initialSyslog = [];
    }

    const shellBase = HOST.shell.split('/').pop(); // "zsh", "bash" vs.

    /* ── Canlı güncelleme için polling hook'ları ── */
    let _processes = [];
    let _mem = { total: 32768, used: 12480, buff: 1480, cache: 8420, free: 10388, swap_total: 8192, swap_used: 320 };
    let _networkIfs = NETWORK_IFS;
    let _connections = CONNECTIONS;
    let _mounts = MOUNTS;
    let _sensors = SENSORS;
    let _gpu = GPU;

    /* Her 2s: süreçler + bellek + disk + ağ */
    async function pollFast() {
      try {
        const [procsData, memData, diskData, netData] = await Promise.allSettled([
          fetchJSON('/api/v1/processes'),
          fetchJSON('/api/v1/memory'),
          fetchJSON('/api/v1/disk'),
          fetchJSON('/api/v1/network'),
        ]);
        if (procsData.status === 'fulfilled')
          _processes = adaptProcesses(procsData.value, shellBase);
        if (memData.status === 'fulfilled')
          _mem = adaptMemory(memData.value);
        if (diskData.status === 'fulfilled')
          _mounts = adaptMounts(diskData.value);
        if (netData.status === 'fulfilled')
          _networkIfs = adaptNetworkIfs(netData.value);
      } catch (e) { /* sessiz hata */ }
    }

    /* Her 10s: bağlantılar + sensörler + GPU */
    async function pollSlow() {
      try {
        const [connsData, sensData, gpuData] = await Promise.allSettled([
          fetchJSON('/api/v1/connections'),
          fetchJSON('/api/v1/sensors'),
          fetchJSON('/api/v1/gpu'),
        ]);
        if (connsData.status === 'fulfilled')
          _connections = adaptConnections(connsData.value);
        if (sensData.status === 'fulfilled')
          _sensors = adaptSensors(sensData.value);
        if (gpuData.status === 'fulfilled')
          _gpu = adaptGpu(gpuData.value);
      } catch (e) { /* sessiz hata */ }
    }

    setInterval(pollFast, 2000);
    setInterval(pollSlow, 10000);
    pollFast(); // hemen başlat

    /* ── Syslog polling ── */
    let _lastSyslogTs = initialSyslog.length > 0 ? initialSyslog[0].ts : 0;
    const _newSyslogQueue = [];

    setInterval(async () => {
      try {
        const data = await fetchJSON('/api/v1/syslog');
        const entries = (data.entries || []).map(adaptSyslogEntry);
        const newEntries = entries.filter(e => e.ts > _lastSyslogTs);
        if (newEntries.length > 0) {
          _lastSyslogTs = newEntries[0].ts;
          newEntries.forEach(e => _newSyslogQueue.push(e));
        }
      } catch (e) {}
    }, 5000);

    /* ── window.NWL ── */
    window.NWL = {
      HOST,
      CORES,
      SENSORS,
      GPU,
      NETWORK_IFS,
      CONNECTIONS,
      SYSLOG_NOTES,
      SYSLOG_NOTE_GENERIC,
      WHISPERS,
      COMMANDS,
      KEYS,
      L,
      CONTAINERS: {},
      PROC_TEMPLATES: [],

      makeProcesses() {
        if (_processes.length > 0) return _processes;
        // fallback: bekleme sırasında boş liste
        return [];
      },

      emitSyslogEvent() {
        return _newSyslogQueue.shift() || null;
      },

      seedSyslog() {
        return initialSyslog;
      },

      /* app.jsx MOUNTS, GPU, SENSORS, NETWORK_IFS, CONNECTIONS için */
      get MOUNTS()      { return _mounts; },
      get GPU()         { return _gpu; },
      get SENSORS()     { return _sensors; },
      get NETWORK_IFS() { return _networkIfs; },
      get CONNECTIONS() { return _connections; },
    };

    /* ── app.jsx'i dinamik olarak yükle ── */
    /* Babel, type="text/babel" script'leri sayfa yüklendikten sonra işler.
       Tüm script'ler zaten index.html'de tanımlı, sadece NWL hazır olduktan
       sonra app.jsx'in render'a ulaşması gerekiyor. Babel her script'i sırayla
       çalıştırdığı için bu init() tamamlanmadan NWL tanımlı olmayacak.
       Ancak Babel async aware değil — bu yüzden tüm init işini sync olmayan
       şekilde tamamlayıp NWL'yi hazırladıktan sonra DOM'a yeni bir script
       etiketi ekliyoruz. */
    console.log('[live.js] NWL hazır, React uygulaması başlatılıyor…');
  }

  /* Babel script'leri DOMContentLoaded'da çalışır.
     live.js'i DOMContentLoaded'dan önce senkron çalıştırmak için
     init()'i hemen çağırıyoruz ama Promise döndürüyor.
     index.html'deki Babel script'leri en sonda olduğu için
     init() resolve olmadan önce çalışmayacak — bu sayede race condition yok. */
  window._nwlReady = init();
})();
```

- [ ] **Adım 2: index.html'de app.jsx'i NWL beklettir**

`internal/server/web/index.html` dosyasında `<script type="text/babel" src="app.jsx"></script>` satırını bul ve şununla değiştir:

```html
<script type="text/babel">
  /* app.jsx'i NWL hazır olduktan sonra çalıştır */
  window._nwlReady.then(() => {
    const s = document.createElement('script');
    s.type = 'text/babel';
    s.src = 'app.jsx';
    document.body.appendChild(s);
  });
</script>
```

- [ ] **Adım 3: Derleme testi**

```bash
cd ~/linux-dashboard && go build ./...
```

- [ ] **Adım 4: Commit**

```bash
git add internal/server/web/live.js internal/server/web/index.html
git commit -m "feat: live.js — full API adapter replacing data.js mock"
```

---

## Task 11: app.jsx — simülasyon intervallarını gerçek API ile değiştir

Mevcut `app.jsx`'teki simülasyon intervalları gerçek veriyle değiştirilir. 5 interval güncellenir.

**Files:**
- Modify: `internal/server/web/app.jsx`

- [ ] **Adım 1: Süreç simülasyon intervalını kaldır, gerçek fetch ekle**

`app.jsx` içinde şu bloğu bul:

```jsx
useEffect(() => {
  if (motionMul === 0) return;
  const id = setInterval(() => {
    setLiveProcesses(prev => prev.map(p => {
      if (Math.random() < 0.04 * motionMul) {
```

Bu `useEffect` bloğunun tamamını (kapanan `}, [motionMul]);` dahil) şununla **değiştir**:

```jsx
/* Gerçek süreç verisi — her 2s API'den */
useEffect(() => {
  const id = setInterval(async () => {
    try {
      const procs = window.NWL.makeProcesses();
      if (procs.length > 0) setLiveProcesses(procs);
    } catch (e) {}
  }, 2000);
  return () => clearInterval(id);
}, []);
```

- [ ] **Adım 2: Bellek simülasyon intervalını kaldır, gerçek fetch ekle**

`app.jsx` içinde şu bloğu bul:

```jsx
useEffect(() => {
  if (motionMul === 0) return;
  const id = setInterval(() => {
    setMem(prev => {
      const drift = (Math.random() - 0.5) * 80 * motionMul;
```

Bu bloğun tamamını şununla **değiştir**:

```jsx
/* Gerçek bellek verisi — her 3s NWL üzerinden */
useEffect(() => {
  const id = setInterval(() => {
    /* live.js NWL güncelleniyor, makeProcesses ile birlikte bellek de gelir */
  }, 3000);
  return () => clearInterval(id);
}, []);
```

Not: Bellek verisi `live.js`'te `_mem` değişkeninde tutuluyor ama app.jsx doğrudan erişemiyor. Şu adımda `window.NWL.getMem` fonksiyonu ekleyelim.

`live.js` dosyasında `window.NWL = { ... }` bloğuna şunu ekle:

```javascript
      getMem() { return _mem; },
```

Sonra `app.jsx`'teki bellek intervalını gerçekten güncelle:

```jsx
useEffect(() => {
  const id = setInterval(() => {
    const m = window.NWL.getMem && window.NWL.getMem();
    if (m && m.total > 0) setMem(m);
  }, 2000);
  return () => clearInterval(id);
}, []);
```

- [ ] **Adım 3: Disk simülasyon intervalını gerçek I/O verisiyle zenginleştir**

`app.jsx` içinde şu bloğu bul (ioSpark state):

```jsx
useEffect(() => {
  if (motionMul === 0) return;
  const id = setInterval(() => {
    setIoSpark(prev => ({
      read:  [...prev.read.slice(1),  Math.max(0, 40 + Math.sin(...)
```

Bu bloğu şununla **değiştir**:

```jsx
/* Disk I/O sparkline — API'den gerçek bps */
useEffect(() => {
  const id = setInterval(async () => {
    try {
      const res = await fetch(window.location.origin + '/api/v1/disk');
      if (!res.ok) return;
      const data = await res.json();
      const drives = data.drives || [];
      const totalRead  = drives.reduce((s, d) => s + (d.read_bps  || 0), 0) / (1024 * 1024);
      const totalWrite = drives.reduce((s, d) => s + (d.write_bps || 0), 0) / (1024 * 1024);
      setIoSpark(prev => ({
        read:  [...prev.read.slice(1),  Math.max(0, totalRead)],
        write: [...prev.write.slice(1), Math.max(0, totalWrite)],
      }));
    } catch (e) {}
  }, 2000);
  return () => clearInterval(id);
}, []);
```

- [ ] **Adım 4: Ağ sparkline intervalını gerçek bps verisiyle güncelle**

`app.jsx` içinde `setIfSparks` kullanan `useEffect` bloğunu bul:

```jsx
useEffect(() => {
  if (motionMul === 0) return;
  const id = setInterval(() => {
    setIfSparks(prev => {
      const next = {};
      window.NWL.NETWORK_IFS.forEach(i => {
```

Bu bloğu şununla **değiştir**:

```jsx
/* Ağ sparkline — API'den gerçek bps */
useEffect(() => {
  const id = setInterval(async () => {
    try {
      const res = await fetch(window.location.origin + '/api/v1/network');
      if (!res.ok) return;
      const data = await res.json();
      const ifaces = data.interfaces || [];
      setIfSparks(prev => {
        const next = { ...prev };
        ifaces.forEach(i => {
          const old = prev[i.name] || { rx: Array(40).fill(0), tx: Array(40).fill(0) };
          const rxMB = i.in_bps  / (1024 * 1024);
          const txMB = i.out_bps / (1024 * 1024);
          next[i.name] = {
            rx: [...old.rx.slice(1), rxMB],
            tx: [...old.tx.slice(1), txMB],
          };
        });
        return next;
      });
    } catch (e) {}
  }, 2000);
  return () => clearInterval(id);
}, []);
```

- [ ] **Adım 5: Syslog emit'ini gerçek API'ye bağla**

`app.jsx` içinde `emitSyslogEvent` kullanan `useEffect` bloğunu bul:

```jsx
useEffect(() => {
  if (motionMul === 0) return;
  const baseDelay = tweaks.motion === "living" ? 1600 : 3000;
  const id = setInterval(() => {
    if (Math.random() < 0.18) return;
    const e = window.NWL.emitSyslogEvent();
```

Bu bloğu şununla **değiştir**:

```jsx
/* Syslog — live.js kuyruğundan gerçek journal olayları */
useEffect(() => {
  const id = setInterval(() => {
    const e = window.NWL.emitSyslogEvent();
    if (!e) return;
    setSyslog(prev => [e, ...prev].slice(0, 400));
    if (e.sev === 'crit') setAlerts(prev => [e, ...prev].slice(0, 6));
  }, 2000);
  return () => clearInterval(id);
}, []);
```

- [ ] **Adım 6: Derleme testi**

```bash
cd ~/linux-dashboard && go build ./...
```

- [ ] **Adım 7: Commit**

```bash
git add internal/server/web/app.jsx internal/server/web/live.js
git commit -m "feat: app.jsx — replace all simulation intervals with real API fetches"
```

---

## Task 12: Binary derle ve uçtan uca test

- [ ] **Adım 1: Binary derle**

```bash
cd ~/linux-dashboard && go build -o ldm ./cmd/ldm/
```

Beklenen: hata yok, `ldm` dosyası oluştu

- [ ] **Adım 2: Çalıştır ve tarayıcıda kontrol et**

```bash
cd ~/linux-dashboard && ./ldm &
sleep 2 && xdg-open http://localhost:19876
```

- [ ] **Adım 3: API endpoint'lerini kontrol et**

```bash
for ep in host cores sensors syslog connections; do
  echo "=== /api/v1/$ep ===" 
  curl -s http://localhost:19876/api/v1/$ep | python3 -m json.tool | head -10
done
```

Beklenen: Her endpoint JSON döndürür, boş liste veya gerçek veri içerir

- [ ] **Adım 4: Dashboard'da kontrol edilecekler**

- [ ] CPU çekirdekleri gerçek governor ve freq gösteriyor
- [ ] Bellek buff/cache ayrımı gerçek değerleri gösteriyor  
- [ ] Sensör sıcaklıkları `/sys/class/hwmon`'dan geliyor
- [ ] Süreç listesi gerçek prosesleri gösteriyor, shell süreci "sen" olarak işaretli
- [ ] Syslog akışı gerçek journalctl loglarını gösteriyor
- [ ] Ağ arayüzleri IP adresleriyle birlikte listeleniyor
- [ ] Bağlantı listesi gerçek `/proc/net/tcp` verilerini gösteriyor

- [ ] **Adım 5: Eski python sunucusunu durdur**

```bash
# Masaüstü/43'teki python sunucusu artık gerekmiyor
pkill -f "python3 -m http.server 8743" 2>/dev/null || true
```

- [ ] **Adım 6: Final commit**

```bash
cd ~/linux-dashboard
git add -A
git commit -m "feat: Gece Nöbeti canlı entegrasyon tamamlandı — tek binary, gerçek veri"
```
