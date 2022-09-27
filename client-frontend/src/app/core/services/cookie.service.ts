import { Injectable } from '@angular/core';

@Injectable()
export class CookieService {

    constructor() { }

    public static getCookie(): string {
        let ca: Array<string> = document.cookie.split(';');
        console.log(document.cookie);
        let caLen: number = ca.length;
        let cookieName = `hanko=`;
        let c: string;
        console.log('ca', ca);

        for (let i: number = 0; i < caLen; i += 1) {
            c = ca[i].replace(/^\s+/g, '');
            if (c.indexOf(cookieName) == 0) {
                return c.substring(cookieName.length, c.length);
            }
        }
        return '';
    }

    public static deleteCookie() {
        this.setCookie({ name: 'hanko', value: '', expireDays: -1 });
    }

    private static setCookie(params: any) {
        let d: Date = new Date();
        d.setTime(
            d.getTime() +
            (params.expireDays ? params.expireDays : 1) * 24 * 60 * 60 * 1000
        );
        document.cookie =
            (params.name ? params.name : '') +
            '=' +
            (params.value ? params.value : '') +
            ';' +
            (params.session && params.session == true
                ? ''
                : 'expires=' + d.toUTCString() + ';') +
            'path=' +
            (params.path && params.path.length > 0 ? params.path : '/') +
            ';' +
            (location.protocol === 'https:' && params.secure && params.secure == true
                ? 'secure'
                : '');
    }


}
