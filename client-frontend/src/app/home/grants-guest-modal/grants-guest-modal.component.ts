import { Component, OnInit } from "@angular/core";
import { MatDialogRef } from "@angular/material/dialog";
import {
  GetAccountSharesResponseDto,
  UserGuestRelationRequest,
  UserService,
} from "src/app/core/services/user.service";

@Component({
  selector: "app-grants-guest-modal",
  templateUrl: "./grants-guest-modal.component.html",
  styleUrls: ["./grants-guest-modal.component.css"],
})
export class GrantsGuestModalComponent implements OnInit {
  public isLoading: boolean = false;
  public data: GetAccountSharesResponseDto[] = [];

  constructor(
    private readonly dialogRef: MatDialogRef<GrantsGuestModalComponent>,
    private readonly userService: UserService
  ) {}

  async ngOnInit() {
    this.isLoading = true;
    const resp = await this.userService.getAccountSharesAsGuest();
    this.isLoading = false;

    if (resp.type === "data") {
      this.data = resp.data;
    }
  }

  public async initiateLogin(relationId: string) {
    const confirmation = confirm("Please provide a biometric");
    if (!confirmation) return;

    const request: UserGuestRelationRequest = {
      relationId: relationId,
    };

    this.isLoading = true;
    const resp = await this.userService.initiateLoginAsGuest(request);
    this.isLoading = false;

    if (resp.type === "data") {
      // TODO: Clean this up, hacky way to refresh the user object
      localStorage.removeItem("user");
      window.location.reload();
    }
  }
}
