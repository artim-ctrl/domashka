export const errorMessage = (e: unknown): string => (e instanceof Error ? e.message : String(e));

export const isAbort = (e: unknown): boolean => e instanceof DOMException && e.name === "AbortError";
