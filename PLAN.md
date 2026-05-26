# Linux Dashboard — Proje Plani

> **Amaç:** Windows Task Manager'in Linux portu — aynı stil, aynı tema, aynı API, sadece Linux için çalışacak.

**Proje Yeri:** `~/linux-dashboard/`
**Port:** `19876` (localhost only)
**Build:** Tek Go binary, `embed.FS` ile frontend gömülü

---

## Genel Mimari

```
linux-dashboard/
├── cmd/ldm/main.go              # Entry point
├── internal/
│   ├── collector/              # Linux /proc, /sys toplayıcılar
│   ├── controller/             # Process kontrol (signals)
│   ├── anomaly/               # Anomali motoru
│   ├── ai/                    # AI advisor (Anthropic, OpenAI, OpenRouter, MiniMax)
│   ├── server/               # REST API + SSE
│   ├── storage/              # Ring buffer
│   ├── config/               # YAML loader
│   ├── event/                # Event emitter
│   └── telegram/            # Telegram bot
├── web/                      # Frontend (kopyalanacak, aynı kalacak)
│   └── ...                   # React + Vite + Tailwind
├── configs/
│   └── default.yaml
├── build.sh                  # Build script
└── README.md
```

---

## Adımlar

### ADIM 1 — Proje iskeleti ve Go modülü ✅
- [x] `~/linux-dashboard/` dizinini oluştur
- [x] `go mod init github.com/burak/linux-dashboard`
- [x] `go get chi/v5, yaml.v3, uuid` — paketler indirildi
- [x] Internal dizinleri oluştur (collector, controller, ai, server, storage, config, event, telegram, anomaly, linuxproc)
- [x] `build.sh` scripti oluştur
- [x] `configs/default.yaml` — Linux'a özel config

### ADIM 2 — Linux /proc alt yapisi (linuxproc) ✅
- [x] `internal/linuxproc/proc.go` — ReadFile, ReadDir, ReadLink, Parse fonksiyonları
- [x] `internal/linuxproc/cpu.go` — /proc/stat CPU toplama, GlobalCPU(), CPUCollector
- [x] `internal/linuxproc/memory.go` — /proc/meminfo MemoryInfo, CollectMemory()
- [x] `internal/linuxproc/network.go` — InterfaceStats, CollectNetwork(), CollectNetworkTotals()
- [x] `internal/linuxproc/disk.go` — DriveInfo, CollectDisk(), CollectDiskStats(), DiskIO
- [x] `internal/linuxproc/process.go` — ProcessInfo, AllProcesses(), PIDsWithName()
- [x] `internal/linuxproc/ports.go` — PortEntry, ParsePorts(), InodeToPID(), PIDToName()

### ADIM 3 — Collector paketi ✅
- [x] `internal/collector/types.go` — tüm metric struct tanımları (API contract korundu)
- [x] `internal/collector/cpu.go` — CPUCollector
- [x] `internal/collector/memory.go` — MemoryCollector
- [x] `internal/collector/disk.go` — DiskCollector (mounts + sys/block)
- [x] `internal/collector/network.go` — NetworkCollector (delta BPS hesaplama)
- [x] `internal/collector/process.go` — ProcessCollector (CPU% hesaplama)
- [x] `internal/collector/tree.go` — BuildProcessTree, orphan detection
- [x] `internal/collector/ports.go` — PortCollector (linuxproc.ParsePorts → []PortBinding)
- [x] `internal/collector/gpu.go` — GPUCollector (nvidia-smi)
- [x] `internal/collector/manager.go` — CollectorManager, 3 ticker loop (1s fast, 2s tree, 3s ports)

**Build: `go build ./...` — ✅ derleniyor (tüm dosyalar hatasız)**

### ADIM 4 — Controller paketi ✅
- [x] `controller/controller.go` — Kill/Suspend/Resume/Nice/Affinity
- [x] `controller/safety.go` — korunan process listesi (Linux kernel processes)

### ADIM 5 — Config paketi ✅
- [x] `config/config.go` — Config struct
- [x] `config/loader.go` — YAML load + validate
- [x] `configs/default.yaml` — Linux'a özel path'ler

### ADIM 6 — Server paketi ✅
- [x] `server/server.go` — HTTP server, chi router, SSE hub, embed.FS static serving
- [x] `server/handlers.go` — tüm REST endpoint'ler
- [x] `server/router.go` — route wiring, index.html serving
- [x] `server/sse.go` — Server-Sent Events

### ADIM 7 — AI paketi ✅
- [x] `ai/advisor.go` — AI advisor
- [x] `ai/anthropic.go` — Anthropic Claude
- [x] `ai/openai.go` — OpenAI + OpenRouter + MiniMax + Groq + DeepSeek
- [x] `ai/prompt.go` — prompt templates
- [x] `ai/ratelimit.go` — rate limiting

