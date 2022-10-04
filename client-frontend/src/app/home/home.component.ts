import { Component, OnInit } from '@angular/core';
import { MatDialog, MatDialogConfig } from '@angular/material/dialog';
import { Router } from '@angular/router';
import { AuthenticationService } from '../core/services/authentication.service';
import { AccountSharingInitializationDialog } from './account-sharing-initialization/account-sharing-initialization.component';

@Component({
  selector: 'app-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.css']
})
export class HomeComponent implements OnInit {

  constructor(
    private readonly router: Router, 
    private readonly matDialog: MatDialog) { }

  async ngOnInit() {
    if (!(await AuthenticationService.isAuthenticated())) {
      this.router.navigate(['']);
    }
  }

  public async logout() {
    await AuthenticationService.logout();
    this.router.navigate(['']);
  }

  public async openShareDialog() {
    this.matDialog.open(AccountSharingInitializationDialog, {
      width: '45em',
      height: '30em'
    });
  }

}
