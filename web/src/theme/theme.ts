import {
    type CSSVariablesResolver,
    createTheme,
    localStorageColorSchemeManager,
    type MantineColorsTuple,
} from "@mantine/core";

const brick: MantineColorsTuple = [
    "#fdf1ee",
    "#f8e0d9",
    "#f0bfb1",
    "#e89a85",
    "#e07a60",
    "#d9654a",
    "#c4502f",
    "#b5412c",
    "#9a3523",
    "#80291a",
];

const teal: MantineColorsTuple = [
    "#e9f6f4",
    "#d0ebe7",
    "#a3d7cf",
    "#7cc8bd",
    "#5ab4a9",
    "#449c91",
    "#378379",
    "#2c6b66",
    "#235652",
    "#1a413e",
];

const serif = '"Iowan Old Style", "Palatino Linotype", Palatino, Georgia, serif';

export const theme = createTheme({
    primaryColor: "brick",
    primaryShade: { light: 7, dark: 4 },
    colors: { brick, teal },
    defaultRadius: "md",
    cursorType: "pointer",
    fontFamily: 'system-ui, -apple-system, "Segoe UI", Roboto, sans-serif',
    headings: { fontFamily: serif, fontWeight: "600" },
    other: { serif },
});

export const cssVariablesResolver: CSSVariablesResolver = () => ({
    variables: { "--app-serif": serif },
    light: { "--mantine-color-body": "#f8f5ef", "--app-surface": "#fffdf9" },
    dark: { "--mantine-color-body": "#141519", "--app-surface": "#1c1e23" },
});

export const THEME_STORAGE_KEY = "domashka.theme";

export const colorSchemeManager = localStorageColorSchemeManager({ key: THEME_STORAGE_KEY });
