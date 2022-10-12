import { Injectable } from "@angular/core";
import { environment } from "src/environments/environment";
import { ServiceResponse } from "../models/service-response.interface";
import { BaseService } from "./service.base";

@Injectable()
export class UserService extends BaseService {
    private usersBase = 'users';
    private sharesBase = 'shares';
    public async getMe(): Promise<ServiceResponse<void>> {
        return await this.getAsync(`me`);
    }
    public async getAccountSharingOverview(): Promise<ServiceResponse<GetAccountSharingOverviewResponseDto>> {
        return await this.getAsync(`${this.usersBase}/${this.sharesBase}/overview`)
    }
    public async getAccountSharesAsGuest(): Promise<ServiceResponse<GetAccountSharesResponseDto[]>> {
        return await this.getAsync(`${this.usersBase}/${this.sharesBase}/guest`)
    }
    public async getAccountSharesAsParent(): Promise<ServiceResponse<GetAccountSharesResponseDto[]>> {
        return await this.getAsync(`${this.usersBase}/${this.sharesBase}/parent`)
    }
}

export interface GetAccountSharingOverviewResponseDto {
    hasGuestGrants: boolean;
    hasParentGrants: boolean;
}

export interface GetAccountSharesResponseDto {
    id: string;
    guestUserId: string;
    parentUserId: string;
    createdAt: Date;
    updatedAt: Date;
    isActive: boolean;
}