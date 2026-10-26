export class ApiError extends Error {
    readonly status: number;

    constructor(status: number, message: string) {
        super(message);
        this.name = "ApiError";
        this.status = status;
    }
}

type Method = "GET" | "POST";

export const request = async <T>(method: Method, path: string, body?: unknown, signal?: AbortSignal): Promise<T> => {
    const res = await fetch(path, {
        method,
        headers: body === undefined ? undefined : { "Content-Type": "application/json" },
        body: body === undefined ? undefined : JSON.stringify(body),
        signal,
    });
    if (!res.ok) throw new ApiError(res.status, await readError(res));
    return (await res.json()) as T;
};

const readError = async (res: Response): Promise<string> => {
    const fallback = `${res.status} ${res.statusText}`.trim();
    try {
        const body = (await res.json()) as { error?: unknown };
        return typeof body.error === "string" && body.error ? body.error : fallback;
    } catch {
        return fallback;
    }
};
