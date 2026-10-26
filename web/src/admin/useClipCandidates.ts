import { useCallback, useEffect, useRef, useState } from "react";
import { chooseCandidate, fetchClipState, generateCandidate } from "../api/admin";
import type { Candidate, Choice, ClipState, Voice } from "../api/types";
import type { ClipId } from "../lib/clip";
import { errorMessage, isAbort } from "../lib/errors";

type ByVoice<T> = Partial<Record<string, T>>;

const upsert = (list: Candidate[], cand: Candidate, order: Voice[]): Candidate[] => {
    const position = (c: Candidate) => order.findIndex((v) => v.id === c.voice);
    return [...list.filter((c) => c.voice !== cand.voice), cand].toSorted((a, b) => position(a) - position(b));
};

export const useClipCandidates = (
    id: ClipId,
    voices: Voice[] | null,
    onChosen: (id: ClipId, url: string, choice: Choice) => void,
) => {
    const [clip, setClip] = useState<ClipState | null>(null);
    const [loadError, setLoadError] = useState<string | null>(null);
    const [busy, setBusy] = useState<ByVoice<boolean>>({});
    const [errors, setErrors] = useState<ByVoice<string>>({});
    const [choosing, setChoosing] = useState<string | null>(null);
    const lifetime = useRef<AbortController | null>(null);

    const generateWith = useCallback(
        async (vs: Voice[], text: string, signal: AbortSignal) => {
            for (const v of vs) {
                if (signal.aborted) return;
                setBusy((b) => ({ ...b, [v.id]: true }));
                setErrors((e) => ({ ...e, [v.id]: undefined }));
                try {
                    const cand = await generateCandidate(id, v.id, text, signal);
                    setClip((s) => s && { ...s, text, candidates: upsert(s.candidates, cand, vs) });
                } catch (e) {
                    if (isAbort(e)) return;
                    setErrors((errs) => ({ ...errs, [v.id]: errorMessage(e) }));
                } finally {
                    if (!signal.aborted) setBusy((b) => ({ ...b, [v.id]: false }));
                }
            }
        },
        [id],
    );

    useEffect(() => {
        if (!voices) return;
        const controller = new AbortController();
        lifetime.current = controller;
        fetchClipState(id, controller.signal)
            .then(async (state) => {
                setClip(state);
                if (state.candidates.length === 0) await generateWith(voices, state.text, controller.signal);
            })
            .catch((e: unknown) => {
                if (!isAbort(e)) setLoadError(errorMessage(e));
            });
        return () => controller.abort();
    }, [id, voices, generateWith]);

    const generating = Object.values(busy).some(Boolean);

    const generate = (text: string) => {
        const signal = lifetime.current?.signal;
        if (voices && signal && !generating) void generateWith(voices, text, signal);
    };

    const choose = async (voice: string) => {
        if (choosing) return;
        setChoosing(voice);
        setErrors((e) => ({ ...e, [voice]: undefined }));
        try {
            const res = await chooseCandidate(id, voice);
            onChosen(id, res.url, res.chosen);
            setClip((s) => s && { ...s, chosen: res.chosen });
        } catch (e) {
            setErrors((errs) => ({ ...errs, [voice]: errorMessage(e) }));
        } finally {
            setChoosing(null);
        }
    };

    return { clip, loadError, busy, errors, generating, choosing, generate, choose };
};
