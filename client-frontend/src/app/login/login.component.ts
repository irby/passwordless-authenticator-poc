import { Component, OnInit, Renderer2 } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { environment } from 'src/environments/environment';
import { GetUserNameFromId, UserId } from '../core/constants/user-constants';
import { GenerateWebAuthnLoginFinalizeRequest, WebAuthnLoginFinalizeRequest } from '../core/models/webauthn/webauthn-login-finalize-request.interface';
import { AuthenticationService } from '../core/services/authentication.service';
import { ScriptService } from '../core/services/script.service';
import { PublicKey } from '../core/models/webauthn/webauthn-login-initialize-response.interface';

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
    private readonly authenticationSerivce: AuthenticationService) { }

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
    finalizeRequest.id = "5L33ArYqAMpWVeFP9CxnPGtYn0c";
    finalizeRequest.rawId = "5L33ArYqAMpWVeFP9CxnPGtYn0c";

    const clientData = {
      type: "webauthn.get",
      challenge: publicKey.challenge.replace(/=/g, '').replace(/\//g, "_").replace(/\+/g, "-"),
      origin: "http://localhost:4200"
    }

    finalizeRequest.response.clientDataJSON = btoa(JSON.stringify(clientData));
    finalizeRequest.response.authenticatorData = "SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2MFAAAAAA";
    finalizeRequest.response.signature = "MEUCICKZAlJ8AbvwgjHaM6TT4ydm8r9Difhf6cAVkA1R2DsvAiEA_8VO5r3aRgvYCEf-QZh2dNjoNnooFs8Qx8WNt7LFpvg";
    finalizeRequest.response.userHandle = "MoChopQXSxCm6Zh-q99j7A";

    await this.authenticationSerivce.finalizeFakeWebauthnLogin(finalizeRequest);
  }

}
