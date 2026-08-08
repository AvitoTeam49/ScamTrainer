import {createContext} from 'react'
import { createRoot } from 'react-dom/client'
import App from './App.tsx'
import Chat from "./store/Chat.ts";
import Messages from "./store/Messages.ts";
import MenuOpen from "./store/MenuOpen.ts";
import Auth from "./store/Auth.ts";

interface Store{
    chat: Chat,
    messages: Messages,
    menuOpen: MenuOpen,
    auth: Auth,
}

const chat = new Chat()
const messages = new Messages()
const menuOpen = new MenuOpen()
const auth = new Auth()

export const Context = createContext<Store>({
    chat,
    messages,
    menuOpen,
    auth,
})


createRoot(document.getElementById('root')!).render(
    <Context.Provider value={{
        chat,
        messages,
        menuOpen,
        auth
    }}>
            <App />
    </Context.Provider>
)
