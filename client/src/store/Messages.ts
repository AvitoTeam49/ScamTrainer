import { makeAutoObservable, runInAction } from "mobx";
import type { IMessage } from "../types/types.tsx";
import ChatService from "../services/ChatService.ts";
import axios from "axios";

class Messages {
    messages: IMessage[] = [];

    constructor() {
        makeAutoObservable(this);
    }


    clearMessages() {
        this.messages = [];
    }

    addMessage(message: IMessage) {
        const exists = this.messages.some(item => item.id === message.id);

        if (exists) {
            return;
        }

        this.messages = [...this.messages, message];
    }

    async getMessages(chatId: number): Promise<{ success: boolean; status?: number; }> {
        try {
            const response = await ChatService.getMessages(chatId);

            runInAction(() => {this.messages = [...response.data.items].reverse();});

            return {success: true};

        } catch (e) {
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined};
        }
    }

    async sendMessage(chatId: number, content: string): Promise<{ success: boolean; status?: number; }> {
        try {
             await ChatService.sendMessage(chatId, content.trim());

            return {success: true};

        } catch (e) {
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined};
        }
    }
}

export default Messages;