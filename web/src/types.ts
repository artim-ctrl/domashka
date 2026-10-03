// Повторяет JSON из server/content.go. Text — строка или переводы {es, en, ru}.

import type { Text } from "./i18n";

export type Page = {
  title: Text;
  sections: Section[];
};

export type Section = {
  id: string;
  title?: Text;
  blocks: Block[];
};

export type Block =
  | { type: "paragraph"; text: Text }
  | { type: "callout"; text: Text }
  | { type: "list"; items: Text[]; style?: "focus" }
  | { type: "table"; columns: Text[]; rows: Text[][] }
  | { type: "alphabet"; letters: LetterRow[] };

export type LetterRow = {
  letter: string;
  sound?: string;
  ipa: Text;
  spanish: Text; // на что похоже в испанском
  english: Text; // на что похоже в английском
  example: string;
  focus?: boolean;
};

export type ClipKind = "s" | "w";

export type PageResponse = {
  page: Page;
  // «s:ш» / «w:дом» → адрес записи; чего нет — ещё не записано.
  recordings: Record<string, string>;
  total: number;
  canRecord: boolean;
};

export type ClipInfo = {
  kind: ClipKind;
  key: string;
  url?: string; // нет — ещё не записано
  ipa?: Text;
  usedIn: string[]; // id разделов страницы
};

export type ClipsResponse = {
  canRecord: boolean;
  sections: Record<string, Text>;
  clips: ClipInfo[];
};
