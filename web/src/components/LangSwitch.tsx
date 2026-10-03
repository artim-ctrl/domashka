import { LANGS, useI18n } from "../i18n";

export function LangSwitch() {
  const { lang, setLang, ui } = useI18n();
  return (
    <div className="segmented" role="group" aria-label={ui.language}>
      {LANGS.map((l) => (
        <button
          key={l.value}
          type="button"
          lang={l.value}
          aria-pressed={lang === l.value}
          title={l.name}
          onClick={() => setLang(l.value)}
        >
          {l.short}
        </button>
      ))}
    </div>
  );
}
