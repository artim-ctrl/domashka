import type { PageResponse } from "./types";

export async function fetchPage(): Promise<PageResponse> {
  const res = await fetch("/api/page");
  if (!res.ok) throw new Error(`GET /api/page: ${res.status}`);
  return res.json();
}
