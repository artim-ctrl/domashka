import type { Block, LetterRow } from "../types";
import { Clip } from "./Clip";
import { Rich } from "./Rich";

export function BlockView({ block }: { block: Block }) {
  switch (block.type) {
    case "paragraph":
      return (
        <p>
          <Rich text={block.text} />
        </p>
      );
    case "callout":
      return (
        <p className="callout">
          <Rich text={block.text} />
        </p>
      );
    case "list":
      return (
        <ul className={block.style === "focus" ? "list list--focus" : "list"}>
          {block.items.map((item, i) => (
            <li key={i}>
              <Rich text={item} />
            </li>
          ))}
        </ul>
      );
    case "table":
      return (
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                {block.columns.map((c) => (
                  <th key={c}>{c}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {block.rows.map((row, i) => (
                <tr key={i}>
                  {row.map((cell, j) => (
                    <td key={j}>
                      <Rich text={cell} />
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      );
    case "alphabet":
      return <Alphabet letters={block.letters} />;
  }
}

function Alphabet({ letters }: { letters: LetterRow[] }) {
  return (
    <div className="alphabet" role="table" aria-label="Alfabeto ruso">
      <div className="alphabet__row alphabet__head" role="row">
        <span role="columnheader">Letra rusa</span>
        <span role="columnheader">Sonido</span>
        <span role="columnheader">Español</span>
        <span role="columnheader">Inglés</span>
        <span role="columnheader">Ejemplo</span>
      </div>
      {letters.map((l) => (
        <div className="alphabet__row" role="row" key={l.letter}>
          <span className="alphabet__letter" role="cell">
            {l.sound ? (
              <Clip kind="s" clipKey={l.sound} variant="tile">
                {l.letter}
              </Clip>
            ) : (
              <span className="tile tile--silent" lang="ru" title="Sin sonido propio">
                {l.letter}
              </span>
            )}
            {l.focus && <span className="focus-dot" title="Hay que practicar más" />}
          </span>
          <span className="alphabet__ipa" role="cell">
            {l.ipa}
          </span>
          <span className="alphabet__text" role="cell" data-label="Español">
            <Rich text={l.es} />
          </span>
          <span className="alphabet__text" role="cell" data-label="Inglés">
            <Rich text={l.en} />
          </span>
          <span className="alphabet__example" role="cell">
            <Clip kind="w" clipKey={l.example} />
          </span>
        </div>
      ))}
    </div>
  );
}
