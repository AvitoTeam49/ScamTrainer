import {type FC, useContext, useEffect, useState} from "react";
import "./styles/app.css"
import {Routes, Route} from "react-router-dom";
import { observer } from "mobx-react-lite";
import { Context } from "./main.tsx";
import MainNewChat from "./components/MainNewChat.tsx";
import Main from "./components/Main.tsx";
import Auth from "./components/Auth.tsx";
import ProtectedRoute from "./components/ProtectedRoute.tsx";
import MainProfile from "./components/MainProfile.tsx";

const App: FC = observer(() => {
    const {auth, user, chat} = useContext(Context);

    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        let cancelled = false;

        const initialize = async () => {
            try {
                const token =
                    localStorage.getItem("token");

                if (!token) {
                    auth.setAuth(false);
                    return;
                }

                const authResult = await auth.checkAuth();

                if (!authResult.success) {
                    auth.setAuth(false);
                    return;
                }

                await Promise.all([user.getProfile(), chat.getChats()]);

                auth.setAuth(true);

            } finally {
                if (!cancelled) {
                    setIsLoading(false);
                }
            }
        };

        initialize();

        return () => {
            cancelled = true;
        };

    }, [auth, user, chat]);

    if (isLoading) {
        return (
            <div className="loading-container">
                <div className="spinner" />
            </div>
        );
    }

    return (
        <Routes>

            <Route path="/auth" element={<Auth />}/>

            <Route element={<ProtectedRoute isAllowed={auth.isAuth}/>}>
                <Route path="/" element={<MainNewChat />}/>
                <Route path="/chat/:id" element={<Main />}/>
                <Route path="/profile" element={<MainProfile />}/>
            </Route>

        </Routes>
    );
});

export default App;