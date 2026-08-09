import Sidebar from "./Sidebar.tsx";
import Profile from "./Profile.tsx";
import LeaderBoard from "./LeaderBoard.tsx";
import {type FC, useContext, useEffect, useState} from "react";
import {Context} from "../main.tsx";
import {observer} from "mobx-react-lite";

const MainProfile:FC = observer(() => {

    const [isLoading, setIsLoading] = useState<boolean>(true)
    const {user} = useContext(Context)

    useEffect(() => {
        const getProgress = async () => {
            await Promise.all([user.getProgress(), user.getLeaderboard()])
            setIsLoading(false)
        }

        getProgress()

    },[user])

    if(isLoading){
        return (
            <div className="loading-container">
                <div className="spinner"></div>
            </div>
        );
    }

    return (
        <div className="app-container">
            <Sidebar />
            <Profile />
            <LeaderBoard />
        </div>
    );
});

export default MainProfile;