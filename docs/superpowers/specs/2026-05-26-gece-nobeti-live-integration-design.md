# Gece Nöbeti — Canlı Entegrasyon Tasarımı

**Tarih:** 2026-05-26  
**Proje:** linux-dashboard + Gece Nöbeti dashboard birleşimi  
**Hedef:** Tek Go binary'si, tek port (19876), tamamen gerçek sistem verisi

---

## Özet

"Gece Nöbeti" adlı Linux sistem izleme arayüzü şu an mock veriyle çalışıyor.  
Bu tasarım: arayüzü `linux-dashboard` Go binary'siyle tam entegre eder.  
Sonuç: `./ldm` komutuyla her şey ayağa kalkar — ayrı dosya sunucusuna gerek yok.

---

## Mimari

```
./ldm (tek binary)
  ├── Go backend  →  /proc, /sys, journalctl'den gerçek veri okur
  ├── REST API    →  /api/v1/* endpoint'leri
  └── Frontend    →  Gece Nöbeti HTML/JSX dosyaları (go:embed ile gömülü)
```

Tarayıcı `http://localhost:19876` adresini açar.  
Frontend aynı origin'den API'ye çağrı yapar — CORS sorunu yok.

---

## Yeni Backend Endpoint'leri

Mevcut linux-dashboard'da eksik olan veriler için yeni okuyucular ve endpoint'ler:

| Endpoint | Veri Kaynağı | Dönen Veri |
|---|---|---|
| `/api/v1/host` | `/proc/version`, `/proc/uptime`, `/etc/os-release`, `hostname` | hostname, kernel, os, arch, uptime, shell, tty |
| `/api/v1/memory` (genişletildi) | `/proc/meminfo` | used, buff, cache, free, swap_used, swap_total (MB) |
| `/api/v1/cores` | `/sys/devices/system/cpu/cpu*/cpufreq/` | her çekirdek için freq, governor, numa node |
| `/api/v1/sensors` | `/sys/class/hwmon/*/temp*`, `/sys/class/thermal/*/temp` | sıcaklık (°C), fan rpm, batarya % |
| `/api/v1/syslog` | `journalctl -n 200 -o json --no-pager` | timestamp, facility, severity, message |
| `/api/v1/connections` | `/proc/net/tcp`, `/proc/net/tcp6`, `/proc/net/udp` | proto, local, remote, state, pid, process adı |

Mevcut endpoint'ler (`/api/v1/cpu`, `/api/v1/processes`, `/api/v1/network`, `/api/v1/disk`, `/api/v1/gpu`) değişmeden kalır.

---

## SSE Stream Güncellemesi

Mevcut `/api/v1/stream` endpoint'i genişletilir:  
- CPU per-core yükü her saniye push edilir  
- Kritik syslog olayları anlık push edilir  
- Süreç listesi değişimleri push edilir

---

## Frontend Değişiklikleri

### Dosya Konumu

Mevcut:
```
~/Masaüstü/43/index.html  (python http.server ile servis ediliyor)
```

Yeni:
```
linux-dashboard/web/  (go:embed ile binary içinde)
  ├── index.html
  ├── live.js         ← data.js'in yerini alır
  ├── app.jsx
  ├── panels.jsx
  ├── cores.jsx
  ├── surveillance.jsx
  ├── palette.jsx
  └── tweaks-panel.jsx
```

### live.js — Adapter Katmanı

`data.js` tamamen kaldırılır. `live.js` şunları yapar:

1. Sayfa yüklenirken `HOST`, `CORES`, `SENSORS`, `GPU` statik bilgilerini backend'den çeker
2. Her 2 saniyede CPU, bellek, disk, ağ, süreç verilerini polling ile günceller
3. SSE stream'e bağlanarak gerçek zamanlı CPU ve syslog güncellemelerini alır
4. Gelen Go JSON yapılarını dashboard'un beklediği formata dönüştürür (adapter fonksiyonları)
5. `window.NWL` nesnesini gerçek veriyle doldurur — JSX dosyaları değişmez

### Adapter Dönüşümleri (data.js → live.js)

| Backend Alanı | Dashboard Beklentisi |
|---|---|
| `cpu.per_core[]` (float64[]) | `CORES[i].load` + `liveCoreLoads[]` |
| `memory.used_phys` (bytes) | `mem.used` (MB) |
| `memory.avail_phys` → hesap | `mem.buff`, `mem.cache`, `mem.free` (MB) |
| `processes[].cpu_percent` | `p.cpuLive` |
| `processes[].working_set` (bytes) | `p.mem` (MB) |
| `processes[].thread_count` | `p.threads` |
| `network.interfaces[].in_bps` | `ifSparks[name].rx` (MB/s) |
| `sensors[].value` | `SENSORS[i].val` |
| `syslog[].message` | syslog stream olayı |

---

## Go Backend — Yeni Dosyalar

```
internal/linuxproc/
  ├── host.go        ← YENİ: hostname, kernel, uptime, os-release
  ├── cores.go       ← YENİ: per-core governor, freq, numa
  ├── sensors.go     ← YENİ: hwmon sıcaklık, fan, batarya
  ├── syslog.go      ← YENİ: journalctl JSON parser
  ├── connections.go ← YENİ: /proc/net/tcp* parser
  └── memory.go      ← DEĞİŞİR: buff/cache ayrımı eklenir

internal/collector/
  ├── host.go        ← YENİ
  ├── cores.go       ← YENİ
  ├── sensors.go     ← YENİ
  ├── syslog.go      ← YENİ
  ├── connections.go ← YENİ
  └── types.go       ← DEĞİŞİR: yeni struct'lar eklenir

internal/server/
  ├── handlers.go    ← DEĞİŞİR: yeni endpoint handler'ları eklenir
  └── router.go      ← DEĞİŞİR: yeni rotalar eklenir
```

---

## Uygulama Sırası

1. Go backend — yeni linuxproc okuyucular
2. Go backend — collector + handler + router güncellemeleri
3. Frontend — dashboard dosyalarını `web/` klasörüne taşı
4. Frontend — `live.js` yaz (adapter katmanı)
5. Binary'yi yeniden derle, test et

---

## Başarı Kriterleri

- `./ldm` komutuyla `http://localhost:19876` açılır
- Tüm veriler gerçek sistem değerlerini gösterir (mock yok)
- CPU, bellek, süreç verileri 2 saniyede güncellenir
- Syslog akışı journalctl'den gerçek log satırlarını gösterir
- Sensör sıcaklıkları gerçek `/sys` değerlerinden okunur
- Ağ bağlantıları gerçek `/proc/net/tcp` verilerinden gelir
