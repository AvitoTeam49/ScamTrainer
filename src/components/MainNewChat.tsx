import Sidebar from "./Sidebar.tsx";
import NewChatArea from "./NewChatArea.tsx";


const MainNewChat = () => {
    return (
        <div className="main">
            <Sidebar></Sidebar>
            <NewChatArea></NewChatArea>
        </div>
    );
};

export default MainNewChat;