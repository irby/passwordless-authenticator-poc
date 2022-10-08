import { Component, Inject, OnInit } from '@angular/core';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import {FormControl, FormGroupDirective, NgForm, Validators} from '@angular/forms';
import {ErrorStateMatcher} from '@angular/material/core';
import axios from 'axios';
import { environment } from 'src/environments/environment';
import { CreateAccountGrantDto, GrantService } from 'src/app/core/services/grant.service';

/** Error when invalid control is dirty, touched, or submitted. */
export class MyErrorStateMatcher implements ErrorStateMatcher {
  isErrorState(control: FormControl | null, form: FormGroupDirective | NgForm | null): boolean {
    const isSubmitted = form && form.submitted;
    return !!(control && control.invalid && (control.dirty || control.touched || isSubmitted));
  }
}

@Component({
  selector: 'app-account-sharing-initialization',
  templateUrl: './account-sharing-initialization.component.html',
  styleUrls: ['./account-sharing-initialization.component.css']
})
export class AccountSharingInitializationDialog implements OnInit {

  emailFormControl = new FormControl('', [Validators.required, Validators.email]);
  loginCountFormControl = new FormControl(0);
  timeMinutesCountFormControl = new FormControl(0);

  matcher = new MyErrorStateMatcher();

  public expireByLogins: boolean = false;
  public loginsAllowed: number = 0;
  public expireByTime: boolean = false;
  public accessLifespanMinutes: Date = new Date();
  public isLoading: boolean = false;

  constructor(
    private readonly dialogRef: MatDialogRef<AccountSharingInitializationDialog>,
    private readonly grantService: GrantService
  ) { }

  ngOnInit() {
  }

  public close() {
    this.dialogRef.close();
  }

  public async submit() {
    this.isLoading = true;
    const dto : CreateAccountGrantDto = {
      email: this.emailFormControl.getRawValue() || ""
    };
    const resp = await this.grantService.createGrant(dto);
    this.isLoading = false;
    if (resp.type === 'error') {
      // TODO: Handle
    }
  }

  public toggleExpireByLogins() {
    this.expireByLogins = !this.expireByLogins;
    if (this.expireByLogins) {
      this.loginCountFormControl.setValidators(Validators.required);

      if (this.expireByTime) {
        this.toggleExpireByTime();
        this.timeMinutesCountFormControl.setValue(0);
      }
    } else {
      this.loginCountFormControl.setValidators([]);
    }
  }

  public toggleExpireByTime() {
    this.expireByTime = !this.expireByTime;
    if (this.expireByTime) {
      this.timeMinutesCountFormControl.setValidators(Validators.required);

      if (this.expireByLogins) {
        this.toggleExpireByLogins();
        this.loginCountFormControl.setValue(0);
      }
    } else {
      this.timeMinutesCountFormControl.setValidators([]);
    }
  }

}
