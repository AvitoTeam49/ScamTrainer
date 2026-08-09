import {
    useState,
    type KeyboardEvent,
    type FC,
    useContext
} from "react";

import {observer} from "mobx-react-lite";
import {Context} from "../main.tsx";
import {useParams} from "react-router-dom";

const InputArea: FC = observer(() => {

    const [value, setValue] = useState("");

    const {messages, chat} = useContext(Context);

    const {id} = useParams<{id: string}>();

    const [isLoading, setIsLoading] = useState<boolean>(false)

    const handleSendMessage = async () => {

        setIsLoading(true)

        if (!value.trim()) {
            setIsLoading(false)
            return;
        }

        if (!id) {
            setIsLoading(false)
            return;
        }

        const chatId = Number(id);

        if (chat.currentChat?.status === "finished" || chat.currentChat?.status === "abandon") {
            setIsLoading(false)
            return;
        }

        const result = await messages.sendMessage(
            chatId,
            value
        );

        if (result.success) {
            setValue("");
        }

        setIsLoading(false)
    };

    const handleKeyDown = (
        e: KeyboardEvent<HTMLInputElement>
    ) => {

        if (e.key === "Enter") {
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
                    placeholder="Сообщение"
                    disabled={isLoading}
                />

            </div>

            <button
                className="send-btn"
                onClick={handleSendMessage}
                disabled={isLoading}
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

                    <polyline points="5 12 12 5 19 12" />
                </svg>
            </button>

        </div>
    );
});

export default InputArea;