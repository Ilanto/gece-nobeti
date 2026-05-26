/* Gözetim katmanı — Türkçe etiketler */

const { useState: useStateS, useEffect: useEffectS, useRef: useRefS } = React;

function FollowingEye({ cursor, idleMs }) {
  const ref = useRefS(null);
  const [pupilXY, setPupilXY] = useStateS({ x: 0, y: 0 });
  const [closed, setClosed] = useStateS(false);

  useEffectS(() => {
    if (!ref.current) return;
    const r = ref.current.getBoundingClientRect();
    const cx = r.left + r.width/2;
    const cy = r.top + r.height/2;
    const dx = cursor.x - cx;
    const dy = cursor.y - cy;
    const dist = Math.sqrt(dx*dx + dy*dy);
    const max = 4;
    const nx = dist === 0 ? 0 : (dx / dist) * Math.min(max, dist / 90);
    const ny = dist === 0 ? 0 : (dy / dist) * Math.min(max, dist / 90);
    setPupilXY({ x: nx, y: ny });
  }, [cursor]);

  useEffectS(() => {
    const blink = () => {
      setClosed(true);
      setTimeout(() => setClosed(false), 130);
    };
    const id = setInterval(blink, 6500 + Math.random() * 4000);
    return () => clearInterval(id);
  }, []);

  return (
    <div ref={ref} style={{ width: 44, height: 22, position:"relative" }}>
      <svg viewBox="0 0 44 22" width="44" height="22">
        <path d="M 2 11 Q 22 1 42 11 Q 22 21 2 11 Z"
              fill="none" stroke="var(--ink-3)" strokeWidth="1" />
        {!closed ? (
          <>
            <circle cx={22 + pupilXY.x} cy={11 + pupilXY.y} r="5" fill="var(--ink-4)" />
            <circle cx={22 + pupilXY.x} cy={11 + pupilXY.y} r="2.5" fill="var(--accent)" />
            <circle cx={22 + pupilXY.x - 1} cy={11 + pupilXY.y - 1} r="0.6" fill="var(--ink)" />
          </>
        ) : (
          <line x1="3" y1="11" x2="41" y2="11" stroke="var(--ink-3)" strokeWidth="1" />
        )}
      </svg>
    </div>
  );
}

function CursorLog({ cursor, idleMs }) {
  const L = window.NWL.L;
  return (
    <div className="mono" style={{
      fontSize: 9, color: "var(--ink-4)", letterSpacing:"0.12em",
      display:"flex", gap:14,
    }}>
      <span>{L.cur}</span>
      <span>{String(cursor.x).padStart(4," ")} ×{String(cursor.y).padStart(4," ")}</span>
      <span style={{ color: idleMs > 30000 ? "var(--warn)" : "var(--ink-4)" }}>
        {L.idle.toUpperCase()} {window.fmtElapsed(idleMs)}
      </span>
      <span>{L.bpm} {Math.round(60 + Math.sin(Date.now()/4000) * 4)}</span>
    </div>
  );
}

function Whisper({ message, onDone }) {
  const [visible, setVisible] = useStateS(false);
  useEffectS(() => {
    setVisible(true);
    const t1 = setTimeout(() => setVisible(false), 7500);
    const t2 = setTimeout(() => onDone && onDone(), 8500);
    return () => { clearTimeout(t1); clearTimeout(t2); };
  }, [message]);
  if (!message) return null;
  const L = window.NWL.L;
  return (
    <div style={{
      position:"fixed", left: 32, bottom: 80,
      maxWidth: 380,
      padding: "14px 18px",
      background: "rgba(10,10,12,0.85)",
      border: "1px solid var(--rule)",
      backdropFilter: "blur(4px)",
      opacity: visible ? 1 : 0,
      transform: visible ? "translateY(0)" : "translateY(8px)",
      transition: "opacity 0.8s, transform 0.8s",
      zIndex: 200,
      pointerEvents:"none",
    }}>
      <div className="dim" style={{ fontSize:9, letterSpacing:"0.24em", marginBottom:6 }}>
        {L.whisper}
      </div>
      <div className="whisper" style={{ fontSize:15, color:"var(--ink-2)", lineHeight:1.45 }}>
        "{message}"
      </div>
    </div>
  );
}

function useCursorTracker() {
  const [cursor, setCursor] = useStateS({ x: 0, y: 0 });
  const [lastMove, setLastMove] = useStateS(Date.now());
  const [, setTick] = useStateS(0);
  useEffectS(() => {
    const onMove = (e) => {
      setCursor({ x: e.clientX, y: e.clientY });
      setLastMove(Date.now());
    };
    const onKey = () => setLastMove(Date.now());
    window.addEventListener("mousemove", onMove);
    window.addEventListener("keydown", onKey);
    return () => {
      window.removeEventListener("mousemove", onMove);
      window.removeEventListener("keydown", onKey);
    };
  }, []);
  useEffectS(() => {
    const id = setInterval(() => setTick(t => t+1), 1000);
    return () => clearInterval(id);
  }, []);
  const idleMs = Date.now() - lastMove;
  return { cursor, idleMs };
}

window.FollowingEye = FollowingEye;
window.CursorLog = CursorLog;
window.Whisper = Whisper;
window.useCursorTracker = useCursorTracker;
