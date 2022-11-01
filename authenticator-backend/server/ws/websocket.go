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
	jwt2 "github.com/teamhanko/hanko/backend/crypto/jwt"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/handler"
	"github.com/teamhanko/hanko/backend/mail"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"net/http"
	"strings"
)

type WebsocketHandler struct {
	renderer              *mail.Renderer
	nanoidGenerator       crypto.NanoidGenerator
	sessionManager        session.Manager
	persister             persistence.Persister
	emailConfig           config.Email
	serviceConfig         config.Service
	cfg                   *config.Config
	accountSharingHandler *handler.AccountSharingHandler
}

type ClientSessionData struct {
	IpAddress        string `json:"ipAddress,omitempty"`
	UserAgent        string `json:"userAgent,omitempty"`
	Email            string `json:"email,omitempty"`
	isAccountHolder  bool
	client           *Client
	grantIdReference uuid.UUID
	userId           uuid.UUID
}

type MessageCode int64

const (
	ConnectedSession     MessageCode = 101
	DisconnectedSession              = 102
	SessionRequest                   = 103
	SessionAlreadyExists             = 104
	TooManySessions                  = 105
	AllPartiesPresent                = 106

	ClientInformation      = 201
	IsPrimaryAccountHolder = 202

	ConfirmGrant = 301
	DenyGrant    = 302

	InitializeGrantConfirm = 401
	FinalizeGrantConfirm   = 402
	CancelGrantConfirm     = 403

	InitializeSubRegistrationConfirm = 501
	FinalizeSubRegistrationConfirm   = 502
	CancelSubRegistrationConfirm     = 503

	AccessGrantSuccess = 601
	AccessGrantFailure = 602
)

