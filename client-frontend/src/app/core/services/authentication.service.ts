import { Injectable } from '@angular/core';
import axios from 'axios';
import { Observable } from 'rxjs';
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
    
    public static async getUser(): Promise<string> {
        const user = localStorage.getItem('user');
        if (user) {
            return user;
        }
        try {
            const userId = await axios.get(`${environment.hankoApiUrl}/me`, { withCredentials: true });
            console.log('userId', userId);
            const userData = await axios.get(`${environment.hankoApiUrl}/users/${userId.data.id}`, { withCredentials: true });
            console.log('user', userData);
            const email = userData.data.email;
            localStorage.setItem('user', email);
            return email;
        } catch (e) {
            return "";
        }
    }

    public static async logout(): Promise<void> {
        localStorage.removeItem('user');
        await axios.post(
            `${environment.hankoApiUrl}/users/logout`,
            { },
            { withCredentials: true }
        );
    }
}
