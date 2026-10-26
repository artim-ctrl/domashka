import {
    ActionIcon,
    Alert,
    AppShell,
    Badge,
    Button,
    Group,
    ScrollArea,
    Stack,
    Text,
    TextInput,
    Title,
    Tooltip,
} from "@mantine/core";
import { IconArrowLeft, IconArrowRight, IconPlayerPlay, IconRefresh, IconSparkles, IconX } from "@tabler/icons-react";
import { useEffect, useState } from "react";
import type { Candidate, Choice } from "../api/types";
import { useI18n } from "../i18n/context";
import { type ClipId, parseClipId } from "../lib/clip";
import { usePlayer } from "../player/context";
import { CandidateList } from "./CandidateList";
import classes from "./ClipDrawer.module.css";
import { useAdmin } from "./context";
import { useClipCandidates } from "./useClipCandidates";
import { useDrawerHotkeys } from "./useDrawerHotkeys";
import { useVoices } from "./useVoices";

type Props = {
    id: ClipId;
    clips: ClipId[];
    onChosen(id: ClipId, url: string, choice: Choice): void;
    onClose(): void;
    onRetry(): void;
};

const candidatePlaybackId = (id: ClipId, voice: string) => `cand:${id}:${voice}`;

export const ClipDrawer = ({ id, clips, onChosen, onClose, onRetry }: Props) => {
    const { ui } = useI18n();
    const { open, choices } = useAdmin();
    const { urlOf, playUrl } = usePlayer();
    const { voices, error: voicesError } = useVoices();
    const { clip, loadError, busy, errors, generating, choosing, generate, choose } = useClipCandidates(
        id,
        voices,
        onChosen,
    );
    const [draft, setDraft] = useState<string | null>(null);

    const { kind, key } = parseClipId(id);
    const index = clips.indexOf(id);
    const previous = index > 0 ? clips[index - 1] : undefined;
    const next = index < clips.length - 1 ? clips[index + 1] : undefined;
    const current = urlOf(id);
    const chosen = choices[id] ?? clip?.chosen;
    const candidates = clip?.candidates ?? [];
    const text = draft ?? clip?.text ?? "";
    const error = voicesError ?? loadError;
    const voiceLabel = (voice: string) => voices?.find((v) => v.id === voice)?.label ?? voice;

    const playCandidate = (c: Candidate) => playUrl(candidatePlaybackId(id, c.voice), c.url);

    useEffect(() => {
        document
            .querySelector(`[data-clip="${CSS.escape(id)}"]`)
            ?.scrollIntoView({ block: "center", behavior: "smooth" });
    }, [id]);

    useDrawerHotkeys({
        onDigit: (i) => {
            const voice = voices?.[i];
            const c = voice && candidates.find((x) => x.voice === voice.id);
            if (c) playCandidate(c);
        },
        onPrevious: () => previous && open(previous),
        onNext: () => next && open(next),
        onClose,
    });

    let nowLabel = ui.admin.noRecording;
    if (chosen) nowLabel = `${voiceLabel(chosen.voice)} · «${chosen.text}»`;
    else if (current) nowLabel = ui.admin.noVoice;

    let generateLabel = ui.admin.generate;
    if (candidates.length) generateLabel = ui.admin.regenerate;

    let generateBlocked: string | undefined;
    if (generating) generateBlocked = ui.admin.generateDisabledBusy;
    else if (!text.trim()) generateBlocked = ui.admin.generateDisabledEmpty;

    return (
        <>
            <AppShell.Section p="md" pb="sm" aria-label={ui.admin.drawer}>
                <Group justify="space-between" wrap="nowrap">
                    <Group gap="sm" align="baseline" wrap="nowrap">
                        <Badge variant="light" color="gray" radius="sm">
                            {kind === "s" ? ui.admin.kindSound : ui.admin.kindWord}
                        </Badge>
                        <Title order={3} lang="ru" className={classes.key}>
                            {key}
                        </Title>
                    </Group>
                    <Group gap={6} wrap="nowrap">
                        <Tooltip label={ui.admin.previous}>
                            <ActionIcon
                                variant="default"
                                radius="xl"
                                disabled={!previous}
                                onClick={() => previous && open(previous)}
                                aria-label={ui.admin.previous}
                            >
                                <IconArrowLeft size={16} />
                            </ActionIcon>
                        </Tooltip>
                        <Text size="xs" c="dimmed" ff="monospace">
                            {index + 1}/{clips.length}
                        </Text>
                        <Tooltip label={ui.admin.next}>
                            <ActionIcon
                                variant="default"
                                radius="xl"
                                disabled={!next}
                                onClick={() => next && open(next)}
                                aria-label={ui.admin.next}
                            >
                                <IconArrowRight size={16} />
                            </ActionIcon>
                        </Tooltip>
                        <Tooltip label={ui.admin.close}>
                            <ActionIcon
                                variant="subtle"
                                color="gray"
                                radius="xl"
                                onClick={onClose}
                                aria-label={ui.admin.close}
                            >
                                <IconX size={16} />
                            </ActionIcon>
                        </Tooltip>
                    </Group>
                </Group>
            </AppShell.Section>

            <AppShell.Section grow component={ScrollArea} px="md" aria-busy={generating}>
                <Stack gap="md" pb="md">
                    <Group gap="sm" wrap="nowrap">
                        <Button
                            variant="light"
                            radius="xl"
                            size="compact-sm"
                            leftSection={<IconPlayerPlay size={14} />}
                            disabled={!current}
                            onClick={() => current && playUrl(id, current)}
                        >
                            {ui.admin.playCurrent}
                        </Button>
                        <Text size="sm" c="dimmed" truncate>
                            {nowLabel}
                        </Text>
                    </Group>

                    <form
                        onSubmit={(e) => {
                            e.preventDefault();
                            if (!generateBlocked) generate(text.trim());
                        }}
                    >
                        <Stack gap="xs">
                            <TextInput
                                label={ui.admin.textLabel}
                                description={ui.admin.textHint}
                                lang="ru"
                                size="md"
                                value={text}
                                onChange={(e) => setDraft(e.currentTarget.value)}
                                spellCheck={false}
                                classNames={{ input: classes.serifInput }}
                            />
                            <Group>
                                <Tooltip label={generateBlocked} disabled={!generateBlocked || generating}>
                                    <Button
                                        type="submit"
                                        radius="xl"
                                        leftSection={<IconSparkles size={16} />}
                                        loading={generating}
                                        disabled={Boolean(generateBlocked) || !clip}
                                    >
                                        {generateLabel}
                                    </Button>
                                </Tooltip>
                            </Group>
                        </Stack>
                    </form>

                    {error && (
                        <Alert variant="light" color="red" radius="md" role="alert">
                            <Stack gap="xs" align="flex-start">
                                <Text size="sm">{error}</Text>
                                <Button
                                    size="compact-sm"
                                    variant="light"
                                    color="red"
                                    leftSection={<IconRefresh size={14} />}
                                    onClick={onRetry}
                                >
                                    {ui.retry}
                                </Button>
                            </Stack>
                        </Alert>
                    )}

                    {voices && (
                        <CandidateList
                            voices={voices}
                            candidates={candidates}
                            chosen={chosen}
                            text={clip?.text}
                            busy={busy}
                            errors={errors}
                            choosing={choosing}
                            playbackId={(voice) => candidatePlaybackId(id, voice)}
                            onPlay={playCandidate}
                            onChoose={(voice) => void choose(voice)}
                        />
                    )}
                </Stack>
            </AppShell.Section>

            <AppShell.Section p="md" pt="xs">
                <Text size="xs" c="dimmed">
                    {ui.admin.hotkeys}
                </Text>
            </AppShell.Section>
        </>
    );
};
