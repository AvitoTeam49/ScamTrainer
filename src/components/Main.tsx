import Sidebar from "./Sidebar.tsx";
import ChatArea from "./ChatArea.tsx";
import {type FC, useState} from "react";

const Main:FC = () => {

    const [isMenuOpen, setIsMenuOpen] = useState<boolean>(false)

    const toggleMenu = () => {
        setIsMenuOpen(prev => !prev)
    }

    return (
        <div className="main">
            <Sidebar isMenuOpen={isMenuOpen} toggleMenu={toggleMenu}/>
            <ChatArea toggleMenu={toggleMenu}/>
        </div>
    );
};

export default Main;