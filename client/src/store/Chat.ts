import type {IChat} from "../types/types.tsx";
import {makeAutoObservable} from "mobx";


class Chat {
    chats: IChat[] = [{id: 1, role: "buyer", difficulty: "easy"}];
    constructor() {
        makeAutoObservable(this);
    }

    addNewChat(newChat: IChat) {
        this.chats.push(newChat);
    }

}

export default new Chat();