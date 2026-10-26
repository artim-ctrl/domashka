import { Alert, Button, Center, Code, Loader, Stack, Text } from "@mantine/core";
import { IconAlertTriangle, IconRefresh } from "@tabler/icons-react";
import { useI18n } from "../i18n/context";

export const LoadingScreen = () => {
    const { ui } = useI18n();
    return (
        <Center mih="100vh" aria-busy>
            <Stack align="center" gap="sm">
                <Loader type="dots" />
                <Text c="dimmed">{ui.loading}</Text>
            </Stack>
        </Center>
    );
};

export const ErrorScreen = ({ error, onRetry }: { error: string; onRetry(): void }) => {
    const { ui } = useI18n();
    return (
        <Center mih="100vh" p="md">
            <Alert
                variant="light"
                color="red"
                radius="lg"
                icon={<IconAlertTriangle />}
                title={ui.loadError}
                maw={480}
                w="100%"
                role="alert"
            >
                <Stack gap="md" align="flex-start">
                    <Code block w="100%">
                        {error}
                    </Code>
                    <Button variant="light" color="red" leftSection={<IconRefresh size={16} />} onClick={onRetry}>
                        {ui.retry}
                    </Button>
                </Stack>
            </Alert>
        </Center>
    );
};
