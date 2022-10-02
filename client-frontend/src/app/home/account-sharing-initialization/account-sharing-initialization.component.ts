import { Component, Inject, OnInit } from '@angular/core';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';

@Component({
  selector: 'app-account-sharing-initialization',
  templateUrl: './account-sharing-initialization.component.html',
  styleUrls: ['./account-sharing-initialization.component.css']
})
export class AccountSharingInitializationDialog implements OnInit {

  constructor(
    private readonly dialogRef: MatDialogRef<AccountSharingInitializationDialog>
  ) { }

  ngOnInit() {
  }

  public close() {
    this.dialogRef.close();
  }

  public async submit() {

  }

}
