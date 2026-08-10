import $api from "../http";
import type {AxiosResponse} from "axios";
import type {IListScenario} from "../types/types.tsx";

export default class ScenariosService {

    static async getScenarios(difficulty: number): Promise<AxiosResponse<IListScenario>> {
        return $api.get<IListScenario>("/scenarios", {
            params: {difficulty},
        });
    }


}

