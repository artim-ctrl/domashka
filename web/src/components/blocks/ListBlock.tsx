import { List, ThemeIcon } from "@mantine/core";
import type { Text } from "../../i18n/languages";
import { Rich } from "../Rich";

export const ListBlock = ({ items, focus }: { items: Text[]; focus: boolean }) => (
    <List
        spacing="xs"
        mb="md"
        maw="72ch"
        center
        icon={focus ? <ThemeIcon size={10} radius="xl" color="red" /> : undefined}
    >
        {items.map((item, i) => (
            <List.Item key={i}>
                <Rich text={item} />
            </List.Item>
        ))}
    </List>
);
