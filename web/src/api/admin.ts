import { type ClipId, parseClipId } from "../lib/clip";
import { request } from "./client";
import type { Candidate, ChooseResult, ClipState, Voice } from "./types";

const clipPath = (id: ClipId) => {
    const { kind, key } = parseClipId(id);
    return `api/admin/clips/${kind}/${encodeURIComponent(key)}`;
};

export const fetchVoices = () => request<Voice[]>("GET", "api/admin/voices");

export const fetchClipState = (id: ClipId, signal?: AbortSignal) =>
    request<ClipState>("GET", clipPath(id), undefined, signal);

export const generateCandidate = (id: ClipId, voice: string, text: string, signal?: AbortSignal) =>
    request<Candidate>("POST", `${clipPath(id)}/candidates/${encodeURIComponent(voice)}`, { text }, signal);

export const chooseCandidate = (id: ClipId, voice: string) =>
    request<ChooseResult>("POST", `${clipPath(id)}/choose`, { voice });
