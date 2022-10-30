import { Component, Inject, OnInit } from '@angular/core';
import { MAT_DIALOG_DATA } from '@angular/material/dialog';
import { NullInt, NullInt32 } from '../../../core/models/null-int';
import { UserModalInfo } from '../../../core/models/user-modal-info.interface';
import { AdminService, UserGuestRelationshipDto } from '../../../core/services/admin.service';
import { NotificationService } from '../../../core/services/notification.service';

@Component({
  selector: 'app-admin-user-grants',
  templateUrl: './admin-user-grants.component.html',
  styleUrls: ['./admin-user-grants.component.css']
})
export class AdminUserGrantsComponent implements OnInit {

  public userEmail!: string;
  public isLoading: boolean = false;
  public userGrants: UserGuestRelationshipDto[] = [];

  constructor(@Inject(MAT_DIALOG_DATA) private readonly user: UserModalInfo, private readonly adminService: AdminService, private readonly notificationService: NotificationService) { }

  async ngOnInit() {
    this.isLoading = true;
    const grantResp = await this.adminService.getGrantsForUser(this.user.userId);
    if (grantResp.type !== 'data') {
      this.notificationService.error("An error occurred fetching grants for user");
      return;
    }
    this.isLoading = false;
    this.userEmail = this.user.userEmail;
    this.userGrants.push(... grantResp.data.grants);
  }

  public getValueOrDefault(value: NullInt32) {
    const converted = new NullInt(value.Int32, value.Valid);
    return converted.getValueOrDefault();
  }

}
