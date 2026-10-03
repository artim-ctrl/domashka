// Повторяет JSON из server/content.go.

export type Page = {
  title: string;
  sections: Section[];
};

export type Section = {
  id: string;
  title?: string;
  blocks: Block[];
};

export type Block =
  | { type: "paragraph"; text: string }
  | { type: "callout"; text: string }
  | { type: "list"; items: string[]; style?: "focus" }
  | { type: "table"; columns: string[]; rows: string[][] }
  | { type: "alphabet"; letters: LetterRow[] };

export type LetterRow = {
  letter: string;
  sound?: string;
  ipa: string;
  es: string;
  en: string;
  example: string;
  focus?: boolean;
};

export type ClipKind = "s" | "w";

export type PageResponse = {
  page: Page;
  // «s:ш» / «w:дом» → адрес записи; чего нет — ещё не записано.
  recordings: Record<string, string>;
  total: number;
};
