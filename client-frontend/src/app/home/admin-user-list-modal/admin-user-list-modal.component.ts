import { HttpStatusCode } from '@angular/common/http';
import { Component, OnInit } from '@angular/core';
import { MatDialog, MatDialogConfig } from '@angular/material/dialog';
import { UserModalInfo } from 'src/app/core/models/user-modal-info.interface';
import { AdminService, UserDto } from 'src/app/core/services/admin.service';
import { NotificationService } from 'src/app/core/services/notification.service';
import { AdminUserGrantsComponent } from './admin-user-grants/admin-user-grants.component';
import { AdminUserLoginAuditComponent } from './admin-user-login-audit/admin-user-login-audit.component';

@Component({
  selector: 'app-admin-user-list-modal',
  templateUrl: './admin-user-list-modal.component.html',
  styleUrls: ['./admin-user-list-modal.component.css']
})
export class AdminUserListModalComponent implements OnInit {
  
  public userList!: UserDto[];
  public isLoading: boolean = false;

  constructor(private readonly adminService: AdminService, private readonly notificationService: NotificationService, private readonly matDialog: MatDialog) { }

  async ngOnInit() {
    this.isLoading = true;
    const userListResp = await this.adminService.getUsers();
    this.isLoading = false;

    if (userListResp.type !== 'data') {
      this.notificationService.error('Unable to retrieve user list');
      return;
    }
    this.userList = userListResp.data;
  }

  public openLoginAuditRecords(user: UserDto) {
    const matDialogConfig: MatDialogConfig = {
      width: '60em',
      data: this.convertToUserModalInfo(user)
    };
    this.matDialog.open(AdminUserLoginAuditComponent, matDialogConfig);
  }

  public openActiveAccountGrants(user: UserDto) {
    const matDialogConfig: MatDialogConfig = {
      width: '60em',
      data: this.convertToUserModalInfo(user)
    };
    this.matDialog.open(AdminUserGrantsComponent, matDialogConfig);
  }

  public async toggleUserIsActive(user: UserDto) {
    const toggleResp = await this.adminService.toggleUserIsActive(user.id);
    if (toggleResp.type !== 'data') {
      switch (toggleResp.statusCode) {
        case HttpStatusCode.Conflict:
          this.notificationService.error("You know you can't deactivate yourself, right?");
          break;
        default:
          this.notificationService.error("An error occurred toggling user status");
      }
      
      return;
    }
    user.is_active = !user.is_active;
    const verb = user.is_active ? 'activated' : 'deactivated';
    this.notificationService.success(`User has been successfully ${verb}.`);
    console.log('here?');
  }

  public async deactivateGrants(user: UserDto) {
    const deactivationResponse = await this.adminService.deactivateGrantsForUser(user.id);
    if (deactivationResponse.type !== 'data') {
      this.notificationService.error("An error occurred while deactivating grants for user");
      return;
    }
    this.notificationService.success("Grants successfully deactivated for user");
  }

  private convertToUserModalInfo(user: UserDto): UserModalInfo {
    return {
      userId: user.id,
      userEmail: user.email
    }
  }
}
