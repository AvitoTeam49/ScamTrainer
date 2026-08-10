export interface IChat{
    id: number;
    user_id: number;
    scenario_id: number;
    session_id: string;
    title: string;
    status: string;
    resume: string;
    score: number;
    current_node_id: string;
    created_at: string;
    finished_at: string;
}

export interface IListChats{
    items: IChat[];
    next_after_id: number | null;
}

export interface IMessage{
    id: number;
    chat_id: number;
    sender_type: string;
    content: string;
    created_at: string;
}

export interface IListMessage{
    items: IMessage[];
    next_after_id: number | null;
}

export interface IDecision{
    id: number;
    chat_id: number;
    node_id: string;
    transition_id: string;
    target_node_id: string;
    score_delta: number;
    feedback: string;
    created_at: string;
}

export interface IListDecision{
    items: IDecision[];
    next_after_id: number | null;
}

export interface IScenario{
    id: number;
    title: string;
    role: string;
    difficulty: number;
}

export interface IListScenario{
    items: IScenario[];
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