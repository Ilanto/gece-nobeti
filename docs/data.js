/* Gece Nöbeti · Ana Sistem — mock data */

const HOST = {
  hostname: "midwatch-01",
  user: "kardas",
  kernel: "6.8.0-31-generic",
  arch: "x86_64",
  os: "Debian GNU/Linux 12 (bookworm)",
  tty: "pts/2",
  shell: "/bin/zsh",
  uid: "1000",
  ssh: "10.0.4.17:22 ← 198.51.100.42",
  sessions: 3,
  boot: Date.now() - (3*86400 + 7*3600 + 22*60 + 18) * 1000,
};

const CORES = Array.from({ length: 8 }, (_, i) => ({
  id: `CPU${i}`,
  freq: 3.6 + (i % 4) * 0.1,
  governor: i < 4 ? "performance" : "powersave",
  microcode: "0xf4",
  numa: i < 4 ? "node0" : "node1",
}));

/* container ID's for some processes — gives a [docker:xxxx] tag */
const CONTAINERS = {
  "nginx":        "4f2a3c",
  "redis-server": "8b1d09",
  "postgres":     "2c447e",
  "node":         "6a91d2",
};

/* zsh is the operator's own shell — marked isOperator */
const PROC_TEMPLATES = [
  { name: "zsh",          kind: "user",   cpu:  0.2, mem:  28,  pid:  3402, threads:  1, isOperator: true },
  { name: "firefox",      kind: "user",   cpu: 35,   mem:1240,  pid: 21847, threads: 64 },
  { name: "chrome",       kind: "user",   cpu: 22,   mem: 980,  pid: 22014, threads: 48 },
  { name: "code",         kind: "user",   cpu: 18,   mem: 720,  pid: 18203, threads: 32 },
  { name: "node",         kind: "user",   cpu: 12,   mem: 540,  pid: 19994, threads: 16 },
  { name: "python3",      kind: "user",   cpu: 28,   mem: 420,  pid: 20771, threads:  4 },
  { name: "postgres",     kind: "system", cpu:  8,   mem: 360,  pid:  1842, threads: 12 },
  { name: "redis-server", kind: "system", cpu:  3,   mem: 120,  pid:  1923, threads:  4 },
  { name: "nginx",        kind: "system", cpu:  2,   mem:  72,  pid:  1247, threads:  8 },
  { name: "docker",       kind: "system", cpu:  6,   mem: 280,  pid:   997, threads: 16 },
  { name: "containerd",   kind: "system", cpu:  4,   mem: 210,  pid:  1003, threads: 12 },
  { name: "gnome-shell",  kind: "user",   cpu: 14,   mem: 680,  pid:  3402, threads: 22 },
  { name: "Xorg",         kind: "user",   cpu:  9,   mem: 320,  pid:  3201, threads:  6 },
  { name: "systemd",      kind: "system", cpu:  1,   mem:  18,  pid:     1, threads:  1 },
  { name: "kworker/u16",  kind: "kernel", cpu:  4,   mem:   0,  pid:    72, threads:  1 },
  { name: "ksoftirqd/0",  kind: "kernel", cpu:  2,   mem:   0,  pid:    14, threads:  1 },
  { name: "rcu_sched",    kind: "kernel", cpu:  1,   mem:   0,  pid:    16, threads:  1 },
  { name: "sshd",         kind: "system", cpu:  1,   mem:  24,  pid:  1119, threads:  3 },
  { name: "java",         kind: "user",   cpu: 42,   mem:1840,  pid: 24008, threads: 88 },
  { name: "rustc",        kind: "user",   cpu: 78,   mem: 920,  pid: 25441, threads:  8 },
  { name: "ffmpeg",       kind: "user",   cpu: 65,   mem: 540,  pid: 26113, threads: 16 },
];

function makeProcesses() {
  const procs = [];
  let id = 1000;
  PROC_TEMPLATES.forEach((t, i) => {
    const primaryCore = i % CORES.length;
    procs.push({
      ...t,
      uid: `P-${id++}`,
      core: primaryCore,
      cpuLive: t.cpu + (Math.random() - 0.5) * 6,
      cpuTarget: t.cpu,
      state: "R",
      container: CONTAINERS[t.name] || null,
      lastActive: Date.now(),
    });
  });
  return procs;
}

