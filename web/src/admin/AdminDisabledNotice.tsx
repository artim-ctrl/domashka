import { Affix, Alert } from "@mantine/core";
import { IconLock } from "@tabler/icons-react";
import { useI18n } from "../i18n/context";

export const AdminDisabledNotice = () => {
    const { ui } = useI18n();
    return (
        <Affix position={{ bottom: 16, left: 16 }}>
            <Alert variant="light" color="yellow" radius="lg" icon={<IconLock />} maw={420}>
                {ui.admin.disabledOnServer}
            </Alert>
        </Affix>
    );
};
