# Gece Nöbeti · Night Watch

**TR** | [EN](#english)

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/platform-Linux-FCC624?style=flat-square&logo=linux&logoColor=black" />
  <img src="https://img.shields.io/github/v/release/Ilanto/gece-nobeti?style=flat-square&color=d99752&label=release" />
  <a href="https://ilanto.github.io/gece-nobeti/"><img src="https://img.shields.io/badge/demo-GitHub%20Pages-222?style=flat-square&logo=github&logoColor=white" /></a>
  <img src="https://img.shields.io/badge/license-MIT-6a8472?style=flat-square" />
</p>

---

## Türkçe

Gece Nöbeti, Linux sistemler için terminal estetiğiyle tasarlanmış bir sistem izleme dashboard'udur.  
Tek bir Go binary'si olarak çalışır — kurulum gerekmez, bağımlılık yok.

**Canlı Demo →** [ilanto.github.io/gece-nobeti](https://ilanto.github.io/gece-nobeti/)

### Demo

<!-- GIF buraya gelecek -->
> 📹 Demo kaydı yakında eklenecek.

![Gece Nöbeti Dashboard](https://raw.githubusercontent.com/Ilanto/gece-nobeti/master/docs/preview.png)

### Özellikler

- **CPU çekirdekleri** — her çekirdeğin yükü, frekansı ve governor'u gerçek zamanlı
- **Bellek** — kullanılan / buffer / cache / swap ayrımıyla
- **Sensörler** — `/sys/class/hwmon` üzerinden gerçek sıcaklıklar
- **Süreçler** — `/proc` kaynaklı, UID ve hangi çekirdekte çalıştığı dahil
- **Syslog** — `journalctl`'dan canlı log akışı
- **Ağ** — arayüz başına IP, gelen/giden bant genişliği sparkline
- **Bağlantılar** — `/proc/net/tcp`'den inode→PID eşleşmesiyle
- **Tema** — 4 renk paleti (Noir · CRT · Mavi Saat · Amber), tarama çizgisi, film tanesi, vinyet efektleri
- **Replay tamponu** — son 60 saniyelik geçmişe geri sar
- **Komut paleti** — `:focus core 3`, `:replay 30`, `:filter crit` vb.

### Kurulum ve Çalıştırma

```bash
git clone https://github.com/Ilanto/gece-nobeti.git
cd gece-nobeti
go build -o ldm ./cmd/ldm/
./ldm
```

Tarayıcıda `http://localhost:19876` adresi otomatik açılır.

Farklı bir adres/port için: `./ldm --bind 0.0.0.0:8080`

**Gereksinimler:**
- Go 1.21+
- Linux (kernel 3.14+ önerilir — `MemAvailable` için)
- `journalctl` — syslog için (systemd); yoksa syslog paneli boş kalır, diğer her şey çalışır
- `ip` komutu — ağ arayüzü IP'leri için

### Klavye Kısayolları

| Tuş | İşlev |
|-----|-------|
| `:` | Komut paletini aç |
| `?` | Yardımı göster |
| `1–8` | Çekirdek 0–7'ye odaklan |
| `j / k` | Syslog'da ↓ / ↑ |
| `← →` | Replay tamponu |
| `Esc` | Kapat / canlıya dön |

### Teknoloji

- **Backend:** Go, chi router, `/proc` + `/sys` okuyucular, SSE stream
- **Frontend:** React 18 (CDN+Babel), JetBrains Mono, Barlow Condensed, Newsreader
- **Gömülü:** `go:embed` — tek binary, dışarıdan dosya gerekmez

---

## English

<a name="english"></a>

Night Watch is a terminal-aesthetic Linux system monitoring dashboard.  
It runs as a single Go binary — no installation, no dependencies.

**Live Demo →** [ilanto.github.io/gece-nobeti](https://ilanto.github.io/gece-nobeti/)

### Demo

<!-- GIF goes here -->
> 📹 Demo recording coming soon.

### Features

- **CPU cores** — per-core load, frequency and governor in real time
- **Memory** — used / buffer / cache / swap breakdown
- **Sensors** — real temperatures via `/sys/class/hwmon`
- **Processes** — sourced from `/proc`, including UID and which core it runs on
- **Syslog** — live log stream from `journalctl`
- **Network** — per-interface IP, inbound/outbound bandwidth sparklines
- **Connections** — from `/proc/net/tcp` with inode→PID mapping
- **Themes** — 4 color palettes (Noir · CRT · Blue Hour · Amber), scanlines, film grain, vignette
- **Replay buffer** — scrub back through the last 60 seconds
- **Command palette** — `:focus core 3`, `:replay 30`, `:filter crit` and more

### Getting Started

```bash
git clone https://github.com/Ilanto/gece-nobeti.git
cd gece-nobeti
go build -o ldm ./cmd/ldm/
./ldm
```

Then open `http://localhost:19876` in your browser.

To bind to a different address/port: `./ldm --bind 0.0.0.0:8080`

**Requirements:**
- Go 1.21+
- Linux (kernel 3.14+ recommended — for `MemAvailable`)
- `journalctl` — for syslog (systemd); if absent, the syslog panel stays empty but everything else works
- `ip` command — for network interface addresses

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `:` | Open command palette |
| `?` | Show help |
| `1–8` | Focus core 0–7 |
| `j / k` | Scroll syslog ↓ / ↑ |
| `← →` | Scrub replay buffer |
| `Esc` | Close / return to live |

### Tech Stack

- **Backend:** Go, chi router, `/proc` + `/sys` readers, SSE stream
- **Frontend:** React 18 (CDN+Babel), JetBrains Mono, Barlow Condensed, Newsreader
- **Embedded:** `go:embed` — single binary, no external files needed

---

*Gece Nöbeti — "Night Watch" in Turkish. The kernel is always awake.*
