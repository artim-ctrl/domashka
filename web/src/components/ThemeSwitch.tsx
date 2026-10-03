import { useEffect, useState } from "react";
import { useI18n } from "../i18n";

type Theme = "light" | "system" | "dark";

const STORAGE_KEY = "domashka.theme"; // его же читает скрипт в index.html

const OPTIONS: { value: Theme; label: "themeLight" | "themeSystem" | "themeDark"; icon: string }[] = [
  { value: "light", label: "themeLight", icon: "M8 4.5a3.5 3.5 0 1 0 0 7 3.5 3.5 0 0 0 0-7ZM8 1v1.6M8 13.4V15M1 8h1.6M13.4 8H15M3 3l1.1 1.1M11.9 11.9 13 13M3 13l1.1-1.1M11.9 4.1 13 3" },
  { value: "system", label: "themeSystem", icon: "M2 3.5h12v7.5H2zM5.5 14h5M8 11v3" },
  { value: "dark", label: "themeDark", icon: "M13.5 9.8A5.8 5.8 0 0 1 6.2 2.5a5.8 5.8 0 1 0 7.3 7.3Z" },
];

export function ThemeSwitch() {
  const { ui } = useI18n();
  const [theme, setTheme] = useState<Theme>(() => (document.documentElement.dataset.theme as Theme) || "system");

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    localStorage.setItem(STORAGE_KEY, theme);
  }, [theme]);

  return (
    <div className="segmented" role="group" aria-label={ui.theme}>
      {OPTIONS.map((o) => {
        const label = ui[o.label];
        return (
          <button key={o.value} type="button" aria-pressed={theme === o.value} title={label} onClick={() => setTheme(o.value)}>
            <svg viewBox="0 0 16 16" width="16" height="16" aria-hidden>
              <path d={o.icon} fill="none" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
            <span className="sr-only">{label}</span>
          </button>
        );
      })}
    </div>
  );
}
