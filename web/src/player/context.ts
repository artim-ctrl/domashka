import { createContext, useContext } from "react";
import type { ClipId } from "../lib/clip";

export type Player = {
    urlOf(id: ClipId): string | undefined;
    playing: string | null;
    play(id: ClipId): void;
    playUrl(playbackId: string, url: string): void;
};

export const PlayerContext = createContext<Player | null>(null);

export const usePlayer = (): Player => {
    const ctx = useContext(PlayerContext);
    if (!ctx) throw new Error("usePlayer must be used inside PlayerProvider");
    return ctx;
};
