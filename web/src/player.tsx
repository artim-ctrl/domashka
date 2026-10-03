import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import type { ClipKind } from "./types";

type Player = {
  urlOf(kind: ClipKind, key: string): string | undefined;
  playing: string | null;
  play(kind: ClipKind, key: string): void;
  playUrl(id: string, url: string): void; // звук не со страницы — например, вариант в админке
};

const PlayerContext = createContext<Player | null>(null);

export const clipId = (kind: ClipKind, key: string) => `${kind}:${key}`;

// Один <audio> на всю страницу: новый звук обрывает предыдущий,
// повторное нажатие проигрывает запись с начала.
export function PlayerProvider({ recordings, children }: { recordings: Record<string, string>; children: ReactNode }) {
  const audio = useRef<HTMLAudioElement | null>(null);
  const attempt = useRef(0);
  const [playing, setPlaying] = useState<string | null>(null);

  useEffect(() => {
    const a = new Audio();
    const stop = () => setPlaying(null);
    a.addEventListener("ended", stop);
    a.addEventListener("error", stop);
    audio.current = a;
    return () => {
      a.pause();
      a.removeEventListener("ended", stop);
      a.removeEventListener("error", stop);
    };
  }, []);

  const urlOf = useCallback((kind: ClipKind, key: string) => recordings[clipId(kind, key)], [recordings]);

  const playUrl = useCallback((id: string, src: string) => {
    const a = audio.current;
    if (!a) return;
    const mine = ++attempt.current;
    a.pause();
    if (a.getAttribute("src") !== src) a.src = src;
    a.currentTime = 0;
    setPlaying(id);
    // Прерванный следующим нажатием play() отклоняется — это не ошибка текущего звука.
    a.play().catch(() => {
      if (attempt.current === mine) setPlaying(null);
    });
  }, []);

  const play = useCallback(
    (kind: ClipKind, key: string) => {
      const src = urlOf(kind, key);
      if (src) playUrl(clipId(kind, key), src);
    },
    [urlOf, playUrl],
  );

  const value = useMemo(() => ({ urlOf, playing, play, playUrl }), [urlOf, playing, play, playUrl]);
  return <PlayerContext.Provider value={value}>{children}</PlayerContext.Provider>;
}

export function usePlayer(): Player {
  const p = useContext(PlayerContext);
  if (!p) throw new Error("usePlayer вне PlayerProvider");
  return p;
}
