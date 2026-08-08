import $api from "../http";
import type {AxiosResponse} from "axios";
import type {LoginResponse, RegistrationResponse, Validate} from "../types/types.tsx";

export default class AuthService {

    static async login(email: string, password: string): Promise<AxiosResponse<LoginResponse>> {
        return $api.post<LoginResponse>("/auth/login", {email, password});
    }

    static async registration(email: string, password: string): Promise<AxiosResponse<RegistrationResponse>>{
        return $api.post<RegistrationResponse>("/auth/register", {email, password});
    }

    static async checkAuth(): Promise<AxiosResponse<Validate>> {
        return $api.post<Validate>("/auth/validate");
    }


}

