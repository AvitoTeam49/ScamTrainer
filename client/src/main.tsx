import {createContext, StrictMode} from 'react'
import { createRoot } from 'react-dom/client'
import App from './App.tsx'
import Chat from "./store/Chat.ts";
import Messages from "./store/Messages.ts";
import MenuOpen from "./store/MenuOpen.ts";

interface Store{
    chat: Chat,
    messages: Messages,
    menuOpen: MenuOpen
}

const chat = new Chat()
const messages = new Messages()
const menuOpen = new MenuOpen()

export const Context = createContext<Store>({
    chat,
    messages,
    menuOpen,
})


createRoot(document.getElementById('root')!).render(
    <Context.Provider value={{
        chat,
        messages,
        menuOpen
    }}>
        <StrictMode>
            <App />
        </StrictMode>
    </Context.Provider>
)
