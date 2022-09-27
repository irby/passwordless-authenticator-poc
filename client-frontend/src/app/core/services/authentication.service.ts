import { Injectable } from '@angular/core';
import axios from 'axios';
import { environment } from 'src/environments/environment';
import { CookieService } from './cookie.service';

@Injectable()
export class AuthenticationService {
    public static async isAuthenticated(): Promise<boolean> {
        try {
            await axios.get(`${environment.hankoApiUrl}/me`, { withCredentials: true });
            return true;
        }
        catch (e) {
            return false;
        }
    }
    public static async logout(): Promise<void> {
        await axios.post(
            `${environment.hankoApiUrl}/users/logout`,
            { },
            { withCredentials: true }
        );
    }
}
