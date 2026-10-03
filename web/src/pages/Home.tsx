import { useEffect, useState } from "react";
import { fetchPage } from "../api";
import { BlockView } from "../components/Blocks";
import { Topbar } from "../components/Topbar";
import { useI18n } from "../i18n";
import { PlayerProvider } from "../player";
import type { PageResponse } from "../types";

export function HomePage() {
  const { ui, t } = useI18n();
  const [data, setData] = useState<PageResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

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

  return (
    <PlayerProvider recordings={recordings}>
      <Topbar>
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
        {data.canRecord && (
          <a className="topbar__action" href="/record">
            <span className="rec-dot" aria-hidden />
            {ui.recordLink}
          </a>
        )}
      </Topbar>

      <main className="page">
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
    </PlayerProvider>
  );
}
