import { Injectable } from "@angular/core";
import { ServiceResponse } from "../models/service-response.interface";
import { WebAuthnLoginFinalizeRequest } from "../models/webauthn/webauthn-login-finalize-request.interface";
import {
  CreateAccountWithGrantResponse,
  WebauthnLoginInitializeResponse,
} from "../models/webauthn/webauthn-login-initialize-response.interface";
import { BaseService } from "./service.base";

@Injectable()
export class GrantService extends BaseService {
  public async createGrant(
    dto: CreateAccountGrantDto
  ): Promise<ServiceResponse<CreateAccountGrantResponseDto>> {
    return await this.postAsync(`access/share/initialize`, dto);
  }

  public async initializeCreateAccountWithGrant(
    request: BeginCreateAccountWithGrantRequest
  ): Promise<ServiceResponse<CreateAccountWithGrantResponse>> {
    return await this.postAsync(
      `access/share/begin-create-account-with-grant`,
      request
    );
  }

  public async finishCreateAccountWithGrant(
    request: FinishCreateAccountWithGrantRequest
  ): Promise<ServiceResponse<any>> {
    return await this.postAsync(
      `access/share/finish-create-account-with-grant`,
      request
    );
  }

  public async getGrantByIdAndToken(
    id: string,
    token: string
  ): Promise<ServiceResponse<void>> {
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

export interface CreateAccountGrantResponseDto {
  url: string;
}

export interface BeginCreateAccountWithGrantRequest {
  guestUserId: string;
  grantId: string;
}

export interface FinishCreateAccountWithGrantRequest
  extends WebAuthnLoginFinalizeRequest {
  guestUserId: string;
  grantId: string;
  grantAttestation?: string;
}
