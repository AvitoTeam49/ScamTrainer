import Sidebar from "./Sidebar.tsx";
import ChatArea from "./ChatArea.tsx";

const Main = () => {

    return (
        <div className="main">
            <Sidebar/>
            <ChatArea/>
        </div>
    );
};

export default Main;