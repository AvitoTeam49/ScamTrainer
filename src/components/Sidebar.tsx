import {observer} from "mobx-react-lite";
import isMenuOpen from "../store/MenuOpen.ts";
import Chat from "../store/Chat.ts";
import type {IChat} from "../types/types.tsx";

const Sidebar = observer(() => {

    return (
        <div className={`sidebar ${isMenuOpen.menu ? 'open' : ''}`}>
            <div className="logo-container">
                <div className="logo-text">
                <span className="logo-icon">
                    <span className="logo-dot dot-blue"></span>
                    <span className="logo-dot dot-red"></span>
                    <span className="logo-dot dot-green"></span>
                </span>
                    Avito
                </div>
                <button className="close-sidebar-btn" onClick={isMenuOpen.setFalse}>✕</button>
            </div>

            <button className="new-chat-btn">
                <span>+</span> Новый чат
            </button>

            <ul className="user-list">
                {Chat.chats.map((chat: IChat) => (
                    <li className="user-item">Chat {chat.role} {chat.difficulty} {chat.id}</li>
                ))}
            </ul>
        </div>
    );
});

export default Sidebar;