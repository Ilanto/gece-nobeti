/* Gece Nöbeti · Ana Sistem — ana uygulama */

const { useState, useEffect, useRef, useMemo, useCallback } = React;

const STAGE_W = 2560;
const STAGE_H = 1080;
const HISTORY_LEN = 60;        // saniye sayısı
const PROC_HISTORY_LEN = 60;   // her süreç için 60 sn CPU geçmişi

function App() {
  const L = window.NWL.L;

  const TWEAK_DEFAULTS = /*EDITMODE-BEGIN*/{
    "palette": "noir",
    "motion": "calm",
    "scanlines": true,
    "grain": true,
    "vignette": true,
    "flicker": true,
    "showSurveillance": true,
    "showWhispers": true,
    "idleDim": true
  }/*EDITMODE-END*/;
  const [tweaks, setTweak] = window.useTweaks(TWEAK_DEFAULTS);

  useEffect(() => {
    document.body.dataset.palette  = tweaks.palette;
    document.body.dataset.scanlines = tweaks.scanlines ? "on" : "off";
    document.body.dataset.grain    = tweaks.grain     ? "on" : "off";
    document.body.dataset.vignette = tweaks.vignette  ? "on" : "off";
    document.body.dataset.flicker  = tweaks.flicker   ? "on" : "off";
  }, [tweaks]);

  /* Stage scaling */
  const [scale, setScale] = useState(1);
  useEffect(() => {
    const handle = () => {
      const sx = window.innerWidth / STAGE_W;
      const sy = window.innerHeight / STAGE_H;
      setScale(Math.min(sx, sy));
    };
    handle();
    window.addEventListener("resize", handle);
    return () => window.removeEventListener("resize", handle);
  }, []);

  /* Clock */
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(id);
  }, []);

  /* Processes */
  const [liveProcesses, setLiveProcesses] = useState(() => window.NWL.makeProcesses());
  const motionMul = tweaks.motion === "still" ? 0 : tweaks.motion === "calm" ? 1 : 2.0;

  useEffect(() => {
    if (motionMul === 0) return;
    const id = setInterval(() => {
      setLiveProcesses(prev => prev.map(p => {
        if (Math.random() < 0.04 * motionMul) {
          const base = p.isOperator ? 0.1 + Math.random() * 1.2
                     : p.kind === "kernel" ? 1 + Math.random()*5
                     : p.kind === "system" ? 1 + Math.random()*15
                     : 4 + Math.random()*78;
          p = { ...p, cpuTarget: base };
        }
        const diff = (p.cpuTarget - p.cpuLive);
        const cpuLive = Math.max(0, Math.min(100, p.cpuLive + diff * 0.12 + (Math.random()-0.5) * 2.5 * motionMul));
        return { ...p, cpuLive };
      }));
    }, 160);
    return () => clearInterval(id);
  }, [motionMul]);

  /* Core loads from live processes */
  const liveCoreLoads = useMemo(() => {
    const cores = window.NWL.CORES;
    const loads = cores.map(() => 0);
    liveProcesses.forEach(p => { loads[p.core] = (loads[p.core] || 0) + p.cpuLive; });
    return loads.map(l => Math.max(0.5, Math.min(100, l + 2)));
  }, [liveProcesses]);

  /* ────────── REPLAY BUFFER ──────────
     Her saniye liveProcesses snapshot'ını + core loads + mem'i kaydet */
  const [snapshots, setSnapshots] = useState([]);  // [{ ts, procs, loads, mem }, ...] (en yenisi en sonda)
  const [replayOffset, setReplayOffset] = useState(0); // saniye geriye

  /* ────────── Diğer state'ler ────────── */
  const [coreHistory, setCoreHistory] = useState(() =>
    window.NWL.CORES.map(() => Array.from({length: 60}, (_, i) => 20 + Math.sin(i/4)*8 + Math.random()*8))
  );
  const [procHistory, setProcHistory] = useState(() => {
    const m = {};
    liveProcesses.forEach(p => { m[p.uid] = Array.from({length: PROC_HISTORY_LEN}, () => Math.max(0, p.cpu + (Math.random()-0.5)*8)); });
    return m;
  });

  // Tick: capture snapshot + update histories
  const [mem, setMem] = useState({
    total: 32768, used: 12480, buff: 1480, cache: 8420, free: 10388,
    swap_total: 8192, swap_used: 320,
  });

  useEffect(() => {
    if (motionMul === 0) return;
    const id = setInterval(() => {
      setMem(prev => {
        const drift = (Math.random() - 0.5) * 80 * motionMul;
        const used = Math.max(8000, Math.min(20000, prev.used + drift));
        const buff = prev.buff + (Math.random()-0.5) * 20;
        const cache = prev.cache + (Math.random()-0.5) * 60;
        const free = prev.total - used - buff - cache;
        const swap_used = Math.max(0, Math.min(prev.swap_total, prev.swap_used + (Math.random()-0.5) * 4));
        return { ...prev, used, buff, cache, free, swap_used };
      });
    }, 1400);
    return () => clearInterval(id);
  }, [motionMul]);

  // 1-saniyelik history & snapshot tick
  useEffect(() => {
    const id = setInterval(() => {
      setCoreHistory(prev => prev.map((arr, i) => [...arr.slice(1), liveCoreLoads[i] || 0]));
      setProcHistory(prev => {
        const next = { ...prev };
        liveProcesses.forEach(p => {
          const arr = next[p.uid] || Array.from({length: PROC_HISTORY_LEN}, () => 0);
          next[p.uid] = [...arr.slice(1), p.cpuLive];
        });
        return next;
      });
      setSnapshots(prev => {
        const snap = {
          ts: Date.now(),
          procs: liveProcesses.map(p => ({ ...p })),  // shallow copy each
          loads: [...liveCoreLoads],
          mem: { ...mem },
        };
        return [...prev.slice(-HISTORY_LEN + 1), snap];
      });
    }, 1000);
    return () => clearInterval(id);
  }, [liveProcesses, liveCoreLoads, mem]);

  /* Replay-aware processes & loads */
  const replayMax = Math.max(0, snapshots.length - 1);
  const usingReplay = replayOffset > 0 && snapshots.length > 0;
  const replaySnap = usingReplay
    ? snapshots[Math.max(0, snapshots.length - 1 - replayOffset)]
    : null;
  const processes  = replaySnap ? replaySnap.procs : liveProcesses;
  const coreLoads  = replaySnap ? replaySnap.loads : liveCoreLoads;
  const displayMem = replaySnap ? replaySnap.mem   : mem;

  /* Syslog */
  const [syslog, setSyslog] = useState(() => window.NWL.seedSyslog(80).reverse());
  const [alerts, setAlerts] = useState([]);
  const [filter, setFilter] = useState("all");

  useEffect(() => {
    if (motionMul === 0) return;
    const baseDelay = tweaks.motion === "living" ? 1600 : 3000;
    const id = setInterval(() => {
      if (Math.random() < 0.18) return;
      const e = window.NWL.emitSyslogEvent();
      setSyslog(prev => [e, ...prev].slice(0, 400));
      if (e.sev === "crit") setAlerts(prev => [e, ...prev].slice(0, 6));
    }, baseDelay);
    return () => clearInterval(id);
  }, [motionMul, tweaks.motion]);

  /* Load avg pulse */
  const [pulseData, setPulseData] = useState(() =>
    Array.from({ length: 64 }, (_, i) => ({ y: 0.4 + Math.sin(i/5) * 0.12 }))
  );
  const [loadavg, setLoadavg] = useState([1.42, 1.36, 1.28]);
  useEffect(() => {
    if (motionMul === 0) return;
    const id = setInterval(() => {
      setPulseData(prev => {
        const t = Date.now() / 1000;
        const totalLoad = liveCoreLoads.reduce((s,v) => s+v, 0) / 100 / 8;
        const base = 0.3 + totalLoad * 0.55 + Math.sin(t/3) * 0.06;
        const noise = (Math.random() - 0.5) * 0.06;
        const next = Math.max(0.05, Math.min(0.98, base + noise));
        return [...prev.slice(1), { y: next }];
      });
      setLoadavg(prev => {
        const totalLoad = liveCoreLoads.reduce((s,v) => s+v, 0) / 100;
        const target = totalLoad;
        return prev.map((v, i) => v + (target - v) * [0.04, 0.02, 0.005][i]);
      });
    }, 400);
    return () => clearInterval(id);
  }, [motionMul, liveCoreLoads]);

  /* Network interface sparklines */
  const [ifSparks, setIfSparks] = useState(() => {
    const m = {};
    window.NWL.NETWORK_IFS.forEach(i => {
      m[i.name] = {
        rx: Array.from({length: 40}, () => i.rxBase + Math.random() * 0.4),
        tx: Array.from({length: 40}, () => i.txBase + Math.random() * 0.2),
      };
    });
    return m;
  });
  useEffect(() => {
    if (motionMul === 0) return;
    const id = setInterval(() => {
      setIfSparks(prev => {
        const next = {};
        window.NWL.NETWORK_IFS.forEach(i => {
          const isWlan = i.name === "wlp3s0";
          const burst = Math.random() < 0.05 ? Math.random() * 3 : 0;
          const rx = (prev[i.name].rx || []);
          const tx = (prev[i.name].tx || []);
          const newRx = i.state === "DOWN" ? 0 : Math.max(0, i.rxBase + Math.random() * 0.6 + burst);
          const newTx = i.state === "DOWN" ? 0 : Math.max(0, i.txBase + Math.random() * 0.3 + burst * 0.4);
          next[i.name] = {
            rx: [...rx.slice(1), newRx],
            tx: [...tx.slice(1), newTx],
          };
        });
        return next;
      });
    }, 700);
    return () => clearInterval(id);
  }, [motionMul]);

  /* Disk IO */
  const [ioSpark, setIoSpark] = useState({
    read:  Array.from({length: 40}, () => Math.random() * 80),
    write: Array.from({length: 40}, () => Math.random() * 40),
  });
  useEffect(() => {
    if (motionMul === 0) return;
    const id = setInterval(() => {
      setIoSpark(prev => ({
        read:  [...prev.read.slice(1),  Math.max(0, 40 + Math.sin(Date.now()/3000) * 30 + Math.random() * 30)],
        write: [...prev.write.slice(1), Math.max(0, 20 + Math.sin(Date.now()/4500) * 18 + Math.random() * 14)],
      }));
    }, 800);
    return () => clearInterval(id);
  }, [motionMul]);

  /* Selection & overlays */
  const [selectedCore, setSelectedCore] = useState(2);
  const [selectedProc, setSelectedProc] = useState(null);
  const [selectedSyslog, setSelectedSyslog] = useState(null);
  const [hoveredProc, setHoveredProc] = useState(null);

  /* Komut paleti + yardım */
  const [paletteOpen, setPaletteOpen] = useState(false);
  const [helpOpen, setHelpOpen] = useState(false);

  /* Syslog seçili index (klavye gezinme) */
  const [syslogIdx, setSyslogIdx] = useState(0);

  /* Cursor & idle */
  const { cursor, idleMs } = window.useCursorTracker();

  /* Whispers */
  const [whisper, setWhisper] = useState(null);
  useEffect(() => {
    if (!tweaks.showWhispers) return;
    const id = setInterval(() => {
      if (Math.random() < 0.5) return;
      const w = window.NWL.WHISPERS;
      const uptime = window.fmtUptime(now - window.NWL.HOST.boot);
      const msg = w[Math.floor(Math.random() * w.length)]
        .replace("{idle}", window.fmtElapsed(idleMs))
        .replace("{uptime}", uptime);
      setWhisper({ msg, key: Date.now() });
    }, 13000);
    return () => clearInterval(id);
  }, [tweaks.showWhispers, idleMs, now]);

  /* Stats / sparks */
  const stats = useMemo(() => ({
    load: loadavg,
    memPct: (displayMem.used / displayMem.total) * 100,
    diskIO: (ioSpark.read[ioSpark.read.length-1] || 0) + (ioSpark.write[ioSpark.write.length-1] || 0),
    netIO: 1.8,
    ctxsw: "42",
  }), [loadavg, displayMem, ioSpark]);

  const sparks = useMemo(() => ({
    load1: pulseData.map(p => p.y),
    load5: pulseData.map((p, i) => p.y * 0.9 + i*0.001),
    mem:   Array.from({length: 40}, (_, i) => 0.4 + Math.sin(i/4)*0.06 + Math.random()*0.04),
    disk:  ioSpark.read,
    net:   Array.from({length: 40}, (_, i) => 0.3 + Math.sin(i/5)*0.15 + Math.random()*0.08),
    ctx:   Array.from({length: 40}, (_, i) => 0.4 + Math.cos(i/3)*0.15 + Math.random()*0.07),
  }), [pulseData, ioSpark]);

  /* Handlers */
  const onAckSyslog = useCallback((id) => {
    setSyslog(prev => prev.map(e => e.id === id ? { ...e, ack: true } : e));
    setAlerts(prev => prev.filter(a => a.id !== id));
  }, []);
  const onDismiss = useCallback((id) => setAlerts(prev => prev.filter(a => a.id !== id)), []);
  const ackAll = useCallback(() => {
    setSyslog(prev => prev.map(e => e.sev === "crit" ? { ...e, ack: true } : e));
    setAlerts([]);
  }, []);

  const runCommand = useCallback((raw) => {
    const cmd = raw.trim().replace(/^:/, "");
    const parts = cmd.split(/\s+/);
    if (parts[0] === "focus" && parts[1] === "core" && parts[2] !== undefined) {
      const n = parseInt(parts[2]);
      if (!isNaN(n) && n >= 0 && n < window.NWL.CORES.length) setSelectedCore(n);
    } else if (parts[0] === "focus" && parts[1] === "proc" && parts[2] !== undefined) {
      const pid = parseInt(parts[2]);
      const proc = liveProcesses.find(p => p.pid === pid);
      if (proc) setSelectedProc(proc);
    } else if (parts[0] === "ack" && parts[1] === "all") {
      ackAll();
    } else if (parts[0] === "replay" && parts[1]) {
      const s = parseInt(parts[1]);
      if (!isNaN(s)) setReplayOffset(Math.min(replayMax, Math.max(0, s)));
    } else if (parts[0] === "live") {
      setReplayOffset(0);
    } else if (parts[0] === "filter" && parts[1]) {
      const f = parts[1].toLowerCase();
      if (["all","info","warn","crit"].includes(f)) setFilter(f);
    } else if (parts[0] === "clear") {
      setAlerts([]);
    } else if (parts[0] === "dim") {
      // toggle vignette deeper
      setTweak("vignette", true);
    } else if (parts[0] === "help" || parts[0] === "?") {
      setHelpOpen(true);
    }
    setPaletteOpen(false);
  }, [liveProcesses, replayMax, ackAll, setTweak]);

  /* Klavye */
  useEffect(() => {
    const onKey = (e) => {
      // input/textarea içindeyken kısayolları görmezden gel
      const tag = (e.target && e.target.tagName) || "";
      if (tag === "INPUT" || tag === "TEXTAREA") return;

      if (e.key === "Escape") {
        if (selectedProc) { setSelectedProc(null); return; }
        if (selectedSyslog) { setSelectedSyslog(null); return; }
        if (paletteOpen) { setPaletteOpen(false); return; }
        if (helpOpen) { setHelpOpen(false); return; }
        if (replayOffset > 0) { setReplayOffset(0); return; }
        return;
      }
      if (e.key === ":") {
        e.preventDefault();
        setPaletteOpen(true);
        return;
      }
      if (e.key === "?") {
        e.preventDefault();
        setHelpOpen(true);
        return;
      }
      if (e.key === "ArrowLeft") {
        e.preventDefault();
        setReplayOffset(o => Math.min(replayMax, o + 1));
        return;
      }
      if (e.key === "ArrowRight") {
        e.preventDefault();
        setReplayOffset(o => Math.max(0, o - 1));
        return;
      }
      if (e.key === "j") {
        e.preventDefault();
        const filtered = filter === "all" ? syslog : syslog.filter(s => s.sev === filter);
        setSyslogIdx(i => Math.min(filtered.length - 1, i + 1));
        return;
      }
      if (e.key === "k") {
        e.preventDefault();
        setSyslogIdx(i => Math.max(0, i - 1));
        return;
      }
      if (e.key === "g") {
        e.preventDefault();
        setSyslogIdx(0);
        return;
      }
      if (e.key === "G") {
        e.preventDefault();
        const filtered = filter === "all" ? syslog : syslog.filter(s => s.sev === filter);
        setSyslogIdx(Math.max(0, filtered.length - 1));
        return;
      }
      if (e.key === "Enter") {
        e.preventDefault();
        const filtered = filter === "all" ? syslog : syslog.filter(s => s.sev === filter);
        if (filtered[syslogIdx]) setSelectedSyslog(filtered[syslogIdx]);
        return;
      }
      if (e.key === "/") {
        e.preventDefault();
        const order = ["all","info","warn","crit"];
        const i = order.indexOf(filter);
        setFilter(order[(i+1) % order.length]);
        return;
      }
      if (/^[0-8]$/.test(e.key)) {
        e.preventDefault();
        const n = parseInt(e.key);
        if (n === 0) setSelectedCore(null);
        else setSelectedCore(n - 1);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [paletteOpen, helpOpen, selectedProc, selectedSyslog, replayOffset, replayMax, syslogIdx, syslog, filter]);

  /* ─── Render ─── */
  return (
    <>
      <div style={{
        position: "absolute",
        left: "50%", top: "50%",
        width: STAGE_W, height: STAGE_H,
        transform: `translate(-50%, -50%) scale(${scale})`,
        transformOrigin: "center center",
        background: "var(--bg)",
        display:"flex", flexDirection:"column",
      }}>
        <window.L_TopBar now={now} boot={window.NWL.HOST.boot} sessions={window.NWL.HOST.sessions} replayOffset={replayOffset} />

        <div style={{
          flex: 1,
          display: "grid",
          gridTemplateColumns: "380px 1fr 640px",
          overflow: "hidden",
        }}>
          {/* SOL */}
          <div style={{ borderRight:"1px solid var(--rule)", display:"flex", flexDirection:"column", overflow:"hidden" }}>
            <window.L_HostBadge   now={now} boot={window.NWL.HOST.boot} idleMs={idleMs} />
            <window.L_LoadPulse   pulseData={pulseData} loadavg={loadavg} />
            <window.L_MemoryMap   mem={displayMem} />
            <window.L_DiskPanel   mounts={window.NWL.MOUNTS} ioSpark={ioSpark} />
            <window.L_GpuPanel    gpu={window.NWL.GPU} />
            <window.L_SensorsPanel sensors={window.NWL.SENSORS} />
          </div>

          {/* ORTA */}
          <div style={{ display:"flex", flexDirection:"column", overflow:"hidden" }}>
            <div style={{ padding:"14px 28px 8px", display:"flex", justifyContent:"space-between", alignItems:"center" }}>
              <div style={{ display:"flex", alignItems:"baseline", gap:14 }}>
                <span className="display" style={{ fontSize:13, letterSpacing:"0.22em", color:"var(--ink-2)" }}>
                  {L.cores}
                </span>
                <span className="mono" style={{ fontSize:10, color:"var(--ink-3)" }}>
                  · 8 çekirdek · {processes.length} süreç · ort. %{(coreLoads.reduce((s,v)=>s+v,0)/8).toFixed(1)}
                </span>
              </div>
              <div style={{ display:"flex", alignItems:"center", gap:14 }}>
                <window.L_ReplayScrubber
                  offset={replayOffset}
                  max={replayMax}
                  onChange={setReplayOffset}
                  onLive={() => setReplayOffset(0)}
                />
                {tweaks.showSurveillance && <window.FollowingEye cursor={cursor} idleMs={idleMs} />}
                <span className="dim" style={{ fontSize:9, letterSpacing:"0.18em" }}>{L.observed}</span>
              </div>
            </div>

            <div style={{ padding:"4px 18px 0", flex:"0 0 auto" }}>
              <window.CoresSchematic
                processes={processes}
                cores={window.NWL.CORES}
                coreLoads={coreLoads}
                selectedCore={selectedCore}
                onSelectCore={setSelectedCore}
                onSelectProcess={setSelectedProc}
                hoveredProc={hoveredProc}
                setHoveredProc={setHoveredProc}
              />
            </div>

            <div style={{ borderTop:"1px solid var(--rule)", flex:"0 0 auto" }}>
              <window.CoreDetail
                coreIdx={selectedCore}
                cores={window.NWL.CORES}
                processes={processes}
                coreLoads={coreLoads}
                history={coreHistory}
              />
            </div>

            <div style={{ marginTop:"auto" }}>
              <window.L_BigNumbers stats={stats} spark={sparks} />
            </div>
          </div>

          {/* SAĞ */}
          <div style={{ borderLeft:"1px solid var(--rule)", display:"flex", flexDirection:"column", overflow:"hidden" }}>
            <window.L_SyslogStream
              events={syslog}
              onSelect={setSelectedSyslog}
              filter={filter}
              setFilter={setFilter}
              selectedIdx={syslogIdx}
            />
            <window.L_ConnectionsPanel
              conns={window.NWL.CONNECTIONS}
              ifs={window.NWL.NETWORK_IFS}
              ifSparks={ifSparks}
            />
            <window.L_AlertQueue
              alerts={alerts}
              onAck={onAckSyslog}
              onDismiss={onDismiss}
            />
          </div>
        </div>

        {/* Alt şerit */}
        <div style={{
          display:"grid",
          gridTemplateColumns:"380px 1fr 640px",
          borderTop:"1px solid var(--rule)",
        }}>
          <div style={{ padding:"6px 24px", display:"flex", alignItems:"center", gap:14 }}>
            {tweaks.showSurveillance && <window.CursorLog cursor={cursor} idleMs={idleMs} />}
          </div>
          <window.L_BottomTicker feed={L.ticker} />
          <div style={{ padding:"6px 24px", display:"flex", alignItems:"center", justifyContent:"flex-end", gap:14,
                          borderLeft:"1px solid var(--rule)" }}>
            <button onClick={() => setHelpOpen(true)}
              className="mono"
              style={{ fontSize:9, color:"var(--ink-3)", letterSpacing:"0.18em",
                       border:"1px solid var(--rule)", padding:"3px 8px" }}>? · YARDIM</button>
            <button onClick={() => setPaletteOpen(true)}
              className="mono"
              style={{ fontSize:9, color:"var(--accent)", letterSpacing:"0.18em",
                       border:"1px solid var(--accent-2)", padding:"3px 8px" }}>: · KOMUT</button>
            <span className="mono" style={{ fontSize:9, color:"var(--ink-4)", letterSpacing:"0.14em" }}>
              {L.session} {window.NWL.HOST.tty}
            </span>
            <span className="mono" style={{ fontSize:9, color:"var(--accent)", letterSpacing:"0.14em" }}>
              {L.recording}
            </span>
          </div>
        </div>
      </div>

      {/* Overlays */}
      <window.L_SyslogOverlay
        event={selectedSyslog}
        onClose={() => setSelectedSyslog(null)}
        onAck={onAckSyslog}
      />
      <window.L_ProcessOverlay
        process={selectedProc}
        history={selectedProc ? procHistory[selectedProc.uid] : null}
        onClose={() => setSelectedProc(null)}
      />

      {tweaks.showWhispers && whisper && (
        <window.Whisper key={whisper.key} message={whisper.msg} onDone={() => setWhisper(null)} />
      )}

      <window.CommandPalette
        open={paletteOpen}
        onClose={() => setPaletteOpen(false)}
        onCommand={runCommand}
      />
      <window.HelpOverlay
        open={helpOpen}
        onClose={() => setHelpOpen(false)}
      />

      {tweaks.idleDim && <window.L_IdleOverlay idleMs={idleMs} threshold={90000} />}

      <window.TweaksPanel title={L.tweakTitle}>
        <window.TweakSection label={L.paletteSec} />
        <window.TweakRadio
          label={L.toneLbl}
          value={tweaks.palette}
          onChange={v => setTweak("palette", v)}
          options={[
            { value:"noir",  label:L.pal_noir },
            { value:"crt",   label:L.pal_crt  },
            { value:"blue",  label:L.pal_blue },
            { value:"amber", label:L.pal_amber },
          ]}
        />
        <window.TweakSection label={L.motionSec} />
        <window.TweakRadio
          label={L.tempoLbl}
          value={tweaks.motion}
          onChange={v => setTweak("motion", v)}
          options={[
            { value:"still",  label:L.motion_still },
            { value:"calm",   label:L.motion_calm },
            { value:"living", label:L.motion_living },
          ]}
        />
        <window.TweakSection label={L.atmosSec} />
        <window.TweakToggle label={L.scanlines} value={tweaks.scanlines} onChange={v => setTweak("scanlines", v)} />
        <window.TweakToggle label={L.grain}     value={tweaks.grain}     onChange={v => setTweak("grain", v)} />
        <window.TweakToggle label={L.vignette}  value={tweaks.vignette}  onChange={v => setTweak("vignette", v)} />
        <window.TweakToggle label={L.flicker}   value={tweaks.flicker}   onChange={v => setTweak("flicker", v)} />
        <window.TweakSection label={L.surveillanceSec} />
        <window.TweakToggle label={L.eyeLbl}      value={tweaks.showSurveillance} onChange={v => setTweak("showSurveillance", v)} />
        <window.TweakToggle label={L.whispersLbl} value={tweaks.showWhispers}     onChange={v => setTweak("showWhispers", v)} />
        <window.TweakToggle label="Boşta karartma" value={tweaks.idleDim}        onChange={v => setTweak("idleDim", v)} />
      </window.TweaksPanel>
    </>
  );
}

ReactDOM.createRoot(document.getElementById("root")).render(<App />);
