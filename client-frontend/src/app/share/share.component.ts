import { HttpStatusCode } from "@angular/common/http";
import { Component, OnDestroy, OnInit } from "@angular/core";
import { MatDialog, MatDialogConfig } from "@angular/material/dialog";
import { ActivatedRoute, Router } from "@angular/router";
import { Subscription } from "rxjs";
import {
  ConfirmBiometricContext,
  ConfirmBiometricData,
  ConfirmBiometricModalComponent,
} from "../core/modals/confirm-biometric-modal/confirm-biometric-modal.component";
import { AuthenticationService } from "../core/services/authentication.service";
import { NotificationService } from "../core/services/notification.service";
import { SocketService } from "../core/services/socket.service";
import { RouteSanitizationUtil } from "../core/utils/route-sanitization-util";

@Component({
  selector: "app-share",
  templateUrl: "./share.component.html",
  styleUrls: ["./share.component.css"],
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
  private token: string | null = null;

  public clientInformation!: ClientInformation;

  public accessGrantSuccessful: boolean | null = null;
  public currentUserId!: string;

  constructor(
    private readonly socket: SocketService,
    private readonly activatedRoute: ActivatedRoute,
    private readonly notificationService: NotificationService,
    private readonly authenticationService: AuthenticationService,
    private readonly matDialog: MatDialog,
    private readonly router: Router
  ) {
    this.messages = [];
    this.chatBox = "";
  }

  ngOnDestroy(): void {
    this.socket?.close();
    this.routeSub.unsubscribe();
    this.querySub.unsubscribe();
  }

  async ngOnInit() {
    this.isLoading = true;
    this.authenticationService
      .getUserAsObservable()
      .subscribe((user) => (this.currentUserId = user.id));
    await this.authenticationService.setLogin();

    this.routeSub = this.activatedRoute.params.subscribe((params) => {
      if (!params) return;
      const sanitizedRoute = RouteSanitizationUtil.sanitizeRoute(
        params["id"] as string
      );
      this.id = sanitizedRoute.grantId;
      this.token ??= sanitizedRoute.token;
      this.fetchGrantByIdAndToken();
    });
    this.querySub = this.activatedRoute.queryParams.subscribe((params) => {
      if (!params) return;
      this.token = params["token"];
      this.fetchGrantByIdAndToken();
    });

    this.socket?.getEventListener().subscribe((event) => {
      if (event.type === "message") {
        const message: Message = JSON.parse(event.data);
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
          case MessageCode.BadSessionToken:
            this.handleBadSessionToken();
            break;
          case MessageCode.InvalidGrantIdOrToken:
            this.handleInvalidGrantIdOrToken(message.parsedContent);
        }

        let data = event.data;
        if (event.data.sender) {
          data = event.data.sender + ": " + data;
        }
        this.messages.push(data);
      }
      if (event.type === "close") {
        this.errorText =
          "Disconnected -- either an error occurred, too many connections, or already connected in another window";
      }
    });
  }

  public async confirmGrant() {
    const data: ConfirmBiometricData = {
      userId: this.currentUserId,
      context: ConfirmBiometricContext.UserConfirmGrant,
      guestUserId: this.clientInformation.userId,
      grantId: this.id ?? "",
    };
    const config: MatDialogConfig = {
      width: "45em",
      height: "20em",
      data: data,
    };
    this.matDialog
      .open(ConfirmBiometricModalComponent, config)
      .afterClosed()
      .subscribe((result) => {
        if (result.data?.isSuccess) {
          this.socket?.send(`${MessageCode.FinalizeGrantConfirm}`);
        }
      });
  }

  public async denyGrant() {
    this.socket?.send(`${MessageCode.DenyGrant}`);
    this.isConnected = false;
  }

  public backToHome() {
    this.router.navigate(["../../home"]);
  }

  private async fetchGrantByIdAndToken() {
    if (!this.id || !this.token) {
      return;
    }
    await this.socket?.createAndAssignSocket(this.id, this.token);
    this.isLoading = false;
  }

  private handleConnectedSession(): void {}

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
    const clientInformation: ClientInformation = JSON.parse(data);
    this.clientInformation = clientInformation;
  }

  private handleIsPrimaryAccountHolder(): void {
    this.isPrimaryAccountHolder = true;
  }

  private handleInitializeGrantConfirmation(): void {
    if (confirm("Provide your biometric to continue")) {
      this.socket?.send(`${MessageCode.FinalizeGrantConfirm}`);
    } else {
      this.socket?.send(`${MessageCode.CancelGrantConfirm}`);
    }
  }

  private handleAccessGrantValue(isSuccess: boolean): void {
    this.accessGrantSuccessful = isSuccess;
  }

  private handleInvalidGrantIdOrToken(message: SocketMessage): void {
    switch (message.message) {
      case `${HttpStatusCode.NotFound}`:
        this.notificationService.dismissibleError(
          "Invalid ID or token",
          "Grant not found"
        );
        break;
      case `${HttpStatusCode.RequestTimeout}`:
        this.notificationService.dismissibleError(
          "Grant has expired. Please submit invite again.",
          "Request expired"
        );
        break;
      default:
        this.notificationService.dismissibleError(
          "An unexpected error occurred. Please try your request again"
        );
    }
  }

  private handleBadSessionToken(): void {
    this.notificationService.dismissibleError(
      "Are you logged in as another user? If so, please log back into your account and try again.",
      "Access not allowed"
    );
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
  userId: string;
}
export enum MessageCode {
  ConnectedSession = 101,
  DisconnectedSession = 102,
  SessionRequest = 103,

  AllPartiesPresent = 106,

  ClientInformation = 201,
  IsPrimaryAccountHolder = 202,

  ConfirmGrant = 301,
  DenyGrant = 302,

  InitializeGrantConfirm = 401,
  FinalizeGrantConfirm = 402,
  CancelGrantConfirm = 403,

  InitializeSubRegistrationConfirm = 501,
  FinalizeSubRegistrationConfirm = 502,
  CancelSubRegistrationConfirm = 503,

  AccessGrantSuccess = 601,
  AccessGrantFailure = 602,

  BadSessionToken = 701,
  InvalidGrantIdOrToken = 702,
}
