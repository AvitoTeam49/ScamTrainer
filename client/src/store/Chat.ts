import { makeAutoObservable, runInAction } from "mobx";
import type { IChat, IDecision } from "../types/types.tsx";
import ChatService, { PAGE_SIZE } from "../services/ChatService.ts";
import axios from "axios";

class Chat {
    chats: IChat[] = [];
    currentChat: IChat | null = null;
    decision: IDecision[] = [];
    chatsNextAfterId: number | null = null;
    isLoadingChats: boolean = false;

    constructor() {
        makeAutoObservable(this);
    }

    get hasMoreChats() {
        return this.chatsNextAfterId !== null;
    }

    setCurrentChat(chat: IChat | null) {
        this.currentChat = chat;
    }


    clearDecision() {
        this.decision = [];
    }

    updateChatInList(updatedChat: IChat) {
        const index = this.chats.findIndex(
            chat => chat.id === updatedChat.id
        );

        if (index === -1) {
            this.chats = [updatedChat, ...this.chats];
            return;
        }

        const newChats = [...this.chats];
        newChats[index] = updatedChat;

        this.chats = newChats;
    }

    async createChat(scenarioId: number, title: string): Promise<{ success: boolean; chat?: IChat; status?: number}> {
        try {
            const response = await ChatService.createChat(scenarioId, title);
            const newChat = response.data;
            runInAction(() => {
                this.chats = [newChat, ...this.chats];
                this.currentChat = newChat;
                this.decision = [];
            });

            return {success: true, chat: newChat};
        } catch (e) {
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined};
        }
    }

    async getChat(chatId: number): Promise<{ success: boolean; chat?: IChat; status?: number}> {
        try {
            const response = await ChatService.getChat(chatId);

            runInAction(() => {
                this.currentChat = response.data;
                this.decision = [];
            });

            return {success: true, chat: response.data};
        } catch (e) {
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined};
        }
    }

    async getChats(): Promise<{ success: boolean; status?: number}> {
        try {
            const response = await ChatService.getChats();

            runInAction(() => {
                this.chats = response.data.items;
                this.chatsNextAfterId = response.data.next_after_id ?? null;
            });

            return {success: true};
        } catch (e) {
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined};
        }
    }

    async loadMoreChats(): Promise<{ success: boolean; status?: number}> {
        const afterId = this.chatsNextAfterId;

        if (afterId === null || this.isLoadingChats) {
            return {success: true};
        }

        this.isLoadingChats = true;

        try {
            const response = await ChatService.getChats(PAGE_SIZE, afterId);

            runInAction(() => {
                const known = new Set(this.chats.map(chat => chat.id));

                const older = response.data.items.filter(chat => !known.has(chat.id));

                this.chats = [...this.chats, ...older];
                this.chatsNextAfterId = response.data.next_after_id ?? null;
            });

            return {success: true};
        } catch (e) {
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined};
        } finally {
            runInAction(() => {this.isLoadingChats = false;});
        }
    }

    async abandonChat(chatId: number): Promise<{ success: boolean; chat?: IChat; status?: number}> {
        try {
            const response = await ChatService.abandonChat(chatId);

            const updatedChat = response.data;

            runInAction(() => {
                this.currentChat = updatedChat;
                this.updateChatInList(updatedChat);
            });

            return {success: true, chat: updatedChat};
        } catch (e) {
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined};
        }
    }

    async getDecision(chatId: number): Promise<{ success: boolean; status?: number}> {
        try {
            const response = await ChatService.getDecisions(chatId);

            runInAction(() => {this.decision = response.data.items;});

            return {success: true};
        } catch (e) {
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined};
        }
    }

    handleSSEChat(updatedChat: IChat) {
        runInAction(() => {
            this.currentChat = updatedChat;
            this.updateChatInList(updatedChat);
        });
    }

    handleSSEDecision(decision: IDecision[]) {
        runInAction(() => {
            this.decision = decision;
        });
    }
}

export default Chat;