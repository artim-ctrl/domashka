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
  // Только если сервер запущен с -admin.
  admin?: boolean;
  clips?: string[]; // id записей в порядке страницы
  choices?: Record<string, Choice>;
};

// Чем озвучена запись: голос и текст для синтезатора («молок+о»).
export type Choice = { voice: string; text: string };

export type Voice = { id: string; label: string; license: string };

export type Candidate = { voice: string; url: string };

export type ClipState = { text: string; candidates: Candidate[]; chosen?: Choice };
