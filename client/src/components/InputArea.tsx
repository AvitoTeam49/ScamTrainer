import {observer} from "mobx-react-lite";
import Messages from "../store/Messages.ts";
import {useState, type KeyboardEvent} from "react";

const InputArea = observer(() => {
    const [value, setValue] = useState<string>("")

    const handleSendMessage = () => {
        if(value.trim() === "") return
        const date = new Date()
        const time_now = date.toLocaleTimeString('ru-RU', {
            hour: '2-digit',
            minute: '2-digit'
        });

        Messages.addNewMessage({content: value, who: "own", time: time_now})

        setValue("")
    }

    const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
        if(e.key === "Enter"){
            e.preventDefault()
            handleSendMessage()
        }

    }

    return (
        <div className="input-area">
            <div className="input-wrapper">
                <input value={value} onChange={(e) => setValue(e.target.value)} onKeyDown={handleKeyDown} type="text" placeholder="Сообщение"/>
            </div>
            <button className="send-btn" onClick={handleSendMessage}> <svg
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2.5"
                strokeLinecap="round"
                strokeLinejoin="round"
            >
                <line x1="12" y1="19" x2="12" y2="5"></line>
                <polyline points="5 12 12 5 19 12"></polyline>
            </svg></button>
        </div>
    );
});

export default InputArea;