import {createContext} from 'react'
import { createRoot } from 'react-dom/client'
import App from './App.tsx'
import Chat from "./store/Chat.ts";
import Messages from "./store/Messages.ts";
import MenuOpen from "./store/MenuOpen.ts";
import Auth from "./store/Auth.ts";
import User from "./store/User.ts";

interface Store{
    chat: Chat,
    messages: Messages,
    menuOpen: MenuOpen,
    auth: Auth,
    user: User,
}

const chat = new Chat()
const messages = new Messages()
const menuOpen = new MenuOpen()
const auth = new Auth()
const user = new User()

export const Context = createContext<Store>({
    chat,
    messages,
    menuOpen,
    auth,
    user
})


createRoot(document.getElementById('root')!).render(
    <Context.Provider value={{
        chat,
        messages,
        menuOpen,
        auth,
        user
    }}>
            <App />
    </Context.Provider>
)
