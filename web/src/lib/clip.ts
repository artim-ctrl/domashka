export type ClipKind = "s" | "w";

export type ClipId = `${ClipKind}:${string}`;

export const clipId = (kind: ClipKind, key: string): ClipId => `${kind}:${key}`;

export const parseClipId = (id: ClipId): { kind: ClipKind; key: string } => ({
    kind: id.slice(0, 1) as ClipKind,
    key: id.slice(2),
});
