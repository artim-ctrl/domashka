import { useEffect, useMemo, useState } from "react";
import { AdminPanel } from "./admin/AdminPanel";
import { AdminContext, adminRequested, type Admin } from "./admin/context";
import { fetchPage } from "./api";
import { BlockView } from "./components/Blocks";
import { LangSwitch } from "./components/LangSwitch";
import { ThemeSwitch } from "./components/ThemeSwitch";
import { useI18n } from "./i18n";
import { PlayerProvider } from "./player";
import type { Choice, PageResponse } from "./types";

export function App() {
  const { ui, t } = useI18n();
  const [data, setData] = useState<PageResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [active, setActive] = useState<string | null>(null);

  useEffect(() => {
    fetchPage().then(setData, (e: unknown) => setError(String(e)));
  }, []);

  // Разделы появляются после загрузки, поэтому к якорю из адреса браузер сам не прокрутит.
  useEffect(() => {
    if (data && location.hash) document.getElementById(location.hash.slice(1))?.scrollIntoView();
  }, [data]);

  useEffect(() => {
    if (data) document.title = t(data.page.title);
  }, [data, t]);

  const admin = useMemo<Admin>(
    () => ({
      enabled: Boolean(data?.admin) && adminRequested(),
      active,
      open: setActive,
      choices: data?.choices ?? {},
    }),
    [data, active],
  );

  if (error) {
    return (
      <main className="page page--status">
        <p>{ui.loadError}</p>
        <code>{error}</code>
      </main>
    );
  }
  if (!data) return <main className="page page--status">{ui.loading}</main>;

  const { page, recordings, total } = data;
  const recorded = Object.keys(recordings).length;

  // Выбранный в админке вариант сразу становится записью на странице.
  const onChosen = (id: string, url: string, choice: Choice) =>
    setData((d) => d && { ...d, recordings: { ...d.recordings, [id]: url }, choices: { ...d.choices, [id]: choice } });

  return (
    <AdminContext.Provider value={admin}>
      <PlayerProvider recordings={recordings}>
        <header className="topbar">
          <nav className="topbar__nav" aria-label={ui.sections}>
            {page.sections.map(
              (s) =>
                s.title && (
                  <a key={s.id} href={`#${s.id}`}>
                    {t(s.title)}
                  </a>
                ),
            )}
          </nav>
          <LangSwitch />
          <ThemeSwitch />
        </header>

        <main className="page" data-admin-open={admin.enabled && active ? "" : undefined}>
          <h1>{t(page.title)}</h1>
          {page.sections.map((s) => (
            <section key={s.id} id={s.id} className="section">
              {s.title && <h2>{t(s.title)}</h2>}
              {s.blocks.map((b, i) => (
                <BlockView key={i} block={b} />
              ))}
            </section>
          ))}
        </main>

        <footer className="footer">{ui.recordings(recorded, total)}</footer>

        {admin.enabled && <AdminPanel clips={data.clips ?? []} onChosen={onChosen} onClose={() => setActive(null)} />}
        {import.meta.env.DEV && adminRequested() && !data.admin && (
          <div className="admin-pill">Админка выключена на сервере — запусти его через make api (флаг -admin).</div>
        )}
      </PlayerProvider>
    </AdminContext.Provider>
  );
}
