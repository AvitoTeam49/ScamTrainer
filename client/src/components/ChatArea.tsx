import ChatHeader from "./ChatHeader.tsx";
import MessageContainer from "./MessageContainer.tsx";
import InputArea from "./InputArea.tsx";
import type {FC} from "react";

const ChatArea:FC = () => {
    return (
        <div className="chat-area">
            <ChatHeader/>
            <MessageContainer/>
            <InputArea/>
        </div>
    );
};

export default ChatArea;