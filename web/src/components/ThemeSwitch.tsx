import {
    type MantineColorScheme,
    SegmentedControl,
    Tooltip,
    useMantineColorScheme,
    VisuallyHidden,
} from "@mantine/core";
import { IconDeviceDesktop, IconMoon, IconSun, type TablerIcon } from "@tabler/icons-react";
import { useI18n } from "../i18n/context";
import type { Messages } from "../i18n/messages";

type Option = {
    value: MantineColorScheme;
    label: keyof Pick<Messages, "themeLight" | "themeSystem" | "themeDark">;
    Icon: TablerIcon;
};

const OPTIONS: readonly Option[] = [
    { value: "light", label: "themeLight", Icon: IconSun },
    { value: "auto", label: "themeSystem", Icon: IconDeviceDesktop },
    { value: "dark", label: "themeDark", Icon: IconMoon },
];

export const ThemeSwitch = () => {
    const { ui } = useI18n();
    const { colorScheme, setColorScheme } = useMantineColorScheme();

    return (
        <SegmentedControl
            aria-label={ui.theme}
            radius="xl"
            size="xs"
            value={colorScheme}
            onChange={(value) => setColorScheme(value as MantineColorScheme)}
            data={OPTIONS.map(({ value, label, Icon }) => ({
                value,
                label: (
                    <Tooltip label={ui[label]} withArrow>
                        <span>
                            <Icon size={16} stroke={1.6} aria-hidden />
                            <VisuallyHidden>{ui[label]}</VisuallyHidden>
                        </span>
                    </Tooltip>
                ),
            }))}
        />
    );
};
