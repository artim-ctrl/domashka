import { useMemo } from "react";
import { useI18n } from "../i18n/context";
import type { Text } from "../i18n/languages";
import { parseMarkup } from "../lib/markup";
import { Clip } from "./Clip";

export const Rich = ({ text }: { text: Text }) => {
    const { t } = useI18n();
    const source = t(text);
    const nodes = useMemo(() => parseMarkup(source), [source]);

    return nodes.map((node, i) => {
        switch (node.type) {
            case "text":
                return node.text;
            case "clip":
                return (
                    <Clip key={i} kind={node.kind} clipKey={node.key}>
                        {node.label}
                    </Clip>
                );
            case "bold":
                return <strong key={i}>{node.text}</strong>;
            default:
                return <em key={i}>{node.text}</em>;
        }
    });
};
