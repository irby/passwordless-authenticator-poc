import { Component, OnInit } from "@angular/core";
import { MatDialogRef } from "@angular/material/dialog";
import {
  GetAccountSharesResponseDto,
  UserGuestRelationRequest,
  UserService,
} from "src/app/core/services/user.service";

@Component({
  selector: "app-grants-parent-modal",
  templateUrl: "./grants-parent-modal.component.html",
  styleUrls: ["./grants-parent-modal.component.css"],
})
export class GrantsParentModalComponent implements OnInit {
  public isLoading: boolean = false;
  public data: GetAccountSharesResponseDto[] = [];

  constructor(
    private readonly dialogRef: MatDialogRef<GrantsParentModalComponent>,
    private readonly userService: UserService
  ) {}

  async ngOnInit() {
    this.isLoading = true;
    const resp = await this.userService.getAccountSharesAsParent();
    this.isLoading = false;

    if (resp.type === "data") {
      this.data = resp.data;
    }
  }

  async revokeAccessOnRelation(relationId: string) {
    const confirmation = confirm("Please provide a biometric");
    if (!confirmation) return;

    const resp = await this.userService.removeAccessToRelation(relationId);
    if (resp.type === "data") {
      window.location.reload();
    }
  }
}
