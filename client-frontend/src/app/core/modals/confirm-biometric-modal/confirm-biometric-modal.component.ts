import { Component, Inject, OnInit } from '@angular/core';
import { MatDialogConfig, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { ActivatedRoute, Router } from '@angular/router';
import { GetUserNameFromId } from '../../constants/user-constants';
import { GenerateWebAuthnLoginFinalizeRequest } from '../../models/webauthn/webauthn-login-finalize-request.interface';
import { AuthenticationService } from '../../services/authentication.service';
import { ChallengeService } from '../../services/challenge.service';
import { NotificationService } from '../../services/notification.service';
import { ChallengeSanitizationUtil } from '../../utils/challenge-sanitization-util';

@Component({
  selector: 'app-confirm-biometric-modal',
  templateUrl: './confirm-biometric-modal.component.html',
  styleUrls: ['./confirm-biometric-modal.component.css']
})
export class ConfirmBiometricModalComponent implements OnInit {

  public userEmail: string = "";
  public isLoading: boolean = false;

  constructor(private readonly dialogRef: MatDialogRef<ConfirmBiometricModalComponent>,
    @Inject(MAT_DIALOG_DATA) private readonly userId: string,
    private readonly authenticationService: AuthenticationService,
    private readonly challengeService: ChallengeService,
    private readonly notificationService: NotificationService,
    private readonly router: Router,
    private readonly route: ActivatedRoute) { }

  ngOnInit() {
    this.userEmail = GetUserNameFromId(this.userId) ?? "";
  }

  public close() {
    this.dialogRef.close();
  }

  public async submitBiometric(isGood: boolean) {
    this.isLoading = true;
    const challengeResp = await this.authenticationService.beginFakeWebauthnLogin(this.userId);
    
    if (challengeResp.type !== 'data') {
      this.notificationService.error('An error occurred fetching challenge', 'Error');
      this.isLoading = false;
      return;
    }

    await this.signChallenge(this.userEmail, challengeResp.data.publicKey.challenge, isGood)
  }

  private async signChallenge(userEmail: string, challenge: string, isGood: boolean) {
    const sanitizedChallenge = ChallengeSanitizationUtil.sanitizeInput(challenge);
    const signResp = await this.challengeService.signChallenge(userEmail, sanitizedChallenge);

    if (signResp.type !== 'data') {
      this.notificationService.error('An error occurred fetching challenge', 'Error');
      this.isLoading = false;
      return;
    }

    const data = signResp.data;

    const finalizeRequest = GenerateWebAuthnLoginFinalizeRequest();
    finalizeRequest.id = data.id;
    finalizeRequest.rawId = data.id;
    finalizeRequest.response.clientDataJSON = data.clientDataJson;
    finalizeRequest.response.authenticatorData = data.authenticatorData;
    finalizeRequest.response.userHandle = data.userHandle;
    finalizeRequest.response.signature = data.signature;

    // If user elected to send a bad biometric, scrabble it up!
    if (!isGood) {
      let newSig = "";
      for (var i = 0; i < finalizeRequest.response.signature.length; i++) {
        const newVal = finalizeRequest.response.signature.charCodeAt(i) + 1;
        newSig += String.fromCharCode(newVal)
      }
      finalizeRequest.response.signature = newSig;
    }

    // Simulate the "waiting"
    await new Promise(f => setTimeout(f, 300));

    const finalizeResp = await this.authenticationService.finalizeFakeWebauthnLogin(finalizeRequest);
    this.isLoading = false;
    if (finalizeResp.type !== 'data') {
      this.notificationService.error('Authentication failed', 'Error');
      return;
    }

    await this.authenticationService.setLogin();
    this.router.navigate([this.route.snapshot.queryParams[`redirect`] || '/'], { replaceUrl: true });
    this.dialogRef.close();
  }

}
