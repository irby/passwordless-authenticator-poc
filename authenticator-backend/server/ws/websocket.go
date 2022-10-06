package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto"
	"github.com/teamhanko/hanko/backend/mail"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/session"
	"net/http"
)

type AccountSharingHandler struct {
	mailer          mail.Mailer
	renderer        *mail.Renderer
	nanoidGenerator crypto.NanoidGenerator
	sessionManager  session.Manager
	persister       persistence.Persister
	emailConfig     config.Email
	serviceConfig   config.Service
	cfg             *config.Config
}

type WebsocketHandler struct {
	renderer        *mail.Renderer
	nanoidGenerator crypto.NanoidGenerator
	sessionManager  session.Manager
	persister       persistence.Persister
	emailConfig     config.Email
	serviceConfig   config.Service
	cfg             *config.Config
}

func NewWebsocketHandler(cfg *config.Config, persister persistence.Persister, sessionManager session.Manager) (*WebsocketHandler, error) {
	renderer, err := mail.NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("failed to create new renderer: %w", err)
	}
	return &WebsocketHandler{
		renderer:        renderer,
		nanoidGenerator: crypto.NewNanoidGenerator(),
		persister:       persister,
		emailConfig:     cfg.Passcode.Email, // TODO: Separate out into its own config value
		serviceConfig:   cfg.Service,
		sessionManager:  sessionManager,
		cfg:             cfg,
	}, nil
}

// Code borrowed from:
// https://www.thepolyglotdeveloper.com/2016/12/create-real-time-chat-app-golang-angular-2-websockets/

type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

type Client struct {
	id     string
	socket *websocket.Conn
	send   chan []byte
}

type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}

var manager = ClientManager{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

func (manager *ClientManager) start() {
	for {
		select {
		case conn := <-manager.register:
			manager.clients[conn] = true
			jsonMessage, _ := json.Marshal(&Message{Content: "/A new socket has connected."})
			manager.send(jsonMessage, conn)
		case conn := <-manager.unregister:
			if _, ok := manager.clients[conn]; ok {
				close(conn.send)
				delete(manager.clients, conn)
				jsonMessage, _ := json.Marshal(&Message{Content: "/A socket has disconnected."})
				manager.send(jsonMessage, conn)
			}
		case message := <-manager.broadcast:
			for conn := range manager.clients {
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(manager.clients, conn)
				}
			}
		}
	}
}

func (manager *ClientManager) send(message []byte, ignore *Client) {
	for conn := range manager.clients {
		if conn != ignore {
			conn.send <- message
		}
	}
}

func (c *Client) read() {
	defer func() {
		manager.unregister <- c
		c.socket.Close()
	}()

	for {
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			manager.unregister <- c
			c.socket.Close()
			break
		}
		jsonMessage, _ := json.Marshal(&Message{Sender: c.id, Content: string(message)})
		manager.broadcast <- jsonMessage
	}
}

func (c *Client) write() {
	defer func() {
		c.socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func (p *WebsocketHandler) WsPage(c echo.Context) error {

	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("missing or malformed jwt")
	}

	go manager.start()

	ipAddr := c.Request().RemoteAddr
	userAgent := c.Request().Header.Get("User-Agent")
	user, err := p.persister.GetUserPersister().Get(uuid.FromStringOrNil(sessionToken.Subject()))
	if err != nil {
		return fmt.Errorf("unable to get user: %w", err)
	}
	userEmail := user.Email

	fmt.Println(ipAddr, userAgent, userEmail)

	conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(c.Response(), c.Request(), nil)
	if error != nil {
		//http.NotFound(res, req)
		return fmt.Errorf("")
	}

	//for existingConn := range manager.clients {
	//	if existingConn.id == sessionToken.Subject() {
	//		fmt.Println("Session already connected for subject")
	//		conn.Close()
	//		conn.WriteMessage(websocket.CloseMessage, []byte{})
	//	}
	//}

	if len(manager.clients) >= 2 {
		fmt.Println("Too many connections")
		conn.Close()
		conn.WriteMessage(websocket.CloseMessage, []byte{})
		return fmt.Errorf("")
	}

	client := &Client{id: sessionToken.Subject(), socket: conn, send: make(chan []byte)}

	manager.register <- client

	go client.read()
	go client.write()

	return nil
}
