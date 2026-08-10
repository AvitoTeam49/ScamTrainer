import {makeAutoObservable} from "mobx";
import AuthService from "../services/AuthService.ts";
import axios from "axios";

export default class Auth {
    isAuth = false

    constructor() {
        makeAutoObservable(this)
    }

    setAuth(bool: boolean) {
        this.isAuth = bool;
    }

    async login(email: string, password: string): Promise<{success: boolean, status?: number}> {
        try{
            const response = await AuthService.login(email, password);
            localStorage.setItem("token", response.data.access_token);
            this.setAuth(true);
            return {success: true};

        }catch(e){
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined};
        }
    }

    async registration(email: string, password: string): Promise<{success: boolean, status?: number}> {
        try{
            await AuthService.registration(email, password);
            return {success: true};
        }catch(e){
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined};
        }
    }

    async checkAuth(): Promise<{success: boolean, status?: number}> {
        try{
            const response = await AuthService.checkAuth();
            if (response.data.is_valid){
                this.setAuth(true);
                return {success: true};
            }else{
                this.setAuth(false)
                return {success: false};
            }
        }catch(e){
            this.setAuth(false)
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined};
        }
    }

}