import {observer} from "mobx-react-lite";
import type {IMessage} from "../types/types.tsx";
import {type FC, useContext, useEffect} from "react";
import {Context} from "../main.tsx";
import {useParams} from "react-router-dom";

const MessageContainer:FC = observer(() => {

    const {messages, chat} = useContext(Context)
    const {id} = useParams<{id: string}>()

    useEffect(() => {
        if (!id) return;

        const eventSource = new EventSource(`http://localhost:8080/api/v1/chats/${id}/events`);

        eventSource.addEventListener('message', (event) => {
            const data = JSON.parse(event.data);
            switch (data.event) {
                case "message":
                    if (data.message) {
                        messages.addMessage(data.message);
                    }
                    break;

                case "decision":
                    if(data.decision){
                        chat.setDecision(data.decision)
                    }
                    break;

                case "chat":
                    if (data.chat) {
                        chat.setCurrentChat(data.chat);
                        if (data.chat.status === "finished") {
                            eventSource.close();
                        }
                    }
                    break;
            }
        });

        eventSource.onerror = () => {
            eventSource.close();
        };

        return () => {
            eventSource.close();
        };
    }, [id, messages, chat]);


    const formatTime = (date: string) => {

        return new Date(date).toLocaleTimeString(
            "ru-RU",
            {
                hour: "2-digit",
                minute: "2-digit"
            }
        );
    };

    return (
        <div className="messages-container">
            {messages.messages.map((mess: IMessage)=> (
                <div key={mess.id} className={`message-wrapper ${mess.sender_type === "agent" ? "other" : "own"}`}>
                    <span className="message-time">{formatTime(mess.created_at)}</span>
                    <div className="message-bubble">
                        {mess.content}
                    </div>
                </div>
            ) )}

        </div>
    );
});

export default MessageContainer;