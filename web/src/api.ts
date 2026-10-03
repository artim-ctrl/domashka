import type { ClipKind, ClipsResponse, PageResponse } from "./types";

async function getJSON<T>(path: string): Promise<T> {
  const res = await fetch(path);
  if (!res.ok) throw new Error(`GET ${path}: ${res.status}`);
  return res.json();
}

export const fetchPage = () => getJSON<PageResponse>("/api/page");

export const fetchClips = () => getJSON<ClipsResponse>("/api/clips");

// Сохраняет запись и возвращает её новый адрес.
export async function uploadClip(kind: ClipKind, key: string, audio: Blob): Promise<string> {
  const res = await fetch(`/api/audio/${kind}/${encodeURIComponent(key)}`, {
    method: "PUT",
    headers: { "Content-Type": audio.type },
    body: audio,
  });
  if (!res.ok) throw new Error(`${res.status}: ${(await res.text()).trim()}`);
  return ((await res.json()) as { url: string }).url;
}
