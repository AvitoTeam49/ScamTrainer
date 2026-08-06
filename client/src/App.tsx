import "./styles/app.css"
import {BrowserRouter, Route, Routes} from "react-router-dom";
import MainNewChat from "./components/MainNewChat.tsx";
import Main from "./components/Main.tsx";

function App() {

  return (
      <BrowserRouter>
            <div>
                <Routes>
                    <Route path="/" element={<MainNewChat/>}/>
                    <Route path="/chat/:id" element={<Main/>}/>
                </Routes>
            </div>
      </BrowserRouter>
  )
}

export default App
