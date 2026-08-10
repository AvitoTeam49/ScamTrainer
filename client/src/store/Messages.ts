import { makeAutoObservable, runInAction } from "mobx";
import type { IMessage } from "../types/types.tsx";
import ChatService, { PAGE_SIZE } from "../services/ChatService.ts";
import axios from "axios";

class Messages {
    messages: IMessage[] = [];
    nextAfterId: number | null = null;
    isLoadingMore: boolean = false;
    loadedChatId: number | null = null;

    constructor() {
        makeAutoObservable(this);
    }

    get hasMore() {
        return this.nextAfterId !== null;
    }

    clearMessages() {
        this.messages = [];
        this.nextAfterId = null;
        this.isLoadingMore = false;
        this.loadedChatId = null;
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

            runInAction(() => {
                this.loadedChatId = chatId;
                this.messages = [...response.data.items].reverse();
                this.nextAfterId = response.data.next_after_id ?? null;
            });

            return {success: true};

        } catch (e) {
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined};
        }
    }

    async loadMoreMessages(chatId: number): Promise<{ success: boolean; status?: number; }> {
        const afterId = this.nextAfterId;

        if (afterId === null || this.isLoadingMore || this.loadedChatId !== chatId) {
            return {success: true};
        }

        this.isLoadingMore = true;

        try {
            const response = await ChatService.getMessages(chatId, PAGE_SIZE, afterId);

            runInAction(() => {
                if (this.loadedChatId !== chatId) {
                    return;
                }

                const known = new Set(this.messages.map(item => item.id));

                const older = response.data.items
                    .filter(item => !known.has(item.id))
                    .reverse();

                this.messages = [...older, ...this.messages];
                this.nextAfterId = response.data.next_after_id ?? null;
            });

            return {success: true};

        } catch (e) {
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined};
        } finally {
            runInAction(() => {this.isLoadingMore = false;});
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
