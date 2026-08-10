import { observer } from "mobx-react-lite";
import { type FC, useContext, useEffect, useRef } from "react";
import { Context } from "../main.tsx";
import type { IDecision } from "../types/types.tsx";

const WARNING_LIFETIME = 12000;
const VISIBLE_LIMIT = 3;

const toneOf = (scoreDelta: number) => {
    if (scoreDelta < 0) {
        return "danger";
    }

    if (scoreDelta > 0) {
        return "safe";
    }

    return "neutral";
};

const titleOf = (tone: string) => {
    if (tone === "danger") {
        return "Осторожно";
    }

    if (tone === "safe") {
        return "Верный ход";
    }

    return "Обратите внимание";
};

const iconOf = (tone: string) => {
    if (tone === "danger") {
        return "!";
    }

    if (tone === "safe") {
        return "✓";
    }

    return "i";
};

const ChatWarnings: FC = observer(() => {
    const {chat} = useContext(Context);

    const timersRef = useRef<Map<number, number>>(new Map());

    useEffect(() => {
        const timers = timersRef.current;

        chat.warnings.forEach(
            (warning: IDecision) => {
                if (timers.has(warning.id)) {
                    return;
                }

                const timer = window.setTimeout(
                    () => {
                        timers.delete(warning.id);

                        chat.dismissWarning(warning.id);
                    },
                    WARNING_LIFETIME
                );

                timers.set(warning.id, timer);
            }
        );
    }, [chat, chat.warnings]);

    useEffect(() => {
        const timers = timersRef.current;

        return () => {
            timers.forEach(timer => window.clearTimeout(timer));

            timers.clear();
        };
    }, []);

    const handleDismiss = (warningId: number) => {
        const timer = timersRef.current.get(warningId);

        if (timer !== undefined) {
            window.clearTimeout(timer);

            timersRef.current.delete(warningId);
        }

        chat.dismissWarning(warningId);
    };

    const visible = chat.warnings.slice(-VISIBLE_LIMIT);

    if (visible.length === 0) {
        return null;
    }

    return (
        <div className="chat-warnings">

            {visible.map(
                (warning: IDecision) => {
                    const tone = toneOf(warning.score_delta);

                    return (
                        <div className={`chat-warning ${tone}`} key={warning.id}>

                            <div className="chat-warning-icon">
                                {iconOf(tone)}
                            </div>

                            <div className="chat-warning-body">

                                <div className="chat-warning-title">
                                    {titleOf(tone)}

                                    {warning.score_delta !== 0 && (
                                        <span className="chat-warning-score">
                                            {warning.score_delta > 0 ? `+${warning.score_delta}` : warning.score_delta}
                                        </span>
                                    )}
                                </div>

                                {warning.feedback && (
                                    <div className="chat-warning-text">
                                        {warning.feedback}
                                    </div>
                                )}

                            </div>

                            <button
                                className="chat-warning-close"
                                onClick={() => handleDismiss(warning.id)}
                            >
                                ×
                            </button>

                        </div>
                    );
                }
            )}

        </div>
    );
});

export default ChatWarnings;
