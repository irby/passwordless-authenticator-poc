import { Injectable } from '@angular/core';
import axios from 'axios';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';

@Injectable()
export class AuthenticationService {

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
        await axios.post(
            `${environment.hankoApiUrl}/users/logout`,
            { },
            { withCredentials: true }
        );
    }

    private async getAndSetUser() : Promise<string> {
        const userId = await axios.get(`${environment.hankoApiUrl}/me`, { withCredentials: true });
        const userData = await axios.get(`${environment.hankoApiUrl}/users/${userId.data.id}`, { withCredentials: true });
        const email = userData.data.email;
        localStorage.setItem(this.userCacheKey, email);
        return email;
    }
}
