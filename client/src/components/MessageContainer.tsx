import { observer } from "mobx-react-lite";
import type { IMessage } from "../types/types.tsx";
import {type FC, useContext, useEffect, useRef} from "react";
import { Context } from "../main.tsx";
import { useParams } from "react-router-dom";

const MessageContainer: FC = observer(() => {
    const {messages, chat, user} = useContext(Context);

    const { id } = useParams<{ id: string }>();

    const messagesEndRef = useRef<HTMLDivElement | null>(null);

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({behavior: "smooth"});
    };

    useEffect(() => {
        if (!id) {return;}

        const chatId = Number(id);

        if (!Number.isFinite(chatId)) {
            return;
        }

        const eventSource = new EventSource(`http://localhost:8080/api/v1/chats/${chatId}/events`);

        const handleMessage = (event: MessageEvent) => {
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

                            eventSource.close();
                        }

                        break;
                    }
                }
            } catch (error) {
                console.error("Ошибка обработки SSE:", error);
            }
        };

        eventSource.addEventListener("message", handleMessage);

        eventSource.onerror = () => {eventSource.close();};

        return () => {
            eventSource.close();
        };

    }, [id, messages, chat, user]);

    useEffect(() => {
        scrollToBottom();
    }, [messages.messages.length]);

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
        <div className="messages-container">

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