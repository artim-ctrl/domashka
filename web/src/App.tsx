import { useEffect, useState } from "react";
import { fetchPage } from "./api";
import { BlockView } from "./components/Blocks";
import { ThemeSwitch } from "./components/ThemeSwitch";
import { PlayerProvider } from "./player";
import type { PageResponse } from "./types";

export function App() {
  const [data, setData] = useState<PageResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchPage().then(setData, (e: unknown) => setError(String(e)));
  }, []);

  // Разделы появляются после загрузки, поэтому к якорю из адреса браузер сам не прокрутит.
  useEffect(() => {
    if (data && location.hash) document.getElementById(location.hash.slice(1))?.scrollIntoView();
  }, [data]);

  if (error) {
    return (
      <main className="page page--status">
        <p>No se pudo cargar la página.</p>
        <code>{error}</code>
      </main>
    );
  }
  if (!data) return <main className="page page--status">Cargando…</main>;

  const { page, recordings, total } = data;
  const titled = page.sections.filter((s) => s.title);
  const recorded = Object.keys(recordings).length;

  return (
    <PlayerProvider recordings={recordings}>
      <header className="topbar">
        <nav className="topbar__nav" aria-label="Secciones">
          {titled.map((s) => (
            <a key={s.id} href={`#${s.id}`}>
              {s.title}
            </a>
          ))}
        </nav>
        <ThemeSwitch />
      </header>

      <main className="page">
        <h1>{page.title}</h1>
        {page.sections.map((s) => (
          <section key={s.id} id={s.id} className="section">
            {s.title && <h2>{s.title}</h2>}
            {s.blocks.map((b, i) => (
              <BlockView key={i} block={b} />
            ))}
          </section>
        ))}
      </main>

      <footer className="footer">
        Grabaciones: {recorded} de {total}
      </footer>
    </PlayerProvider>
  );
}