### ADIM 8 — Anomaly, storage, event, telegram ✅
- [x] `anomaly/engine.go` — CPU, Memory, Disk, Network, Process anomaly detector'ları
- [x] `storage/store.go` — ring buffer
- [x] `event/emitter.go` — event emitter (Latest() fonksiyonu dahil)
- [x] `telegram/bot.go` — Telegram bot (/start, /status, /alerts, /help)

### ADIM 9 — Main.go ve wiring ✅
- [x] `cmd/ldm/main.go` — tüm parçaları birleştir
- [x] `cmd/ldm_debug/main.go` — debug binary (collector test)
- [x] `ldm_debug` çalışıyor, `ldm` shell background'da crash oluyor (sonra bakılacak)
- [ ] Single instance: Unix socket veya PID dosyası
- [ ] Browser: `xdg-open` ile
- [ ] Config path: `~/.config/linux-dashboard/config.yaml`

### ADIM 10 — Frontend entegrasyonu ✅
- [x] Frontend dosyalarını `web/` altına kopyala
- [x] Vite build et → `dist/` klasörü
- [x] `embed.FS` ile binary'e göm (`//go:embed all:web`)
- [x] `/` → index.html served
- [x] `/assets/*` → CSS/JS static files
- [x] `/api/v1/*` → REST API (static'ten önce öncelikli)
- [x] **Frontend TAM TÜRKÇE** — 17 sayfa çevrildi (Gösterge Paneli, İşlemler, Uyarılar, Ayarlar, AI, Hakkında, Diskler, Portlar, Ağaç, Kurallar)
- [x] **Assets 404** düzeltildi — embedded web dizini güncellendi (web/ → internal/server/web/)

### ADIM 11 — Build ve test ✅
- [x] `go build -o ldm_debug ./cmd/ldm_debug/` — başarılı
- [x] `go build -o ldm ./cmd/ldm/` — başarılı (embed ile)
- [x] `./ldm_debug` başlatıldı, port 19876'da çalışıyor
- [x] `http://127.0.0.1:19876/` — HTML dönüyor ✅
- [x] `http://127.0.0.1:19876/assets/*` — CSS/JS OK ✅ (404 çözüldü)
- [x] `http://127.0.0.1:19876/api/v1/system` — JSON OK
- [ ] `./ldm` shell'de crash — nedeni araştırılması gerekiyor (sonra bakılacak)

---

## API Endpoints (Aynı Kalacak)

| Method | Endpoint | Aciklama |
|--------|----------|----------|
| GET | `/api/v1/system` | Full system snapshot |
| GET | `/api/v1/cpu` | CPU metrics |
| GET | `/api/v1/memory` | Memory metrics |
| GET | `/api/v1/gpu` | GPU metrics |
| GET | `/api/v1/disk` | Disk metrics |
| GET | `/api/v1/network` | Network metrics |
| GET | `/api/v1/processes` | Process list |
| GET | `/api/v1/processes/tree` | Process tree |
| GET | `/api/v1/ports` | Port bindings |
| GET | `/api/v1/alerts` | Active alerts |
| GET | `/api/v1/ai/status` | AI durumu |
| POST | `/api/v1/ai/chat` | AI chat |
| GET | `/api/v1/stream` | SSE stream |
| GET/POST | `/api/v1/config` | Config get/update |
| POST | `/api/v1/processes/:pid/kill` | Kill process |
| POST | `/api/v1/processes/:pid/suspend` | Suspend process |
| POST | `/api/v1/processes/:pid/resume` | Resume process |

---

## AI Providerlar

| Provider | Endpoint | Model |
|----------|----------|-------|
| anthropic | `https://api.anthropic.com/v1/messages` | claude-sonnet-4 |
| openai | `https://api.openai.com/v1/chat/completions` | gpt-4o |
| openrouter | `https://openrouter.ai/api/v1/chat/completions` | herhangi |
| minimax | `https://api.minimax.chat/v1/chat/completions` | MiniMax-01 |
| groq | `https://api.groq.com/openai/v1/chat/completions` | llama-3.3 |
| deepseek | `https://api.deepseek.com/v1/chat/completions` | deepseek-chat |

---

## Data Tipleri (Aynı Kalacak)

```go
type CPUMetrics struct {
    TotalPercent float64   `json:"total_percent"`
    PerCore      []float64 `json:"per_core"`
    NumLogical   int       `json:"num_logical"`
    Name         string    `json:"name"`
    FreqMHz      uint32    `json:"freq_mhz"`
}

type MemoryMetrics struct {
    TotalPhys     uint64  `json:"total_phys"`
    AvailPhys     uint64  `json:"avail_phys"`
    UsedPhys      uint64  `json:"used_phys"`
    UsedPercent   float64 `json:"used_percent"`
}

type ProcessInfo struct {
    PID           uint32  `json:"pid"`
    ParentPID     uint32  `json:"parent_pid"`
    Name          string  `json:"name"`
    ExePath       string  `json:"exe_path"`
    CPUPercent    float64 `json:"cpu_percent"`
    WorkingSet    uint64  `json:"working_set"`
    ThreadCount   uint32  `json:"thread_count"`
    IsCritical    bool    `json:"is_critical"`
    Connections   int     `json:"connections"`
}

type PortBinding struct {
    Protocol   string `json:"protocol"`
    LocalAddr  string `json:"local_addr"`
    LocalPort  uint16 `json:"local_port"`
    RemoteAddr string `json:"remote_addr"`
    RemotePort uint16 `json:"remote_port"`
    State      string `json:"state"`
    PID        uint32 `json:"pid"`
    Process    string `json:"process"`
    Label      string `json:"label"`
}

type SystemSnapshot struct {
    Timestamp    time.Time      `json:"timestamp"`
    CPU          CPUMetrics     `json:"cpu"`
    Memory       MemoryMetrics  `json:"memory"`
    GPU          GPUMetrics     `json:"gpu"`
    Disk         DiskMetrics    `json:"disk"`
    Network      NetworkMetrics `json:"network"`
    Processes    []ProcessInfo  `json:"processes"`
    ProcessTree  []*ProcessNode `json:"process_tree,omitempty"`
    PortBindings []PortBinding  `json:"port_bindings,omitempty"`
}
```

---

## Protected Processes (Linux)

```
init (PID 1)
systemd (PID 1)
kthreadd (PID 2)
rcu_sched
ksoftirqd
migration
kswapd
fsync
crypt
jbd2
loop
systemd-journal
systemd-logind
dbus
sshd
containerd
docker
```

---

## Config Path

Linux'ta config yeri: `~/.config/linux-dashboard/config.yaml`

```yaml
server:
  host: 127.0.0.1
  port: 19876
  open_browser: true

monitoring:
  interval: 1s
  process_tree_interval: 2s
  port_scan_interval: 3s
  gpu_interval: 2s
  history_duration: 10m
  max_processes: 2000

controller:
  protected_processes:
    - systemd
    - kthreadd
    - init
    - sshd
    - dbus
  confirm_kill_system: true

well_known_ports:
  22: SSH
  80: HTTP
  443: HTTPS
  3000: Dev Server
  5432: PostgreSQL
  6379: Redis
  8080: HTTP Alt
  9090: Prometheus
  27017: MongoDB

ai:
  enabled: false
  provider: minimax  # or "anthropic", "openai", "openrouter", "deepseek", "groq"
  api_key: ""
  model: MiniMax-01
  endpoint: ""
  auto_analyze_on_critical: true
  max_tokens: 1024

telegram:
  enabled: false
  bot_token: ""
  allowed_chat_ids: []
  notify_on_critical: true
  require_confirm: true

ui:
  theme: system
  default_sort: cpu
  sparkline_points: 60
```

---

## Build Komutu

```bash
go build -ldflags="-s -w -X main.version=1.0.0" -o linux-dashboard ./cmd/ldm/
```

---

## Hedefler

1. API contract bozulmayacak — frontend aynı endpoint'leri bekliyor
2. Data tipleri aynı — JSON field'lar degismeyecek
3. Tema/stil aynı — CSS, React component'lar degismeyecek
4. Platform-specific sadece collector + controller
5. AI: MiniMax, OpenRouter, Anthropic, OpenAI, Groq, DeepSeek desteği

---

## Kalan İşler (Yarına)

- [ ] **VisionAPI modülü** — `/home/mayata/linux-dashboard/internal/ai/` konumunda resim yükleme API endpoint'leri (upload, process, retrieve)
- [ ] **Telegram bot** — `TG_BOT_TOKEN` ve `TG_ALLOWED_CHAT_IDS` config'leri gerekiyor
- [ ] **Logging** — ADIM 6 (log dosyası, stdout redirect)
- [ ] **Database** — ADIM 7 (PostgreSQL/SQLite entegrasyonu)
- [ ] **Docker** — ADIM 11 (Dockerfile, docker-compose)
- [ ] **./ldm** shell crash — arka planda crash oluyor, foreground'da çalışıyor
- [ ] **./ldm_debug** port değişti — 19876 yerine 8080 veya başka port kullanıyor

---

## Biten İşler (10 Mayıs 2026)

| Tarih | İş |
|--------|----|
| 10 Mayıs | Frontend TAM TÜRKÇE — 17 sayfa çevrildi |
| 10 Mayıs | Assets 404 düzeltildi (internal/server/web/ güncellendi) |
| 10 Mayıs | Sunucu port 19876'da çalışıyor |
| 10 Mayıs | API endpoint'leri test edildi (system, cpu, memory) |
| 10 Mayıs | ldm_debug binary yeniden build edildi |

---

*Plan oluşturuldu: 10 Mayıs 2026*
*Güncellendi: 10 Mayıs 2026*
*GitHub: github.com/burak/linux-dashboard*