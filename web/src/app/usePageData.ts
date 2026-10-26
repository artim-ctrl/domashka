import { useCallback, useEffect, useState } from "react";
import { fetchPage } from "../api/page";
import type { Choice, PageResponse } from "../api/types";
import type { ClipId } from "../lib/clip";
import { errorMessage, isAbort } from "../lib/errors";

export type PageData =
    | { status: "loading" }
    | { status: "error"; error: string }
    | { status: "ready"; data: PageResponse };

export const usePageData = () => {
    const [state, setState] = useState<PageData>({ status: "loading" });

    useEffect(() => {
        const controller = new AbortController();
        fetchPage(controller.signal).then(
            (data) => setState({ status: "ready", data }),
            (e: unknown) => {
                if (!isAbort(e)) setState({ status: "error", error: errorMessage(e) });
            },
        );
        return () => controller.abort();
    }, []);

    const applyChoice = useCallback((id: ClipId, url: string, choice: Choice) => {
        setState((s) =>
            s.status === "ready"
                ? {
                      ...s,
                      data: {
                          ...s.data,
                          recordings: { ...s.data.recordings, [id]: url },
                          choices: { ...s.data.choices, [id]: choice },
                      },
                  }
                : s,
        );
    }, []);

    return { state, applyChoice };
};
