import { SegmentedControl } from "@mantine/core";
import { useI18n } from "../i18n/context";
import { isLang, LANGS } from "../i18n/languages";

export const LangSwitch = () => {
    const { lang, setLang, ui } = useI18n();

    return (
        <SegmentedControl
            aria-label={ui.language}
            radius="xl"
            size="xs"
            fw={650}
            value={lang}
            onChange={(value) => isLang(value) && setLang(value)}
            data={LANGS.map((l) => ({
                value: l.value,
                label: (
                    <span lang={l.value} title={l.name}>
                        {l.short}
                    </span>
                ),
            }))}
        />
    );
};
