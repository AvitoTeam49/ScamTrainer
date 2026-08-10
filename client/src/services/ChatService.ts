import $api from "../http";
import type {AxiosResponse} from "axios";
import type {IChat, IListChats, IListDecision, IListMessage, IMessage} from "../types/types.tsx";

export default class ChatService {

    static async createChat(scenario_id: number, title: string): Promise<AxiosResponse<IChat>> {
        return $api.post<IChat>("/chats", {scenario_id, title});
    }

    static async getChats(limit: number = 50,  afterId: number = 0): Promise<AxiosResponse<IListChats>>{

        return $api.get<IListChats>("/chats", {
            params: {
                limit,
                after_id: afterId,
            }
        });
    }

    static async getChat(chatId: number): Promise<AxiosResponse<IChat>> {
        return $api.get<IChat>(`/chats/${chatId}`);
    }

    static async abandonChat(chatId: number): Promise<AxiosResponse<IChat>>{
        return $api.post<IChat>(`/chats/${chatId}/abandon`)
    }

    static async sendMessage(chatId: number, content: string): Promise<AxiosResponse<IMessage>> {
        return $api.post<IMessage>(`/chats/${chatId}/messages`, {content});
    }

    static async getMessages(chatId: number, limit: number = 50,  afterId: number = 0): Promise<AxiosResponse<IListMessage>> {
        return $api.get<IListMessage>(`/chats/${chatId}/messages`, {
            params: {
                limit,
                after_id: afterId,
            }
        })
    }

    static async getDecisions(chatId: number, limit: number = 50,  afterId: number = 0): Promise<AxiosResponse<IListDecision>> {
        return $api.get<IListDecision>(`/chats/${chatId}/decisions`, {
            params: {
                limit,
                after_id: afterId,
            }
        })
    }

}

