import {observer} from "mobx-react-lite";
import menuOpen from "../store/MenuOpen.ts";
import Chat from "../store/Chat.ts";
import type {IChat} from "../types/types.tsx";
import {useNavigate} from "react-router-dom";


interface SidebarProps {
    id?: string
}

const Sidebar = observer(({id}: SidebarProps) => {

    const navigate = useNavigate()

    return (
        <div className={`sidebar ${menuOpen.menu ? 'open' : ''}`}>
            <div className="logo-container">
                <div className="logo-text" onClick={() => navigate("/")}>
                <span className="logo-icon">
                    <span className="logo-dot dot-blue"></span>
                    <span className="logo-dot dot-red"></span>
                    <span className="logo-dot dot-green"></span>
                </span>
                    Avito
                </div>
                <button className="close-sidebar-btn" onClick={menuOpen.setFalse}>✕</button>
            </div>

            <button className="new-chat-btn" onClick={() => navigate("/")}>
                <span>+</span> Новый чат
            </button>

            <ul className="chat-list">
                {Chat.chats.map((chat: IChat) => (
                    <li key={chat.id}
                        className={`chat-item ${String(chat.id) === id ? "active" : ""}`} onClick={() => navigate(`/chat/${chat.id}`)}>Chat {chat.role} {chat.difficulty} {chat.id}</li>
                ))}
            </ul>
        </div>
    );
});

export default Sidebar;