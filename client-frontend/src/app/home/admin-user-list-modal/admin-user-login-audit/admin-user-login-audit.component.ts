import { Component, Inject, OnInit } from '@angular/core';
import { MAT_DIALOG_DATA } from '@angular/material/dialog';
import { AdminService, LoginAuditLogResponseDto } from 'src/app/core/services/admin.service';
import { NotificationService } from 'src/app/core/services/notification.service';

@Component({
  selector: 'app-admin-user-login-audit',
  templateUrl: './admin-user-login-audit.component.html',
  styleUrls: ['./admin-user-login-audit.component.css']
})
export class AdminUserLoginAuditComponent implements OnInit {

  public loginAuditRecords!: LoginAuditLogResponseDto;
  public isLoading: boolean = false;

  constructor(@Inject(MAT_DIALOG_DATA) private readonly userId: string, private readonly adminService: AdminService, private readonly notificationService: NotificationService) { }

  async ngOnInit() {
    this.isLoading = true;
    const auditLogResp = await this.adminService.getLoginAuditLogForUser(this.userId);
    this.isLoading = false;

    if (auditLogResp.type !== 'data') {
      this.notificationService.error('failed to retrieve login audits for user');
      return;
    }
    console.log(auditLogResp.data);
    this.loginAuditRecords = auditLogResp.data;
  }

  public cleanIpAddress(ipAddress: string) {
    if (ipAddress?.includes(':')) {
      return ipAddress.substring(0, ipAddress.indexOf(':'));
    }
    return ipAddress;
  }

  public cleanBrowser(userAgent: string) {
    if (userAgent?.length > 15) {
      return userAgent.substring(0, 15) + '...';
    }
    return userAgent;
  }

}
