import type { Candidate, Choice, ClipState, PageResponse, Voice } from "./types";

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(path, {
    method,
    headers: body ? { "Content-Type": "application/json" } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) throw new Error((await res.text()).trim() || `${method} ${path}: ${res.status}`);
  return res.json();
}

export const fetchPage = () => request<PageResponse>("GET", "/api/page");

// ── Админка: id записи вида «w:дом» ──

const clipPath = (id: string) => {
  const [kind, key] = [id.slice(0, 1), id.slice(2)];
  return `/api/admin/clips/${kind}/${encodeURIComponent(key)}`;
};

export const fetchVoices = () => request<Voice[]>("GET", "/api/admin/voices");

export const fetchClipState = (id: string) => request<ClipState>("GET", clipPath(id));

export const generateCandidate = (id: string, voice: string, text: string) =>
  request<Candidate>("POST", `${clipPath(id)}/candidates/${voice}`, { text });

export const chooseCandidate = (id: string, voice: string) =>
  request<{ url: string; chosen: Choice }>("POST", `${clipPath(id)}/choose`, { voice });
