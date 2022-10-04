import { Injectable } from "@angular/core";
import { ActivatedRouteSnapshot, CanActivate, Router, RouterStateSnapshot, UrlTree } from "@angular/router";
import { Observable } from "rxjs";
import { AuthenticationService } from "../services/authentication.service";

@Injectable({ providedIn: 'root' })
export class AuthenticationGuard implements CanActivate {
    constructor(protected router: Router) {}
    
    public canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): boolean | UrlTree | Observable<boolean | UrlTree> | Promise<boolean | UrlTree> {
        return AuthenticationService.isAuthenticated().then((result) => {
            if (result) {
                return true;
            }
            this.router.navigate(['/login'], { queryParams: { redirect: state.url }, replaceUrl: true });
            return false;
        });
    }
}