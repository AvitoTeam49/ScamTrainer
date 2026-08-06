import Sidebar from "./Sidebar.tsx";
import ChatArea from "./ChatArea.tsx";
import {useParams} from 'react-router-dom'

const Main = () => {

    const {id} = useParams<{id: string}>();
    return (
        <div className="main">
            <Sidebar id={id}/>
            <ChatArea/>
        </div>
    );
};

export default Main;