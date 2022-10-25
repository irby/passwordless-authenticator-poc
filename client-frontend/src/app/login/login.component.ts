import { Component, OnInit, Renderer2 } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { environment } from 'src/environments/environment';
import { GetUserNameFromId, UserId } from '../core/constants/user-constants';
import { AuthenticationService } from '../core/services/authentication.service';
import { ScriptService } from '../core/services/script.service';
import { ChallengeService } from '../core/services/challenge.service';
import { NotificationService } from '../core/services/notification.service';
import { MatDialog, MatDialogConfig } from '@angular/material/dialog';
import { ConfirmBiometricModalComponent } from '../core/modals/confirm-biometric-modal/confirm-biometric-modal.component';
// import { EccUtil } from '../core/utils/ecc-util';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css'],
  providers: [ ScriptService ]
})
export class LoginComponent implements OnInit {

  public hankoElementUrl = environment.hankoElementUrl;
  public hankoApiUrl = environment.hankoApiUrl;
  public UserId = UserId;

  public isHankoElementsLoaded = false;

  constructor(private renderer: Renderer2,
    private scriptService: ScriptService,
    private readonly router: Router,
    private readonly route: ActivatedRoute,
    private readonly authenticationSerivce: AuthenticationService,
    private readonly challengeService: ChallengeService,
    private readonly notificationService: NotificationService,
    private readonly matDialog: MatDialog) { }

  ngOnInit() {
    const scriptElement = this.scriptService.loadJsScript(this.renderer, `${environment.hankoElementUrl}/element.hanko-auth.js`);
    scriptElement.onload = () => {
      this.isHankoElementsLoaded = true;
    };
    scriptElement.onerror = () => {
      console.error('Error loading elements');
    }
  }

  public async redirectToIndex(e : any) {
    await this.authenticationSerivce.setLogin();
    this.router.navigate([this.route.snapshot.queryParams[`redirect`] || '/'], { replaceUrl: true });
  }

  public openConfirmBiometricDialog(userId: string) {
    const matDialogConfig: MatDialogConfig = {
      width: '45em',
      height: '20em',
      data: userId
    };
    this.matDialog.open(ConfirmBiometricModalComponent, matDialogConfig);
  }

}
