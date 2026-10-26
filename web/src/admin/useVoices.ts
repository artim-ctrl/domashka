import { useEffect, useState } from "react";
import { fetchVoices } from "../api/admin";
import type { Voice } from "../api/types";
import { errorMessage } from "../lib/errors";

let cache: Promise<Voice[]> | null = null;

const loadVoices = (): Promise<Voice[]> => {
    cache ??= fetchVoices().catch((e: unknown) => {
        cache = null;
        throw e;
    });
    return cache;
};

export const useVoices = () => {
    const [voices, setVoices] = useState<Voice[] | null>(null);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        let alive = true;
        loadVoices().then(
            (vs) => {
                if (alive) setVoices(vs);
            },
            (e: unknown) => {
                if (alive) setError(errorMessage(e));
            },
        );
        return () => {
            alive = false;
        };
    }, []);

    return { voices, error };
};
