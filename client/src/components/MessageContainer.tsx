import { observer } from "mobx-react-lite";
import type { IMessage } from "../types/types.tsx";
import {type FC, useContext, useEffect, useLayoutEffect, useRef} from "react";
import { Context } from "../main.tsx";
import { useParams } from "react-router-dom";
import { API_URL, refreshAccessToken } from "../http";

const SSE_EVENT_TYPES = ["message", "decision", "chat"];
const SSE_REOPEN_DELAY = 3000;
const SSE_MAX_REOPEN_ATTEMPTS = 5;

const MessageContainer: FC = observer(() => {
    const {messages, chat, user} = useContext(Context);

    const { id } = useParams<{ id: string }>();

    const messagesEndRef = useRef<HTMLDivElement | null>(null);
    const containerRef = useRef<HTMLDivElement | null>(null);

    const offsetFromBottomRef = useRef<number | null>(null);
    const lastScrollTopRef = useRef(0);

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({behavior: "smooth"});
    };

    const loadOlderThreshold = 120;

    const handleScroll = () => {
        const container = containerRef.current;

        if (!container || !id) {
            return;
        }

        const chatId = Number(id);

        if (!Number.isFinite(chatId)) {
            return;
        }

        const previousScrollTop = lastScrollTopRef.current;

        lastScrollTopRef.current = container.scrollTop;

        if (container.scrollTop > previousScrollTop) {
            return;
        }

        if (container.scrollTop > loadOlderThreshold || !messages.hasMore || messages.isLoadingMore) {
            return;
        }

        offsetFromBottomRef.current = container.scrollHeight - container.scrollTop;

        messages.loadMoreMessages(chatId);
    };

    useEffect(() => {
        lastScrollTopRef.current = 0;
        offsetFromBottomRef.current = null;
    }, [id]);

    useEffect(() => {
        if (!id) {return;}

        const chatId = Number(id);

        if (!Number.isFinite(chatId)) {
            return;
        }

        let source: EventSource | null = null;
        let reopenTimer: number | undefined;
        let reopenAttempts = 0;
        let stopped = false;

        const stop = () => {
            stopped = true;

            window.clearTimeout(reopenTimer);

            source?.close();
        };

        const handleEvent = (event: MessageEvent) => {
            try {
                const data = JSON.parse(event.data);

                switch (data.type) {

                    case "message": {
                        if (data.message) {
                            messages.addMessage(data.message);
                        }

                        break;
                    }

                    case "decision": {
                        if (data.decision) {
                            chat.handleSSEDecision(data.decision);
                        }
                        break;
                    }

                    case "chat": {
                        if (!data.chat) {
                            break;
                        }

                        chat.handleSSEChat(data.chat);

                        const status = data.chat.status;

                        if (status === "finished" || status === "abandoned") {
                            chat.getDecision(chatId);

                            user.refreshAfterChat();

                            stop();
                        }

                        break;
                    }
                }
            } catch (error) {
                console.error("Ошибка обработки SSE:", error);
            }
        };

        const open = () => {
            if (stopped) {
                return;
            }

            const current = new EventSource(`${API_URL}/chats/${chatId}/events`, {withCredentials: true});

            source = current;

            current.onopen = () => {
                reopenAttempts = 0;
            };

            SSE_EVENT_TYPES.forEach(type => current.addEventListener(type, handleEvent));

            current.onerror = () => {
                if (current.readyState !== EventSource.CLOSED) {
                    return;
                }

                current.close();

                if (stopped || reopenAttempts >= SSE_MAX_REOPEN_ATTEMPTS) {
                    return;
                }

                reopenAttempts += 1;

                reopenTimer = window.setTimeout(
                    () => {
                        refreshAccessToken()
                            .catch(() => undefined)
                            .then(open);
                    },
                    SSE_REOPEN_DELAY
                );
            };
        };

        open();

        return stop;

    }, [id, messages, chat, user]);

    const firstMessageId = messages.messages[0]?.id ?? null;
    const lastMessageId = messages.messages[messages.messages.length - 1]?.id ?? null;

    useEffect(() => {
        scrollToBottom();
    }, [lastMessageId]);

    useLayoutEffect(() => {
        const container = containerRef.current;
        const offsetFromBottom = offsetFromBottomRef.current;

        if (!container || offsetFromBottom === null) {
            return;
        }

        container.scrollTop = container.scrollHeight - offsetFromBottom;

        if (!messages.isLoadingMore) {
            offsetFromBottomRef.current = null;
        }
    }, [firstMessageId, messages.isLoadingMore]);

    const formatTime = (date: string) => {
        return new Date(date).toLocaleTimeString(
            "ru-RU",
            {
                hour: "2-digit",
                minute: "2-digit"
            }
        );
    };

    return (
        <div className="messages-container" ref={containerRef} onScroll={handleScroll}>

            {messages.isLoadingMore && (
                <div className="messages-loader">Загружаем историю…</div>
            )}

            {messages.messages.map(
                (mess: IMessage) => (
                    <div key={mess.id} className={`message-wrapper ${mess.sender_type === "agent" ? "other" : "own"}`}>
                        <span className="message-time">
                            {formatTime(
                                mess.created_at
                            )}
                        </span>

                        <div className="message-bubble">
                            {mess.content}
                        </div>
                    </div>
                )
            )}

            <div ref={messagesEndRef} />

        </div>
    );
});

export default MessageContainer;