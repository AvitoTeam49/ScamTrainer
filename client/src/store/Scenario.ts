import type {IScenario} from "../types/types.tsx";
import {makeAutoObservable} from "mobx";
import axios from "axios";
import ScenariosService from "../services/ScenariosService.ts";


class Scenario {
    scenarios: IScenario[] = []

    constructor() {
        makeAutoObservable(this)
    }

    serScenarios(scenarios: IScenario[]) {
        this.scenarios = scenarios;
    }

    async getScenarios(difficulty: number): Promise<{success: boolean, scenarios?: IScenario[] ,status?: number}> {
        try{
            const response = await ScenariosService.getScenarios(difficulty);
            this.serScenarios(response.data.items);
            return {success: true, scenarios: response.data.items};
        }catch(e){
            return {success: false, status: axios.isAxiosError(e) ? e.response?.status : undefined}
        }
    }


}

export default Scenario;