/* Komut paleti + yardım — Türkçe arayüz, komutlar İngilizce */

const { useState: useStateCP, useEffect: useEffectCP, useRef: useRefCP } = React;

function CommandPalette({ open, onClose, onCommand }) {
  const L = window.NWL.L;
  const COMMANDS = window.NWL.COMMANDS;
  const KEYS = window.NWL.KEYS;
  const [value, setValue] = useStateCP("");
  const [history, setHistory] = useStateCP([]);
  const [histIdx, setHistIdx] = useStateCP(-1);
  const inputRef = useRefCP(null);

  useEffectCP(() => {
    if (open) {
      setTimeout(() => inputRef.current && inputRef.current.focus(), 50);
      setValue("");
      setHistIdx(-1);
    }
  }, [open]);

  if (!open) return null;

  const submit = () => {
    const v = value.trim();
    if (!v) return;
    setHistory(h => [v, ...h].slice(0, 20));
    onCommand(v);
    setValue("");
  };

  const onKeyDown = (e) => {
    if (e.key === "Enter") { e.preventDefault(); submit(); }
    else if (e.key === "ArrowUp")   {
      e.preventDefault();
      const idx = Math.min(history.length - 1, histIdx + 1);
      setHistIdx(idx);
      if (history[idx]) setValue(history[idx]);
    }
    else if (e.key === "ArrowDown") {
      e.preventDefault();
      const idx = Math.max(-1, histIdx - 1);
      setHistIdx(idx);
      setValue(idx === -1 ? "" : history[idx]);
    }
    else if (e.key === "Escape") { e.preventDefault(); onClose(); }
  };

  // suggestions
  const v = value.trim().toLowerCase();
  const sugg = !v ? COMMANDS : COMMANDS.filter(c => c.cmd.toLowerCase().includes(v) || c.tr.toLowerCase().includes(v));

  return (
    <div onClick={onClose} style={{
      position:"fixed", inset:0, background:"rgba(0,0,0,0.65)",
      zIndex: 8500, display:"flex", alignItems:"flex-start", justifyContent:"center",
      paddingTop: "12vh",
      backdropFilter:"blur(2px)",
    }}>
      <div onClick={e => e.stopPropagation()} style={{
        width: 640, background:"var(--bg-2)", border:"1px solid var(--rule)",
      }}>
        <div style={{ padding:"14px 18px", borderBottom:"1px solid var(--rule)",
                       display:"flex", alignItems:"center", justifyContent:"space-between" }}>
          <span className="dim" style={{ fontSize:9, letterSpacing:"0.24em" }}>{L.cmdTitle}</span>
          <span className="dim" style={{ fontSize:9, letterSpacing:"0.14em" }}>{L.cmdHint}</span>
        </div>
        <div style={{ padding:"14px 18px", borderBottom:"1px solid var(--rule)",
                       display:"flex", alignItems:"baseline", gap:10 }}>
          <span className="display" style={{ fontSize:18, color:"var(--accent)" }}>:</span>
          <input
            ref={inputRef}
            value={value}
            onChange={e => setValue(e.target.value)}
            onKeyDown={onKeyDown}
            placeholder="örn. focus core 2"
            style={{
              flex:1, background:"transparent", color:"var(--ink)",
              border:"none", outline:"none",
              fontFamily:"JetBrains Mono", fontSize:16, letterSpacing:"0.02em",
            }}
          />
        </div>
        <div style={{ maxHeight: 320, overflowY:"auto" }}>
          {sugg.map((s, i) => (
            <button key={i}
              onClick={() => { setValue(s.example.replace(/^:/, "")); inputRef.current && inputRef.current.focus(); }}
              style={{
                width:"100%", textAlign:"left",
                padding:"10px 18px",
                borderBottom:"1px solid var(--rule-2)",
                display:"grid", gridTemplateColumns:"180px 1fr 110px", gap:10, alignItems:"baseline",
              }}
              onMouseEnter={ev => ev.currentTarget.style.background = "var(--bg)"}
              onMouseLeave={ev => ev.currentTarget.style.background = "transparent"}
            >
              <span className="mono" style={{ fontSize:11, color:"var(--accent)" }}>{s.cmd}</span>
              <span className="whisper" style={{ fontSize:13, color:"var(--ink-2)" }}>{s.tr}</span>
              <span className="mono" style={{ fontSize:9, color:"var(--ink-4)", textAlign:"right" }}>{s.example}</span>
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}

function HelpOverlay({ open, onClose }) {
  const L = window.NWL.L;
  const KEYS = window.NWL.KEYS;
  const COMMANDS = window.NWL.COMMANDS;
  if (!open) return null;
  return (
    <div onClick={onClose} style={{
      position:"fixed", inset:0, background:"rgba(0,0,0,0.72)",
      zIndex: 8500, display:"flex", alignItems:"center", justifyContent:"center",
      backdropFilter:"blur(2px)",
    }}>
      <div onClick={e => e.stopPropagation()} style={{
        width: 760, background:"var(--bg-2)", border:"1px solid var(--rule)",
        padding: "32px 36px",
      }}>
        <div className="dim" style={{ fontSize:9, letterSpacing:"0.22em" }}>{L.helpTitle}</div>
        <div className="display" style={{ fontSize:32, color:"var(--accent)", letterSpacing:"0.06em", marginTop:6, lineHeight:1 }}>
          tuşlar ve komutlar
        </div>

        <div style={{ display:"grid", gridTemplateColumns:"1fr 1fr", gap:36, marginTop:24 }}>
          <div>
            <div className="dim" style={{ fontSize:9, letterSpacing:"0.22em", marginBottom:10 }}>KISAYOLLAR</div>
            <div style={{ display:"grid", gridTemplateColumns:"80px 1fr", gap:"8px 14px", fontSize:11 }}>
              {KEYS.map(k => (
                <React.Fragment key={k.k}>
                  <span className="mono" style={{ color:"var(--accent)" }}>{k.k}</span>
                  <span className="whisper" style={{ color:"var(--ink-2)" }}>{k.tr}</span>
                </React.Fragment>
              ))}
            </div>
          </div>
          <div>
            <div className="dim" style={{ fontSize:9, letterSpacing:"0.22em", marginBottom:10 }}>KOMUTLAR</div>
            <div style={{ display:"grid", gridTemplateColumns:"170px 1fr", gap:"8px 12px", fontSize:11 }}>
              {COMMANDS.map(c => (
                <React.Fragment key={c.cmd}>
                  <span className="mono" style={{ color:"var(--accent)" }}>{c.cmd}</span>
                  <span className="whisper" style={{ color:"var(--ink-2)" }}>{c.tr}</span>
                </React.Fragment>
              ))}
            </div>
          </div>
        </div>

        <div style={{ marginTop:30, paddingTop:18, borderTop:"1px solid var(--rule)" }}>
          <div className="whisper" style={{ fontSize:13, color:"var(--ink-3)", lineHeight:1.5 }}>
            "Bu sistem komutları İngilizce konuşur, açıklamayı Türkçe yapar. Hangi dili konuşursan konuş, çekirdek aynı şeyi hatırlar."
          </div>
        </div>

        <div style={{ display:"flex", justifyContent:"flex-end", marginTop:24 }}>
          <button onClick={onClose} style={{
            padding:"8px 16px", fontSize:10, letterSpacing:"0.22em",
            color:"var(--ink-3)", border:"1px solid var(--rule)",
          }}>{L.close}  ·  Esc</button>
        </div>
      </div>
    </div>
  );
}

window.CommandPalette = CommandPalette;
window.HelpOverlay = HelpOverlay;
