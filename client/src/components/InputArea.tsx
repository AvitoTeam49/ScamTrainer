import {useState, type KeyboardEvent, type FC, useContext} from "react";
import { observer } from "mobx-react-lite";
import { Context } from "../main.tsx";
import { useParams } from "react-router-dom";

const InputArea: FC = observer(() => {
    const [value, setValue] = useState("");
    const [isLoading, setIsLoading] = useState(false);

    const {messages, chat} = useContext(Context);

    const { id } = useParams<{ id: string }>();

    const isChatFinished = chat.currentChat?.status === "finished" || chat.currentChat?.status === "abandoned";

    const handleSendMessage = async () => {
        const content = value.trim();

        if (isLoading || !content || !id || isChatFinished) {
            return;
        }

        const chatId = Number(id);

        if (!Number.isFinite(chatId)) {
            return;
        }

        setIsLoading(true);

        try {
            const result = await messages.sendMessage(chatId, content);
            if (result.success) {
                setValue("");
            }

        } finally {
            setIsLoading(false);
        }
    };

    const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
        if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            handleSendMessage();
        }
    };

    return (
        <div className="input-area">

            <div className="input-wrapper">
                <input
                    value={value}
                    onChange={e =>
                        setValue(e.target.value)
                    }
                    onKeyDown={handleKeyDown}
                    type="text"
                    placeholder={isChatFinished ? "Чат завершен" : "Сообщение"}
                    disabled={isLoading || isChatFinished}
                />
            </div>

            <button
                className="send-btn"
                onClick={handleSendMessage}
                disabled={isLoading || isChatFinished || !value.trim()}
            >
                <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2.5"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                >
                    <line
                        x1="12"
                        y1="19"
                        x2="12"
                        y2="5"
                    />

                    <polyline
                        points="5 12 12 5 19 12"
                    />
                </svg>
            </button>

        </div>
    );
});

export default InputArea;