import { Injectable } from '@angular/core';
import { Observable, Subject } from 'rxjs';
import { ServiceResponse } from '../models/service-response.interface';
import { WebAuthnLoginFinalizeRequest } from '../models/webauthn/webauthn-login-finalize-request.interface';
import { WebauthnLoginInitializeResponse } from '../models/webauthn/webauthn-login-initialize-response.interface';
import { BaseService } from './service.base';

@Injectable()
export class AuthenticationService extends BaseService {

    private readonly userCacheKey: string = 'user';
    private userSubject = new Subject<User>();

    public async setLogin(): Promise<void> {
        const user = await this.getAndSetUser();
        localStorage.setItem(this.userCacheKey, user.email);
        this.userSubject.next(user);
    }

    public getUserAsObservable(): Observable<User> {
        return this.userSubject.asObservable();
    }

    public async logout(): Promise<void> {
        localStorage.removeItem(this.userCacheKey);
        await this.postAsync(
            `users/logout`,
            { }
        );
    }

    public async logoutAsGuest(): Promise<void> {
        localStorage.removeItem(this.userCacheKey);
        await this.postAsync(
            `users/logout-guest`,
            { }
        );
    }

    public async beginWebauthnLogin(userId: string): Promise<ServiceResponse<WebauthnLoginInitializeResponse>> {
        return await this.postAsync(`webauthn/login/initialize`, {"user_id": userId});
    }

    public async finalizeWebauthnLogin(request: WebAuthnLoginFinalizeRequest): Promise<ServiceResponse<any>> {
        return await this.postAsync(`webauthn/login/finalize`, request);
    }

    private async getAndSetUser() : Promise<User> {
        const userId = await this.getAsync<GetMeResponse>(`me`);
        if (userId.type !== 'data') {
            throw new Error("Unable to fetch user");
        }
        const email = userId.data.email;
        localStorage.setItem(this.userCacheKey, email);
        return {
            id: userId.data.id,
            email: userId.data.email,
            isAccountHolder: userId.data.isAccountHolder,
            isAdmin: userId.data.isAdmin
        };
    }
}

export interface GetMeResponse {
    email: string;
    id: string;
    isAccountHolder: boolean;
    isAdmin: boolean;
}

export interface User {
    id: string;
    email: string;
    isAccountHolder: boolean;
    isAdmin: boolean;
}