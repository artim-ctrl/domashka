import { useState } from "react";
import { PageScreen } from "./PageScreen";
import { ErrorScreen, LoadingScreen } from "./StatusScreen";
import { usePageData } from "./usePageData";

const PageLoader = ({ onRetry }: { onRetry(): void }) => {
    const { state, applyChoice } = usePageData();

    switch (state.status) {
        case "loading":
            return <LoadingScreen />;
        case "error":
            return <ErrorScreen error={state.error} onRetry={onRetry} />;
        case "ready":
            return <PageScreen data={state.data} onChosen={applyChoice} />;
    }
};

export const App = () => {
    const [attempt, setAttempt] = useState(0);
    return <PageLoader key={attempt} onRetry={() => setAttempt((n) => n + 1)} />;
};
