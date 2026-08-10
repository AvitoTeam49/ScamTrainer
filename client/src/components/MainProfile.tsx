import Sidebar from "./Sidebar.tsx";
import Profile from "./Profile.tsx";
import LeaderBoard from "./LeaderBoard.tsx";
import {type FC, useContext, useEffect, useState} from "react";
import { Context } from "../main.tsx";
import { observer } from "mobx-react-lite";

const MainProfile: FC = observer(() => {
    const [isLoading, setIsLoading] = useState(true);

    const { user } = useContext(Context);

    useEffect(() => {
        const loadProfile = async () => {
            setIsLoading(true);

            try {
                await Promise.all([user.getProfile(), user.getProgress(), user.getLeaderboard()]);
            } finally {
                setIsLoading(false);
            }
        };

        loadProfile();
    }, [user]);

    if (isLoading) {
        return (
            <div className="loading-container">
                <div className="spinner" />
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