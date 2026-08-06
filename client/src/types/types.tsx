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