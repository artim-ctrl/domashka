import type { ReactNode } from "react";
import { useI18n, type Text } from "../i18n";
import type { ClipKind } from "../types";
import { Clip } from "./Clip";

// Разметка из server/content/page.json: **жирный**, *курсив*,
// {{s:ш}} — звук, {{w:дом}} — слово, {{w:молоко|молоко́}} — ключ записи|подпись.
const MARKUP = /\{\{([sw]):([^}|]+)(?:\|([^}]*))?\}\}|\*\*([^*]+)\*\*|\*([^*]+)\*/g;

export function Rich({ text: source }: { text: Text }) {
  const text = useI18n().t(source);
  const out: ReactNode[] = [];
  let last = 0;
  for (const m of text.matchAll(MARKUP)) {
    const at = m.index ?? 0;
    if (at > last) out.push(text.slice(last, at));
    const [, kind, key, label, bold, italic] = m;
    if (kind) {
      out.push(
        <Clip key={at} kind={kind as ClipKind} clipKey={key}>
          {label || key}
        </Clip>,
      );
    } else if (bold) {
      out.push(<strong key={at}>{bold}</strong>);
    } else {
      out.push(<em key={at}>{italic}</em>);
    }
    last = at + m[0].length;
  }
  if (last < text.length) out.push(text.slice(last));
  return <>{out}</>;
}
