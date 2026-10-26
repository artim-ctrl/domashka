import "@mantine/core/styles.css";
import "./global.css";
import { MantineProvider } from "@mantine/core";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./app/App";
import { I18nProvider } from "./i18n/I18nProvider";
import { colorSchemeManager, cssVariablesResolver, theme } from "./theme/theme";

const root = document.getElementById("root");
if (!root) throw new Error("#root element is missing in index.html");

createRoot(root).render(
    <StrictMode>
        <MantineProvider
            theme={theme}
            colorSchemeManager={colorSchemeManager}
            defaultColorScheme="auto"
            cssVariablesResolver={cssVariablesResolver}
        >
            <I18nProvider>
                <App />
            </I18nProvider>
        </MantineProvider>
    </StrictMode>,
);
