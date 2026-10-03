import { useEffect, useMemo, useRef, useState } from "react";
import { fetchClips, uploadClip } from "../api";
import { useTake, type Phase } from "../audio/useTake";
import { Clip, Speaker } from "../components/Clip";
import { Topbar } from "../components/Topbar";
import { useI18n, type Messages } from "../i18n";
import { clipId, PlayerProvider } from "../player";
import type { ClipInfo, ClipKind, ClipsResponse } from "../types";

type StatusFilter = "missing" | "done" | "all";
type KindFilter = ClipKind | "all";

const idOf = (c: ClipInfo) => clipId(c.kind, c.key);
const busy = (p: Phase) => p.name === "recording" || p.name === "processing" || p.name === "saving";

export function RecordPage() {
  const { ui, t, lang } = useI18n();
  const [data, setData] = useState<ClipsResponse | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [status, setStatus] = useState<StatusFilter>("missing");
  const [kind, setKind] = useState<KindFilter>("all");
  const [active, setActive] = useState<string | null>(null);
  const take = useTake();

  useEffect(() => {
    fetchClips().then(setData, (e: unknown) => setLoadError(String(e)));
  }, []);

  useEffect(() => {
    document.title = ui.recordTitle;
  }, [ui]);

  const clips = data?.clips ?? [];
  const ofKind = clips.filter((c) => kind === "all" || c.kind === kind);
  const visible = ofKind.filter((c) => status === "all" || (status === "done") === Boolean(c.url));
  const doneCount = ofKind.filter((c) => c.url).length;

  const recordings = useMemo(
    () => Object.fromEntries(clips.flatMap((c) => (c.url ? [[idOf(c), c.url]] : []))),
    [clips],
  );

  // Выбранная строка пропала из списка (сменили фильтр) — выбираем первую видимую.
  useEffect(() => {
    if (!data || busy(take.phase)) return;
    if (!active || !visible.some((c) => idOf(c) === active)) setActive(visible[0] ? idOf(visible[0]) : null);
  });

  const select = (id: string) => {
    if (id === active || busy(take.phase)) return;
    take.cancel();
    setActive(id);
    document.getElementById(`clip-${id}`)?.scrollIntoView({ block: "nearest" });
  };

  const move = (delta: number) => {
    const i = visible.findIndex((c) => idOf(c) === active);
    const next = visible[i + delta];
    if (next) select(idOf(next));
  };

  const saveActive = async () => {
    const i = visible.findIndex((c) => idOf(c) === active);
    const clip = visible[i];
    if (!clip) return;
    const next = visible[i + 1];
    const saved = await take.save(async (audio) => {
      const url = await uploadClip(clip.kind, clip.key, audio);
      setData((d) => d && { ...d, clips: d.clips.map((c) => (idOf(c) === idOf(clip) ? { ...c, url } : c)) });
    });
    // Записанное уходит из «Не записаны», так что дальше — следующая строка списка.
    if (saved && next) {
      setActive(idOf(next));
      document.getElementById(`clip-${idOf(next)}`)?.scrollIntoView({ block: "nearest" });
    }
  };

  const toggleRecording = () => {
    if (!data?.canRecord || !active) return;
    if (take.phase.name === "recording") void take.stop();
    else void take.start();
  };

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      const p = take.phase.name;
      switch (e.key) {
        case " ":
          e.preventDefault();
          (document.activeElement as HTMLElement | null)?.blur();
          toggleRecording();
          break;
        case "Enter":
          if (p === "review") {
            e.preventDefault();
            void saveActive();
          }
          break;
        case "Escape":
          take.cancel();
          break;
        case "ArrowDown":
        case "ArrowUp":
          e.preventDefault();
          move(e.key === "ArrowDown" ? 1 : -1);
          break;
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  });

  return (
    <PlayerProvider recordings={recordings}>
      <Topbar>
        <a className="topbar__back" href="/">
          ← {ui.backHome}
        </a>
      </Topbar>

      <main className="page">
        <h1>{ui.recordTitle}</h1>
        {loadError && (
          <p className="callout">
            {ui.loadError} <code>{loadError}</code>
          </p>
        )}
        {data && !data.canRecord && <p className="callout">{ui.recordDisabled}</p>}
        {data?.canRecord && (
          <>
            <p className="rec-intro">{ui.recordIntro}</p>
            <p className="rec-keys">{ui.keysHint}</p>
          </>
        )}

        {data && (
          <div className="rec-filters">
            <div className="segmented" role="group">
              {(
                [
                  ["missing", ui.filterMissing, ofKind.length - doneCount],
                  ["done", ui.filterDone, doneCount],
                  ["all", ui.filterAll, ofKind.length],
                ] as const
              ).map(([value, label, count]) => (
                <button key={value} type="button" aria-pressed={status === value} onClick={() => setStatus(value)}>
                  {label} <span className="count">{count}</span>
                </button>
              ))}
            </div>
            <div className="segmented" role="group">
              {(
                [
                  ["all", ui.kindAll],
                  ["s", ui.kindSounds],
                  ["w", ui.kindWords],
                ] as const
              ).map(([value, label]) => (
                <button key={value} type="button" aria-pressed={kind === value} onClick={() => setKind(value)}>
                  {label}
                </button>
              ))}
            </div>
          </div>
        )}

        {data && visible.length === 0 && <p className="rec-empty">{ui.emptyList}</p>}

        <ul className="rec-list">
          {visible.map((c) => {
            const id = idOf(c);
            const isActive = id === active;
            const phase = isActive ? take.phase : null;
            return (
              <li
                key={id}
                id={`clip-${id}`}
                className="rec-row"
                aria-current={isActive || undefined}
                onClick={() => select(id)}
              >
                <span className="rec-row__status" data-done={Boolean(c.url)} title={c.url ? ui.filterDone : ui.filterMissing} />
                <span className="rec-row__key" lang="ru">
                  {c.key}
                </span>
                <span className="rec-row__meta">
                  <span className="rec-row__kind">{c.kind === "s" ? ui.kindSound : ui.kindWord}</span>
                  {c.ipa && <span className="rec-row__ipa">{t(c.ipa)}</span>}
                  {c.usedIn.map(
                    (s) =>
                      data?.sections[s] && (
                        <a key={s} className="rec-row__where" href={`/#${s}`} target="_blank" onClick={(e) => e.stopPropagation()}>
                          {t(data.sections[s])}
                        </a>
                      ),
                  )}
                </span>
                <span className="rec-row__actions" onClick={(e) => e.stopPropagation()}>
                  {c.url && (
                    <Clip kind={c.kind} clipKey={c.key} lang={lang}>
                      {ui.listen}
                    </Clip>
                  )}
                  {data?.canRecord && (
                    <button
                      type="button"
                      className="btn btn--rec"
                      data-recording={phase?.name === "recording" || undefined}
                      disabled={phase ? phase.name === "processing" || phase.name === "saving" : busy(take.phase)}
                      onClick={() => {
                        if (!isActive) {
                          take.cancel();
                          setActive(id);
                          void take.start();
                        } else toggleRecording();
                      }}
                    >
                      <span className="rec-dot" aria-hidden />
                      {phase?.name === "recording" ? ui.stop : c.url ? ui.rerecordAction : ui.recordAction}
                    </button>
                  )}
                </span>
                {phase && phase.name !== "idle" && (
                  <TakePanel phase={phase} ui={ui} level={take.level} onPlay={() => take.play()} onSave={saveActive} onAgain={() => void take.start()} onCancel={take.cancel} />
                )}
              </li>
            );
          })}
        </ul>
      </main>
    </PlayerProvider>
  );
}

