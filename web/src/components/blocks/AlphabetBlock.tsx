import { Indicator, Paper, Table, Text, useMantineTheme } from "@mantine/core";
import { useMediaQuery } from "@mantine/hooks";
import type { LetterRow } from "../../api/types";
import { useI18n } from "../../i18n/context";
import { Clip } from "../Clip";
import { Rich } from "../Rich";
import classes from "./blocks.module.css";

const LetterTile = ({ row }: { row: LetterRow }) => {
    const { ui } = useI18n();
    const tile = row.sound ? (
        <Clip kind="s" clipKey={row.sound} variant="tile">
            {row.letter}
        </Clip>
    ) : (
        <span className={classes.silent} lang="ru" title={ui.silent}>
            {row.letter}
        </span>
    );

    return (
        <Indicator disabled={!row.focus} color="red" size={12} offset={4} withBorder title={ui.focus}>
            {tile}
        </Indicator>
    );
};

const AlphabetTable = ({ letters }: { letters: LetterRow[] }) => {
    const { ui } = useI18n();
    return (
        <Table verticalSpacing="sm" horizontalSpacing="md" highlightOnHover aria-label={ui.alphabet}>
            <Table.Thead className={classes.head}>
                <Table.Tr>
                    <Table.Th>{ui.colLetter}</Table.Th>
                    <Table.Th>{ui.colSound}</Table.Th>
                    <Table.Th>{ui.colSpanish}</Table.Th>
                    <Table.Th>{ui.colEnglish}</Table.Th>
                    <Table.Th>{ui.colExample}</Table.Th>
                </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
                {letters.map((row) => (
                    <Table.Tr key={row.letter}>
                        <Table.Td className={classes.letter}>
                            <LetterTile row={row} />
                        </Table.Td>
                        <Table.Td className={classes.ipa}>
                            <Rich text={row.ipa} />
                        </Table.Td>
                        <Table.Td fz="sm">
                            <Rich text={row.spanish} />
                        </Table.Td>
                        <Table.Td fz="sm">
                            <Rich text={row.english} />
                        </Table.Td>
                        <Table.Td>
                            <Clip kind="w" clipKey={row.example} />
                        </Table.Td>
                    </Table.Tr>
                ))}
            </Table.Tbody>
        </Table>
    );
};

const AlphabetCards = ({ letters }: { letters: LetterRow[] }) => {
    const { ui } = useI18n();
    return (
        <ul className={classes.cards} aria-label={ui.alphabet}>
            {letters.map((row) => (
                <li key={row.letter} className={classes.card}>
                    <div className={classes.cardLetter}>
                        <LetterTile row={row} />
                    </div>
                    <div className={classes.ipa}>
                        <Rich text={row.ipa} />
                    </div>
                    <div className={classes.cardExample}>
                        <Clip kind="w" clipKey={row.example} />
                    </div>
                    <Text size="sm" className={classes.cardSpanish}>
                        <Text span size="xs" fw={600} c="dimmed" tt="uppercase">
                            {ui.colSpanish} ·{" "}
                        </Text>
                        <Rich text={row.spanish} />
                    </Text>
                    <Text size="sm" className={classes.cardEnglish}>
                        <Text span size="xs" fw={600} c="dimmed" tt="uppercase">
                            {ui.colEnglish} ·{" "}
                        </Text>
                        <Rich text={row.english} />
                    </Text>
                </li>
            ))}
        </ul>
    );
};

export const AlphabetBlock = ({ letters }: { letters: LetterRow[] }) => {
    const theme = useMantineTheme();
    const compact = useMediaQuery(`(max-width: ${theme.breakpoints.sm})`, false, { getInitialValueInEffect: false });

    return (
        <Paper withBorder radius="lg" mb="md" className={classes.surface}>
            {compact ? <AlphabetCards letters={letters} /> : <AlphabetTable letters={letters} />}
        </Paper>
    );
};
