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

export interface ProfileResponse{
    id: number;
    username: string;
    score: number;
    created_at: string;
    updated_at: string;
}

export interface ProgressResponse{
    user_id: number;
    scenarios_completed: number;
    scams_detected: number;
    failed_attempts: number;
}

export interface ITopUser{
    rank: number;
    user_id: number;
    username: string;
    score: number;
}

export interface LeaderBoardResponse{
    top_users: ITopUser[];
    user_rank: number;
}