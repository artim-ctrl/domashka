import type { ClipKind } from "./clip";

export type MarkupNode =
    | { type: "text"; text: string }
    | { type: "clip"; kind: ClipKind; key: string; label: string }
    | { type: "bold"; text: string }
    | { type: "italic"; text: string };

const MARKUP = /\{\{([sw]):([^}|]+)(?:\|([^}]*))?\}\}|\*\*([^*]+)\*\*|\*([^*]+)\*/g;

export const parseMarkup = (source: string): MarkupNode[] => {
    const nodes: MarkupNode[] = [];
    let last = 0;
    for (const match of source.matchAll(MARKUP)) {
        const at = match.index;
        if (at > last) nodes.push({ type: "text", text: source.slice(last, at) });
        const [whole, kind, key, label, bold, italic] = match;
        if (kind) nodes.push({ type: "clip", kind: kind as ClipKind, key, label: label || key });
        else if (bold) nodes.push({ type: "bold", text: bold });
        else nodes.push({ type: "italic", text: italic });
        last = at + whole.length;
    }
    if (last < source.length) nodes.push({ type: "text", text: source.slice(last) });
    return nodes;
};
