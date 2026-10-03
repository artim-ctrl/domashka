import type { ReactNode } from "react";
import { clipId, usePlayer } from "../player";
import type { ClipKind } from "../types";

type Props = {
  kind: ClipKind;
  clipKey: string;
  children?: ReactNode;
  variant?: "inline" | "tile";
  className?: string;
};

// Кнопка, которая проигрывает запись звука или слова. Без записи — видна, но неактивна.
export function Clip({ kind, clipKey, children, variant = "inline", className = "" }: Props) {
  const { urlOf, playing, play } = usePlayer();
  const available = urlOf(kind, clipKey) !== undefined;
  const isPlaying = playing === clipId(kind, clipKey);
  const what = kind === "s" ? "sonido" : "palabra";

  return (
    <button
      type="button"
      className={`clip clip--${variant} clip--${kind === "s" ? "sound" : "word"} ${className}`}
      data-state={isPlaying ? "playing" : available ? "ready" : "missing"}
      disabled={!available}
      onClick={() => play(kind, clipKey)}
      title={available ? `Escuchar ${what}: ${clipKey}` : `Grabación pendiente (${clipKey})`}
      lang="ru"
    >
      <span className="clip__label">{children ?? clipKey}</span>
      <span className="clip__icon" aria-hidden>
        {isPlaying ? <Bars /> : <Speaker />}
      </span>
    </button>
  );
}

function Speaker() {
  return (
    <svg viewBox="0 0 16 16" width="1em" height="1em">
      <path d="M2.5 6h2.2L8 3.2v9.6L4.7 10H2.5z" fill="currentColor" />
      <path
        d="M10.4 5.6a3.4 3.4 0 0 1 0 4.8M12.2 3.8a6 6 0 0 1 0 8.4"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.4"
        strokeLinecap="round"
      />
    </svg>
  );
}

function Bars() {
  return (
    <span className="bars">
      <i />
      <i />
      <i />
    </span>
  );
}
