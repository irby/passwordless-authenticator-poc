package ws

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

// Code borrowed from:
// https://www.thepolyglotdeveloper.com/2016/12/create-real-time-chat-app-golang-angular-2-websockets/

//type ClientManager struct {
//	clients    map[*Client]bool
//	broadcast  chan []byte
//	register   chan *Client
//	unregister chan *Client
//}
//
//type Client struct {
//	id     string
//	socket *websocket.Conn
//	send   chan []byte
//}
//
//type Message struct {
//	Sender    string `json:"sender,omitempty"`
//	Recipient string `json:"recipient,omitempty"`
//	Content   string `json:"content,omitempty"`
//}
//
//var manager = ClientManager{
//	broadcast:  make(chan []byte),
//	register:   make(chan *Client),
//	unregister: make(chan *Client),
//	clients:    make(map[*Client]bool),
//}
//
//func (manager *ClientManager) start() {
//	for {
//		select {
//		case conn := <-manager.register:
//			manager.clients[conn] = true
//			jsonMessage, _ := json.Marshal(&Message{Content: "/A new socket has connected."})
//			manager.send(jsonMessage, conn)
//		case conn := <-manager.unregister:
//			if _, ok := manager.clients[conn]; ok {
//				close(conn.send)
//				delete(manager.clients, conn)
//				jsonMessage, _ := json.Marshal(&Message{Content: "/A socket has disconnected."})
//				manager.send(jsonMessage, conn)
//			}
//		case message := <-manager.broadcast:
//			for conn := range manager.clients {
//				select {
//				case conn.send <- message:
//				default:
//					close(conn.send)
//					delete(manager.clients, conn)
//				}
//			}
//		}
//	}
//}
//
//func (manager *ClientManager) send(message []byte, ignore *Client) {
//	for conn := range manager.clients {
//		if conn != ignore {
//			conn.send <- message
//		}
//	}
//}
//
//func (c *Client) read() {
//	defer func() {
//		manager.unregister <- c
//		c.socket.Close()
//	}()
//
//	for {
//		_, message, err := c.socket.ReadMessage()
//		if err != nil {
//			manager.unregister <- c
//			c.socket.Close()
//			break
//		}
//		jsonMessage, _ := json.Marshal(&Message{Sender: c.id, Content: string(message)})
//		manager.broadcast <- jsonMessage
//	}
//}
//
//func (c *Client) write() {
//	defer func() {
//		c.socket.Close()
//	}()
//
//	for {
//		select {
//		case message, ok := <-c.send:
//			if !ok {
//				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
//				return
//			}
//
//			c.socket.WriteMessage(websocket.TextMessage, message)
//		}
//	}
//}
//
//func WsPage(res http.ResponseWriter, req *http.Request) {
//	manager.start()
//	conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
//	if error != nil {
//		http.NotFound(res, req)
//		return
//	}
//	uId, err := uuid.NewV4()
//	if err != nil {
//		fmt.Println("Unable to generate a UID")
//		return
//	}
//	client := &Client{id: uId.String(), socket: conn, send: make(chan []byte)}
//
//	manager.register <- client
//
//	go client.read()
//	go client.write()
//}

func Hello(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		for {
			// Write
			err := websocket.Message.Send(ws, "Hello, Client!")
			if err != nil {
				c.Logger().Error(err)
			}

			// Read
			msg := ""
			err = websocket.Message.Receive(ws, &msg)
			if err != nil {
				c.Logger().Error(err)
			}
			fmt.Printf("%s\n", msg)
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
