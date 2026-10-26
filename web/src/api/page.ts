import { request } from "./client";
import type { PageResponse } from "./types";

export const fetchPage = (signal?: AbortSignal) => request<PageResponse>("GET", "api/page", undefined, signal);
