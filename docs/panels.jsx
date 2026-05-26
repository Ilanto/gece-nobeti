/* Panels — Türkçe etiketler + per-process sparkline + canlı ağ sparkline'ları */

const { useMemo: useMemoP, useState: useStateP, useEffect: useEffectP } = React;

function pad2(n) { return String(n).padStart(2, "0"); }
function fmtClock(d) { return `${pad2(d.getHours())}:${pad2(d.getMinutes())}:${pad2(d.getSeconds())}`; }
function fmtElapsed(ms) {
  const s = Math.floor(ms / 1000);
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const sec = s % 60;
  return `${pad2(h)}:${pad2(m)}:${pad2(sec)}`;
}
function fmtUptime(ms) {
  const s = Math.floor(ms / 1000);
  const d = Math.floor(s / 86400);
  const h = Math.floor((s % 86400) / 3600);
  const m = Math.floor((s % 3600) / 60);
  return `${d}g ${pad2(h)}sa ${pad2(m)}dk`;
}
function fmtTime(ts) {
  const d = new Date(ts);
  return `${pad2(d.getHours())}:${pad2(d.getMinutes())}:${pad2(d.getSeconds())}`;
}

/* ────────── Üst Çubuk ────────── */
function TopBar({ now, boot, sessions, replayOffset }) {
  const HOST = window.NWL.HOST;
  const L = window.NWL.L;
  const displayTime = now - replayOffset * 1000;
  return (
    <div style={{
      display:"grid",
      gridTemplateColumns:"380px 1fr auto auto",
      gap:24, alignItems:"baseline",
      padding:"10px 28px",
      borderBottom:"1px solid var(--rule)",
    }}>
      <div style={{ display:"flex", gap:14, alignItems:"baseline" }}>
        <span className="display" style={{ fontSize:18, letterSpacing:"0.22em", color:"var(--ink)" }}>
          {L.brand}
        </span>
        <span className="dim" style={{ fontSize:10, letterSpacing:"0.16em" }}>
          / {L.subbrand} · {HOST.hostname}
        </span>
      </div>
      <div style={{ color:"var(--ink-3)", fontSize:10, letterSpacing:"0.14em" }}>
        <span className="dim">{L.kernel} </span>
        <span style={{ color:"var(--ink-2)" }}>{HOST.kernel}</span>
        <span className="dim"> · {HOST.arch} · </span>
        <span style={{ color:"var(--ink-2)" }}>{HOST.os}</span>
        <span className="dim"> · ssh </span>
        <span style={{ color:"var(--accent)" }}>{sessions}</span>
        <span className="dim"> {L.sessions} · {L.tty} </span>
        <span style={{ color:"var(--ink-2)" }}>{HOST.tty}</span>
      </div>
      <div style={{ fontSize:10, color:"var(--ink-3)", letterSpacing:"0.14em" }}>
        {L.uptime} <span style={{ color:"var(--ink)" }}>{fmtUptime(now - boot)}</span>
      </div>
      <div className="display" style={{ fontSize:22, color: replayOffset > 0 ? "var(--warn)" : "var(--ink)", letterSpacing:"0.12em" }}>
        {fmtClock(new Date(displayTime))}
        <span className="dim" style={{ fontSize:11, marginLeft:8 }}>
          {replayOffset > 0 ? `−${replayOffset}sn` : "YEREL"}
        </span>
      </div>
    </div>
  );
}

/* ────────── Makine kimliği ────────── */
function HostBadge({ now, boot, idleMs }) {
  const HOST = window.NWL.HOST;
  const L = window.NWL.L;
  return (
    <div style={{ padding:"22px 24px 18px", borderBottom:"1px solid var(--rule)" }}>
      <div style={{ display:"flex", justifyContent:"space-between", alignItems:"baseline" }}>
        <span className="dim" style={{ fontSize:9, letterSpacing:"0.22em" }}>{L.host}</span>
        <span className="dim" style={{ fontSize:9, letterSpacing:"0.18em" }}>{L.uid} {HOST.uid}</span>
      </div>
      <div className="display" style={{ fontSize:30, color:"var(--ink)", letterSpacing:"0.04em", marginTop:6, lineHeight:1 }}>
        {HOST.user}<span style={{ color:"var(--ink-3)" }}>@</span>{HOST.hostname}
      </div>
      <div style={{ display:"grid", gridTemplateColumns:"auto 1fr", gap:"4px 14px", marginTop:16, fontSize:10 }}>
        <span className="dim">{L.uptime}</span> <span style={{ color:"var(--accent)" }}>{fmtUptime(now - boot)}</span>
        <span className="dim">{L.kernel}</span> <span>{HOST.kernel}</span>
        <span className="dim">{L.shell}</span>  <span>{HOST.shell}</span>
        <span className="dim">{L.tty}</span>    <span>{HOST.tty}</span>
        <span className="dim">{L.ssh}</span>    <span style={{ color:"var(--ink-2)" }}>{HOST.ssh}</span>
        <span className="dim">{L.idle}</span>   <span style={{ color: idleMs > 30000 ? "var(--warn)" : "var(--ink)" }}>{fmtElapsed(idleMs)}</span>
      </div>
    </div>
  );
}

