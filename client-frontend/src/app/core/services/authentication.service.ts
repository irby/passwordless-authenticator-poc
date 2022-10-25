import { Injectable } from '@angular/core';
import axios from 'axios';
import { environment } from 'src/environments/environment';
import { ServiceResponse } from '../models/service-response.interface';
import { WebAuthnLoginFinalizeRequest } from '../models/webauthn/webauthn-login-finalize-request.interface';
import { WebauthnLoginInitializeResponse } from '../models/webauthn/webauthn-login-initialize-response.interface';
import { BaseService } from './service.base';

@Injectable()
export class AuthenticationService extends BaseService {

    private readonly userCacheKey: string = 'user';

    public async setLogin(): Promise<void> {
        const email = await this.getAndSetUser();
        localStorage.setItem(this.userCacheKey, email);
    }

    public async getUser(): Promise<string> {
        const user = localStorage.getItem(this.userCacheKey);
        if (user) {
            return user;
        }
        try {
            return await this.getAndSetUser();
        } catch (e) {
            return "";
        }
    }

    public async logout(): Promise<void> {
        localStorage.removeItem(this.userCacheKey);
        await this.postAsync(
            `users/logout`,
            { }
        );
    }

    public async beginFakeWebauthnLogin(userId: string): Promise<ServiceResponse<WebauthnLoginInitializeResponse>> {
        return await this.postAsync(`webauthn/login/initialize`, {"user_id": userId});
    }

    public async finalizeFakeWebauthnLogin(request: WebAuthnLoginFinalizeRequest): Promise<ServiceResponse<any>> {
        return await this.postAsync(`webauthn/login/finalize`, request);
    }

    private async getAndSetUser() : Promise<string> {
        const userId = await this.getAsync<any>(`me`);
        if (userId.type !== 'data') {
            return "";
        }
        const email = userId.data.email;
        localStorage.setItem(this.userCacheKey, email);
        return email;
    }
}
