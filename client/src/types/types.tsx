export interface IChat{
    id: number;
    role: string;
    difficulty: string;
}

export interface IMessage{
    content: string;
    who: string;
    time: string;
}

export interface LoginResponse{
    access_token: string;
}

export interface RegistrationResponse{
    user_id: number;
    message: string;
}

export interface Validate{
    is_valid: boolean;
    user_id: number;
}