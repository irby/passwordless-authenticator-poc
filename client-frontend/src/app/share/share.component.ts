import { Component, OnDestroy, OnInit } from '@angular/core';
import { SocketService } from '../core/services/socket.service';

@Component({
  selector: 'app-share',
  templateUrl: './share.component.html',
  styleUrls: ['./share.component.css']
})
export class ShareComponent implements OnInit, OnDestroy {

  public messages: Array<any>;
  public chatBox: string;

  constructor(private readonly socket: SocketService) { 
    this.messages = [];
    this.chatBox = "";
  }
  ngOnDestroy(): void {
    this.socket.close();
  }

  ngOnInit() {
    this.socket.getEventListener().subscribe(event => {
      console.log(event);
      if (event.type === 'message') {
        let data = event.data;
        if (event.data.sender) {
          data = event.data.sender + ": " + data;
        }
        this.messages.push(data)
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

}
