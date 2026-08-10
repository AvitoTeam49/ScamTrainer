import { observer } from "mobx-react-lite";
import { type FC, useContext, useEffect, useState } from "react";
import { Context } from "../main.tsx";
import { useNavigate, useParams } from "react-router-dom";

const ChatHeader: FC = observer(() => {
    const { id } = useParams<{ id: string }>();

    const {menuOpen, user, chat} = useContext(Context);

    const navigate = useNavigate();

    const [isExplanationOpen, setIsExplanationOpen] = useState(false);
    const [isEnding, setIsEnding] = useState(false);

    useEffect(() => {
        if (!id) {
            return;
        }

        const chatId = Number(id);

        if (!Number.isFinite(chatId)) {
            return;
        }

        const loadChat = async () => {
            await chat.getChat(chatId);
        };

        loadChat();
    }, [id, chat]);

    useEffect(() => {
        if (!id) {
            return;
        }

        const currentChat = chat.currentChat;

        if (currentChat?.status === "finished" || currentChat?.status === "abandoned") {
            chat.getDecision(Number(id));
        }
    }, [id, chat, chat.currentChat?.status]);

    const handleEndChat = async () => {
        if (!id || isEnding) {
            return;
        }

        setIsEnding(true);

        const result = await chat.abandonChat(Number(id));

        if (result.success) {
            await chat.getDecision(Number(id));

            await user.refreshAfterChat();
        }

        setIsEnding(false);
    };

    const isFinished =
        chat.currentChat?.status === "finished" ||
        chat.currentChat?.status === "abandoned";

    return (
        <div className="chat-header">

            <button className="menu-toggle" onClick={menuOpen.setTrue}>
                ☰
            </button>

            <div className="user-avatar">
                U
            </div>

            <div className="chat-info">

                <div className="chat-title">
                    {chat.currentChat?.title}
                </div>

                {chat.currentChat?.status === "active" && (
                    <button
                        onClick={handleEndChat}
                        disabled={isEnding}
                    >
                        {isEnding ? "Завершение..." : "Завершить чат"}
                    </button>
                )}

                {isFinished && (
                    <div style={{
                            display: "flex",
                            alignItems: "center",
                            flexWrap: "wrap",
                            marginTop: "4px"
                        }}
                    >
                        <span style={{
                                fontSize: "13px",
                                color: "#666"
                            }}
                        >
                            Чат завершен
                        </span>

                        <button className="explain-btn" onClick={() => setIsExplanationOpen(true)}>
                            Показать объяснения
                        </button>
                    </div>
                )}

            </div>

            <button className="header-username" onClick={() => navigate("/profile")}>
                {user.user?.username}
            </button>

            {isExplanationOpen && (
                <div className="modal-overlay" onClick={() => setIsExplanationOpen(false)}>
                    <div className="modal-content" onClick={e => e.stopPropagation()}>
                        <button className="modal-close-btn" onClick={() => setIsExplanationOpen(false)}>
                            ×
                        </button>

                        <h2 className="modal-title">
                            Объяснения решений
                        </h2>

                        {chat.decision.length === 0 ? (
                            <div className="modal-body">
                                Объяснения пока отсутствуют.
                            </div>
                        ) : (
                            chat.decision.map((des, index) => (
                                <div className="modal-body" key={`${des.chat_id}-${index}`}>
                                    <div>
                                        {des.feedback}
                                    </div>

                                    <div>
                                        Счет: {des.score_delta}
                                    </div>
                                </div>
                            ))
                        )}

                        <button className="modal-footer-btn" onClick={() => setIsExplanationOpen(false)}>
                            Закрыть
                        </button>
                    </div>
                </div>
            )}

        </div>
    );
});

export default ChatHeader;