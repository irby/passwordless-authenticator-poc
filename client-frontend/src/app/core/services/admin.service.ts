import { Injectable } from "@angular/core";
import { NullInt32 } from "../models/null-int";
import { ServiceResponse } from "../models/service-response.interface";
import { BaseService } from "./service.base";

@Injectable()
export class AdminService extends BaseService {
  public async getUsers(): Promise<ServiceResponse<UserDto[]>> {
    return await this.getAsync(`admin/users`);
  }
  public async getLoginAuditLogForUser(
    userId: string
  ): Promise<ServiceResponse<LoginAuditLogResponseDto>> {
    return await this.postAsync(`admin/login-audit`, { userId: userId });
  }
  public async getGrantsForUser(
    userId: string
  ): Promise<ServiceResponse<GetGrantsForUserResponse>> {
    return await this.getAsync(`admin/grants/${userId}`);
  }
  public async toggleUserIsActive(
    userId: string
  ): Promise<ServiceResponse<void>> {
    return await this.putAsync(`admin/users/active/${userId}`);
  }
  public async deactivateGrantsForUser(
    userId: string
  ): Promise<ServiceResponse<void>> {
    return await this.deleteAsync(`admin/grants/${userId}`);
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
  LoginsAsGuest: LoginAuditLogAccountLoginDto[];
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

export interface GetGrantsForUserResponse {
  userId: string;
  userEmail: string;
  grants: UserGuestRelationshipDto[];
}

export interface UserGuestRelationshipDto {
  guestUserEmail: string;
  guestUserId: string;
  createdAt: Date;
  isActive: boolean;
  loginRemaining: NullInt32;
  minutesRemaining: NullInt32;
}
