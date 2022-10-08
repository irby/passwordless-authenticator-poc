import { Injectable } from "@angular/core";
import { environment } from "src/environments/environment";
import { ServiceResponse } from "../models/service-response.interface";
import { BaseService } from "./service.base";

@Injectable()
export class UserService extends BaseService {
    public async getMe(): Promise<ServiceResponse<void>> {
        return await this.getAsync(`me`);
    }
}