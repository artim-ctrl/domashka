import { Affix, Badge, Button, Group, Paper, Text } from "@mantine/core";
import { IconArrowRight } from "@tabler/icons-react";
import { useI18n } from "../i18n/context";
import type { ClipId } from "../lib/clip";
import { useAdmin } from "./context";

export const AdminStatus = ({ clips }: { clips: ClipId[] }) => {
    const { ui } = useI18n();
    const { active, open, choices } = useAdmin();
    const done = clips.filter((id) => choices[id]).length;

    const openNextUnchosen = () => {
        const from = active ? clips.indexOf(active) + 1 : 0;
        const next = [...clips.slice(from), ...clips.slice(0, from)].find((id) => !choices[id]);
        if (next) open(next);
    };

    return (
        <Affix position={{ bottom: 16, left: 16 }}>
            <Paper shadow="md" radius="xl" withBorder py={6} pl="md" pr={done < clips.length ? 6 : "md"}>
                <Group gap="sm" wrap="nowrap">
                    <Badge variant="light" radius="sm">
                        {ui.admin.title}
                    </Badge>
                    <Text size="sm">{ui.admin.chosenCount(done, clips.length)}</Text>
                    {done < clips.length && (
                        <Button
                            size="compact-sm"
                            radius="xl"
                            variant="light"
                            rightSection={<IconArrowRight size={14} />}
                            onClick={openNextUnchosen}
                        >
                            {ui.admin.nextUnchosen}
                        </Button>
                    )}
                </Group>
            </Paper>
        </Affix>
    );
};
