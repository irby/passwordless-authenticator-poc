import { Component, Inject, OnInit } from '@angular/core';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { ActivatedRoute, Router } from '@angular/router';
import { GetUserNameFromId } from '../../constants/user-constants';
import { GenerateWebAuthnLoginFinalizeRequest, WebAuthnLoginFinalizeRequest } from '../../models/webauthn/webauthn-login-finalize-request.interface';
import { AuthenticationService } from '../../services/authentication.service';
import { ChallengeService, SignChallengeAsUserResponse } from '../../services/challenge.service';
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
    @Inject(MAT_DIALOG_DATA) private readonly data: ConfirmBiometricData,
    private readonly authenticationService: AuthenticationService,
    private readonly challengeService: ChallengeService,
    private readonly notificationService: NotificationService,
    private readonly router: Router,
    private readonly route: ActivatedRoute) { }

  ngOnInit() {
    this.userEmail = GetUserNameFromId(this.data.userId) ?? "user";
  }

  public close() {
    this.dialogRef.close();
  }

  public async submitBiometric(isGood: boolean) {
    this.isLoading = true;
    const challengeResp = await this.authenticationService.beginWebauthnLogin(this.data.userId);
    
    if (challengeResp.type !== 'data') {
      this.notificationService.error('An error occurred fetching challenge', 'Error');
      this.isLoading = false;
      return;
    }

    const finalizeRequest = await this.signChallenge(this.data.userId, challengeResp.data.publicKey.challenge, isGood);

    if (!finalizeRequest)
      return;

    await this.submitSignature(finalizeRequest);

    await this.handlePostConfirm();

    this.dialogRef.close();
  }

  private async signChallenge(userId: string, challenge: string, isGood: boolean): Promise<WebAuthnLoginFinalizeRequest | null> {
    const sanitizedChallenge = ChallengeSanitizationUtil.sanitizeInput(challenge);
    const signResp = await this.challengeService.signChallenge(userId, sanitizedChallenge);

    if (signResp.type !== 'data') {
      this.notificationService.error('An error occurred signing challenge', 'Error');
      this.isLoading = false;
      return null;
    }

    return this.generateRequestModel(signResp.data, isGood);
  }

  private generateRequestModel(data: SignChallengeAsUserResponse, isGood: boolean): WebAuthnLoginFinalizeRequest {
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

    return finalizeRequest;
  }

  private async handlePostConfirm(): Promise<void> {
    if (this.data.context !== ConfirmBiometricContext.Login) {
      return;
    }

    await this.authenticationService.setLogin();
    this.router.navigate([this.route.snapshot.queryParams[`redirect`] || '/'], { replaceUrl: true });
  }

  private async submitSignature(finalizeRequest: WebAuthnLoginFinalizeRequest): Promise<void> {
    // Simulate the "waiting"
    await new Promise(f => setTimeout(f, 300));

    const finalizeResp = await this.authenticationService.finalizeWebauthnLogin(finalizeRequest);
    this.isLoading = false;
    if (finalizeResp.type !== 'data') {
      this.notificationService.error('Authentication failed', 'Error');
      return;
    }
  }
}

export enum ConfirmBiometricContext {
  Login = 10,
  AdminDeactivateAccount = 20,
  AdminDeactivateGrants = 21,
  UserConfirmGrant = 30,
  UserRemoveGrant = 31,
  UserAssumeGrant = 32
}

export interface ConfirmBiometricData {
  userId: string;
  context: ConfirmBiometricContext;
}