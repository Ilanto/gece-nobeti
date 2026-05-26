/* live.js — Gece Nöbeti canlı veri katmanı */
/* data.js'in yerini alır — window.NWL'yi gerçek API'den doldurur */

(function () {
  const API = window.location.origin;

  /* ── Statik UI metinleri ── */
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
      uid:        'P-' + p.pid,
      name:       p.name,
      kind:       classifyProcess(p),
      cpu:        p.cpu_percent,
      mem:        Math.round(p.working_set / (1024 * 1024)),
      pid:        p.pid,
      threads:    p.thread_count,
      isOperator: p.name === shellName,
      core:       p.core_id || 0,
      cpuLive:    p.cpu_percent,
      cpuTarget:  p.cpu_percent,
      state:      p.status || 'R',
      container:  null,
      lastActive: Date.now(),
    }));
  }

  function adaptMemory(apiMem) {
    const MB = 1024 * 1024;
    return {
      total:      Math.round(apiMem.total_phys / MB),
      used:       Math.round(apiMem.used_phys / MB),
      buff:       Math.round(apiMem.buffers / MB),
      cache:      Math.round(apiMem.cached / MB),
      free:       Math.round(apiMem.free_phys / MB),
      swap_total: Math.round((apiMem.total_page_file - apiMem.total_phys) / MB) || 8192,
      swap_used:  Math.round(apiMem.swap_used / MB),
    };
  }

  function adaptGpu(apiGpu) {
    if (!apiGpu || !apiGpu.available) {
      return { name: '—', driver: '—', util: 0, mem_used: 0, mem_total: 1,
               temp: 0, power: 0, power_max: 1, fan: 0, procs: [] };
    }
    const GB = 1024 * 1024 * 1024;
    return {
      name:      apiGpu.name,
      driver:    apiGpu.driver || '—',
      util:      apiGpu.utilization,
      mem_used:  apiGpu.vram_used / GB,
      mem_total: apiGpu.vram_total / GB,
      temp:      apiGpu.temperature,
      power:     0,
      power_max: 300,
      fan:       0,
      procs:     [],
    };
  }

  function adaptSensors(apiSensors) {
    return (apiSensors || []).map(s => ({
      name: s.name,
      tr:   s.name,
      val:  Math.round(s.value * 10) / 10,
      unit: s.unit,
      crit: Math.round(s.critical),
    }));
  }

  function adaptNetworkIfs(apiNetwork) {
    return (apiNetwork.interfaces || []).map(i => ({
      name:   i.name,
      ip:     i.address || '—',
      state:  i.status,
      rxBase: i.in_bps  / (1024 * 1024),
      txBase: i.out_bps / (1024 * 1024),
    }));
  }

  function adaptConnections(apiConns) {
    return (apiConns || []).slice(0, 20).map(c => ({
      proto:  c.protocol,
      local:  c.local_addr + ':' + c.local_port,
      remote: c.remote_port > 0 ? c.remote_addr + ':' + c.remote_port : '*',
      state:  c.state,
      proc:   c.process || '—',
    }));
  }

  function adaptMounts(apiDisk) {
    const GB = 1024 * 1024 * 1024;
    return (apiDisk.drives || []).map(d => ({
      mp:    d.letter,
      fs:    d.fs_type,
      used:  d.used_bytes / GB,
      total: d.total_bytes / GB,
      hot:   d.used_pct > 80,
    }));
  }

  function adaptCores(apiCores) {
    return (apiCores || []).map((c, i) => ({
      id:        'CPU' + i,
      freq:      c.freq_mhz / 1000,
      governor:  c.governor || 'unknown',
      microcode: c.microcode || '—',
      numa:      'node' + (c.numa_node || 0),
    }));
  }

  function adaptSyslogEntry(e, idx) {
    return {
      id:  'S-live-' + (e.timestamp || idx),
      ts:  e.timestamp,
      fac: e.facility,
      sev: e.severity,
      text: e.message,
      ack: false,
    };
  }

  /* ── Ana başlatma fonksiyonu ── */
  async function init() {
    let HOST, CORES, SENSORS, GPU, MOUNTS, CONNECTIONS, NETWORK_IFS, initialSyslog;

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
      user:     h.user     || 'user',
      kernel:   h.kernel_version || '—',
      arch:     h.arch     || 'x86_64',
      os:       h.os       || 'Linux',
      tty:      h.tty      || 'pts/0',
      shell:    h.shell    || '/bin/sh',
      uid:      String(h.uid || '1000'),
      ssh:      '—',
      sessions: 1,
      boot:     Date.now() - ((h.uptime_seconds || 3600) * 1000),
    };

    const rawCores = coresData.status === 'fulfilled' ? coresData.value : [];
    CORES = adaptCores(Array.isArray(rawCores) ? rawCores : []);
    if (CORES.length === 0) {
      const n = navigator.hardwareConcurrency || 4;
      for (let i = 0; i < n; i++)
        CORES.push({ id: 'CPU' + i, freq: 2.4, governor: 'unknown', microcode: '—', numa: 'node0' });
    }

    SENSORS     = adaptSensors(sensorsData.status === 'fulfilled' ? sensorsData.value : []);
    GPU         = adaptGpu(gpuData.status === 'fulfilled' ? gpuData.value : null);
    MOUNTS      = adaptMounts(diskData.status === 'fulfilled' ? diskData.value : { drives: [] });
    NETWORK_IFS = adaptNetworkIfs(netData.status === 'fulfilled' ? netData.value : { interfaces: [] });
    CONNECTIONS = adaptConnections(connsData.status === 'fulfilled' ? connsData.value : []);

    const sl = syslogData.status === 'fulfilled' ? (syslogData.value.entries || []) : [];
    initialSyslog = sl.map(adaptSyslogEntry);

    const shellBase = HOST.shell.split('/').pop();

    /* ── Canlı güncelleme state'i ── */
    let _processes  = [];
    let _mem        = { total: 32768, used: 12480, buff: 1480, cache: 8420, free: 10388, swap_total: 8192, swap_used: 320 };
    let _networkIfs = NETWORK_IFS;
    let _connections = CONNECTIONS;
    let _mounts     = MOUNTS;
    let _sensors    = SENSORS;
    let _gpu        = GPU;

    async function pollFast() {
      try {
        const [procsData, memData, diskData, netData] = await Promise.allSettled([
          fetchJSON('/api/v1/processes'),
          fetchJSON('/api/v1/memory'),
          fetchJSON('/api/v1/disk'),
          fetchJSON('/api/v1/network'),
        ]);
        if (procsData.status === 'fulfilled') _processes  = adaptProcesses(procsData.value, shellBase);
        if (memData.status   === 'fulfilled') _mem        = adaptMemory(memData.value);
        if (diskData.status  === 'fulfilled') _mounts     = adaptMounts(diskData.value);
        if (netData.status   === 'fulfilled') _networkIfs = adaptNetworkIfs(netData.value);
      } catch (e) {}
    }

    async function pollSlow() {
      try {
        const [connsData, sensData, gpuData] = await Promise.allSettled([
          fetchJSON('/api/v1/connections'),
          fetchJSON('/api/v1/sensors'),
          fetchJSON('/api/v1/gpu'),
        ]);
        if (connsData.status === 'fulfilled') _connections = adaptConnections(connsData.value);
        if (sensData.status  === 'fulfilled') _sensors     = adaptSensors(sensData.value);
        if (gpuData.status   === 'fulfilled') _gpu         = adaptGpu(gpuData.value);
      } catch (e) {}
    }

    setInterval(pollFast, 2000);
    setInterval(pollSlow, 10000);
    pollFast();

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
        return _processes.length > 0 ? _processes : [];
      },

      getMem() {
        return _mem;
      },

      emitSyslogEvent() {
        return _newSyslogQueue.shift() || null;
      },

      seedSyslog() {
        return initialSyslog;
      },

      get MOUNTS()      { return _mounts; },
      get GPU()         { return _gpu; },
      get SENSORS()     { return _sensors; },
      get NETWORK_IFS() { return _networkIfs; },
      get CONNECTIONS() { return _connections; },
    };

    console.log('[live.js] NWL hazır — ' + CORES.length + ' çekirdek, ' + SENSORS.length + ' sensör');
  }

  window._nwlReady = init();
})();
