import { Component, OnInit, Renderer2 } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { environment } from 'src/environments/environment';
import { GetPrivateKeyFromId, GetUserNameFromId, UserId } from '../core/constants/user-constants';
import { GenerateWebAuthnLoginFinalizeRequest, WebAuthnLoginFinalizeRequest } from '../core/models/webauthn/webauthn-login-finalize-request.interface';
import { AuthenticationService } from '../core/services/authentication.service';
import { ScriptService } from '../core/services/script.service';
import { PublicKey } from '../core/models/webauthn/webauthn-login-initialize-response.interface';
import { ChallengeSanitizationUtil } from '../core/utils/challenge-sanitization-util';
import { CryptoUtil } from '../core/utils/crypto-util';
import { ChallengeService } from '../core/services/challenge.service';

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
    private readonly challengeService: ChallengeService) { }

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

  public async beginFakeWebauthnLogin(userId: string) {
    var resp = await this.authenticationSerivce.beginFakeWebauthnLogin(userId);
    if (resp.type === 'data') {
      const confirmation = confirm(`Provide biometric for ${GetUserNameFromId(userId)}?`);

      if (confirmation) {
        this.finalizeFakeWebAuthnLogin(userId, resp.data.publicKey);
      }

      console.log(resp.data);
    }
  }

  private async finalizeFakeWebAuthnLogin(userId: string, publicKey: PublicKey) {
    const finalizeRequest = GenerateWebAuthnLoginFinalizeRequest();
    finalizeRequest.id = "V-Xjt3TuMNWo-D8YR5BjNOUnTRE";
    finalizeRequest.rawId = "V-Xjt3TuMNWo-D8YR5BjNOUnTRE";

    const resp = await this.challengeService.signChallenge(GetUserNameFromId(userId) ?? "", publicKey.challenge);

    const clientData = {
      type: "webauthn.get",
      challenge: ChallengeSanitizationUtil.sanitizeInput(publicKey.challenge),
      origin: "http://localhost:4200"
    };

    console.log(resp);

    const signature = resp.data.signature;

    finalizeRequest.response.clientDataJSON = btoa(JSON.stringify(clientData));
    finalizeRequest.response.authenticatorData = "SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2MFAAAAAA";
    finalizeRequest.response.signature = ChallengeSanitizationUtil.sanitizeInput(signature);
    finalizeRequest.response.userHandle = "MoChopQXSxCm6Zh-q99j7A";

    await this.authenticationSerivce.finalizeFakeWebauthnLogin(finalizeRequest);
  }

}
