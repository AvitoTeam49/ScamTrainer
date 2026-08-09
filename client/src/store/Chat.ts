import type {IChat, IDecision} from "../types/types.tsx";
import {makeAutoObservable} from "mobx";
import ChatService from "../services/ChatService.ts";
import axios from "axios";


class Chat {
    chats: IChat[] = []
    currentChat: IChat | null = null
    decision: IDecision[] | null = null

    constructor() {
        makeAutoObservable(this);
    }

    setChats(chats: IChat[]) {
        this.chats = chats;
    }

    setCurrentChat(chat: IChat) {
        this.currentChat = chat;
    }

    setDecision(decision: IDecision[]) {
        this.decision = decision;
    }

    async createChat(scenarioId: number, title: string): Promise<{success: boolean, chat?: IChat, status?: number}> {
        try{
            const response = await ChatService.createChat(scenarioId, title);

            const newChat = response.data

            this.chats.push(newChat)

            this.setCurrentChat(newChat)

            return {success: true, chat: newChat}
        }catch(e){
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined}
        }
    }

    async getChat(chatId: number): Promise<{success: boolean, chat?: IChat, status?: number}> {
        try{
            const response = await ChatService.getChat(chatId);
            this.setCurrentChat(response.data)
            return {success: true, chat: response.data}
        }catch(e){

        return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined}
        }
    }

    async getChats(): Promise<{success: boolean, status?: number}> {
        try{
            const response = await ChatService.getChats();
            this.setChats(response.data.items);
            return {success: true}
        }catch(e){
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined}
        }
    }

    async abandonChat(chatId: number): Promise<{success: boolean, status?: number}> {
        try{
            const response = await ChatService.abandonChat(chatId);
            this.setCurrentChat(response.data)
            const index = this.chats.findIndex(chat => chat.id === chatId);
            if(index !== - 1) {
                this.chats[index] = response.data
            }
            return {success: true}
        }catch(e){
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined}
        }
    }

    async getDecision(chatId: number): Promise<{success: boolean, status?: number}> {
        try{
            const response = await ChatService.getDecisions(chatId);
            this.setDecision(response.data.items);
            return {success: true}
        }catch(e){
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined}
        }
    }

}

export default Chat;