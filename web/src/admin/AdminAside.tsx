import { useState } from "react";
import type { Choice } from "../api/types";
import type { ClipId } from "../lib/clip";
import { ClipDrawer } from "./ClipDrawer";
import { useAdmin } from "./context";

type Props = {
    clips: ClipId[];
    onChosen(id: ClipId, url: string, choice: Choice): void;
    onClose(): void;
};

export const AdminAside = ({ clips, onChosen, onClose }: Props) => {
    const { active } = useAdmin();
    const [attempt, setAttempt] = useState(0);

    if (!active) return null;
    return (
        <ClipDrawer
            key={`${active}#${attempt}`}
            id={active}
            clips={clips}
            onChosen={onChosen}
            onClose={onClose}
            onRetry={() => setAttempt((n) => n + 1)}
        />
    );
};
