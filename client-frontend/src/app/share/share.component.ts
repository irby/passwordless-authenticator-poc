import { HttpStatusCode } from '@angular/common/http';
import { Component, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { Subscription } from 'rxjs';
import { GrantService } from '../core/services/grant.service';
import { SocketService } from '../core/services/socket.service';

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

  public errorText: string = "";

  private routeSub!: Subscription;
  private querySub!: Subscription;

  private id: string | null = null;
  private token: string | null =  null;

  constructor(private readonly socket: SocketService, private readonly grantService: GrantService, private activatedRoute: ActivatedRoute) { 
    this.messages = [];
    this.chatBox = "";
  }

  ngOnDestroy(): void {
    this.socket.close();
    this.routeSub.unsubscribe();
    this.querySub.unsubscribe();
  }

  ngOnInit() {
    this.isLoading = true;

    this.routeSub = this.activatedRoute.params.subscribe(params => {
      if (!params)
        return;
      this.id = params['id'];
      this.fetchGrantByIdAndToken();
    })
    this.querySub = this.activatedRoute.queryParams.subscribe(params => {
      if (!params)
        return;
      this.token = params['token'];
      this.fetchGrantByIdAndToken();
    })

    this.socket.getEventListener().subscribe(event => {
      console.log(event);
      if (event.type === 'message') {
        if (event.data.includes('A new socket has connected')) { // TODO: Handle this better
          this.isConnected = true;
          // this.socket.send('A new socket has connected');
        } else if (event.data === '{\"content\":\"/A socket has disconnected.\"}') {
          this.isConnected = false;
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

  public send() {
    if (this.chatBox) {
      this.socket.send(this.chatBox);
      this.chatBox = "";
    }
  }

  public isSystemMessage(message: string) {
    return message.startsWith("/") ? "<strong>" + message.substring(1) + "</strong>" : message;
  }

  private async fetchGrantByIdAndToken() {
    if (!this.id || !this.token) {
      return;
    }
    const grantData = await this.grantService.getGrantByIdAndToken(this.id, this.token);
    this.isLoading = false;

    if (grantData.type === 'data') {
      return;
    }

    console.log('grantdata', grantData);

    switch (grantData.statusCode) {
      case HttpStatusCode.RequestTimeout:
        this.errorText = "Request timed out. Please submit a new request";
        break;
      case HttpStatusCode.NotFound:
        this.errorText = "Invalid id or token.";
        break;
      default:
        this.errorText = "Unknown error occurred.";
    }
  }

}
