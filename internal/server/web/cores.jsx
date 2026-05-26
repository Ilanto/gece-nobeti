/* CPU çekirdekleri şeması — Türkçe + operatör sürecini "(sen)" olarak işaretle */

const { useMemo: useMemoC } = React;

function CoresSchematic({ processes, cores, coreLoads, selectedCore, onSelectCore, onSelectProcess, hoveredProc, setHoveredProc }) {
  const L = window.NWL.L;
  const W = 1480;
  const ROW_H = 56;
  const PAD_TOP = 28;
  const X0 = 160;
  const X1 = W - 28;
  const H = PAD_TOP + cores.length * ROW_H + 16;

  const procsByCore = useMemoC(() => {
    const m = {};
    processes.forEach(p => { (m[p.core] ||= []).push(p); });
    return m;
  }, [processes]);

  const thresholds = [25, 50, 75, 100];

  return (
    <svg viewBox={`0 0 ${W} ${H}`}
         style={{ width: "100%", height: "auto", display: "block" }}
         preserveAspectRatio="xMidYMid meet">
      <defs>
        <filter id="procGlow" x="-200%" y="-200%" width="500%" height="500%">
          <feGaussianBlur stdDeviation="2" />
        </filter>
      </defs>

      {thresholds.map(t => {
        const x = X0 + (t/100) * (X1 - X0);
        return (
          <g key={t}>
            <line x1={x} y1={PAD_TOP} x2={x} y2={H-16}
                  stroke="var(--rule-2)" strokeWidth="1" strokeDasharray="2 4" />
            <text x={x} y={PAD_TOP - 8} fill="var(--ink-4)"
                  fontFamily="JetBrains Mono" fontSize="8" textAnchor="middle">%{t}</text>
          </g>
        );
      })}

      {cores.map((core, i) => {
        const y = PAD_TOP + i * ROW_H + ROW_H/2;
        const procs = procsByCore[i] || [];
        const load = coreLoads[i] || 0;
        const active = selectedCore === i;
        const dim = selectedCore !== null && !active;
        const loadX = X0 + (load/100) * (X1 - X0);

        return (
          <g key={core.id}
             opacity={dim ? 0.32 : 1}
             style={{ cursor: "pointer", transition: "opacity 0.3s" }}
             onClick={() => onSelectCore(active ? null : i)}>
            <rect x="0" y={y - ROW_H/2} width={W} height={ROW_H} fill="transparent" />

            <text x="8" y={y - 8} fill="var(--ink-3)"
                  fontFamily="JetBrains Mono" fontSize="9" letterSpacing="0.1em">
              {core.id} · {core.governor}
            </text>
            <text x="8" y={y + 12} fill={active ? "var(--accent)" : "var(--ink)"}
                  fontFamily="Barlow Condensed" fontSize="20" fontWeight="500" letterSpacing="0.04em">
              %{load.toFixed(0).padStart(3," ")}
            </text>
            <text x="78" y={y + 12} fill="var(--ink-4)"
                  fontFamily="JetBrains Mono" fontSize="9">
              {core.freq.toFixed(2)}GHz
            </text>

            <line x1={X0} y1={y} x2={X1} y2={y}
                  stroke={active ? "var(--accent-2)" : "var(--rule)"} strokeWidth="1" />

            <line x1={X0} y1={y} x2={loadX} y2={y}
                  stroke={active ? "var(--accent)" : "var(--ink-3)"} strokeWidth="1.5" opacity="0.55" />

            <circle cx={loadX} cy={y} r={active ? 3.4 : 2.4}
                    fill={active ? "var(--accent)" : "var(--ink-2)"} />

            {procs.map((p, pi) => {
              const px = X0 + Math.min(100, Math.max(0, p.cpuLive)) / 100 * (X1 - X0);
              const py = y - 10 - (pi % 2) * 8;
              const isHovered = hoveredProc === p.uid;
              const color = p.isOperator    ? "var(--accent)"
                          : p.kind === "kernel" ? "var(--ink-3)"
                          : p.kind === "system" ? "var(--ink-2)"
                          : "var(--accent)";
              const label = p.isOperator ? `${p.name} · ${window.NWL.L.you}` : p.name;
              return (
                <g key={p.uid}
                   onMouseEnter={() => setHoveredProc(p.uid)}
                   onMouseLeave={() => setHoveredProc(null)}
                   onClick={(e) => { e.stopPropagation(); onSelectProcess(p); }}
                   style={{ cursor: "pointer" }}>
                  <line x1={px} y1={py} x2={px} y2={y - 2}
                        stroke={color} strokeWidth="0.5" opacity="0.4" strokeDasharray="1 2" />
                  {/* operator gets a halo */}
                  {p.isOperator && (
                    <circle cx={px} cy={py} r="7" fill="none" stroke="var(--accent)" strokeWidth="0.5" opacity="0.5">
                      <animate attributeName="r" values="6;9;6" dur="3s" repeatCount="indefinite" />
                      <animate attributeName="opacity" values="0.5;0.15;0.5" dur="3s" repeatCount="indefinite" />
                    </circle>
                  )}
                  <circle cx={px} cy={py} r={isHovered ? 4 : 2.4}
                          fill={color} opacity="0.95" />
                  <text x={px + 6} y={py + 3} fill={color}
                        fontFamily="JetBrains Mono" fontSize="8" opacity={isHovered ? 1 : 0.75}>
                    {label}
                    {p.container && <tspan fill="var(--ink-4)" fontStyle="italic"> [docker:{p.container}]</tspan>}
                  </text>
                  {isHovered && (
                    <g>
                      <rect x={px + 8} y={py - 32} width="180" height="42"
                            fill="var(--bg-2)" stroke="var(--rule)" />
                      <text x={px + 16} y={py - 18} fill="var(--ink)"
                            fontFamily="JetBrains Mono" fontSize="9">
                        {p.name} · pid {p.pid} {p.isOperator && "· bu sensin"}
                      </text>
                      <text x={px + 16} y={py - 6} fill="var(--ink-3)"
                            fontFamily="JetBrains Mono" fontSize="8">
                        %{p.cpuLive.toFixed(1)} · {p.mem}MB · {p.threads} iş.p.
                      </text>
                    </g>
                  )}
                </g>
              );
            })}

            <line x1="0" y1={y + ROW_H/2 - 0.5} x2={W} y2={y + ROW_H/2 - 0.5}
                  stroke="var(--rule-2)" strokeWidth="1" />
          </g>
        );
      })}
    </svg>
  );
}

