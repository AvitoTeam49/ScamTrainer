import "./styles/app.css"
import {BrowserRouter, Route, Routes} from "react-router-dom";
import MainNewChat from "./components/MainNewChat.tsx";
import Main from "./components/Main.tsx";
import Auth from "./components/Auth.tsx";
import {type FC, useContext, useEffect, useState} from "react";
import {Context} from "./main.tsx";
import {observer} from "mobx-react-lite";
import ProtectedRoute from "./components/ProtectedRoute.tsx";

const App:FC = observer(()=> {

    const {auth} = useContext(Context);

    const [isLoading, setIsLoading] = useState<boolean>(true);

    useEffect(() => {
        const checkAuth = async () => {
            if(localStorage.getItem('token')){
                await auth.checkAuth();
            }
            setIsLoading(false);
        };

        checkAuth();
    }, []);

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
                    </Route>
                </Routes>
            </div>
      </BrowserRouter>
  )
});

export default App
