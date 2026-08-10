import { makeAutoObservable, runInAction } from "mobx";
import type {LeaderBoardResponse, ProfileResponse, ProgressResponse} from "../types/types.tsx";
import UsersService from "../services/UsersService.ts";
import axios from "axios";

export default class User {
    user: ProfileResponse | null = null;
    progress: ProgressResponse | null = null;
    leaderboard: LeaderBoardResponse | null = null;

    constructor() {
        makeAutoObservable(this);
    }


    async createProfile(username: string): Promise<{ success: boolean; status?: number; }> {
        try {
            const response = await UsersService.createProfile(username);

            runInAction(() => {
                this.user = response.data;
            });

            return {success: true};
        } catch (e) {
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined};
        }
    }

    async getProfile(): Promise<{ success: boolean; status?: number; }> {
        try {
            const response = await UsersService.getUser();

            runInAction(() => {
                this.user = response.data;
            });

            return {success: true};
        } catch (e) {
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined};
        }
    }

    async getProgress(): Promise<{ success: boolean; status?: number; }> {
        try {
            const response = await UsersService.getProgressUser();

            runInAction(() => {
                this.progress = response.data;
            });

            return {success: true};
        } catch (e) {
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined};
        }
    }

    async getLeaderboard(): Promise<{ success: boolean; status?: number; }> {
        try {
            const response = await UsersService.getLeaderboard();

            runInAction(() => {
                this.leaderboard = response.data;
            });

            return {success: true};
        } catch (e) {
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined};
        }
    }

    async refreshAfterChat() {
        await Promise.all([this.getProfile(), this.getProgress(), this.getLeaderboard()]);
    }
}