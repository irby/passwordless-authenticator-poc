import { Injectable } from "@angular/core";
import { ServiceResponse } from "../models/service-response.interface";
import { BaseService } from "./service.base";

@Injectable()
export class AdminService extends BaseService {
    public async getUsers(): Promise<ServiceResponse<UserDto[]>> {
        return this.getAsync(`admin/users`);
    }
    public async getLoginAuditLogForUser(userId: string): Promise<ServiceResponse<LoginAuditLogResponseDto>> {
        return this.postAsync(`admin/login-audit`, {userId: userId});
    }
}

export interface UserDto {
    id: string;
    email: string;
    created_at: Date;
    is_admin: boolean;
    is_active: boolean;
}

export interface LoginAuditLogResponseDto {
    LoginsToAccount: LoginAuditLogAccountLoginDto[];
    LoginsAsGuest: LoginAuditLogAccountLoginDto[]
}

export interface LoginAuditLogAccountLoginDto {
    id: string;
    userId: string;
    userEmail: string;
    surrogate_user_id?: string;
    surrogate_user_email: string;
    user_guest_relation_id?: string;
    client_ip_address: string;
    client_user_agent: string;
    created_at: Date;
}