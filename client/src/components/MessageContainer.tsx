import {observer} from "mobx-react-lite";
import type {IMessage} from "../types/types.tsx";
import {type FC, useContext} from "react";
import {Context} from "../main.tsx";

const MessageContainer:FC = observer(() => {

    const {messages} = useContext(Context)

    return (
        <div className="messages-container">
            {messages.messages.map((mess: IMessage)=> (
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