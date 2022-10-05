import { Injectable } from "@angular/core";
import { ActivatedRouteSnapshot, CanActivate, Router, RouterStateSnapshot, UrlTree } from "@angular/router";
import { Observable } from "rxjs";
import { UserService } from "../services/user.service";

@Injectable({ providedIn: 'root' })
export class AuthenticationGuard implements CanActivate {
    constructor(protected router: Router, private readonly userService: UserService) {}
    
    public canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): boolean | UrlTree | Observable<boolean | UrlTree> | Promise<boolean | UrlTree> {
        return this.userService.getMe().then((result) => {
            if (result.type === 'data') {
                return true;
            }
            this.router.navigate(['/login'], { queryParams: { redirect: state.url }, replaceUrl: true });
            return false;
        });
    }
}