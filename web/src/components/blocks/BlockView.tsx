import { Alert, Text } from "@mantine/core";
import { IconInfoCircle } from "@tabler/icons-react";
import type { Block } from "../../api/types";
import { Rich } from "../Rich";
import { AlphabetBlock } from "./AlphabetBlock";
import { ListBlock } from "./ListBlock";
import { TableBlock } from "./TableBlock";

export const BlockView = ({ block }: { block: Block }) => {
    switch (block.type) {
        case "paragraph":
            return (
                <Text maw="72ch" lh={1.65} mb="md">
                    <Rich text={block.text} />
                </Text>
            );
        case "callout":
            return (
                <Alert variant="light" color="teal" radius="lg" icon={<IconInfoCircle />} mb="md" maw="72ch">
                    <Text size="sm" lh={1.6}>
                        <Rich text={block.text} />
                    </Text>
                </Alert>
            );
        case "list":
            return <ListBlock items={block.items} focus={block.style === "focus"} />;
        case "table":
            return <TableBlock columns={block.columns} rows={block.rows} />;
        case "alphabet":
            return <AlphabetBlock letters={block.letters} />;
    }
};
