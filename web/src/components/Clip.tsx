import { Indicator, Loader, UnstyledButton } from "@mantine/core";
import { IconVolume } from "@tabler/icons-react";
import type { ReactNode } from "react";
import { type Admin, useAdmin } from "../admin/context";
import { useI18n } from "../i18n/context";
import { type ClipId, type ClipKind, clipId } from "../lib/clip";
import { usePlayer } from "../player/context";
import classes from "./Clip.module.css";

type Props = {
    kind: ClipKind;
    clipKey: string;
    children?: ReactNode;
    variant?: "inline" | "tile";
};

type AdminState = "active" | "chosen" | "unchosen";

const adminState = (admin: Admin, id: ClipId): AdminState | undefined => {
    if (!admin.enabled) return undefined;
    if (admin.active === id) return "active";
    return admin.choices[id] ? "chosen" : "unchosen";
};

export const Clip = ({ kind, clipKey, children, variant = "inline" }: Props) => {
    const { urlOf, playing, play } = usePlayer();
    const { ui } = useI18n();
    const admin = useAdmin();
    const id = clipId(kind, clipKey);
    const available = urlOf(id) !== undefined;
    const isPlaying = playing === id;
    const inAdmin = adminState(admin, id);

    let title: string;
    if (!available) title = ui.pending(clipKey);
    else if (kind === "s") title = ui.listenSound(clipKey);
    else title = ui.listenWord(clipKey);

    let state: "playing" | "ready" | "missing";
    if (isPlaying) state = "playing";
    else if (available) state = "ready";
    else state = "missing";

    const onClick = () => {
        play(id);
        if (admin.enabled) admin.open(id);
    };

    return (
        <Indicator disabled={inAdmin !== "unchosen"} size={9} offset={3} withBorder processing={false}>
            <UnstyledButton
                className={classes.clip}
                data-variant={variant}
                data-kind={kind}
                data-state={state}
                data-admin={inAdmin}
                data-clip={id}
                disabled={!available && !admin.enabled}
                onClick={onClick}
                title={title}
                lang="ru"
            >
                <span className={classes.label}>{children ?? clipKey}</span>
                <span className={classes.icon} aria-hidden>
                    {isPlaying ? <Loader type="bars" size={12} color="currentColor" /> : <IconVolume size={14} />}
                </span>
            </UnstyledButton>
        </Indicator>
    );
};
