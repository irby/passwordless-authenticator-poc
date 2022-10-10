import { Injectable } from "@angular/core";
import { ServiceResponse } from "../models/service-response.interface";
import { BaseService } from "./service.base";

@Injectable()
export class GrantService extends BaseService {
    public async createGrant(dto : CreateAccountGrantDto) : Promise<ServiceResponse<void>> {
        return await this.postAsync(`access/share/initialize`, dto);
    }

    public async getGrantByIdAndToken(id: string, token: string) : Promise<ServiceResponse<void>> {
        return await this.getAsync(`access/share/grant/${id}?token=${token}`);
    }
}

export interface CreateAccountGrantDto {
    email: string;
    expireByLogin?: boolean;
    loginsAllowed?: number;
    expireByTime?: boolean;
    minutesAllowed?: number;
}