function CoreDetail({ coreIdx, cores, processes, coreLoads, history }) {
  const L = window.NWL.L;
  const core = coreIdx == null ? null : cores[coreIdx];
  if (!core) {
    return (
      <div style={{ padding:"22px 24px", color:"var(--ink-3)", fontFamily:"JetBrains Mono", fontSize:11 }}>
        <span className="dim">{L.selectCore}</span>
      </div>
    );
  }
  const procs = processes.filter(p => p.core === coreIdx).sort((a,b) => b.cpuLive - a.cpuLive);
  const load = coreLoads[coreIdx] || 0;
  const hist = history[coreIdx] || [];

  const W = 1480, H = 220;
  const histX0 = 760, histX1 = W - 28;
  const histY0 = 30, histY1 = H - 30;
  const histPath = hist.map((v, i) => {
    const x = histX0 + (i / Math.max(1, hist.length - 1)) * (histX1 - histX0);
    const y = histY1 - (v/100) * (histY1 - histY0);
    return `${i === 0 ? "M" : "L"}${x.toFixed(1)},${y.toFixed(1)}`;
  }).join(" ");
  const histArea = histPath ? `${histPath} L${histX1},${histY1} L${histX0},${histY1} Z` : "";

  return (
    <svg viewBox={`0 0 ${W} ${H}`} style={{ width:"100%", height:"auto", display:"block" }}>
      <text x="8" y="20" fill="var(--ink-3)" fontFamily="JetBrains Mono" fontSize="10" letterSpacing="0.1em">
        {core.id} · {L.detail}
      </text>
      <text x="8" y="56" fill="var(--accent)" fontFamily="Barlow Condensed" fontSize="48" fontWeight="500" letterSpacing="0.04em">
        %{load.toFixed(1)}
      </text>
      <text x="8" y="76" fill="var(--ink-4)" fontFamily="JetBrains Mono" fontSize="9">
        {core.freq.toFixed(2)} GHz · {core.governor} · {core.numa} · µcode {core.microcode}
      </text>

      <text x="280" y="20" fill="var(--ink-3)" fontFamily="JetBrains Mono" fontSize="10" letterSpacing="0.1em">
        {L.topProcs}
      </text>
      {procs.slice(0, 8).map((p, i) => {
        const y = 40 + i * 22;
        const barW = Math.max(2, (p.cpuLive / 100) * 360);
        const color = p.isOperator ? "var(--accent)"
                    : p.kind === "kernel" ? "var(--ink-3)"
                    : p.kind === "system" ? "var(--ink-2)"
                    : "var(--accent)";
        return (
          <g key={p.uid}>
            <text x="280" y={y} fill="var(--ink-2)" fontFamily="JetBrains Mono" fontSize="10">
              {p.name.padEnd(14, " ").slice(0,14)}{p.isOperator ? " ·sen" : ""}
            </text>
            <text x="380" y={y} fill="var(--ink-4)" fontFamily="JetBrains Mono" fontSize="9">
              {String(p.pid).padStart(6, " ")}
            </text>
            <line x1="430" y1={y - 3} x2={430 + barW} y2={y - 3} stroke={color} strokeWidth="2" />
            <text x="700" y={y} fill={color} fontFamily="JetBrains Mono" fontSize="10" textAnchor="end">
              %{p.cpuLive.toFixed(1)}
            </text>
          </g>
        );
      })}

      <text x={histX0} y="20" fill="var(--ink-3)" fontFamily="JetBrains Mono" fontSize="10" letterSpacing="0.1em">
        {L.load60}
      </text>
      <line x1={histX0} y1={histY1} x2={histX1} y2={histY1} stroke="var(--rule)" strokeWidth="1" />
      {[25, 50, 75].map(t => {
        const y = histY1 - (t/100) * (histY1 - histY0);
        return (
          <g key={t}>
            <line x1={histX0} y1={y} x2={histX1} y2={y} stroke="var(--rule-2)" strokeDasharray="2 4" />
            <text x={histX0 - 6} y={y + 3} fill="var(--ink-4)" fontFamily="JetBrains Mono" fontSize="8" textAnchor="end">{t}</text>
          </g>
        );
      })}
      <path d={histArea} fill="var(--accent)" opacity="0.12" />
      <path d={histPath} fill="none" stroke="var(--accent)" strokeWidth="1" />
    </svg>
  );
}

window.CoresSchematic = CoresSchematic;
window.CoreDetail = CoreDetail;