const SYSLOG_TEMPLATES = [
  { fac:"kern",   sev:"info", text:"usb 1-1: new high-speed USB device number 4" },
  { fac:"kern",   sev:"warn", text:"EXT4-fs: warning: maximal mount count reached" },
  { fac:"kern",   sev:"crit", text:"Out of memory: Killed process 24008 (java) total-vm:8421MB" },
  { fac:"kern",   sev:"warn", text:"thermal: cpu_thermal trip point 0 reached 87°C" },
  { fac:"kern",   sev:"info", text:"audit: USER_LOGIN op=login uid=1000 res=success" },
  { fac:"auth",   sev:"warn", text:"sshd: invalid user root from 198.51.100.7" },
  { fac:"auth",   sev:"info", text:"PAM: session opened for user kardas by (uid=0)" },
  { fac:"auth",   sev:"crit", text:"sshd: maximum authentication attempts exceeded" },
  { fac:"daemon", sev:"info", text:"systemd: Started apt-daily.timer" },
  { fac:"daemon", sev:"info", text:"systemd: journal-flush.service complete" },
  { fac:"daemon", sev:"warn", text:"NetworkManager: connection lost on wlp3s0" },
  { fac:"user",   sev:"info", text:"cron: (kardas) CMD (backup.sh --incremental)" },
  { fac:"kern",   sev:"info", text:"nvme0n1: SMART self-test routine started" },
  { fac:"kern",   sev:"warn", text:"TCP: out of memory -- consider tuning tcp_mem" },
  { fac:"kern",   sev:"info", text:"Bluetooth: hci0 link mode reset" },
  { fac:"audit",  sev:"info", text:"audit: type=1300 ARCH=c000003e SYSCALL=execve" },
  { fac:"kern",   sev:"crit", text:"segfault at 0 ip 00007f2b3c rsp 00007ffd1e error 4" },
  { fac:"kern",   sev:"warn", text:"i915: GPU HANG detected on render ring" },
  { fac:"daemon", sev:"info", text:"dockerd: container 4f2a3c started" },
  { fac:"daemon", sev:"warn", text:"systemd-resolved: DNSSEC validation failed" },
  { fac:"kern",   sev:"info", text:"nf_conntrack: table full, dropping packet" },
  { fac:"cron",   sev:"info", text:"CRON[2341]: pam_unix(cron:session): session closed" },
];

/* Türkçe açıklamalar — overlay'lerde yan yana gösterilir */
const SYSLOG_NOTES = {
  "Out of memory: Killed process 24008 (java) total-vm:8421MB":
    "Çekirdek bellek baskısı altında. OOM-killer en obur süreci seçti ve sessizce vurdu. Java süreci geride bir vm boşluğu bıraktı.",
  "thermal: cpu_thermal trip point 0 reached 87°C":
    "İşlemcinin termal eşiği aşıldı. Fanlar duyarsa hızlanır. Sen duyarsan kuliste bir uğultu olur.",
  "sshd: invalid user root from 198.51.100.7":
    "Tanımadığımız bir IP root denemesi yaptı. Üç kere. Sonra vazgeçti. Gözledim, kayıt aldım.",
  "sshd: maximum authentication attempts exceeded":
    "Birisi denemekten yoruldu. Ya da yeni bir yöntem deneyecek. fail2ban'i kontrol et.",
  "i915: GPU HANG detected on render ring":
    "GPU bir kareyi çevirmekte tereddüt etti. Genelde kendiliğinden toparlanır. Genelde.",
  "TCP: out of memory -- consider tuning tcp_mem":
    "Soket tamponları doluyor. Upstream sıkışmış olabilir. /proc/sys/net/ipv4/tcp_mem.",
  "nf_conntrack: table full, dropping packet":
    "Bağlantı izleme tablosu doldu. Yeni bağlantılar düşüyor. nf_conntrack_max'i büyütmek gerekebilir.",
  "EXT4-fs: warning: maximal mount count reached":
    "Dosya sistemi yeterince mount edildi. Yakın bir bakım pencerelerinde fsck düşünülebilir.",
  "segfault at 0 ip 00007f2b3c rsp 00007ffd1e error 4":
    "Bir süreç adres uzayında olmayan bir yere uzandı. Çekirdek nazikçe sonlandırdı.",
  "NetworkManager: connection lost on wlp3s0":
    "Kablosuz bağlantı kesildi. Otomatik geri dönüş aktif. Yağmur olabilir.",
  "systemd-resolved: DNSSEC validation failed":
    "İmza doğrulaması başarısız. Yukarıdaki resolver güvenilmez olabilir.",
};
const SYSLOG_NOTE_GENERIC = "Kayda alındı. Çekirdek hatırlar. Sen unutsan da o unutmaz.";

