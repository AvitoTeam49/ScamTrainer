import "./styles/app.css"
import {BrowserRouter, Route, Routes} from "react-router-dom";
import MainNewChat from "./components/MainNewChat.tsx";
import Main from "./components/Main.tsx";
import Auth from "./components/Auth.tsx";
import {type FC, useContext, useEffect, useState} from "react";
import {Context} from "./main.tsx";
import {observer} from "mobx-react-lite";
import ProtectedRoute from "./components/ProtectedRoute.tsx";
import MainProfile from "./components/MainProfile.tsx";

const App:FC = observer(()=> {

    const {auth, user} = useContext(Context);

    const [isLoading, setIsLoading] = useState<boolean>(true);

    useEffect(() => {
        const initialize = async () => {

            const token = localStorage.getItem("token");

            if(!token){
                auth.setAuth(false)
                setIsLoading(false)
                return
            }

            const isValid = await auth.checkAuth()

            if(isValid){
                await user.getProfile()
                auth.setAuth(true)
            }

            setIsLoading(false);
        };

        initialize()
    }, [auth, user]);

    if(isLoading){
        return (
          <div>Загрузка...</div>
        );
    }



  return (
      <BrowserRouter>
            <div>
                <Routes>

                    <Route path="/auth" element={<Auth/>}/>

                    <Route element={<ProtectedRoute isAllowed={auth.isAuth}/>}>
                        <Route path="/" element={<MainNewChat/>}/>
                        <Route path="/chat/:id" element={<Main/>}/>
                        <Route path="/profile" element={<MainProfile/>}/>
                    </Route>
                </Routes>
            </div>
      </BrowserRouter>
  )
});

export default App