/* ────────── Yük nabzı ────────── */
function LoadPulse({ pulseData, loadavg }) {
  const L = window.NWL.L;
  const W = 332, H = 88;
  const N = pulseData.length;
  const path = pulseData.map((p, i) => {
    const x = (i / (N - 1)) * W;
    const y = H - 10 - p.y * (H - 20);
    return `${i === 0 ? "M" : "L"}${x.toFixed(1)},${y.toFixed(1)}`;
  }).join(" ");
  const last = pulseData[pulseData.length - 1] || { y: 0.5 };
  const lastY = H - 10 - last.y * (H - 20);
  return (
    <div style={{ padding:"16px 24px", borderBottom:"1px solid var(--rule)" }}>
      <div style={{ display:"flex", justifyContent:"space-between", alignItems:"baseline", marginBottom:6 }}>
        <span className="dim" style={{ fontSize:9, letterSpacing:"0.22em" }}>{L.loadavg}</span>
        <span className="mono" style={{ fontSize:9, color:"var(--accent)" }}>● {L.live}</span>
      </div>
      <div style={{ display:"flex", gap:18, alignItems:"baseline" }}>
        <span className="display" style={{ fontSize:28, color:"var(--ink)", lineHeight:1 }}>{loadavg[0].toFixed(2)}</span>
        <span className="mono" style={{ fontSize:10, color:"var(--ink-3)" }}>
          5dk {loadavg[1].toFixed(2)} · 15dk {loadavg[2].toFixed(2)}
        </span>
      </div>
      <svg viewBox={`0 0 ${W} ${H}`} width="100%" height={H} style={{ display:"block", marginTop:4 }}>
        {[0.25, 0.5, 0.75].map(p => (
          <line key={p} x1="0" y1={H * p} x2={W} y2={H * p} stroke="var(--rule-2)" strokeWidth="1" strokeDasharray="2 4" />
        ))}
        <path d={path} fill="none" stroke="var(--ink-2)" strokeWidth="1" />
        <circle cx={W - 1} cy={lastY} r="2" fill="var(--accent)" />
        <line x1={W-1} y1="0" x2={W-1} y2={H} stroke="var(--accent)" strokeWidth="0.5" opacity="0.4" />
      </svg>
    </div>
  );
}

/* ────────── Bellek haritası ────────── */
function MemoryMap({ mem }) {
  const L = window.NWL.L;
  const total = mem.total;
  const segs = [
    { k:"used",  tr:L.used,  v: mem.used,  c:"var(--accent)" },
    { k:"buff",  tr:L.buff,  v: mem.buff,  c:"var(--ink-2)" },
    { k:"cache", tr:L.cache, v: mem.cache, c:"var(--ink-3)" },
    { k:"free",  tr:L.free,  v: mem.free,  c:"var(--ink-4)" },
  ];
  const W = 332;
  let cx = 0;
  return (
    <div style={{ padding:"16px 24px", borderBottom:"1px solid var(--rule)" }}>
      <div style={{ display:"flex", justifyContent:"space-between", alignItems:"baseline", marginBottom:8 }}>
        <span className="dim" style={{ fontSize:9, letterSpacing:"0.22em" }}>{L.mem} · {(total/1024).toFixed(1)} GiB</span>
        <span className="mono" style={{ fontSize:9, color:"var(--ink-3)" }}>
          %{((mem.used / total) * 100).toFixed(1)} kullanım
        </span>
      </div>
      <svg viewBox={`0 0 ${W} 10`} width="100%" height="10" style={{ display:"block" }}>
        {segs.map((s) => {
          const w = (s.v / total) * W;
          const el = (
            <rect key={s.k} x={cx} y="0" width={Math.max(0, w-1)} height="10" fill={s.c} opacity={s.k==="used" ? 0.9 : 0.6} />
          );
          cx += w;
          return el;
        })}
      </svg>
      <div style={{ display:"grid", gridTemplateColumns:"repeat(4, 1fr)", gap:4, marginTop:8, fontSize:9 }}>
        {segs.map(s => (
          <div key={s.k} style={{ display:"flex", flexDirection:"column", gap:2 }}>
            <span style={{ display:"flex", alignItems:"center", gap:6 }}>
              <span style={{ width:8, height:8, background:s.c, opacity:s.k==="used"?0.9:0.6, display:"inline-block" }}/>
              <span className="dim" style={{ letterSpacing:"0.12em" }}>{s.tr}</span>
            </span>
            <span className="mono" style={{ color:"var(--ink-2)", fontSize:10 }}>{(s.v/1024).toFixed(2)}G</span>
          </div>
        ))}
      </div>
      <div style={{ marginTop:14 }}>
        <div style={{ display:"flex", justifyContent:"space-between", alignItems:"baseline", marginBottom:4 }}>
          <span className="dim" style={{ fontSize:9, letterSpacing:"0.22em" }}>{L.swap} · {(mem.swap_total/1024).toFixed(1)} GiB</span>
          <span className="mono" style={{ fontSize:9, color:"var(--ink-3)" }}>
            %{((mem.swap_used / mem.swap_total) * 100).toFixed(1)}
          </span>
        </div>
        <svg viewBox={`0 0 ${W} 6`} width="100%" height="6" style={{ display:"block" }}>
          <rect x="0" y="0" width={W} height="6" fill="var(--ink-4)" opacity="0.4" />
          <rect x="0" y="0" width={(mem.swap_used / mem.swap_total) * W} height="6" fill="var(--warn)" opacity="0.7" />
        </svg>
      </div>
    </div>
  );
}

