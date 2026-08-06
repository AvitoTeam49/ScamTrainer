import type {IMessage} from "../types/types.tsx";
import {makeAutoObservable} from "mobx";


class Messages {
    messages: IMessage[] = [
        {content: "Привет, хотел бы купить комплект!", who:"other", time: "15:36"},
        {content: "Здравствуйте, хорошо.", who:"own", time: "15:37"},
    ]
    constructor() {
        makeAutoObservable(this)
    }

    addNewMessage = (newMessage: IMessage) => {
        this.messages.push(newMessage)
    }

}

export default new Messages()