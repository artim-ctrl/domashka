import { type ReactNode, useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { Recordings } from "../api/types";
import type { ClipId } from "../lib/clip";
import { type Player, PlayerContext } from "./context";

type Props = {
    recordings: Recordings;
    children: ReactNode;
};

export const PlayerProvider = ({ recordings, children }: Props) => {
    const audioRef = useRef<HTMLAudioElement | null>(null);
    const attempt = useRef(0);
    const [playing, setPlaying] = useState<string | null>(null);

    useEffect(() => {
        const audio = new Audio();
        const stop = () => setPlaying(null);
        audio.addEventListener("ended", stop);
        audio.addEventListener("error", stop);
        audioRef.current = audio;
        return () => {
            audio.pause();
            audio.removeEventListener("ended", stop);
            audio.removeEventListener("error", stop);
        };
    }, []);

    const urlOf = useCallback((id: ClipId) => recordings[id], [recordings]);

    const playUrl = useCallback((playbackId: string, src: string) => {
        const audio = audioRef.current;
        if (!audio) return;
        const mine = ++attempt.current;
        audio.pause();
        if (audio.getAttribute("src") !== src) audio.src = src;
        audio.currentTime = 0;
        setPlaying(playbackId);
        audio.play().catch(() => {
            if (attempt.current === mine) setPlaying(null);
        });
    }, []);

    const play = useCallback(
        (id: ClipId) => {
            const src = urlOf(id);
            if (src) playUrl(id, src);
        },
        [urlOf, playUrl],
    );

    const value = useMemo<Player>(() => ({ urlOf, playing, play, playUrl }), [urlOf, playing, play, playUrl]);
    return <PlayerContext.Provider value={value}>{children}</PlayerContext.Provider>;
};
