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
    recordLink: "Grabar",
    recordTitle: "Grabar sonidos y palabras",
    recordIntro:
      "Elige una fila y pulsa Espacio para grabar. Al parar se recortan los silencios y se ajusta el volumen: escucha el resultado y guárdalo con Enter.",
    backHome: "Pronunciación",
    recordDisabled: "La grabación está desactivada. Inicia el servidor con -record (make api).",
    filterAll: "Todos",
    filterMissing: "Sin grabar",
    filterDone: "Grabados",
    kindAll: "Todo",
    kindSounds: "Sonidos",
    kindWords: "Palabras",
    kindSound: "sonido",
    kindWord: "palabra",
    recordAction: "Grabar",
    rerecordAction: "Regrabar",
    stop: "Parar",
    save: "Guardar",
    again: "Otra vez",
    cancel: "Cancelar",
    listen: "Escuchar",
    recordingNow: "Grabando…",
    processing: "Procesando…",
    saving: "Guardando…",
    silence: "No se oye nada. ¿Funciona el micrófono?",
    micDenied: "Sin acceso al micrófono.",
    saveFailed: "No se pudo guardar.",
    emptyList: "No hay nada con este filtro.",
    keysHint: "Espacio — grabar / parar · Enter — guardar · Esc — cancelar · ↑ ↓ — elegir",
    seconds: (s: number) => `${s.toFixed(1)} s`,
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
    recordLink: "Record",
    recordTitle: "Record sounds and words",
    recordIntro:
      "Pick a row and press Space to record. When you stop, silence is trimmed and the volume is levelled: listen to the result and save it with Enter.",
    backHome: "Pronunciation",
    recordDisabled: "Recording is turned off. Start the server with -record (make api).",
    filterAll: "All",
    filterMissing: "Not recorded",
    filterDone: "Recorded",
    kindAll: "All",
    kindSounds: "Sounds",
    kindWords: "Words",
    kindSound: "sound",
    kindWord: "word",
    recordAction: "Record",
    rerecordAction: "Re-record",
    stop: "Stop",
    save: "Save",
    again: "Again",
    cancel: "Cancel",
    listen: "Listen",
    recordingNow: "Recording…",
    processing: "Processing…",
    saving: "Saving…",
    silence: "Nothing was heard. Is the microphone working?",
    micDenied: "No access to the microphone.",
    saveFailed: "Couldn't save.",
    emptyList: "Nothing matches this filter.",
    keysHint: "Space — record / stop · Enter — save · Esc — cancel · ↑ ↓ — select",
    seconds: (s: number) => `${s.toFixed(1)} s`,
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
    recordLink: "Запись",
    recordTitle: "Запись звуков и слов",
    recordIntro:
      "Выбери строку и нажми пробел, чтобы записать. После остановки тишина по краям обрежется, а громкость выровняется: послушай результат и сохрани его клавишей Enter.",
    backHome: "Произношение",
    recordDisabled: "Запись выключена. Запусти сервер с -record (make api).",
    filterAll: "Все",
    filterMissing: "Не записаны",
    filterDone: "Записаны",
    kindAll: "Всё",
    kindSounds: "Звуки",
    kindWords: "Слова",
    kindSound: "звук",
    kindWord: "слово",
    recordAction: "Записать",
    rerecordAction: "Перезаписать",
    stop: "Стоп",
    save: "Сохранить",
    again: "Ещё раз",
    cancel: "Отмена",
    listen: "Послушать",
    recordingNow: "Идёт запись…",
    processing: "Обработка…",
    saving: "Сохранение…",
    silence: "Ничего не слышно. Микрофон работает?",
    micDenied: "Нет доступа к микрофону.",
    saveFailed: "Не удалось сохранить.",
    emptyList: "С таким фильтром ничего нет.",
    keysHint: "Пробел — запись / стоп · Enter — сохранить · Esc — отмена · ↑ ↓ — выбор",
    seconds: (s: number) => `${s.toFixed(1)} с`,
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
