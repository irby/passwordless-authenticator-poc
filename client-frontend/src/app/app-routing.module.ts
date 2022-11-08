import { NgModule } from "@angular/core";
import { RouterModule, Routes } from "@angular/router";
import { AuthenticationGuard } from "./core/guards/authentication.guard";
import { DefaultComponent } from "./default/default.component";
import { HomeComponent } from "./home/home.component";
import { LoginComponent } from "./login/login.component";
import { ShareComponent } from "./share/share.component";

const routes: Routes = [
  { path: "", component: DefaultComponent },
  { path: "login", component: LoginComponent },
  {
    path: "home",
    component: HomeComponent,
    canActivate: [AuthenticationGuard],
  },
  {
    path: "share/:id",
    component: ShareComponent,
    canActivate: [AuthenticationGuard],
  },
];

@NgModule({
  imports: [RouterModule.forRoot(routes, { useHash: true })],
  exports: [RouterModule],
})
export class AppRoutingModule {}
