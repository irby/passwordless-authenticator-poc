import { Injectable } from "@angular/core";
import { environment } from "src/environments/environment";
import { ServiceResponse } from "../models/service-response.interface";
import { BaseService } from "./service.base";

@Injectable()
export class UserService extends BaseService {
  private usersBase = "users";
  private sharesBase = "shares";
  public async getMe(): Promise<ServiceResponse<void>> {
    return await this.getAsync(`me`);
  }
  public async getAccountSharingOverview(): Promise<
    ServiceResponse<GetAccountSharingOverviewResponseDto>
  > {
    return await this.getAsync(`${this.usersBase}/${this.sharesBase}/overview`);
  }
  public async getAccountSharesAsGuest(): Promise<
    ServiceResponse<GetAccountSharesResponseDto[]>
  > {
    return await this.getAsync(`${this.usersBase}/${this.sharesBase}/guest`);
  }
  public async getAccountSharesAsParent(): Promise<
    ServiceResponse<GetAccountSharesResponseDto[]>
  > {
    return await this.getAsync(`${this.usersBase}/${this.sharesBase}/parent`);
  }
  public async initiateLoginAsGuest(
    request: UserGuestRelationRequest
  ): Promise<ServiceResponse<void>> {
    return await this.postAsync(`login/guest`, request);
  }
  public async removeAccessToRelation(
    relationId: string
  ): Promise<ServiceResponse<void>> {
    return await this.deleteAsync(
      `${this.usersBase}/${this.sharesBase}/${relationId}`
    );
  }
}

export interface GetAccountSharingOverviewResponseDto {
  hasGuestGrants: boolean;
  hasParentGrants: boolean;
}

export interface GetAccountSharesResponseDto {
  relationId: string;
  guestUserId: string;
  guestUserEmail: string;
  parentUserId: string;
  parentUserEmail: string;
  createdAt: Date;
  isActive: boolean;
}

export interface UserGuestRelationRequest {
  relationId: string;
}