/* ────────── Disk ────────── */
function DiskPanel({ mounts, ioSpark }) {
  const L = window.NWL.L;
  return (
    <div style={{ padding:"14px 24px", borderBottom:"1px solid var(--rule)" }}>
      <div style={{ display:"flex", justifyContent:"space-between", alignItems:"baseline", marginBottom:8 }}>
        <span className="dim" style={{ fontSize:9, letterSpacing:"0.22em" }}>{L.disk} · {mounts.length} {L.mounts}</span>
        <span className="mono" style={{ fontSize:9, color:"var(--ink-3)" }}>
          O {ioSpark.read[ioSpark.read.length-1]?.toFixed(0) || 0}MB/s · Y {ioSpark.write[ioSpark.write.length-1]?.toFixed(0) || 0}MB/s
        </span>
      </div>
      <div style={{ display:"flex", flexDirection:"column", gap:4 }}>
        {mounts.map(m => {
          const pct = (m.used / m.total) * 100;
          const hot = m.hot || pct > 85;
          return (
            <div key={m.mp} style={{ display:"grid", gridTemplateColumns:"60px 1fr auto", gap:8, alignItems:"center", fontSize:10 }}>
              <span className="mono" style={{ color:"var(--ink-2)" }}>{m.mp}</span>
              <div style={{ position:"relative", height:5, background:"var(--ink-4)", opacity:0.5 }}>
                <div style={{
                  position:"absolute", left:0, top:0, bottom:0,
                  width:`${pct}%`,
                  background: hot ? "var(--accent)" : "var(--ink-2)", opacity:0.85,
                }} />
              </div>
              <span className="mono" style={{ color:"var(--ink-3)", fontSize:9, minWidth:80, textAlign:"right" }}>
                {m.used.toFixed(0)}/{m.total.toFixed(0)}G · {m.fs}
              </span>
            </div>
          );
        })}
      </div>
      <div style={{ display:"grid", gridTemplateColumns:"1fr 1fr", gap:12, marginTop:10 }}>
        <MiniSpark label={L.read}  data={ioSpark.read}  color="var(--ink-2)" />
        <MiniSpark label={L.write} data={ioSpark.write} color="var(--accent)" />
      </div>
    </div>
  );
}

function MiniSpark({ label, data, color }) {
  const W = 150, H = 20;
  if (!data || data.length === 0) return null;
  const max = Math.max(...data, 1);
  const path = data.map((v, i) => {
    const x = (i / (data.length - 1)) * W;
    const y = H - 1 - (v / max) * (H - 2);
    return `${i === 0 ? "M" : "L"}${x.toFixed(1)},${y.toFixed(1)}`;
  }).join(" ");
  return (
    <div>
      <div className="dim" style={{ fontSize:8, letterSpacing:"0.18em", marginBottom:2 }}>{label}</div>
      <svg viewBox={`0 0 ${W} ${H}`} width="100%" height={H} style={{ display:"block" }}>
        <path d={path} fill="none" stroke={color} strokeWidth="1" />
      </svg>
    </div>
  );
}

