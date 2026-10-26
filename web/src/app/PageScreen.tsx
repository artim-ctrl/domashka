import { AppShell, Container, Text, Title } from "@mantine/core";
import { useEffect, useMemo, useState } from "react";
import { AdminAside } from "../admin/AdminAside";
import { AdminDisabledNotice } from "../admin/AdminDisabledNotice";
import { AdminStatus } from "../admin/AdminStatus";
import { type Admin, AdminContext, adminRequested } from "../admin/context";
import type { Choice, PageResponse } from "../api/types";
import { BlockView } from "../components/blocks/BlockView";
import { useI18n } from "../i18n/context";
import type { ClipId } from "../lib/clip";
import { PlayerProvider } from "../player/PlayerProvider";
import classes from "./PageScreen.module.css";
import { TopBar } from "./TopBar";

type Props = {
    data: PageResponse;
    onChosen(id: ClipId, url: string, choice: Choice): void;
};

const HEADER_HEIGHT = 56;
const ASIDE_WIDTH = 420;

export const PageScreen = ({ data, onChosen }: Props) => {
    const { ui, t } = useI18n();
    const [active, setActive] = useState<ClipId | null>(null);
    const [requested] = useState(adminRequested);
    const { page, recordings, total } = data;
    const recorded = Object.keys(recordings).length;
    const clips = data.clips ?? [];

    useEffect(() => {
        if (location.hash) document.getElementById(decodeURIComponent(location.hash.slice(1)))?.scrollIntoView();
    }, []);

    useEffect(() => {
        document.title = t(page.title);
    }, [page.title, t]);

    const admin = useMemo<Admin>(
        () => ({
            enabled: Boolean(data.admin) && requested,
            active,
            open: setActive,
            choices: data.choices ?? {},
        }),
        [data.admin, data.choices, requested, active],
    );
    const asideOpen = admin.enabled && active !== null;

    return (
        <AdminContext.Provider value={admin}>
            <PlayerProvider recordings={recordings}>
                <AppShell
                    header={{ height: HEADER_HEIGHT }}
                    aside={{
                        width: ASIDE_WIDTH,
                        breakpoint: "md",
                        collapsed: { desktop: !asideOpen, mobile: !asideOpen },
                    }}
                    padding={0}
                >
                    <AppShell.Header className={classes.header}>
                        <TopBar sections={page.sections} />
                    </AppShell.Header>

                    <AppShell.Main>
                        <Container size={1040} px="md" pb="xl">
                            <Title order={1} className={classes.title}>
                                {t(page.title)}
                            </Title>
                            {page.sections.map((s) => (
                                <section key={s.id} id={s.id} className={classes.section}>
                                    {s.title && (
                                        <Title order={2} mb="md">
                                            {t(s.title)}
                                        </Title>
                                    )}
                                    {s.blocks.map((b, i) => (
                                        <BlockView key={i} block={b} />
                                    ))}
                                </section>
                            ))}
                            <Text component="footer" ta="center" c="dimmed" size="sm" py="xl">
                                {ui.recordings(recorded, total)}
                            </Text>
                        </Container>
                    </AppShell.Main>

                    {admin.enabled && (
                        <AppShell.Aside>
                            <AdminAside clips={clips} onChosen={onChosen} onClose={() => setActive(null)} />
                        </AppShell.Aside>
                    )}
                </AppShell>

                {admin.enabled && <AdminStatus clips={clips} />}
                {import.meta.env.DEV && requested && !data.admin && <AdminDisabledNotice />}
            </PlayerProvider>
        </AdminContext.Provider>
    );
};
