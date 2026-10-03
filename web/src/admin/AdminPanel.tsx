import { useEffect, useRef, useState } from "react";
import { chooseCandidate, fetchClipState, fetchVoices, generateCandidate } from "../api";
import { usePlayer } from "../player";
import type { Candidate, Choice, ClipState, Voice } from "../types";
import { useAdmin } from "./context";

type Props = {
  clips: string[]; // порядок записей на странице
  onChosen(id: string, url: string, choice: Choice): void;
  onClose(): void;
};

// Плашка со счётчиком и, если запись открыта, панель выбора варианта озвучки.
export function AdminPanel({ clips, onChosen, onClose }: Props) {
  const { active, open, choices } = useAdmin();
  const done = clips.filter((id) => choices[id]).length;
  const nextUnchosen = () => {
    const from = active ? clips.indexOf(active) + 1 : 0;
    const next = [...clips.slice(from), ...clips.slice(0, from)].find((id) => !choices[id]);
    if (next) open(next);
  };

  return (
    <>
      <div className="admin-pill">
        <strong>Админка</strong>
        <span>
          выбрано {done} из {clips.length}
        </span>
        {done < clips.length && (
          <button type="button" className="btn" onClick={nextUnchosen}>
            Следующий без выбора →
          </button>
        )}
      </div>
      {active && <Drawer key={active} id={active} clips={clips} onChosen={onChosen} onClose={onClose} />}
    </>
  );
}

let voicesCache: Promise<Voice[]> | null = null;