type PanelProps = {
  phase: Phase;
  ui: Messages;
  level(): number;
  onPlay(): void;
  onSave(): void;
  onAgain(): void;
  onCancel(): void;
};

function TakePanel({ phase, ui, level, onPlay, onSave, onAgain, onCancel }: PanelProps) {
  return (
    <div className="take" onClick={(e) => e.stopPropagation()}>
      {phase.name === "recording" && (
        <>
          <span className="take__live">
            <span className="rec-dot rec-dot--live" aria-hidden />
            {ui.recordingNow}
          </span>
          <LevelMeter level={level} />
          <span className="take__keys">
            <kbd>␣</kbd> {ui.stop} · <kbd>Esc</kbd> {ui.cancel}
          </span>
        </>
      )}
      {phase.name === "processing" && <span className="take__note">{ui.processing}</span>}
      {(phase.name === "review" || phase.name === "saving") && (
        <>
          <button type="button" className="btn" onClick={onPlay}>
            <Speaker /> {ui.listen}
          </button>
          <span className="take__note">{ui.seconds(phase.take.duration)}</span>
          <button type="button" className="btn btn--primary" disabled={phase.name === "saving"} onClick={onSave}>
            {phase.name === "saving" ? ui.saving : ui.save} <kbd>Enter</kbd>
          </button>
          <button type="button" className="btn" disabled={phase.name === "saving"} onClick={onAgain}>
            {ui.again} <kbd>␣</kbd>
          </button>
          <button type="button" className="btn btn--quiet" disabled={phase.name === "saving"} onClick={onCancel}>
            {ui.cancel} <kbd>Esc</kbd>
          </button>
          {phase.name === "review" && phase.saveError && (
            <span className="take__error">
              {ui.saveFailed} {phase.saveError}
            </span>
          )}
        </>
      )}
      {phase.name === "error" && (
        <>
          <span className="take__error">
            {phase.reason === "mic" ? ui.micDenied : phase.reason === "silence" ? ui.silence : `${ui.saveFailed} ${phase.detail ?? ""}`}
          </span>
          <button type="button" className="btn" onClick={onAgain}>
            {ui.again} <kbd>␣</kbd>
          </button>
        </>
      )}
    </div>
  );
}

// Уровень микрофона рисуется напрямую в стиль, без перерисовки React на каждый кадр.
function LevelMeter({ level }: { level(): number }) {
  const bar = useRef<HTMLSpanElement>(null);
  useEffect(() => {
    let frame = 0;
    const tick = () => {
      const v = Math.min(1, Math.sqrt(level()));
      if (bar.current) {
        bar.current.style.transform = `scaleX(${v})`;
        bar.current.dataset.hot = String(v > 0.97);
      }
      frame = requestAnimationFrame(tick);
    };
    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, [level]);
  return (
    <span className="meter" aria-hidden>
      <span ref={bar} className="meter__bar" />
    </span>
  );
}