function emitSyslogEvent() {
  const t = SYSLOG_TEMPLATES[Math.floor(Math.random() * SYSLOG_TEMPLATES.length)];
  return {
    id: `S-${Date.now()}-${Math.floor(Math.random()*9999)}`,
    ts: Date.now(), fac: t.fac, sev: t.sev, text: t.text, ack: false,
  };
}
function seedSyslog(n=80) {
  const now = Date.now();
  const out = [];
  for (let i=n; i>0; i--) {
    const t = SYSLOG_TEMPLATES[Math.floor(Math.random() * SYSLOG_TEMPLATES.length)];
    out.push({
      id: `S-init-${i}`,
      ts: now - i * (3000 + Math.random() * 14000),
      fac: t.fac, sev: t.sev, text: t.text,
      ack: Math.random() > 0.2,
    });
  }
  return out;
}

const WHISPERS = [
  "takas alanı bu gece sığ",
  "92 dosya tanımlayıcı açık. hangileri olduğunu biliyorsun.",
  "çekirdek 6.8.0 · son açılış {uptime} önce",
  "PID 1 boot'tan beri gözünü kırpmadı",
  "bağlam değişimi 4.2M/sn · vızıltı",
  "son tuş vuruşun {idle} önce",
  "fanlar 1240 rpm · sen hâlâ boştasın",
  "sen nefes aldın. inotify de aldı.",
  "denetim günlüğü dakikada 0.4MB büyüyor",
  "/proc her şeyi hatırlar",
  "çekirdek uyanık. her zaman.",
  "TCP tekrar göndermeleri tırmanıyor",
  "zsh geçmişin 18.402 satır boyunda",
  "uid 0 uyuyor",
  "yük ritmine kavuşuyor",
  "önbellek sıcak",
  "epoll boşuna 12.4sn bekledi",
  "takas kullanımı %4 — sabırlı",
  "sen kıpırdamadın. ben de.",
  "PID 3402 — bu sensin",
];

const NETWORK_IFS = [
  { name: "lo",      ip: "127.0.0.1/8",     state: "UP",   rxBase: 0.0,  txBase: 0.0 },
  { name: "wlp3s0",  ip: "192.168.1.42/24", state: "UP",   rxBase: 1.4,  txBase: 0.3 },
  { name: "enp2s0",  ip: "—",               state: "DOWN", rxBase: 0.0,  txBase: 0.0 },
  { name: "docker0", ip: "172.17.0.1/16",   state: "UP",   rxBase: 0.02, txBase: 0.01 },
  { name: "wg0",     ip: "10.8.0.4/24",     state: "UP",   rxBase: 0.21, txBase: 0.08 },
];

const CONNECTIONS = [
  { proto:"tcp",  local:"0.0.0.0:22",       remote:"198.51.100.42:51284", state:"ESTABLISHED", proc:"sshd" },
  { proto:"tcp",  local:"127.0.0.1:5432",   remote:"127.0.0.1:38922",     state:"ESTABLISHED", proc:"postgres" },
  { proto:"tcp",  local:"127.0.0.1:6379",   remote:"127.0.0.1:39014",     state:"ESTABLISHED", proc:"redis-server" },
  { proto:"tcp",  local:"0.0.0.0:80",       remote:"203.0.113.7:48910",   state:"ESTABLISHED", proc:"nginx" },
  { proto:"tcp",  local:"0.0.0.0:443",      remote:"203.0.113.7:48911",   state:"ESTABLISHED", proc:"nginx" },
  { proto:"tcp",  local:"0.0.0.0:22",       remote:"*",                   state:"LISTEN",      proc:"sshd" },
  { proto:"tcp",  local:"0.0.0.0:80",       remote:"*",                   state:"LISTEN",      proc:"nginx" },
  { proto:"tcp",  local:"127.0.0.1:5432",   remote:"*",                   state:"LISTEN",      proc:"postgres" },
  { proto:"tcp",  local:"127.0.0.1:6379",   remote:"*",                   state:"LISTEN",      proc:"redis-server" },
  { proto:"udp",  local:"0.0.0.0:53",       remote:"*",                   state:"LISTEN",      proc:"systemd-resolved" },
  { proto:"tcp",  local:"192.168.1.42:52331", remote:"140.82.114.4:443",  state:"ESTABLISHED", proc:"firefox" },
  { proto:"tcp",  local:"192.168.1.42:52340", remote:"104.16.132.229:443",state:"ESTABLISHED", proc:"chrome" },
];