type SocketMessage struct {
	Code    MessageCode `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
}

func NewWebsocketHandler(cfg *config.Config, persister persistence.Persister, sessionManager session.Manager, accountSharingHandler *handler.AccountSharingHandler) (*WebsocketHandler, error) {
	renderer, err := mail.NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("failed to create new renderer: %w", err)
	}
	return &WebsocketHandler{
		renderer:              renderer,
		nanoidGenerator:       crypto.NewNanoidGenerator(),
		persister:             persister,
		emailConfig:           cfg.Passcode.Email, // TODO: Separate out into its own config value
		serviceConfig:         cfg.Service,
		sessionManager:        sessionManager,
		cfg:                   cfg,
		accountSharingHandler: accountSharingHandler,
	}, nil
}

// Code borrowed from:
// https://www.thepolyglotdeveloper.com/2016/12/create-real-time-chat-app-golang-angular-2-websockets/

type ClientManager struct {
	clients          map[*Client]bool
	broadcast        chan []byte
	register         chan *Client
	unregister       chan *Client
	websocketHandler *WebsocketHandler
}

type ClientSessionDataManager struct {
	clients map[string]ClientSessionData
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

var clientSessionDataManager = ClientSessionDataManager{
	clients: make(map[string]ClientSessionData),
}

func (manager *ClientManager) start(grant *models.AccountAccessGrant) {
	for {
		select {
		case conn := <-manager.register:
			manager.clients[conn] = true
			jsonMessage, _ := json.Marshal(&SocketMessage{Code: ConnectedSession})
			jsonMessage, _ = json.Marshal(&Message{Content: string(jsonMessage)})
			manager.send(jsonMessage, conn, true)

			if len(manager.clients) > 1 {
				// Notify all parties that session is in state
				jsonMessage, _ := json.Marshal(&SocketMessage{Code: AllPartiesPresent})
				jsonMessage, _ = json.Marshal(&Message{Content: string(jsonMessage)})
				manager.send(jsonMessage, conn, false)

				var guestAccount ClientSessionData
				var primaryAccountClient *Client
				for conn2 := range clientSessionDataManager.clients {
					conn3 := clientSessionDataManager.clients[conn2]

					if conn3.grantIdReference != grant.ID {
						continue
					}

					if conn3.isAccountHolder {
						primaryAccountClient = conn3.client
					} else {
						guestAccount = conn3
					}
				}

				fmt.Println("New account ID: ", primaryAccountClient.id)
				guestDataMessage, _ := json.Marshal(&guestAccount)
				fmt.Println("Guest data message: ", string(guestDataMessage))

				jsonMessage, _ = json.Marshal(&SocketMessage{Code: IsPrimaryAccountHolder})
				jsonMessage, _ = json.Marshal(&Message{Content: string(jsonMessage)})
				primaryAccountClient.send <- jsonMessage

				jsonMessage, _ = json.Marshal(&SocketMessage{Code: ClientInformation, Message: string(guestDataMessage)})
				jsonMessage, _ = json.Marshal(&Message{Content: string(jsonMessage)})
				primaryAccountClient.send <- jsonMessage

			}
		case conn := <-manager.unregister:
			if _, ok := manager.clients[conn]; ok {
				close(conn.send)
				delete(manager.clients, conn)
				jsonMessage, _ := json.Marshal(&SocketMessage{Code: DisconnectedSession})
				jsonMessage, _ = json.Marshal(&Message{Content: string(jsonMessage)})
				manager.send(jsonMessage, conn, false)
			}
		case message := <-manager.broadcast:
			var broadcastMessages = true
			var parsedMessage Message
			json.Unmarshal(message, &parsedMessage)
			fmt.Println("Message received: ", parsedMessage)

			// If deny grant, close out the guest session
			if parsedMessage.Content == fmt.Sprintf("%d", DenyGrant) {
				handleDenyGrant(grant)
			}

			// If confirm grant, prompt for biometric by account holder
			if parsedMessage.Content == fmt.Sprintf("%d", ConfirmGrant) {
				broadcastMessages = handleConfirmGrant(grant)
			}

			if parsedMessage.Content == fmt.Sprintf("%d", FinalizeGrantConfirm) {
				_ = handleFinalizeGrantConfirm(grant)
				broadcastMessages = false
			}

			if broadcastMessages {
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
}

func (manager *ClientManager) send(message []byte, ignore *Client, enforceIgnore bool) {
	for conn := range manager.clients {
		if !enforceIgnore || conn != ignore {
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
	grantId := c.Param("id")
	grant, err := p.persister.GetAccountAccessGrantPersister().Get(uuid.FromStringOrNil(grantId))

	if err != nil {
		return fmt.Errorf("unable to find grant %s: %w", grantId, err)
	}

	sessionToken, err := p.getSessionTokenFromContext(c)
	if err != nil {
		return err
	}

	go manager.start(grant)

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

	clientKey := createClientKey(sessionToken.Subject(), grant.ID)

	for existingConn := range manager.clients {
		if existingConn.id == clientKey {
			fmt.Println("Session already connected for subject")
			conn.Close()
			conn.WriteMessage(websocket.CloseMessage, []byte{})
		}
	}

	if len(manager.clients) >= 2 {
		fmt.Println("Too many connections")
		conn.Close()
		conn.WriteMessage(websocket.CloseMessage, []byte{})
		return fmt.Errorf("")
	}

	if len(manager.clients) > 0 {
		for clientSessionKey := range clientSessionDataManager.clients {
			if !strings.Contains(clientSessionKey, "::"+grantId) {
				continue
			}
			if clientSessionDataManager.clients[clientSessionKey].isAccountHolder {
				continue
			}
			if user.ID != grant.UserId {
				fmt.Println("Invalid session: Need to satisfy a primary account holder and a guest")
				conn.Close()
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return fmt.Errorf("")
			}
		}
	}

	client := &Client{id: clientKey, socket: conn, send: make(chan []byte)}

	clientSessionDataManager.clients[clientKey] = ClientSessionData{IpAddress: ipAddr, UserAgent: userAgent, Email: userEmail,
		isAccountHolder: user.ID == grant.UserId, client: client, grantIdReference: grant.ID, userId: user.ID}
	manager.register <- client

	if manager.websocketHandler == nil {
		manager.websocketHandler = p
	}

	go client.read()
	go client.write()

	return nil
}

func createClientKey(subject string, grantId uuid.UUID) string {
	return createClientKeyFromString(subject, grantId.String())
}

func createClientKeyFromString(subject string, grantId string) string {
	return subject + "::" + grantId
}

func handleDenyGrant(grant *models.AccountAccessGrant) {
	fmt.Println("Denied!")
	jsonMessage, _ := json.Marshal(&SocketMessage{Code: DenyGrant})
	jsonMessage, _ = json.Marshal(&Message{Content: string(jsonMessage)})

	for conn2 := range clientSessionDataManager.clients {
		conn3 := clientSessionDataManager.clients[conn2]

		if conn3.grantIdReference != grant.ID || conn3.isAccountHolder {
			continue
		}

		conn3.client.send <- jsonMessage
		close(conn3.client.send)
		delete(manager.clients, conn3.client)
		delete(clientSessionDataManager.clients, conn3.client.id)
	}
}

func handleConfirmGrant(grant *models.AccountAccessGrant) bool {
	fmt.Println("Confirming grant!")
	jsonMessage, _ := json.Marshal(&SocketMessage{Code: InitializeGrantConfirm})
	jsonMessage, _ = json.Marshal(&Message{Content: string(jsonMessage)})

	for conn2 := range clientSessionDataManager.clients {
		conn3 := clientSessionDataManager.clients[conn2]

		if conn3.grantIdReference != grant.ID || !conn3.isAccountHolder {
			continue
		}

		conn3.client.send <- jsonMessage
		return false
	}
	return true
}

func handleFinalizeGrantConfirm(grant *models.AccountAccessGrant) error {
	var primaryAccountHolderSession *ClientSessionData
	var guestSession *ClientSessionData
	for conn2 := range clientSessionDataManager.clients {
		conn3 := clientSessionDataManager.clients[conn2]

		if conn3.grantIdReference != grant.ID {
			continue
		}

		if conn3.isAccountHolder {
			primaryAccountHolderSession = &conn3
		} else {
			guestSession = &conn3
		}
	}

	if primaryAccountHolderSession == nil || guestSession == nil {
		return errors.New("both primary account holder and guest sessions are required")
	}

	err := manager.websocketHandler.accountSharingHandler.CreateAccountWithGrant(grant.ID, primaryAccountHolderSession.userId, guestSession.userId)

	if err != nil {
		return fmt.Errorf("an error occurred when creating account with grant: %w", err)
	}

	jsonMessage, _ := json.Marshal(&SocketMessage{Code: AccessGrantSuccess})
	jsonMessage, _ = json.Marshal(&Message{Content: string(jsonMessage)})
	primaryAccountHolderSession.client.send <- jsonMessage
	guestSession.client.send <- jsonMessage

	return nil
}

func (p *WebsocketHandler) getSessionTokenFromContext(c echo.Context) (jwt.Token, error) {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return nil, dto.NewHTTPError(http.StatusUnauthorized, "invalid or expired session token")
	}
	surrogateKey, err := jwt2.GetSurrogateKeyFromToken(sessionToken)
	if err != nil {
		return nil, dto.NewHTTPError(http.StatusInternalServerError, "could not extract surrogate key from session token")
	}
	if sessionToken.Subject() != surrogateKey {
		return nil, dto.NewHTTPError(http.StatusForbidden, "surrogate ID must match session subject")
	}
	return sessionToken, nil
}
