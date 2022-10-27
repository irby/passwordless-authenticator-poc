import { Component, OnInit } from '@angular/core';
import { MatDialog, MatDialogConfig } from '@angular/material/dialog';
import { Router } from '@angular/router';
import { GrantsContext } from '../core/enums/grants-context.enum';
import { AuthenticationService, User } from '../core/services/authentication.service';
import { UserService } from '../core/services/user.service';
import { AccountSharingInitializationDialog } from './account-sharing-initialization/account-sharing-initialization.component';
import { AdminUserListModalComponent } from './admin-user-list-modal/admin-user-list-modal.component';
import { GrantsGuestModalComponent } from './grants-guest-modal/grants-guest-modal.component';
import { GrantsParentModalComponent } from './grants-parent-modal/grants-parent-modal.component';

@Component({
  selector: 'app-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.css']
})
export class HomeComponent implements OnInit {

  public user!: User;

  constructor(
    private readonly router: Router, 
    private readonly matDialog: MatDialog,
    private readonly authenticationService: AuthenticationService,
    private readonly userService: UserService) { }

    public hasGuestGrants: boolean = false;
    public hasParentGrants: boolean = false;
    public GrantsContext = GrantsContext;

  async ngOnInit() {
    this.authenticationService.getUserAsObservable().subscribe(data => {
      this.user = data;
    });
    await this.authenticationService.setLogin();
    const relationsOverviewResponse = await this.userService.getAccountSharingOverview();
    if (relationsOverviewResponse.type === 'data') {
      this.hasGuestGrants = relationsOverviewResponse.data.hasGuestGrants;
      this.hasParentGrants = relationsOverviewResponse.data.hasParentGrants;
    }
  }

  public async logout() {
    await this.authenticationService.logout();
    this.router.navigate(['']);
  }

  public async openShareDialog() {
    this.matDialog.open(AccountSharingInitializationDialog, {
      width: '45em',
      height: '30em'
    });
  }

  public async openGrantsDialog(context: GrantsContext) {
    const config: MatDialogConfig = {
      width: '45em',
      height: '30em'
    }
    switch (context) {
      case GrantsContext.Guest:
        this.matDialog.open(GrantsGuestModalComponent, config);
        break;
      case GrantsContext.Parent:
        this.matDialog.open(GrantsParentModalComponent, config);
        break;
    }
  }

  public openUserListDialog() {
    const config: MatDialogConfig = {
    }
    this.matDialog.open(AdminUserListModalComponent, config);
  }

}
