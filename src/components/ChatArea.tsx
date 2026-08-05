import ChatHeader from "./ChatHeader.tsx";
import MessageContainer from "./MessageContainer.tsx";
import InputArea from "./InputArea.tsx";


interface ChatAreaProps {
    toggleMenu: () => void
}

const ChatArea = ({toggleMenu}: ChatAreaProps) => {
    return (
        <div className="chat-area">
            <ChatHeader toggleMenu={toggleMenu}/>
            <MessageContainer/>
            <InputArea/>
        </div>
    );
};

export default ChatArea;