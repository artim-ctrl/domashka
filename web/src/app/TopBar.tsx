import { Button, Group, ScrollArea } from "@mantine/core";
import type { Section } from "../api/types";
import { LangSwitch } from "../components/LangSwitch";
import { ThemeSwitch } from "../components/ThemeSwitch";
import { useI18n } from "../i18n/context";
import classes from "./TopBar.module.css";

export const TopBar = ({ sections }: { sections: Section[] }) => {
    const { ui, t } = useI18n();

    return (
        <Group h="100%" px="md" gap="md" wrap="nowrap" maw={1100} mx="auto">
            <ScrollArea type="never" className={classes.nav} component="nav" aria-label={ui.sections}>
                <Group gap={4} wrap="nowrap">
                    {sections.map(
                        (s) =>
                            s.title && (
                                <Button
                                    key={s.id}
                                    className={classes.link}
                                    component="a"
                                    href={`#${s.id}`}
                                    variant="subtle"
                                    color="gray"
                                    radius="xl"
                                    size="compact-sm"
                                    fw={500}
                                >
                                    {t(s.title)}
                                </Button>
                            ),
                    )}
                </Group>
            </ScrollArea>
            <LangSwitch />
            <ThemeSwitch />
        </Group>
    );
};
