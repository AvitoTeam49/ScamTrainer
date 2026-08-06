import {observer} from "mobx-react-lite/src/observer.ts";
import menuOpen from "../store/MenuOpen.ts";


const ChatHeader = observer(() => {
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