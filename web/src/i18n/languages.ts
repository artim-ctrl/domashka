export type Lang = "es" | "en" | "ru";

export const LANGS: readonly { value: Lang; short: string; name: string }[] = [
    { value: "es", short: "ES", name: "Español" },
    { value: "en", short: "EN", name: "English" },
    { value: "ru", short: "RU", name: "Русский" },
];

export const FALLBACK_LANG: Lang = "es";

export type Text = string | Record<Lang, string>;

export const isLang = (s: string | null | undefined): s is Lang => LANGS.some((l) => l.value === s);

export const translate = (text: Text, lang: Lang): string =>
    typeof text === "string" ? text : (text[lang] ?? text[FALLBACK_LANG]);
