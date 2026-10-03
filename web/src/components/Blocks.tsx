import { useI18n } from "../i18n";
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
                {block.columns.map((c, i) => (
                  <th key={i}>
                    <Rich text={c} />
                  </th>
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
  const { ui } = useI18n();
  return (
    <div className="alphabet" role="table" aria-label={ui.alphabet}>
      <div className="alphabet__row alphabet__head" role="row">
        <span role="columnheader">{ui.colLetter}</span>
        <span role="columnheader">{ui.colSound}</span>
        <span role="columnheader">{ui.colSpanish}</span>
        <span role="columnheader">{ui.colEnglish}</span>
        <span role="columnheader">{ui.colExample}</span>
      </div>
      {letters.map((l) => (
        <div className="alphabet__row" role="row" key={l.letter}>
          <span className="alphabet__letter" role="cell">
            {l.sound ? (
              <Clip kind="s" clipKey={l.sound} variant="tile">
                {l.letter}
              </Clip>
            ) : (
              <span className="tile tile--silent" lang="ru" title={ui.silent}>
                {l.letter}
              </span>
            )}
            {l.focus && <span className="focus-dot" title={ui.focus} />}
          </span>
          <span className="alphabet__ipa" role="cell">
            <Rich text={l.ipa} />
          </span>
          <span className="alphabet__text" role="cell" data-label={ui.colSpanish}>
            <Rich text={l.spanish} />
          </span>
          <span className="alphabet__text" role="cell" data-label={ui.colEnglish}>
            <Rich text={l.english} />
          </span>
          <span className="alphabet__example" role="cell">
            <Clip kind="w" clipKey={l.example} />
          </span>
        </div>
      ))}
    </div>
  );
}
