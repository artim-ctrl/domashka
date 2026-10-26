import { type ReactNode, useCallback, useEffect, useMemo, useState } from "react";
import { type I18n, I18nContext } from "./context";
import { FALLBACK_LANG, isLang, type Lang, type Text, translate } from "./languages";
import { en } from "./locales/en";
import { es } from "./locales/es";
import { ru } from "./locales/ru";
import type { Messages } from "./messages";

const STORAGE_KEY = "domashka.lang";

const MESSAGES: Record<Lang, Messages> = { es, en, ru };

const initialLang = (): Lang => {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (isLang(saved)) return saved;
    for (const tag of navigator.languages) {
        const base = tag.slice(0, 2).toLowerCase();
        if (isLang(base)) return base;
    }
    return FALLBACK_LANG;
};

export const I18nProvider = ({ children }: { children: ReactNode }) => {
    const [lang, setLangState] = useState<Lang>(initialLang);

    useEffect(() => {
        document.documentElement.lang = lang;
    }, [lang]);

    const setLang = useCallback((next: Lang) => {
        localStorage.setItem(STORAGE_KEY, next);
        setLangState(next);
    }, []);

    const t = useCallback((text: Text) => translate(text, lang), [lang]);

    const value = useMemo<I18n>(() => ({ lang, setLang, ui: MESSAGES[lang], t }), [lang, setLang, t]);
    return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
};
