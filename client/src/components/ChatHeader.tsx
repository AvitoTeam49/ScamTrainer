import {observer} from "mobx-react-lite";
import {type FC, useContext} from "react";
import {Context} from "../main.tsx";


const ChatHeader:FC = observer(() => {

    const {menuOpen} = useContext(Context)

    return (
        <div className="chat-header">
            <button className="menu-toggle" onClick={menuOpen.setTrue}>☰</button>
            <div className="user-avatar">U</div>
            <div className="chat-info">
                <div className="chat-title">Комплект GPU</div>
                <div className="chat-price">500₽</div>
            </div>
        </div>
    );
});

export default ChatHeader;