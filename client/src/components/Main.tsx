import Sidebar from "./Sidebar.tsx";
import ChatArea from "./ChatArea.tsx";
import {useParams} from 'react-router-dom'
import {type FC, useContext, useEffect} from "react";
import {Context} from "../main.tsx";

const Main:FC = () => {

    const {id} = useParams<{id: string}>();
    const {chat,messages} = useContext(Context)

    useEffect(() => {
        if(!id) return

        const chatId = Number(id)
        chat.getChat(chatId)
        messages.getMessages(chatId)


    }, [id, chat, messages]);


    return (
        <div className="main">
            <Sidebar id={id}/>
            <ChatArea/>
        </div>
    );
};

export default Main;