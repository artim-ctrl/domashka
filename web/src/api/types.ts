import type { Text } from "../i18n/languages";
import type { ClipId } from "../lib/clip";

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
    spanish: Text;
    english: Text;
    example: string;
    focus?: boolean;
};

export type Recordings = Partial<Record<ClipId, string>>;

export type Choices = Partial<Record<ClipId, Choice>>;

export type PageResponse = {
    page: Page;
    recordings: Recordings;
    total: number;
    admin?: boolean;
    clips?: ClipId[];
    choices?: Choices;
};

export type Choice = { voice: string; text: string };

export type Voice = { id: string; label: string; license: string };

export type Candidate = { voice: string; url: string };

export type ClipState = { text: string; candidates: Candidate[]; chosen?: Choice };

export type ChooseResult = { url: string; chosen: Choice };
