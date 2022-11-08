import { Component, Inject, OnInit } from '@angular/core';
import { MatDialogConfig, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { ActivatedRoute, Router } from '@angular/router';
import { GetUserNameFromId } from '../../constants/user-constants';
import { ServiceResponse } from '../../models/service-response.interface';
import { GenerateWebAuthnLoginFinalizeRequest, WebAuthnLoginFinalizeRequest } from '../../models/webauthn/webauthn-login-finalize-request.interface';
import { CreateAccountWithGrantResponse, GrantAttestationObject, WebauthnLoginInitializeResponse } from '../../models/webauthn/webauthn-login-initialize-response.interface';
import { AuthenticationService } from '../../services/authentication.service';
import { ChallengeService, SignChallengeAsUserResponse } from '../../services/challenge.service';
import { BeginCreateAccountWithGrantRequest, FinishCreateAccountWithGrantRequest, GrantService } from '../../services/grant.service';
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
    private readonly grantService: GrantService,
    private readonly route: ActivatedRoute) { }

  ngOnInit() {
    this.userEmail = GetUserNameFromId(this.data.userId) ?? "user";
  }

  public close() {
    this.dialogRef.close();
  }

  public async submitBiometric(isGood: boolean) {
    this.isLoading = true;
    const challengeResp = await this.initializeRequest(this.data, isGood);

    if (!challengeResp) {
      return;
    }

    if (await this.submitSignature(challengeResp)) {
      await this.handlePostConfirm();
      this.dialogRef.close({ data: {isSuccess: true} as ConfirmBiometricModalData });
    }
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

  private async submitSignature(finalizeRequest: InitializeChallengeResponse): Promise<boolean> {
    // Simulate the "waiting"
    await new Promise(f => setTimeout(f, 300));

    const finalizeResp = await this.finalizeRequest(this.data, finalizeRequest);
    this.isLoading = false;
    if (finalizeResp.type !== 'data') {
      this.notificationService.error('Authentication failed', 'Error');
      return false;
    }
    return true;
  }

  private async initializeRequest(data: ConfirmBiometricData, isGood: boolean): Promise<InitializeChallengeResponse | null> {
    switch (data.context) {
      case ConfirmBiometricContext.Login:
        const loginResp = await this.initializeWebauthnLogin(data.userId);

        if (loginResp.type !== 'data') {
          this.notificationService.error('An error occurred fetching challenge', 'Error');
          this.isLoading = false;
          return null;
        }

        const signedChallenge = await this.signChallenge(this.data.userId, loginResp.data.publicKey.challenge, isGood);

        return {signature: signedChallenge}
      case ConfirmBiometricContext.UserConfirmGrant:
        const grantResp = await this.initializeCreateAccountWithGrant(data.guestUserId ?? "", data.grantId ?? "");

        if (grantResp.type !== 'data') {
          this.notificationService.error('An error occurred fetching challenge', 'Error');
          this.isLoading = false;
          return null;
        }

        const signature = await this.signChallenge(this.data.userId, grantResp.data.options.publicKey.challenge, isGood);
        const attestation = await this.signChallenge(this.data.userId, this.createAttestationObject(grantResp.data.grant), isGood);

        return {signature: signature, attestation: attestation?.response.signature}
      default:
        throw new Error(`unknown context ${data.context}`);
    }
  }

  private async initializeWebauthnLogin(userId: string): Promise<ServiceResponse<WebauthnLoginInitializeResponse>> {
    return await this.authenticationService.beginWebauthnLogin(userId);
  }

  private async initializeCreateAccountWithGrant(guestUserId: string, grantId: string): Promise<ServiceResponse<CreateAccountWithGrantResponse>> {
    const request: BeginCreateAccountWithGrantRequest = {
      guestUserId: guestUserId,
      grantId: grantId
    }
    return await this.grantService.initializeCreateAccountWithGrant(request);
  }




  private async finalizeRequest(data: ConfirmBiometricData, finalizeRequest: InitializeChallengeResponse): Promise<ServiceResponse<any>> {
    if (!finalizeRequest.signature) {
      throw new Error();
    }

    switch (data.context) {
      case ConfirmBiometricContext.Login:
        return await this.finalizeWebauthnLogin(finalizeRequest.signature);
      case ConfirmBiometricContext.UserConfirmGrant:
        if (!finalizeRequest.attestation) {
          throw new Error("Attestation is missing");
        }
        return await this.finalizeCreateAccountWithGrant(data.guestUserId ?? "", data.grantId ?? "", finalizeRequest.attestation, finalizeRequest.signature);
      default:
        throw new Error(`unknown context ${data.context}`);
    }
  }

  private async finalizeWebauthnLogin(finalizeRequest: WebAuthnLoginFinalizeRequest): Promise<ServiceResponse<any>> {
    return await this.authenticationService.finalizeWebauthnLogin(finalizeRequest);
  }

  private async finalizeCreateAccountWithGrant(guestUserId: string, grantId: string, grantAttestation: string, finalizeRequest: WebAuthnLoginFinalizeRequest): Promise<ServiceResponse<WebauthnLoginInitializeResponse>> {
    const request: FinishCreateAccountWithGrantRequest = {
      guestUserId: guestUserId,
      grantId: grantId,
      grantAttestation: grantAttestation,
      id: finalizeRequest.id,
      clientExtensionResults: finalizeRequest.clientExtensionResults,
      response: finalizeRequest.response,
      rawId: finalizeRequest.rawId,
      authenticationAttachment: finalizeRequest.authenticationAttachment,
      type: finalizeRequest.type
    }
    return await this.grantService.finishCreateAccountWithGrant(request);
  }

  private createAttestationObject(grantAttestationObject: GrantAttestationObject) {
    return btoa(JSON.stringify(grantAttestationObject));
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
  guestUserId?: string;
  grantId?: string;
}

export interface ConfirmBiometricModalData {
  isSuccess: boolean;
}

export interface InitializeChallengeResponse {
  signature: WebAuthnLoginFinalizeRequest | null;
  attestation?: string | null;
}