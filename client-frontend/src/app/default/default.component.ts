import { Component, OnInit } from "@angular/core";
import { Router } from "@angular/router";
import { AuthenticationService } from "../core/services/authentication.service";
import { UserService } from "../core/services/user.service";

@Component({
  selector: "app-default",
  templateUrl: "./default.component.html",
  styleUrls: ["./default.component.css"],
})
export class DefaultComponent implements OnInit {
  constructor(
    private readonly router: Router,
    private readonly userService: UserService
  ) {}

  async ngOnInit() {
    const user = await this.userService.getMe();
    if (user.type === "data") {
      this.router.navigate(["/home"], { replaceUrl: true });
      return;
    }

    this.router.navigate(["/login"], { replaceUrl: true });
  }
}
