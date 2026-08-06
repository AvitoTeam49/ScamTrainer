import {observer} from "mobx-react-lite";
import Messages from "../store/Messages.ts";
import type {IMessage} from "../types/types.tsx";

const MessageContainer = observer(() => {

    return (
        <div className="messages-container">
            {Messages.messages.map((mess: IMessage)=> (
                <div className={`message-wrapper ${mess.who === "other" ? "other" : "own"}`}>
                    <span className="message-time">{mess.time}</span>
                    <div className="message-bubble">
                        {mess.content}
                    </div>
                </div>
            ) )}

        </div>
    );
});

export default MessageContainer;