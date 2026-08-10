import Sidebar from "./Sidebar.tsx";
import NewChatArea from "./NewChatArea.tsx";
import type {FC} from "react";


const MainNewChat:FC = () => {
    return (
        <div className="main">
            <Sidebar />
            <NewChatArea />
        </div>
    );
};

export default MainNewChat;