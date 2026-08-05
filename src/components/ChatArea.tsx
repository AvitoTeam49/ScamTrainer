import ChatHeader from "./ChatHeader.tsx";
import MessageContainer from "./MessageContainer.tsx";
import InputArea from "./InputArea.tsx";

const ChatArea = () => {
    return (
        <div className="chat-area">
            <ChatHeader/>
            <MessageContainer/>
            <InputArea/>
        </div>
    );
};

export default ChatArea;