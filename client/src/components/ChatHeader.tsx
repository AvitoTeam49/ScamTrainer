import {observer} from "mobx-react-lite";
import {type FC, useContext} from "react";
import {Context} from "../main.tsx";
import {useNavigate} from "react-router-dom";


const ChatHeader:FC = observer(() => {

    const {menuOpen, user} = useContext(Context)
    const navigate = useNavigate()

    return (
        <div className="chat-header">
            <button className="menu-toggle" onClick={menuOpen.setTrue}>☰</button>
            <div className="user-avatar">U</div>
            <div className="chat-info">
                <div className="chat-title">Комплект GPU</div>
                <div className="chat-price">500₽</div>
            </div>
            <button className="header-username" onClick={() => navigate("/profile")}>{user.user.username}</button>
        </div>
    );
});

export default ChatHeader;