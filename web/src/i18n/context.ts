import { createContext, useContext } from "react";
import type { Lang, Text } from "./languages";
import type { Messages } from "./messages";

export type I18n = {
    lang: Lang;
    setLang(lang: Lang): void;
    ui: Messages;
    t(text: Text): string;
};

export const I18nContext = createContext<I18n | null>(null);

export const useI18n = (): I18n => {
    const ctx = useContext(I18nContext);
    if (!ctx) throw new Error("useI18n must be used inside I18nProvider");
    return ctx;
};
