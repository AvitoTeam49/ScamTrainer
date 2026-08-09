import {makeAutoObservable} from "mobx";
import type {LeaderBoardResponse, ProfileResponse, ProgressResponse} from "../types/types.tsx";
import UsersService from "../services/UsersService.ts";
import axios from "axios";

export default class User {
    user: ProfileResponse | null = null;
    progress: ProgressResponse | null = null;
    leaderboard: LeaderBoardResponse | null = null;

    constructor() {
        makeAutoObservable(this)
    }

    setUser(user: ProfileResponse) {
        this.user = user;
    }

    setProgress(progress: ProgressResponse) {
        this.progress = progress;
    }

    setLeaderBoard(leaderboard: LeaderBoardResponse) {
        this.leaderboard = leaderboard;
    }

    async createProfile (username: string): Promise<{success: boolean, status?: number}> {
        try{
            const res = await UsersService.createProfile(username);
            this.setUser(res.data)
            return {success: true}
        }catch(e){
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined}
        }
    }

    async getProfile(): Promise<{success: boolean, status?: number}>{
        try{
            const res = await UsersService.getUser()
            this.setUser(res.data)
            return {success: true}
        }catch(e){
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined}
        }
    }

    async getProgress(): Promise<{success: boolean, status?: number}>{
        try{
            const res = await UsersService.getProgressUser()
            this.setProgress(res.data)
            return {success: true}
        }catch(e){
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined}
        }
    }

    async getLeaderboard(): Promise<{success: boolean, status?: number}>{
        try{
            const res = await UsersService.getLeaderboard()
            this.setLeaderBoard(res.data)
            return {success: true}
        }catch(e){
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined}
        }
    }

}