function Drawer({ id, clips, onChosen, onClose }: { id: string } & Props) {
  const { open, choices } = useAdmin();
  const { urlOf, playUrl, playing } = usePlayer();
  const [voices, setVoices] = useState<Voice[]>([]);
  const [state, setState] = useState<ClipState | null>(null);
  const [text, setText] = useState("");
  const [busy, setBusy] = useState<Record<string, boolean>>({});
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [loadError, setLoadError] = useState<string | null>(null);
  // Токен запуска: ответы, пришедшие после ухода с записи (или после повторного
  // запуска эффекта в StrictMode), не должны трогать состояние.
  const run = useRef(0);

  const kind = id.slice(0, 1) as "s" | "w";
  const key = id.slice(2);
  const index = clips.indexOf(id);
  const current = urlOf(kind, key);
  const chosen = choices[id] ?? state?.chosen;

  const generate = async (vs: Voice[], forText: string) => {
    const mine = run.current;
    const here = () => run.current === mine;
    for (const v of vs) {
      if (!here()) return;
      setBusy((b) => ({ ...b, [v.id]: true }));
      setErrors((e) => ({ ...e, [v.id]: "" }));
      try {
        const cand = await generateCandidate(id, v.id, forText);
        if (here()) setState((s) => s && { ...s, text: forText, candidates: upsert(s.candidates, cand, vs) });
      } catch (e) {
        if (here()) setErrors((errs) => ({ ...errs, [v.id]: String(e instanceof Error ? e.message : e) }));
      } finally {
        if (here()) setBusy((b) => ({ ...b, [v.id]: false }));
      }
    }
  };

  // Открыли запись: подгружаем варианты; если их ещё нет — сразу делаем всеми голосами.
  useEffect(() => {
    const mine = ++run.current;
    (async () => {
      try {
        voicesCache ??= fetchVoices();
        const [vs, st] = await Promise.all([voicesCache, fetchClipState(id)]);
        if (run.current !== mine) return;
        setVoices(vs);
        setState(st);
        setText(st.text);
        if (st.candidates.length === 0) await generate(vs, st.text);
      } catch (e) {
        if (run.current === mine) setLoadError(String(e));
      }
    })();
    document.querySelector(`[data-clip="${CSS.escape(id)}"]`)?.scrollIntoView({ block: "center", behavior: "smooth" });
    return () => {
      run.current++;
    };
  }, [id]);

  const choose = async (voice: string) => {
    try {
      const res = await chooseCandidate(id, voice);
      onChosen(id, res.url, res.chosen);
      setState((s) => s && { ...s, chosen: res.chosen });
    } catch (e) {
      setErrors((errs) => ({ ...errs, [voice]: String(e instanceof Error ? e.message : e) }));
    }
  };

  const candidates = state?.candidates ?? [];
  const playCandidate = (c: Candidate) => playUrl(`cand:${id}:${c.voice}`, c.url);

  // 1…9, 0 — послушать вариант; ← → — соседние записи (когда не печатаем в поле).
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.metaKey || e.ctrlKey || e.altKey || e.target instanceof HTMLInputElement) return;
      if (/^[0-9]$/.test(e.key)) {
        const v = voices[(Number(e.key) + 9) % 10];
        const c = v && candidates.find((x) => x.voice === v.id);
        if (c) playCandidate(c);
      } else if (e.key === "ArrowLeft" && index > 0) open(clips[index - 1]);
      else if (e.key === "ArrowRight" && index < clips.length - 1) open(clips[index + 1]);
      else if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  });

  const anyBusy = Object.values(busy).some(Boolean);

  return (
    <aside className="drawer" aria-label="Озвучка">
      <header className="drawer__head">
        <span className="drawer__kind">{kind === "s" ? "звук" : "слово"}</span>
        <span className="drawer__key" lang="ru">
          {key}
        </span>
        <span className="drawer__nav">
          <button type="button" className="btn btn--icon" disabled={index <= 0} onClick={() => open(clips[index - 1])} title="Предыдущая (←)">
            ←
          </button>
          <span className="drawer__pos">
            {index + 1}/{clips.length}
          </span>
          <button
            type="button"
            className="btn btn--icon"
            disabled={index >= clips.length - 1}
            onClick={() => open(clips[index + 1])}
            title="Следующая (→)"
          >
            →
          </button>
          <button type="button" className="btn btn--icon" onClick={onClose} title="Закрыть (Esc)">
            ×
          </button>
        </span>
      </header>

      <section className="drawer__now">
        <button type="button" className="btn" disabled={!current} onClick={() => current && playUrl(id, current)}>
          ▶ Сейчас
        </button>
        <span className="drawer__muted">
          {chosen ? `${voiceLabel(voices, chosen.voice)} · «${chosen.text}»` : current ? "голос не выбран" : "записи нет"}
        </span>
      </section>

      <form
        className="drawer__text"
        onSubmit={(e) => {
          e.preventDefault();
          if (text.trim() && !anyBusy) void generate(voices, text.trim());
        }}
      >
        <label>
          Что произносить <span className="drawer__muted">(+ перед ударной гласной)</span>
          <input lang="ru" value={text} onChange={(e) => setText(e.target.value)} spellCheck={false} />
        </label>
        <button type="submit" className="btn" disabled={anyBusy || !text.trim()}>
          {anyBusy ? "Генерирую…" : candidates.length ? "Сгенерировать заново" : "Сгенерировать"}
        </button>
      </form>

      {loadError && <p className="drawer__error">{loadError}</p>}

      <ol className="cands">
        {voices.map((v, i) => {
          const c = candidates.find((x) => x.voice === v.id);
          const isChosen = chosen?.voice === v.id && chosen.text === state?.text;
          const isPlaying = playing === `cand:${id}:${v.id}`;
          return (
            <li key={v.id} className="cand" data-chosen={isChosen || undefined}>
              <button
                type="button"
                className="btn btn--icon"
                disabled={!c || busy[v.id]}
                onClick={() => c && playCandidate(c)}
                title={`Послушать (${(i + 1) % 10})`}
              >
                {busy[v.id] ? "…" : isPlaying ? "■" : "▶"}
              </button>
              <span className="cand__name">
                {v.label}
                <span className="cand__license">{v.license}</span>
                {errors[v.id] && <span className="drawer__error">{errors[v.id]}</span>}
              </span>
              {isChosen ? (
                <span className="cand__chosen">✓ выбран</span>
              ) : (
                <button type="button" className="btn" disabled={!c || busy[v.id]} onClick={() => void choose(v.id)}>
                  Выбрать
                </button>
              )}
            </li>
          );
        })}
      </ol>
      <p className="drawer__muted drawer__keys">1…0 — послушать вариант · ← → — соседние записи · Esc — закрыть</p>
    </aside>
  );
}

function upsert(list: Candidate[], cand: Candidate, order: Voice[]): Candidate[] {
  const rest = list.filter((c) => c.voice !== cand.voice);
  const pos = (c: Candidate) => order.findIndex((v) => v.id === c.voice);
  return [...rest, cand].sort((a, b) => pos(a) - pos(b));
}

const voiceLabel = (voices: Voice[], id: string) => voices.find((v) => v.id === id)?.label ?? id;