/* ────────── GPU ────────── */
function GpuPanel({ gpu }) {
  const L = window.NWL.L;
  const vramPct = (gpu.mem_used / gpu.mem_total) * 100;
  const powerPct = (gpu.power / gpu.power_max) * 100;
  return (
    <div style={{ padding:"14px 24px", borderBottom:"1px solid var(--rule)" }}>
      <div style={{ display:"flex", justifyContent:"space-between", alignItems:"baseline", marginBottom:8 }}>
        <span className="dim" style={{ fontSize:9, letterSpacing:"0.22em" }}>{L.gpu}</span>
        <span className="mono" style={{ fontSize:9, color:"var(--ink-3)" }}>{L.driver} {gpu.driver}</span>
      </div>
      <div className="display" style={{ fontSize:14, color:"var(--ink-2)", letterSpacing:"0.04em" }}>
        {gpu.name}
      </div>
      <div style={{ display:"grid", gridTemplateColumns:"1fr 1fr", gap:14, marginTop:10 }}>
        <Gauge label={L.gpu_util} pct={gpu.util}     unit="%" />
        <Gauge label={L.vram}     pct={vramPct}      unit={`${gpu.mem_used}/${gpu.mem_total}G`} />
        <Gauge label={L.power}    pct={powerPct}     unit={`${gpu.power}W`} />
        <Gauge label={L.temp}     pct={(gpu.temp/87)*100} unit={`${gpu.temp}°C`} />
      </div>
      <div style={{ marginTop:10 }}>
        {gpu.procs.slice(0, 3).map(p => (
          <div key={p.pid} style={{ display:"grid", gridTemplateColumns:"1fr auto auto", gap:8, fontSize:9, color:"var(--ink-3)" }}>
            <span className="mono">{p.name}</span>
            <span className="mono">{p.pid}</span>
            <span className="mono">{p.vram}MB</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function Gauge({ label, pct, unit }) {
  const v = Math.max(0, Math.min(100, pct));
  return (
    <div>
      <div className="dim" style={{ fontSize:9, letterSpacing:"0.16em" }}>{label}</div>
      <div style={{ display:"flex", alignItems:"baseline", gap:6 }}>
        <span className="display" style={{ fontSize:18, color:"var(--ink)", lineHeight:1 }}>{v.toFixed(0)}</span>
        <span className="dim" style={{ fontSize:9 }}>{unit}</span>
      </div>
      <div style={{ height:3, background:"var(--ink-4)", opacity:0.5, marginTop:4 }}>
        <div style={{ height:"100%", width:`${v}%`, background: v > 80 ? "var(--accent)" : "var(--ink-2)", opacity:0.85 }} />
      </div>
    </div>
  );
}

/* ────────── Algılayıcılar ────────── */
function SensorsPanel({ sensors }) {
  const L = window.NWL.L;
  return (
    <div style={{ padding:"14px 24px", flex:1, display:"flex", flexDirection:"column" }}>
      <div className="dim" style={{ fontSize:9, letterSpacing:"0.22em", marginBottom:8 }}>{L.sensors}</div>
      <div style={{ display:"grid", gridTemplateColumns:"1fr 1fr", gap:"6px 16px" }}>
        {sensors.map(s => {
          const hot = s.crit > 0 && s.val > s.crit * 0.85;
          return (
            <div key={s.name} style={{ display:"flex", justifyContent:"space-between", alignItems:"baseline", fontSize:10 }}>
              <span className="dim" style={{ letterSpacing:"0.1em" }}>{s.tr}</span>
              <span className="mono" style={{ color: hot ? "var(--warn)" : "var(--ink-2)" }}>
                {s.val}{s.unit}
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
}

/* ────────── Büyük rakamlar ────────── */
function BigNumbersLinux({ stats, spark }) {
  const items = [
    { k:"YÜK · 1DK",      v: stats.load[0].toFixed(2), u:"",     s: spark.load1 },
    { k:"YÜK · 5DK",      v: stats.load[1].toFixed(2), u:"",     s: spark.load5 },
    { k:"BELLEK · DOLULUK", v: stats.memPct.toFixed(1), u:"%",   s: spark.mem },
    { k:"DİSK G/Ç",       v: stats.diskIO.toFixed(0),  u:"MB/SN",s: spark.disk },
    { k:"AĞ · ALMA/GÖND.",v: stats.netIO.toFixed(1),   u:"MB/SN",s: spark.net },
    { k:"BAĞLAM DEĞ.",    v: stats.ctxsw,              u:"BIN/SN",s: spark.ctx },
  ];
  return (
    <div style={{
      display:"grid",
      gridTemplateColumns:"repeat(6, 1fr)",
      borderTop:"1px solid var(--rule)",
    }}>
      {items.map((it, i) => (
        <div key={i} style={{
          padding:"14px 16px",
          borderRight: i < items.length-1 ? "1px solid var(--rule)" : "none",
          display:"flex", flexDirection:"column", gap:4,
        }}>
          <div className="dim" style={{ fontSize:9, letterSpacing:"0.16em" }}>{it.k}</div>
          <div style={{ display:"flex", alignItems:"baseline", gap:8 }}>
            <span className="display" style={{ fontSize:34, color:"var(--ink)", lineHeight:1, letterSpacing:"0.02em" }}>{it.v}</span>
            <span className="dim" style={{ fontSize:9, letterSpacing:"0.16em" }}>{it.u}</span>
          </div>
          <Spark data={it.s} />
        </div>
      ))}
    </div>
  );
}

function Spark({ data }) {
  const W = 160, H = 18;
  if (!data || data.length === 0) return <div style={{ height: H }} />;
  const min = Math.min(...data), max = Math.max(...data);
  const range = max - min || 1;
  const path = data.map((v, i) => {
    const x = (i / (data.length - 1)) * W;
    const y = H - 2 - ((v - min) / range) * (H - 4);
    return `${i === 0 ? "M" : "L"}${x.toFixed(1)},${y.toFixed(1)}`;
  }).join(" ");
  return (
    <svg viewBox={`0 0 ${W} ${H}`} width="100%" height={H} style={{ display:"block", marginTop:2 }}>
      <path d={path} fill="none" stroke="var(--ink-3)" strokeWidth="1" />
    </svg>
  );
}

/* ────────── Olay akışı ────────── */
function SyslogStream({ events, onSelect, filter, setFilter, selectedIdx }) {
  const L = window.NWL.L;
  const filtered = useMemoP(() => {
    if (filter === "all") return events;
    return events.filter(e => e.sev === filter);
  }, [events, filter]);

  const sevLabel = { all: L.all, info: L.info, warn: L.warn, crit: L.crit };
  const listRef = React.useRef(null);

  // scroll to selectedIdx
  useEffectP(() => {
    if (!listRef.current || selectedIdx == null) return;
    const child = listRef.current.children[selectedIdx];
    if (child && typeof child.scrollIntoView === "function") {
      // scroll within container only
      const c = listRef.current;
      const t = child.offsetTop - c.offsetTop;
      c.scrollTop = Math.max(0, t - 80);
    }
  }, [selectedIdx]);

  return (
    <div style={{ display:"flex", flexDirection:"column", height:"100%", overflow:"hidden" }}>
      <div style={{ padding:"18px 24px 12px", borderBottom:"1px solid var(--rule)" }}>
        <div style={{ display:"flex", justifyContent:"space-between", alignItems:"baseline" }}>
          <div>
            <span className="dim" style={{ fontSize:9, letterSpacing:"0.22em" }}>{L.syslog}</span>
            <span className="dim" style={{ fontSize:8, letterSpacing:"0.14em", marginLeft:10 }}>{L.syslogSub}</span>
          </div>
          <span className="mono" style={{ fontSize:9, color:"var(--ink-3)" }}>{filtered.length} {L.lines}</span>
        </div>
        <div style={{ display:"flex", gap:14, marginTop:10, fontSize:9, letterSpacing:"0.14em" }}>
          {["all","info","warn","crit"].map(f => (
            <button key={f}
              onClick={() => setFilter(f)}
              style={{
                color: filter===f ? "var(--accent)" : "var(--ink-3)",
                borderBottom: filter===f ? "1px solid var(--accent)" : "1px solid transparent",
                paddingBottom: 2,
                letterSpacing:"0.18em",
              }}>{sevLabel[f]}</button>
          ))}
        </div>
      </div>
      <div ref={listRef} style={{ flex:1, overflowY:"auto" }}>
        {filtered.slice(0, 100).map((e, idx) => {
          const c = e.sev === "crit" ? "var(--crit)" : e.sev === "warn" ? "var(--warn)" : "var(--ink-2)";
          const mark = e.sev === "crit" ? "▮▮" : e.sev === "warn" ? "▮ " : "  ";
          const isSelected = idx === selectedIdx;
          return (
            <button key={e.id}
              onClick={() => onSelect(e)}
              style={{
                width:"100%", textAlign:"left",
                padding:"6px 24px",
                display:"grid",
                gridTemplateColumns:"68px 26px 54px 1fr",
                gap:8, alignItems:"baseline",
                borderBottom:"1px solid var(--rule-2)",
                opacity: e.ack ? 0.55 : 1,
                background: isSelected ? "var(--bg-2)" : "transparent",
                borderLeft: isSelected ? "2px solid var(--accent)" : "2px solid transparent",
              }}
              onMouseEnter={ev => { if (!isSelected) ev.currentTarget.style.background = "var(--bg-2)"; }}
              onMouseLeave={ev => { if (!isSelected) ev.currentTarget.style.background = "transparent"; }}
            >
              <span className="mono" style={{ fontSize:9, color:"var(--ink-3)" }}>{fmtTime(e.ts)}</span>
              <span className="mono" style={{ fontSize:9, color:c }}>{mark}</span>
              <span className="mono" style={{ fontSize:9, color:"var(--ink-3)", letterSpacing:"0.1em" }}>{e.fac}</span>
              <span className="mono" style={{ fontSize:10, color:c, textOverflow:"ellipsis", overflow:"hidden", whiteSpace:"nowrap" }}>
                {e.text}
              </span>
            </button>
          );
        })}
      </div>
    </div>
  );
}

/* ────────── Ağ ────────── */
function ConnectionsPanel({ conns, ifs, ifSparks }) {
  const L = window.NWL.L;
  return (
    <div style={{ borderTop:"1px solid var(--rule)" }}>
      <div style={{ padding:"14px 24px 8px", display:"flex", justifyContent:"space-between", alignItems:"baseline" }}>
        <span className="dim" style={{ fontSize:9, letterSpacing:"0.22em" }}>{L.network}</span>
        <span className="mono" style={{ fontSize:9, color:"var(--ink-3)" }}>{conns.length} {L.conn}</span>
      </div>
      <div style={{ padding:"0 24px 8px", display:"flex", flexDirection:"column", gap:4 }}>
        {ifs.map(i => {
          const spark = ifSparks[i.name] || { rx: [], tx: [] };
          const last_rx = spark.rx[spark.rx.length-1] || 0;
          const last_tx = spark.tx[spark.tx.length-1] || 0;
          const down = i.state === "DOWN";
          return (
            <div key={i.name} style={{ display:"grid", gridTemplateColumns:"60px 110px 1fr 80px", gap:8, fontSize:10, alignItems:"center" }}>
              <span className="mono" style={{ color: down ? "var(--ink-4)" : "var(--ink-2)" }}>{i.name}</span>
              <span className="mono" style={{ color:"var(--ink-3)", fontSize:9 }}>{i.ip}</span>
              <NetSpark rx={spark.rx} tx={spark.tx} />
              <span className="mono" style={{ color:"var(--ink-3)", fontSize:9, textAlign:"right" }}>
                ↓{last_rx.toFixed(1)} ↑{last_tx.toFixed(1)}
              </span>
            </div>
          );
        })}
      </div>
      <div style={{ maxHeight: 130, overflowY:"auto", borderTop:"1px solid var(--rule-2)" }}>
        {conns.slice(0, 12).map((c, i) => (
          <div key={i} style={{
            padding:"4px 24px",
            display:"grid",
            gridTemplateColumns:"32px 1fr 1fr 80px 80px",
            gap:8, alignItems:"baseline",
            fontSize:9,
            borderBottom:"1px solid var(--rule-2)",
          }}>
            <span className="mono dim">{c.proto}</span>
            <span className="mono" style={{ color:"var(--ink-2)" }}>{c.local}</span>
            <span className="mono" style={{ color:"var(--ink-3)" }}>{c.remote}</span>
            <span className="mono" style={{ color: c.state==="ESTABLISHED" ? "var(--accent)" : "var(--ink-4)", letterSpacing:"0.1em" }}>
              {c.state}
            </span>
            <span className="mono" style={{ color:"var(--ink-3)", textAlign:"right" }}>{c.proc}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function NetSpark({ rx, tx }) {
  const W = 200, H = 18;
  if (rx.length === 0) return <svg viewBox={`0 0 ${W} ${H}`} width="100%" height={H} />;
  const max = Math.max(0.1, ...rx, ...tx);
  const mid = H / 2;
  // rx above midline (mirrored), tx below
  const rxPath = rx.map((v, i) => {
    const x = (i / (rx.length - 1)) * W;
    const y = mid - (v / max) * (mid - 1);
    return `${i === 0 ? "M" : "L"}${x.toFixed(1)},${y.toFixed(1)}`;
  }).join(" ");
  const txPath = tx.map((v, i) => {
    const x = (i / (tx.length - 1)) * W;
    const y = mid + (v / max) * (mid - 1);
    return `${i === 0 ? "M" : "L"}${x.toFixed(1)},${y.toFixed(1)}`;
  }).join(" ");
  return (
    <svg viewBox={`0 0 ${W} ${H}`} width="100%" height={H} preserveAspectRatio="none" style={{ display:"block" }}>
      <line x1="0" y1={mid} x2={W} y2={mid} stroke="var(--rule-2)" strokeWidth="0.5" />
      <path d={rxPath} fill="none" stroke="var(--ink-2)" strokeWidth="1" />
      <path d={txPath} fill="none" stroke="var(--accent)" strokeWidth="1" />
    </svg>
  );
}

/* ────────── Alarm sırası ────────── */
function AlertQueue({ alerts, onAck, onDismiss }) {
  const L = window.NWL.L;
  return (
    <div style={{ borderTop:"1px solid var(--rule)" }}>
      <div style={{ padding:"14px 24px 10px", display:"flex", justifyContent:"space-between", alignItems:"baseline" }}>
        <span className="dim" style={{ fontSize:9, letterSpacing:"0.22em" }}>{L.alerts}</span>
        <span className="mono" style={{ fontSize:9, color: alerts.length ? "var(--crit)" : "var(--ink-3)" }}>
          {alerts.length === 0 ? L.quiet : `${alerts.length} ${L.pending}`}
        </span>
      </div>
      <div style={{ maxHeight: 150, overflowY:"auto" }}>
        {alerts.length === 0 && (
          <div style={{ padding:"18px 24px", color:"var(--ink-4)", fontFamily:"Newsreader", fontStyle:"italic", fontSize:12 }}>
            {L.noAlerts}
          </div>
        )}
        {alerts.map(a => (
          <div key={a.id} style={{
            padding:"8px 24px",
            borderBottom:"1px solid var(--rule-2)",
            display:"grid",
            gridTemplateColumns:"1fr auto auto",
            gap:8, alignItems:"center",
          }}>
            <div>
              <div style={{ display:"flex", gap:8, alignItems:"baseline" }}>
                <span className="mono" style={{ fontSize:9, color:"var(--crit)", letterSpacing:"0.16em" }}>{L.crit}</span>
                <span className="mono" style={{ fontSize:10, color:"var(--ink-3)", letterSpacing:"0.08em" }}>{a.fac}</span>
                <span className="mono" style={{ fontSize:10, color:"var(--ink-2)" }}>{fmtTime(a.ts)}</span>
              </div>
              <div className="mono" style={{ fontSize:10, color:"var(--ink)", marginTop:2, textOverflow:"ellipsis", overflow:"hidden", whiteSpace:"nowrap" }}>
                {a.text}
              </div>
            </div>
            <button onClick={() => onAck(a.id)} style={{
              padding:"4px 10px", fontSize:9, letterSpacing:"0.18em",
              color:"var(--accent)", border:"1px solid var(--accent-2)",
            }}>{L.acknowledge}</button>
            <button onClick={() => onDismiss(a.id)} style={{
              padding:"4px 8px", fontSize:9, letterSpacing:"0.18em",
              color:"var(--ink-3)", border:"1px solid var(--rule)",
            }}>{L.drop}</button>
          </div>
        ))}
      </div>
    </div>
  );
}

/* ────────── Alt ticker ────────── */
function BottomTicker({ feed }) {
  return (
    <div style={{
      borderTop:"1px solid var(--rule)",
      padding:"6px 0",
      overflow:"hidden",
      height: 28,
    }}>
      <div style={{
        display:"flex", gap:48, whiteSpace:"nowrap",
        animation: "scroll-x 110s linear infinite",
      }}>
        {[...feed, ...feed].map((f, i) => (
          <span key={i} className="mono" style={{ fontSize:10, color:"var(--ink-3)", letterSpacing:"0.1em" }}>
            <span style={{ color:"var(--ink-4)" }}>$</span> {f}
          </span>
        ))}
      </div>
      <style>{`
        @keyframes scroll-x {
          from { transform: translateX(0); }
          to   { transform: translateX(-50%); }
        }
      `}</style>
    </div>
  );
}

/* ────────── Süreç dosyası ────────── */
function ProcessOverlay({ process, onClose, history }) {
  const L = window.NWL.L;
  if (!process) return null;
  const histPath = (history || []).map((v, i, arr) => {
    const W = 560, H = 40;
    const x = (i / Math.max(1, arr.length - 1)) * W;
    const y = H - 2 - (v / 100) * (H - 4);
    return `${i === 0 ? "M" : "L"}${x.toFixed(1)},${y.toFixed(1)}`;
  }).join(" ");

  const operatorNote = "PID 3402 — bu süreç sensin. Çekirdeğin senin için tuttuğu kayıt, şu anki bu satırı da içeriyor.";
  const genericNote = `"${process.name} bir süredir uyanık. Çekirdeğe sistem çağrılarıyla konuşuyor, aralarda kibarca bekliyor."`;

  return (
    <div onClick={onClose} style={{
      position:"fixed", inset:0, background:"rgba(0,0,0,0.7)",
      zIndex: 8000, display:"flex", alignItems:"center", justifyContent:"center",
      backdropFilter:"blur(2px)",
    }}>
      <div onClick={e => e.stopPropagation()} style={{
        width:680, background:"var(--bg-2)", border:"1px solid var(--rule)",
        padding:"34px 36px",
      }}>
        <div className="dim" style={{ fontSize:9, letterSpacing:"0.22em" }}>{L.process} · {L.dossier}</div>
        <div style={{ display:"flex", gap:14, alignItems:"baseline", marginTop:6 }}>
          <span className="display" style={{ fontSize:42, color:"var(--accent)", letterSpacing:"0.04em", lineHeight:1 }}>
            {process.name}
          </span>
          <span className="mono" style={{ fontSize:14, color:"var(--ink-3)" }}>pid {process.pid}</span>
          {process.isOperator && (
            <span className="mono" style={{ fontSize:11, color:"var(--accent)", letterSpacing:"0.2em",
              border:"1px solid var(--accent-2)", padding:"2px 8px" }}>BU SENSİN</span>
          )}
          {process.container && (
            <span className="mono" style={{ fontSize:10, color:"var(--ink-3)", fontStyle:"italic" }}>
              [docker:{process.container}]
            </span>
          )}
        </div>
        <div className="dim" style={{ fontSize:10, letterSpacing:"0.14em", marginTop:6 }}>
          {process.kind === "kernel" ? "ÇEKİRDEK" : process.kind === "system" ? "SİSTEM" : "KULLANICI"}
          {" · "}ÇEKİRDEK CPU{process.core}{" · "}DURUM {process.state}
        </div>
        <div style={{ marginTop:24, display:"grid", gridTemplateColumns:"1fr 1fr 1fr 1fr", gap:18 }}>
          <Dossier label={L.cpu}      value={`%${process.cpuLive.toFixed(1)}`} />
          <Dossier label={L.memTab}   value={`${process.mem} MB`} />
          <Dossier label={L.threads}  value={`${process.threads}`} />
          <Dossier label={L.pid}      value={`${process.pid}`} />
        </div>

        {history && history.length > 0 && (
          <div style={{ marginTop:24 }}>
            <div className="dim" style={{ fontSize:9, letterSpacing:"0.22em", marginBottom:6 }}>
              CPU% · SON 60 SANİYE
            </div>
            <svg viewBox="0 0 560 40" width="100%" height="40" style={{ display:"block" }}>
              <line x1="0" y1="38" x2="560" y2="38" stroke="var(--rule)" strokeWidth="1" />
              {[25,50,75].map(t => (
                <line key={t} x1="0" y1={40 - 2 - (t/100)*36} x2="560" y2={40 - 2 - (t/100)*36} stroke="var(--rule-2)" strokeDasharray="2 4" />
              ))}
              <path d={histPath} fill="none" stroke="var(--accent)" strokeWidth="1" />
            </svg>
          </div>
        )}

        <div style={{ marginTop:24, paddingTop:18, borderTop:"1px solid var(--rule)" }}>
          <div className="dim" style={{ fontSize:9, letterSpacing:"0.22em", marginBottom:8 }}>{L.notes}</div>
          <div className="whisper" style={{ fontSize:14, color:"var(--ink-2)", lineHeight:1.5 }}>
            {process.isOperator ? operatorNote : genericNote}
          </div>
          {process.isOperator && (
            <div className="mono" style={{ fontSize:9, color:"var(--ink-4)", marginTop:10, letterSpacing:"0.08em" }}>
              /proc/{process.pid}/status · /proc/{process.pid}/cmdline · /proc/{process.pid}/fd/
            </div>
          )}
        </div>
        <div style={{ display:"flex", justifyContent:"flex-end", gap:10, marginTop:28 }}>
          <button onClick={onClose} style={{
            padding:"8px 16px", fontSize:10, letterSpacing:"0.22em",
            color:"var(--ink-3)", border:"1px solid var(--rule)",
          }}>{L.close}</button>
        </div>
      </div>
    </div>
  );
}

function Dossier({ label, value }) {
  return (
    <div>
      <div className="dim" style={{ fontSize:9, letterSpacing:"0.18em" }}>{label}</div>
      <div className="mono" style={{ fontSize:14, color:"var(--ink)", marginTop:4 }}>{value}</div>
    </div>
  );
}

/* ────────── Olay dosyası ────────── */
function SyslogOverlay({ event, onClose, onAck }) {
  const L = window.NWL.L;
  if (!event) return null;
  const c = event.sev === "crit" ? "var(--crit)" : event.sev === "warn" ? "var(--warn)" : "var(--ink-2)";
  const sevLabel = { info: L.info, warn: L.warn, crit: L.crit };
  const note = window.NWL.SYSLOG_NOTES[event.text] || window.NWL.SYSLOG_NOTE_GENERIC;
  return (
    <div onClick={onClose} style={{
      position:"fixed", inset:0, background:"rgba(0,0,0,0.7)",
      zIndex: 8000, display:"flex", alignItems:"center", justifyContent:"center",
      backdropFilter:"blur(2px)",
    }}>
      <div onClick={e => e.stopPropagation()} style={{
        width:760, background:"var(--bg-2)", border:"1px solid var(--rule)",
        padding:"34px 36px",
      }}>
        <div className="dim" style={{ fontSize:9, letterSpacing:"0.22em" }}>{L.syslogEntry}</div>
        <div style={{ display:"flex", gap:18, alignItems:"baseline", marginTop:6 }}>
          <span className="mono" style={{ fontSize:11, color:c, letterSpacing:"0.18em" }}>
            {sevLabel[event.sev] || event.sev.toUpperCase()}
          </span>
          <span className="display" style={{ fontSize:24, color:"var(--accent)", letterSpacing:"0.08em" }}>
            {event.fac.toUpperCase()}
          </span>
          <span className="mono" style={{ fontSize:14, color:"var(--ink-3)" }}>{fmtTime(event.ts)}</span>
          {event.ack && (
            <span className="mono" style={{ fontSize:10, color:"var(--ok)", letterSpacing:"0.2em", padding:"2px 8px", border:"1px solid var(--ok)" }}>
              ONAYLANDI
            </span>
          )}
        </div>
        <div className="mono" style={{ fontSize:14, color:"var(--ink)", marginTop:18, padding:"14px 16px",
            background:"var(--bg)", border:"1px solid var(--rule)" }}>
          {event.text}
        </div>
        <div style={{ marginTop:24, paddingTop:18, borderTop:"1px solid var(--rule)" }}>
          <div className="dim" style={{ fontSize:9, letterSpacing:"0.22em", marginBottom:8 }}>ÇÖZÜMLEME · NÖBET</div>
          <div className="whisper" style={{ fontSize:14, color:"var(--ink-2)", lineHeight:1.6 }}>
            {note}
          </div>
        </div>
        <div style={{ display:"flex", justifyContent:"flex-end", gap:10, marginTop:28 }}>
          <button onClick={onClose} style={{
            padding:"8px 16px", fontSize:10, letterSpacing:"0.22em",
            color:"var(--ink-3)", border:"1px solid var(--rule)",
          }}>{L.close}</button>
          {!event.ack && (
            <button onClick={() => { onAck(event.id); onClose(); }} style={{
              padding:"8px 16px", fontSize:10, letterSpacing:"0.22em",
              color:"var(--accent)", border:"1px solid var(--accent-2)",
            }}>{L.acknowledge}</button>
          )}
        </div>
      </div>
    </div>
  );
}

/* ────────── Replay scrubber ────────── */
function ReplayScrubber({ offset, max, onChange, onLive }) {
  const L = window.NWL.L;
  if (max <= 0) return null;
  const live = offset === 0;
  return (
    <div style={{
      display:"flex", alignItems:"center", gap:12,
      padding:"6px 10px",
      border:"1px solid " + (live ? "var(--rule)" : "var(--accent-2)"),
      background: live ? "transparent" : "var(--bg-2)",
    }}>
      <span className="mono" style={{ fontSize:9, color: live ? "var(--ink-3)" : "var(--accent)", letterSpacing:"0.2em" }}>
        {live ? L.live2.toUpperCase() : L.replay}
      </span>
      <input
        type="range" min="0" max={max} step="1" value={offset}
        onChange={(e) => onChange(parseInt(e.target.value))}
        style={{
          width: 220, height: 4, accentColor: "var(--accent)",
          background: "var(--rule)", appearance:"none", outline:"none",
        }}
      />
      <span className="mono" style={{ fontSize:9, color: live ? "var(--ink-4)" : "var(--warn)", minWidth: 56 }}>
        {live ? "—" : `−${offset} ${L.secAgo}`}
      </span>
      {!live && (
        <button onClick={onLive} style={{
          fontSize:9, letterSpacing:"0.18em", color:"var(--accent)",
          border:"1px solid var(--accent-2)", padding:"2px 8px",
        }}>{L.live2.toUpperCase()}</button>
      )}
    </div>
  );
}

/* ────────── Boşta karartma ────────── */
function IdleOverlay({ idleMs, threshold = 90000 }) {
  if (idleMs < threshold) return null;
  const lines = [
    "sistem hâlâ burada.",
    "sistem hâlâ izliyor.",
    "kıpırda. yoksa kayıt devam eder.",
  ];
  const idx = Math.floor(idleMs / 12000) % lines.length;
  return (
    <div style={{
      position:"fixed", inset:0, zIndex:5000,
      background:"rgba(0,0,0,0.78)",
      display:"flex", alignItems:"center", justifyContent:"center",
      pointerEvents:"none",
      backdropFilter:"blur(1px)",
    }}>
      <div className="whisper" style={{
        fontSize: 38, color:"var(--ink-2)", letterSpacing:"0.08em",
        textAlign:"center",
        textShadow:"0 0 32px rgba(0,0,0,0.8)",
      }}>
        "{lines[idx]}"
      </div>
    </div>
  );
}

Object.assign(window, {
  L_TopBar: TopBar,
  L_HostBadge: HostBadge,
  L_LoadPulse: LoadPulse,
  L_MemoryMap: MemoryMap,
  L_DiskPanel: DiskPanel,
  L_GpuPanel: GpuPanel,
  L_SensorsPanel: SensorsPanel,
  L_BigNumbers: BigNumbersLinux,
  L_SyslogStream: SyslogStream,
  L_ConnectionsPanel: ConnectionsPanel,
  L_AlertQueue: AlertQueue,
  L_BottomTicker: BottomTicker,
  L_ProcessOverlay: ProcessOverlay,
  L_SyslogOverlay: SyslogOverlay,
  L_ReplayScrubber: ReplayScrubber,
  L_IdleOverlay: IdleOverlay,
  fmtClock, fmtTime, fmtElapsed, fmtUptime,
});
