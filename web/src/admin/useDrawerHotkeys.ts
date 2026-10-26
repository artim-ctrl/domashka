import { useEffect, useEffectEvent } from "react";

type Handlers = {
    onDigit(index: number): void;
    onPrevious(): void;
    onNext(): void;
    onClose(): void;
};

const isTyping = (target: EventTarget | null) =>
    target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement;

export const useDrawerHotkeys = (handlers: Handlers) => {
    const onKeyDown = useEffectEvent((e: KeyboardEvent) => {
        if (e.metaKey || e.ctrlKey || e.altKey || isTyping(e.target)) return;
        if (/^[0-9]$/.test(e.key)) handlers.onDigit((Number(e.key) + 9) % 10);
        else if (e.key === "ArrowLeft") handlers.onPrevious();
        else if (e.key === "ArrowRight") handlers.onNext();
        else if (e.key === "Escape") handlers.onClose();
    });

    useEffect(() => {
        window.addEventListener("keydown", onKeyDown);
        return () => window.removeEventListener("keydown", onKeyDown);
    }, []);
};
