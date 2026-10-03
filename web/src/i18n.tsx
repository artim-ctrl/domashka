import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";

export type Lang = "es" | "en" | "ru";

export const LANGS: { value: Lang; short: string; name: string }[] = [
  { value: "es", short: "ES", name: "Español" },
  { value: "en", short: "EN", name: "English" },
  { value: "ru", short: "RU", name: "Русский" },
];

// Строка из page.json: общая для всех языков или с переводами.
export type Text = string | Record<Lang, string>;

const STORAGE_KEY = "domashka.lang";
const FALLBACK: Lang = "es";

const UI = {
  es: {
    sections: "Secciones",
    language: "Idioma",
    theme: "Tema",
    themeLight: "Claro",
    themeSystem: "Sistema",
    themeDark: "Oscuro",
    loading: "Cargando…",
    loadError: "No se pudo cargar la página.",
    alphabet: "Alfabeto ruso",
    colLetter: "Letra rusa",
    colSound: "Sonido",
    colSpanish: "Español",
    colEnglish: "Inglés",
    colExample: "Ejemplo",
    silent: "Sin sonido propio",
    focus: "Hay que practicar más",
    listenSound: (k: string) => `Escuchar el sonido: ${k}`,
    listenWord: (k: string) => `Escuchar la palabra: ${k}`,
    pending: (k: string) => `Grabación pendiente (${k})`,
    recordings: (n: number, total: number) => `Grabaciones: ${n} de ${total}`,
  },
  en: {
    sections: "Sections",
    language: "Language",
    theme: "Theme",
    themeLight: "Light",
    themeSystem: "System",
    themeDark: "Dark",
    loading: "Loading…",
    loadError: "Couldn't load the page.",
    alphabet: "Russian alphabet",
    colLetter: "Russian letter",
    colSound: "Sound",
    colSpanish: "Spanish",
    colEnglish: "English",
    colExample: "Example",
    silent: "No sound of its own",
    focus: "Needs extra practice",
    listenSound: (k: string) => `Listen to the sound: ${k}`,
    listenWord: (k: string) => `Listen to the word: ${k}`,
    pending: (k: string) => `Recording pending (${k})`,
    recordings: (n: number, total: number) => `Recordings: ${n} of ${total}`,
  },
  ru: {
    sections: "Разделы",
    language: "Язык",
    theme: "Тема",
    themeLight: "Светлая",
    themeSystem: "Как в системе",
    themeDark: "Тёмная",
    loading: "Загрузка…",
    loadError: "Не удалось загрузить страницу.",
    alphabet: "Русский алфавит",
    colLetter: "Буква",
    colSound: "Звук",
    colSpanish: "Испанский",
    colEnglish: "Английский",
    colExample: "Пример",
    silent: "Своего звука нет",
    focus: "Стоит потренировать",
    listenSound: (k: string) => `Послушать звук: ${k}`,
    listenWord: (k: string) => `Послушать слово: ${k}`,
    pending: (k: string) => `Записи пока нет (${k})`,
    recordings: (n: number, total: number) => `Записи: ${n} из ${total}`,
  },
} satisfies Record<Lang, unknown>;

export type Messages = (typeof UI)["es"];

type I18n = {
  lang: Lang;
  setLang(l: Lang): void;
  ui: Messages;
  t(text: Text): string;
};

const I18nContext = createContext<I18n | null>(null);

const isLang = (s: string | null | undefined): s is Lang => LANGS.some((l) => l.value === s);

// Сохранённый выбор, иначе язык браузера, иначе испанский.
function initialLang(): Lang {
  const saved = localStorage.getItem(STORAGE_KEY);
  if (isLang(saved)) return saved;
  for (const tag of navigator.languages) {
    const base = tag.slice(0, 2).toLowerCase();
    if (isLang(base)) return base;
  }
  return FALLBACK;
}

export function I18nProvider({ children }: { children: ReactNode }) {
  const [lang, setLangState] = useState<Lang>(initialLang);

  useEffect(() => {
    document.documentElement.lang = lang;
  }, [lang]);

  const setLang = useCallback((l: Lang) => {
    localStorage.setItem(STORAGE_KEY, l);
    setLangState(l);
  }, []);

  const t = useCallback((text: Text) => (typeof text === "string" ? text : (text[lang] ?? text[FALLBACK])), [lang]);

  const value = useMemo(() => ({ lang, setLang, ui: UI[lang], t }), [lang, setLang, t]);
  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n(): I18n {
  const ctx = useContext(I18nContext);
  if (!ctx) throw new Error("useI18n вне I18nProvider");
  return ctx;
}
