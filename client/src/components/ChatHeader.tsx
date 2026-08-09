import {observer} from "mobx-react-lite";
import {type FC, useContext, useState, useEffect} from "react";
import {Context} from "../main.tsx";
import {useNavigate, useParams} from "react-router-dom";

const ChatHeader: FC = observer(() => {
    const {id} = useParams<{id: string}>()
    const {menuOpen, user, chat} = useContext(Context);
    const navigate = useNavigate();
    const [isExplanationOpen, setIsExplanationOpen] = useState<boolean>(false);

    useEffect(() => {
        if (isExplanationOpen) {
            document.body.style.overflow = 'hidden';
        } else {
            document.body.style.overflow = 'unset';
        }
        return () => {
            document.body.style.overflow = 'unset';
        };
    }, [isExplanationOpen]);

    const handleEndChat = async() => {
        await chat.abandonChat(Number(id))
    }


    return (
        <div className="chat-header">
            <button className="menu-toggle" onClick={menuOpen.setTrue}>☰</button>
            <div className="user-avatar">U</div>
            <div className="chat-info">
                <div className="chat-title">{chat.currentChat?.title}</div>
                {chat.currentChat?.status === "active" ? <button onClick={handleEndChat}>Завершить чат</button> : null}
                {chat.currentChat?.status === "finished" || chat.currentChat?.status === "abandon" ? (
                    <div style={{ display: 'flex', alignItems: 'center', flexWrap: 'wrap', marginTop: '4px' }}>
                        <span style={{ fontSize: '13px', color: '#666' }}>Чат завершен</span>
                        <button
                            className="explain-btn"
                            onClick={() => setIsExplanationOpen(true)}
                        >
                            Показать объяснения
                        </button>
                    </div>
                ) : null}
            </div>
            <button className="header-username" onClick={() => navigate("/profile")}>{user.user?.username}</button>

            {isExplanationOpen && (
                <div className="modal-overlay" onClick={() => setIsExplanationOpen(false)}>
                    <div className="modal-content" onClick={(e) => e.stopPropagation()}>
                        <button
                            className="modal-close-btn"
                            onClick={() => setIsExplanationOpen(false)}
                        >
                            ×
                        </button>

                        <h2 className="modal-title">Объяснения решений</h2>
                        {chat.decision?.map((des) => (
                            <div className="modal-body">
                                {des.feedback} счет: {des.score_delta}
                            </div>
                        ))}
                        <button
                            className="modal-footer-btn"
                            onClick={() => setIsExplanationOpen(false)}
                        >
                            Закрыть
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
});

export default ChatHeader;