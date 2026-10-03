import { useCallback, useEffect, useRef, useState } from "react";
import { Mic, type Recorder } from "./mic";
import { processTake, SilenceError, type Take } from "./process";

const MAX_SECONDS = 10; // забытая запись остановится сама

export type Phase =
  | { name: "idle" }
  | { name: "recording" }
  | { name: "processing" }
  | { name: "review"; take: Take; url: string; saveError?: string }
  | { name: "saving"; take: Take; url: string }
  | { name: "error"; reason: "mic" | "silence" | "failed"; detail?: string };

// Один дубль: запись → обработка → прослушивание → сохранение.
export function useTake() {
  const [phase, setPhaseState] = useState<Phase>({ name: "idle" });
  const phaseRef = useRef(phase);
  const mic = useRef<Mic | null>(null);
  const recorder = useRef<Recorder | null>(null);
  const timer = useRef<number | undefined>(undefined);
  const preview = useRef<HTMLAudioElement | null>(null);

  // Ссылку на прослушиваемый дубль освобождаем, как только он больше не нужен.
  const setPhase = useCallback((next: Phase) => {
    const prev = phaseRef.current;
    const prevUrl = "url" in prev ? prev.url : null;
    if (prevUrl && !("url" in next && next.url === prevUrl)) {
      preview.current?.pause();
      URL.revokeObjectURL(prevUrl);
    }
    phaseRef.current = next;
    setPhaseState(next);
  }, []);

  useEffect(
    () => () => {
      window.clearTimeout(timer.current);
      recorder.current?.cancel();
      mic.current?.close();
      preview.current?.pause();
    },
    [],
  );

  const play = useCallback((url?: string) => {
    const p = phaseRef.current;
    const src = url ?? ("url" in p ? p.url : null);
    if (!src) return;
    preview.current ??= new Audio();
    preview.current.src = src;
    void preview.current.play().catch(() => {});
  }, []);

  const stop = useCallback(async () => {
    window.clearTimeout(timer.current);
    const r = recorder.current;
    if (!r) return;
    recorder.current = null;
    setPhase({ name: "processing" });
    try {
      const take = await processTake(await r.stop());
      const url = URL.createObjectURL(take.blob);
      setPhase({ name: "review", take, url });
      play(url);
    } catch (e) {
      setPhase(e instanceof SilenceError ? { name: "error", reason: "silence" } : { name: "error", reason: "failed", detail: String(e) });
    }
  }, [play, setPhase]);

  const start = useCallback(async () => {
    const p = phaseRef.current.name;
    if (p === "recording" || p === "processing" || p === "saving") return;
    setPhase({ name: "idle" });
    try {
      mic.current ??= await Mic.open();
    } catch (e) {
      setPhase({ name: "error", reason: "mic", detail: String(e) });
      return;
    }
    recorder.current = mic.current.record();
    setPhase({ name: "recording" });
    timer.current = window.setTimeout(() => void stop(), MAX_SECONDS * 1000);
  }, [setPhase, stop]);

  const cancel = useCallback(() => {
    const p = phaseRef.current.name;
    if (p === "processing" || p === "saving") return;
    window.clearTimeout(timer.current);
    recorder.current?.cancel();
    recorder.current = null;
    setPhase({ name: "idle" });
  }, [setPhase]);

  // upload отправляет дубль на сервер; при ошибке дубль остаётся, можно попробовать ещё.
  const save = useCallback(
    async (upload: (audio: Blob) => Promise<void>): Promise<boolean> => {
      const p = phaseRef.current;
      if (p.name !== "review") return false;
      setPhase({ name: "saving", take: p.take, url: p.url });
      try {
        await upload(p.take.blob);
        setPhase({ name: "idle" });
        return true;
      } catch (e) {
        setPhase({ name: "review", take: p.take, url: p.url, saveError: String(e) });
        return false;
      }
    },
    [setPhase],
  );

  const level = useCallback(() => mic.current?.level() ?? 0, []);

  return { phase, start, stop, cancel, save, play, level };
}
