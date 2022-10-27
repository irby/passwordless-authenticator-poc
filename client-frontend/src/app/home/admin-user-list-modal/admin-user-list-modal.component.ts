import { Component, OnInit } from '@angular/core';
import { MatDialog, MatDialogConfig } from '@angular/material/dialog';
import { AdminService, UserDto } from 'src/app/core/services/admin.service';
import { NotificationService } from 'src/app/core/services/notification.service';
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

  public openLoginAuditRecords(userId: string) {
    const matDialogConfig: MatDialogConfig = {
      width: '60em',
      data: userId
    };
    this.matDialog.open(AdminUserLoginAuditComponent, matDialogConfig);
  }
}
