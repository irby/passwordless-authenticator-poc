import { HttpStatusCode } from '@angular/common/http';
import { Component, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { Subscription } from 'rxjs';
import { GrantService } from '../core/services/grant.service';
import { SocketService } from '../core/services/socket.service';
import { RouteSanitizationUtil } from '../core/utils/route-sanitization-util';

@Component({
  selector: 'app-share',
  templateUrl: './share.component.html',
  styleUrls: ['./share.component.css']
})
export class ShareComponent implements OnInit, OnDestroy {

  public messages: Array<any>;
  public chatBox: string;

  public isLoading: boolean = false;
  public isConnected: boolean = false;
  public isPrimaryAccountHolder: boolean = false;
  public hasSentConnectionConfirmation: boolean = false;

  public errorText: string = "";

  private routeSub!: Subscription;
  private querySub!: Subscription;

  private id: string | null = null;
  private token: string | null =  null;

  public clientInformation!: ClientInformation;

  public accessGrantSuccessful: boolean | null = null;

  constructor(
    private readonly socket: SocketService, 
    private readonly grantService: GrantService, 
    private readonly activatedRoute: ActivatedRoute,
    private readonly router: Router) { 
    this.messages = [];
    this.chatBox = "";
  }

  ngOnDestroy(): void {
    this.socket?.close();
    this.routeSub.unsubscribe();
    this.querySub.unsubscribe();
  }

  ngOnInit() {
    this.isLoading = true;

    this.routeSub = this.activatedRoute.params.subscribe(params => { 
      if (!params)
        return;
      const sanitizedRoute = RouteSanitizationUtil.sanitizeRoute(params['id'] as string);
      this.id = sanitizedRoute.grantId;
      this.token ??= sanitizedRoute.token;
      this.fetchGrantByIdAndToken();
    })
    this.querySub = this.activatedRoute.queryParams.subscribe(params => {
      if (!params)
        return;
      this.token = params['token'];
      this.fetchGrantByIdAndToken();
    })

    this.socket?.getEventListener().subscribe(event => {
      if (event.type === 'message') {
        let message: Message;
        message = JSON.parse(event.data);
        message.parsedContent = JSON.parse(message.content);

        switch (message.parsedContent.code) {
          case MessageCode.ConnectedSession:
            this.handleConnectedSession();
            break;
          case MessageCode.DisconnectedSession:
            this.handleDisconnectedSession();
            break;
          case MessageCode.AllPartiesPresent:
            this.handleAllPartiesPresent();
            break;
          case MessageCode.ClientInformation:
            this.handleClientInformation(message.parsedContent.message);
            break;
          case MessageCode.IsPrimaryAccountHolder:
            this.handleIsPrimaryAccountHolder();
            break;
          case MessageCode.DenyGrant:
            this.handleDenyGrant();
            break;
          case MessageCode.InitializeGrantConfirm:
            this.handleInitializeGrantConfirmation();
            break;
          case MessageCode.AccessGrantSuccess:
            this.handleAccessGrantValue(true);
            break;
          case MessageCode.AccessGrantFailure:
            this.handleAccessGrantValue(false);
            break;
        }

        let data = event.data;
        if (event.data.sender) {
          data = event.data.sender + ": " + data;
        }
        this.messages.push(data)
      }
      if (event.type === 'close') {
        this.errorText = 'Disconnected -- either too many connections or already connected in another window'
      }
    })
  }

  public async confirmGrant() {
    this.socket?.send(`${MessageCode.ConfirmGrant}`);
  }

  public async denyGrant() {
    this.socket?.send(`${MessageCode.DenyGrant}`);
    this.isConnected = false;
  }

  public backToHome() {
    this.router.navigate(['../../home'])
  }


  private async fetchGrantByIdAndToken() {
    if (!this.id || !this.token) {
      return;
    }
    const grantData = await this.grantService.getGrantByIdAndToken(this.id, this.token);
    this.isLoading = false;

    if (grantData.type === 'data') {
      this.socket?.createAndAssignSocket(this.id);
      return;
    }

    switch (grantData.statusCode) {
      case HttpStatusCode.RequestTimeout:
        this.errorText = "Request timed out. Please submit a new request";
        break;
      case HttpStatusCode.Forbidden:
        this.errorText = "Accessing grant not allowed. Are you logged in as another user? If so, log back into your account and try again.";
        break;
      case HttpStatusCode.NotFound:
        this.errorText = "Invalid id or token.";
        break;
      default:
        this.errorText = "Unknown error occurred.";
    }
  }

  private handleConnectedSession(): void {
  }

  private handleAllPartiesPresent(): void {
    this.isConnected = true;
  }

  private handleDisconnectedSession(): void {
    this.isConnected = false;
    this.hasSentConnectionConfirmation = false;
  }

  private handleDenyGrant(): void {
    this.isConnected = false;
  }

  private handleClientInformation(data: string): void {
    let clientInformation: ClientInformation;
    clientInformation = JSON.parse(data);
    this.clientInformation = clientInformation;
  }

  private handleIsPrimaryAccountHolder(): void {
    this.isPrimaryAccountHolder = true;
  }

  private handleInitializeGrantConfirmation(): void {
    if (confirm('Provide your biometric to continue')) {
      this.socket?.send(`${MessageCode.FinalizeGrantConfirm}`);
    } else {
      this.socket?.send(`${MessageCode.CancelGrantConfirm}`);
    }
  }

  private handleAccessGrantValue(isSuccess : boolean): void {
    this.accessGrantSuccessful = isSuccess;
  }
}

export interface Message {
  content: string;
  parsedContent: SocketMessage;
}
export interface SocketMessage {
  code: MessageCode;
  message?: any;
}
export interface ClientInformation {
  ipAddress: string;
  userAgent: string;
  email: string;
}
export enum MessageCode {
  ConnectedSession = 101,
	DisconnectedSession = 102,
	SessionRequest  = 103,

  AllPartiesPresent = 106,

	ClientInformation = 201,
  IsPrimaryAccountHolder = 202,

  ConfirmGrant = 301,
  DenyGrant = 302,

  InitializeGrantConfirm = 401,
	FinalizeGrantConfirm   = 402,
	CancelGrantConfirm     = 403,

	InitializeSubRegistrationConfirm = 501,
	FinalizeSubRegistrationConfirm   = 502,
	CancelSubRegistrationConfirm     = 503,

  AccessGrantSuccess = 601,
  AccessGrantFailure = 602
}