const MOUNTS = [
  { mp:"/",        fs:"btrfs", used: 312,  total: 512,  hot: true },
  { mp:"/home",    fs:"ext4",  used: 580,  total: 1024, hot: false },
  { mp:"/var",     fs:"ext4",  used:  84,  total: 128,  hot: true },
  { mp:"/tmp",     fs:"tmpfs", used:   1,  total:  32,  hot: false },
  { mp:"/boot",    fs:"vfat",  used: 0.2,  total:   1,  hot: false },
];

const SENSORS = [
  { name:"CPU PKG",   tr:"İŞLEMCİ",     val: 64,   unit:"°C",  crit: 95 },
  { name:"CORE0",     tr:"ÇKR-0",       val: 62,   unit:"°C",  crit: 95 },
  { name:"GPU",       tr:"GPU",         val: 51,   unit:"°C",  crit: 87 },
  { name:"NVME0",     tr:"NVME-0",      val: 44,   unit:"°C",  crit: 70 },
  { name:"ANAKART",   tr:"ANAKART",     val: 38,   unit:"°C",  crit: 70 },
  { name:"FAN-1",     tr:"FAN-1",       val:1240,  unit:"rpm", crit: 0 },
  { name:"FAN-2",     tr:"FAN-2",       val: 980,  unit:"rpm", crit: 0 },
  { name:"BATARYA",   tr:"BATARYA",     val: 87,   unit:"%",   crit: 0 },
];

const GPU = {
  name: "NVIDIA RTX 4070",
  driver: "545.29.06",
  util: 23,
  mem_used: 4.2,
  mem_total: 12,
  temp: 51,
  power: 84,
  power_max: 200,
  fan: 38,
  procs: [
    { pid: 21847, name: "firefox", vram: 380 },
    { pid: 24008, name: "java",    vram: 1840 },
    { pid: 18203, name: "code",    vram: 220 },
    { pid: 26113, name: "ffmpeg",  vram: 540 },
  ],
};

/* Komut paleti — komutlar İngilizce, açıklamaları Türkçe */
const COMMANDS = [
  { cmd: ":focus core <n>",   tr: "Çekirdek <n>'e odaklan",          example: ":focus core 2" },
  { cmd: ":focus proc <pid>", tr: "PID'e göre süreç dosyasını aç",   example: ":focus proc 25441" },
  { cmd: ":ack all",          tr: "Tüm kritik alarmları onayla",     example: ":ack all" },
  { cmd: ":replay <s>",       tr: "<s> saniye geri sar",             example: ":replay 30" },
  { cmd: ":live",             tr: "Canlıya dön",                     example: ":live" },
  { cmd: ":filter <sev>",     tr: "Olay akışını seviyeye göre süz",  example: ":filter crit" },
  { cmd: ":clear",            tr: "Alarm sırasını temizle",          example: ":clear" },
  { cmd: ":dim",              tr: "Ekranı kıs",                      example: ":dim" },
  { cmd: ":help",             tr: "Yardımı aç",                      example: ":help" },
];

const KEYS = [
  { k: ":",          tr: "Komut paletini aç" },
  { k: "?",          tr: "Yardımı göster" },
  { k: "j / k",      tr: "Olay akışında ↓ / ↑" },
  { k: "g / G",      tr: "En başa / en sona" },
  { k: "Enter",      tr: "Seçili olayı aç" },
  { k: "/",          tr: "Olay süzgecini değiştir" },
  { k: "1-8",        tr: "Çekirdek 0-7'ye odaklan" },
  { k: "0",          tr: "Odağı bırak" },
  { k: "Esc",        tr: "Kapat · canlıya dön" },
];

