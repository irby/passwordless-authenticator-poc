import { EventEmitter, Injectable } from "@angular/core";
import axios from "axios";

@Injectable()
export class SocketService {

    private socket!: WebSocket;
    private listener: EventEmitter<any> = new EventEmitter();

    public constructor() {
        
    }

    public createAndAssignSocket(id: string, token: string) {
        this.socket = new WebSocket(`ws://localhost:8000/ws/${id}?token=${token}`);
        this.socket.onopen = event => {
            this.listener.emit({"type": "open", "data": event});
        }
        this.socket.onclose = event => {
            this.listener.emit({"type": "close", "data": event});
        }
        this.socket.onmessage = event => {
            this.listener.emit({"type": "message", "data": event.data});
        }
    }

    public send(data: string) {
        this.socket?.send(data);
    }

    public close() {
        this.socket?.close();
    }

    public getEventListener() {
        return this.listener;
    }

}