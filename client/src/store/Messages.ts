import type {IMessage} from "../types/types.tsx";
import {makeAutoObservable} from "mobx";
import ChatService from "../services/ChatService.ts";
import axios from "axios";


class Messages {
    messages: IMessage[] = []

    constructor() {
        makeAutoObservable(this)
    }

    setMessages(messages: IMessage[]) {
        this.messages = messages;
    }

    addMessage(message: IMessage) {
        this.messages.push(message);
    }

    async getMessages(chatId: number): Promise<{success: boolean, status?: number}> {
        try{
            const response = await ChatService.getMessages(chatId);
            this.setMessages([...response.data.items].reverse());
            return {success: true}
        }catch(e){
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined}
        }
    }

    async sendMessage(chatId: number, content: string): Promise<{success: boolean, status?: number}> {
        try{
            await ChatService.sendMessage(chatId, content.trim());
            return {success: true}
        }catch(e){
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined}
        }
    }

}

export default Messages;