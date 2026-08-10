import type {AxiosResponse} from "axios";
import type {LeaderBoardResponse, ProfileResponse, ProgressResponse} from "../types/types.tsx";
import $api from "../http";

export default class UsersService {

   static async createProfile(username: string): Promise<AxiosResponse<ProfileResponse>> {
       return $api.post<ProfileResponse>("/users", {username});
   }

   static async getUser(): Promise<AxiosResponse<ProfileResponse>>{
       return $api.get<ProfileResponse>("/users/me");
   }

   static async getProgressUser(): Promise<AxiosResponse<ProgressResponse>> {
       return $api.get<ProgressResponse>("/users/me/progress");
   }

   static async getLeaderboard(): Promise<AxiosResponse<LeaderBoardResponse>>{
       return $api.get<LeaderBoardResponse>("/users/leaderboard");
   }

}

