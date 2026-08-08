import {makeAutoObservable} from "mobx";
import AuthService from "../services/AuthService.ts";
import type {RegistrationResponse} from "../types/types.tsx";

export default class Auth {
    isAuth = false
    user = {} as RegistrationResponse;

    constructor() {
        makeAutoObservable(this)
    }

    setAuth(bool: boolean) {
        this.isAuth = bool;
    }

    setUser (user: RegistrationResponse) {
        this.user = user;
    }

    async login(email: string, password: string): Promise<{success: boolean, status?: number}> {
        try{
            const response = await AuthService.login(email, password);
            localStorage.setItem("token", response.data.access_token);
            this.setAuth(true);
            return {success: true};

        }catch(e){
            console.error(e.response?.data?.error);
            return {success: false, status: e.response?.status};
        }
    }

    async registration(email: string, password: string): Promise<{success: boolean, status?: number}> {
        try{
            const response = await AuthService.registration(email, password);
            const res = await this.login(email, password);
            this.setUser(response.data)
            return res
        }catch(e){
            console.error(e.response?.data?.error);
            return {success: false, status: e.response?.status};
        }
    }

    async checkAuth(): Promise<boolean> {
        try{
            const response = await AuthService.checkAuth();
            if (response.data.is_valid){
                this.setAuth(true);
                return true;
            }else{
                this.setAuth(false)
                return false;
            }
        }catch(e){
            console.error(e.response?.data?.error);
            this.setAuth(false)
            return false;
        }
    }

}