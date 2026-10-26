import { ActionIcon, Badge, Button, Group, Loader, Paper, Stack, Text, Tooltip } from "@mantine/core";
import { IconCheck, IconPlayerPlayFilled, IconPlayerStopFilled } from "@tabler/icons-react";
import type { Candidate, Choice, Voice } from "../api/types";
import { useI18n } from "../i18n/context";
import { usePlayer } from "../player/context";
import classes from "./CandidateList.module.css";

type Props = {
    voices: Voice[];
    candidates: Candidate[];
    chosen: Choice | undefined;
    text: string | undefined;
    busy: Partial<Record<string, boolean>>;
    errors: Partial<Record<string, string>>;
    choosing: string | null;
    playbackId(voice: string): string;
    onPlay(candidate: Candidate): void;
    onChoose(voice: string): void;
};

const PlayIcon = ({ busy, playing }: { busy: boolean; playing: boolean }) => {
    if (busy) return <Loader size={14} />;
    if (playing) return <IconPlayerStopFilled size={14} />;
    return <IconPlayerPlayFilled size={14} />;
};

export const CandidateList = ({
    voices,
    candidates,
    chosen,
    text,
    busy,
    errors,
    choosing,
    playbackId,
    onPlay,
    onChoose,
}: Props) => {
    const { ui } = useI18n();
    const { playing } = usePlayer();

    return (
        <Paper withBorder radius="lg" className={classes.list} component="ol">
            {voices.map((v, i) => {
                const candidate = candidates.find((c) => c.voice === v.id);
                const isBusy = Boolean(busy[v.id]);
                const isChosen = chosen?.voice === v.id && chosen.text === text;
                const isPlaying = playing === playbackId(v.id);
                const unavailable = !candidate || isBusy;
                const error = errors[v.id];

                return (
                    <li key={v.id} className={classes.row} data-chosen={isChosen || undefined}>
                        <Group gap="sm" wrap="nowrap" align="center">
                            <Tooltip label={candidate ? ui.admin.listen((i + 1) % 10) : ui.admin.noCandidate}>
                                <ActionIcon
                                    variant={isPlaying ? "filled" : "light"}
                                    radius="xl"
                                    size="lg"
                                    disabled={unavailable}
                                    onClick={() => candidate && onPlay(candidate)}
                                    aria-label={ui.admin.listen((i + 1) % 10)}
                                >
                                    <PlayIcon busy={isBusy} playing={isPlaying} />
                                </ActionIcon>
                            </Tooltip>
                            <Stack gap={2} flex={1} miw={0}>
                                <Group gap={6} wrap="nowrap">
                                    <Text size="sm" fw={500} truncate>
                                        {v.label}
                                    </Text>
                                    <Badge size="xs" variant="outline" color="gray" radius="sm">
                                        {v.license}
                                    </Badge>
                                </Group>
                                {error && (
                                    <Text size="xs" c="red">
                                        {error}
                                    </Text>
                                )}
                            </Stack>
                            {isChosen ? (
                                <Badge variant="filled" radius="xl" leftSection={<IconCheck size={12} />}>
                                    {ui.admin.chosen}
                                </Badge>
                            ) : (
                                <Button
                                    size="compact-sm"
                                    radius="xl"
                                    variant="default"
                                    loading={choosing === v.id}
                                    disabled={unavailable || (choosing !== null && choosing !== v.id)}
                                    title={candidate ? undefined : ui.admin.noCandidate}
                                    onClick={() => onChoose(v.id)}
                                >
                                    {ui.admin.choose}
                                </Button>
                            )}
                        </Group>
                    </li>
                );
            })}
        </Paper>
    );
};