/* Türkçe etiket sözlüğü — UI metinleri */
const L = {
  brand:        "GECE NÖBETİ",
  subbrand:     "ana sistem",
  kernel:       "çekirdek",
  uptime:       "çalışma",
  sessions:     "oturum",
  tty:          "tty",
  observed:     "GÖZLEM AKTİF",
  live:         "CANLI",
  recording:    "▮ KAYITTA",
  session:      "OTURUM",
  host:         "MAKİNE",
  uid:          "UID",
  shell:        "kabuk",
  ssh:          "ssh",
  idle:         "boşta",
  loadavg:      "YÜK ORTALAMASI · 1DK",
  mem:          "BELLEK",
  swap:         "TAKAS",
  used:         "KULLANIM",
  buff:         "TAMPON",
  cache:        "ÖNBELLEK",
  free:         "BOŞ",
  disk:         "DİSK",
  mounts:       "bağlama",
  read:         "OKUMA",
  write:        "YAZMA",
  gpu:          "GPU",
  driver:       "sürücü",
  gpu_util:     "SM YÜK",
  vram:         "VRAM",
  power:        "GÜÇ",
  temp:         "SICAKLIK",
  sensors:      "ALGILAYICILAR",
  cores:        "İŞLEMCİ ÇEKİRDEKLERİ",
  topProcs:     "EN AKTİF SÜREÇLER · CPU%",
  load60:       "YÜK · SON 60 SANİYE",
  syslog:       "OLAY AKIŞI",
  syslogSub:    "journalctl · dmesg · auth",
  lines:        "satır",
  all:          "HEPSİ",
  info:         "BİLGİ",
  warn:         "UYARI",
  crit:         "KRİTİK",
  network:      "AĞ · ARAYÜZLER VE BAĞLANTILAR",
  conn:         "BAĞLANTI",
  alerts:       "ALARM SIRASI",
  quiet:        "— sessiz —",
  noAlerts:     "kritik olay yok. çekirdek eşit ritimle nefes alıyor.",
  pending:      "bekleyen",
  process:      "SÜREÇ",
  dossier:      "DOSYA",
  syslogEntry:  "OLAY · KAYIT",
  notes:        "NOTLAR · NÖBET",
  whisper:      "▮ SİSTEM FISILTISI",
  cur:          "İML",
  bpm:          "BPM",
  close:        "KAPAT",
  ack:          "ONAY",
  acknowledge:  "ONAYLA",
  dismiss:      "AT",
  drop:         "DÜŞÜR",
  cpu:          "CPU",
  memTab:       "BELLEK",
  threads:      "İŞ PARÇACIĞI",
  pid:          "PID",
  state:        "DURUM",
  selectCore:   "bir çekirdek seç · yukarıdaki herhangi bir rayı tıkla",
  detail:       "AYRINTI",
  tweakTitle:   "Ayarlar",
  paletteSec:   "Palet",
  toneLbl:      "Ton",
  motionSec:    "Hareket",
  tempoLbl:     "Tempo",
  atmosSec:     "Atmosfer",
  scanlines:    "Tarama çizgileri",
  grain:        "Film tanesi",
  vignette:     "Vinyet",
  flicker:      "Titreme",
  surveillanceSec: "Gözetim",
  eyeLbl:       "Göz + imleç kaydı",
  whispersLbl:  "Sistem fısıltıları",
  motion_still: "Hareketsiz",
  motion_calm:  "Sakin",
  motion_living:"Canlı",
  pal_noir:     "Noir",
  pal_crt:      "CRT",
  pal_blue:     "Mavi Saat",
  pal_amber:    "Amber",
  replay:       "GERİ SARMA",
  replayHint:   "← / → · 1sn ileri-geri  ·  Esc · canlıya dön",
  live2:        "canlı",
  secAgo:       "sn önce",
  you:          "sen",
  cmdTitle:     "KOMUT PALETİ",
  cmdHint:      "komutu yaz · Enter çalıştır · Esc kapat · ↑↓ önceki",
  helpTitle:    "KISAYOLLAR",
  ticker:       [
    "kardas@midwatch-01 ~/code $ git fetch --all",
    "systemctl status nginx · etkin (çalışıyor) açılıştan beri",
    "cron · backup.sh --incremental · 4dk 12sn'de tamamlandı",
    "apt list --upgradable · 14 paket güncellenebilir",
    "tail -f /var/log/syslog · akıyor",
    "ssh kardas@198.51.100.42 · oturum kuruldu",
    "journalctl -f -u sshd · akıyor",
    "kardas@midwatch-01 ~ $ htop · tty2'de",
    "iotop · 7 süreç G/Ç yapıyor",
    "free -h · 12G kullanım / 32G toplam / 320M takas",
    "uptime · yük ortalaması 1.42, 1.36, 1.28",
    "nvidia-smi · %23 yük · 4.2G/12G vram",
    "lsof -i :22 · sshd'ye 3 bağlantı",
    "ps -ef | grep -c kworker · 24 çekirdek işçisi",
  ],
  nightNotes: "Vardiyanın altıncı saati. Çekirdek kendi kendine çalışıyor. Süreçler usulca konuşuyor. Dışarıda şehir izlendiğini bilmiyor. Sen de bilmiyorsun.",
};

window.NWL = {
  HOST, CORES, PROC_TEMPLATES, NETWORK_IFS, CONNECTIONS, MOUNTS, SENSORS, GPU,
  CONTAINERS, SYSLOG_TEMPLATES, SYSLOG_NOTES, SYSLOG_NOTE_GENERIC,
  WHISPERS, COMMANDS, KEYS, L,
  makeProcesses, emitSyslogEvent, seedSyslog,
};
