import {observer} from "mobx-react-lite";
import type {IChat} from "../types/types.tsx";
import {useNavigate} from "react-router-dom";
import {useContext} from "react";
import {Context} from "../main.tsx";


interface SidebarProps {
    id?: string
}

const Sidebar= observer(({id}: SidebarProps) => {

    const navigate = useNavigate()
    const {menuOpen, chat} = useContext(Context)

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
                {chat.chats.map((ch: IChat) => (
                    <li key={ch.id}
                        className={`chat-item ${String(ch.id) === id ? "active" : ""}`} onClick={() => navigate(`/chat/${ch.id}`)}>{ch.title}</li>
                ))}
            </ul>
        </div>
    );
});

export default Sidebar;