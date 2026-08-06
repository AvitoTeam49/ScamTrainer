import "./styles/app.css"
import {BrowserRouter, Route, Routes} from "react-router-dom";
import MainNewChat from "./components/MainNewChat.tsx";
import Main from "./components/Main.tsx";
import Auth from "./components/Auth.tsx";

function App() {

  return (
      <BrowserRouter>
            <div>
                <Routes>
                    <Route path="/" element={<MainNewChat/>}/>
                    <Route path="/auth" element={<Auth/>}/>
                    <Route path="/chat/:id" element={<Main/>}/>
                </Routes>
            </div>
      </BrowserRouter>
  )
}

export default App
