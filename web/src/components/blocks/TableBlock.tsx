import { Paper, Table } from "@mantine/core";
import type { Text } from "../../i18n/languages";
import { Rich } from "../Rich";
import classes from "./blocks.module.css";

export const TableBlock = ({ columns, rows }: { columns: Text[]; rows: Text[][] }) => (
    <Paper withBorder radius="lg" mb="md" className={classes.surface}>
        <Table.ScrollContainer minWidth={480}>
            <Table verticalSpacing="sm" horizontalSpacing="md" highlightOnHover>
                <Table.Thead className={classes.head}>
                    <Table.Tr>
                        {columns.map((column, i) => (
                            <Table.Th key={i}>
                                <Rich text={column} />
                            </Table.Th>
                        ))}
                    </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                    {rows.map((row, i) => (
                        <Table.Tr key={i}>
                            {row.map((cell, j) => (
                                <Table.Td key={j}>
                                    <Rich text={cell} />
                                </Table.Td>
                            ))}
                        </Table.Tr>
                    ))}
                </Table.Tbody>
            </Table>
        </Table.ScrollContainer>
    </Paper>
);
