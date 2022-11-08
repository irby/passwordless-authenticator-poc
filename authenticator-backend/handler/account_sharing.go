package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/teamhanko/hanko/backend/dto/intern"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto"
	jwt2 "github.com/teamhanko/hanko/backend/crypto/jwt"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/mail"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
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
	webauthn        *webauthn.WebAuthn
}

const TimeToLiveMinutes = 15 // TODO: make into a config value

func NewAccountSharingHandler(cfg *config.Config, persister persistence.Persister, sessionManager session.Manager, mailer mail.Mailer) (*AccountSharingHandler, error) {
	renderer, err := mail.NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("failed to create new renderer: %w", err)
	}
	f := false
	wa, _ := webauthn.New(&webauthn.Config{
		RPDisplayName:         cfg.Webauthn.RelyingParty.DisplayName,
		RPID:                  cfg.Webauthn.RelyingParty.Id,
		RPOrigin:              cfg.Webauthn.RelyingParty.Origin,
		AttestationPreference: protocol.PreferNoAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			RequireResidentKey: &f,
			ResidentKey:        protocol.ResidentKeyRequirementDiscouraged,
			UserVerification:   protocol.VerificationRequired,
		},
		Timeout: cfg.Webauthn.Timeout,
		Debug:   false,
	})

	return &AccountSharingHandler{
		mailer:          mailer,
		renderer:        renderer,
		nanoidGenerator: crypto.NewNanoidGenerator(),
		persister:       persister,
		emailConfig:     cfg.Passcode.Email, // TODO: Separate out into its own config value
		serviceConfig:   cfg.Service,
		sessionManager:  sessionManager,
		cfg:             cfg,
		webauthn:        wa,
	}, nil
}

type AccountShareRequest struct {
	Email           string `json:"email" validate:"required,email"`
	ExpireByTime    bool   `json:"expireByTime"`
	LifetimeMinutes int32  `json:"minutesAllowed"`
	ExpireByLogins  bool   `json:"expireByLogin"`
	LoginsAllowed   int32  `json:"loginsAllowed"`
}

func (h *AccountSharingHandler) BeginShare(c echo.Context) error {

	// Parse and validate request
	var request AccountShareRequest
	if err := (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return dto.ToHttpError(err)
	}
	if err := c.Validate(request); err != nil {
		return dto.ToHttpError(err)
	}

	// Parse UID from token
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("failed to cast session object")
	}
	uId, err := uuid.FromString(sessionToken.Subject())
	if err != nil {
		return fmt.Errorf("failed to parse userId from JWT subject:%w", err)
	}

	surrogateId, err := jwt2.GetSurrogateKeyFromToken(sessionToken)
	if err != nil {
		return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(fmt.Errorf("unable to get surrogate ID from token: %w", err))
	}

	if sessionToken.Subject() != surrogateId {
		return dto.NewHTTPError(http.StatusForbidden)
	}

	user, err := h.persister.GetUserPersister().Get(uId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return dto.NewHTTPError(http.StatusNotFound).SetInternal(errors.New("user not found"))
	}

	nanoidGenerator := crypto.NewNanoidGenerator()
	accessToken, err := nanoidGenerator.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate an access token: %w", err)
	}

	grantId, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to create grantId: %w", err)
	}
	now := time.Now().UTC()
	hashedAccessToken, err := bcrypt.GenerateFromPassword([]byte(accessToken), 12)
	if err != nil {
		return fmt.Errorf("failed to hash access token: %w", err)
	}
	accessGrantModel := models.AccountAccessGrant{
		ID:             grantId,
		UserId:         uId,
		Ttl:            60 * TimeToLiveMinutes,
		Token:          string(hashedAccessToken),
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
		ExpireByLogins: request.ExpireByLogins,
		LoginsAllowed:  sql.NullInt32{Int32: request.LoginsAllowed, Valid: request.ExpireByLogins},
		ExpireByTime:   request.ExpireByTime,
		MinutesAllowed: sql.NullInt32{Int32: request.LifetimeMinutes, Valid: request.ExpireByTime},
	}

	err = h.persister.GetAccountAccessGrantPersister().Create(accessGrantModel)
	if err != nil {
		return fmt.Errorf("failed to create access grant: %w", err)
	}

	lang := c.Request().Header.Get("Accept-Language")

	data := map[string]interface{}{
		"BaseUrl": "http://localhost:4200/#",
		"GrantId": grantId.String(),
		"Token":   accessToken,
		"TTL":     strconv.Itoa(TimeToLiveMinutes),
	}
	linkUrl := fmt.Sprintf("%s/share/%s?token=%s", data["BaseUrl"], data["GrantId"], data["Token"])
	str1, err := h.renderer.Render("accountShareSenderMail", lang, data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	messageToUser := gomail.NewMessage(gomail.SetEncoding(gomail.Base64))
	messageToUser.SetAddressHeader("To", user.Email, "")
	messageToUser.SetAddressHeader("From", "no-reply@hanko.io", "Hanko")
	messageToUser.SetHeader("Subject", "Access request provisioned for your account")
	messageToUser.SetBody("text/html", str1)

	str2, err := h.renderer.Render("accountShareReceiverMail", lang, data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	messageToReceiver := gomail.NewMessage(gomail.SetEncoding(gomail.Base64))
	messageToReceiver.SetAddressHeader("To", request.Email, "")
	messageToReceiver.SetAddressHeader("From", "no-reply@hanko.io", "Hanko")
	messageToReceiver.SetHeader("Subject", "You have been invited to access an account!")
	messageToReceiver.SetBody("text/html", str2)

	err = h.mailer.Send(messageToUser)
	if err != nil {
		return fmt.Errorf("failed to send passcode: %w", err)
	}

	err = h.mailer.Send(messageToReceiver)
	if err != nil {
		return fmt.Errorf("failed to send passcode: %w", err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"url": linkUrl,
	})
}

func (h *AccountSharingHandler) GetAccountShareGrantWithToken(grantId string, token string) error {
	startTime := time.Now().UTC()

	grantUid, err := uuid.FromString(grantId)
	if err != nil {
		return dto.NewHTTPError(http.StatusBadRequest, "failed to parse id as uuid").SetInternal(err)
	}

	var businessError error
	transactionError := h.persister.Transaction(func(tx *pop.Connection) error {
		grantPersister := h.persister.GetAccountAccessGrantPersister()
		grant, err := grantPersister.Get(grantUid)
		if err != nil || grant == nil {
			businessError = dto.NewHTTPError(http.StatusNotFound, "grant not found")
			return nil
		}

		if !grant.IsActive {
			businessError = dto.NewHTTPError(http.StatusNotFound, "grant is no longer active")
			return nil
		}

		expirationTime := grant.CreatedAt.Add(time.Duration(grant.Ttl) * time.Second)
		if expirationTime.Before(startTime) {
			businessError = dto.NewHTTPError(http.StatusRequestTimeout, "grant request timed out").SetInternal(fmt.Errorf("createdAt: %s -> lastVerificationTime: %s", grant.CreatedAt, expirationTime))
			return nil
		}

		err = bcrypt.CompareHashAndPassword([]byte(grant.Token), []byte(token))

		// Return same HTTP code for (grant ID not found) and (token invalid) to prevent disclosing which condition failed
		if err != nil {
			businessError = dto.NewHTTPError(http.StatusNotFound, "grant not found")
		}

		return nil
	})

	if businessError != nil {
		return businessError
	}

	return transactionError
}

type BeginCreateAccountWithGrantRequest struct {
	GuestUserId string `param:"guestUserId" validate:"required,uuid4"`
	GrantId     string `param:"grantId" validate:"required,uuid4"`
}

type BeginCreateAccountWithGrantResponse struct {
	Options *protocol.CredentialAssertion `json:"options"`
	Grant   GrantAttestationObject        `json:"grantAttestation"`
}

type GrantAttestationObject struct {
	AccountAccessGrantId uuid.UUID `json:"accountAccessGrantId"`
	GuestUserId          uuid.UUID `json:"guestUserId"`
	CreatedAt            time.Time `json:"createdAt"`
	ExpireByTime         bool      `json:"expireByTime"`
	ExpireByLogins       bool      `json:"expireByLogins"`
	MinutesAllowed       int       `json:"minutesAllowed"`
	LoginsAllowed        int       `json:"loginsAllowed"`
}

func (h *AccountSharingHandler) BeginCreateAccountWithGrant(c echo.Context) error {
	var request BeginCreateAccountWithGrantRequest
	if err := c.Bind(&request); err != nil {
		return dto.ToHttpError(err)
	}
	if err := c.Validate(request); err != nil {
		return dto.ToHttpError(err)
	}
	sessionToken, err := h.validateTokenForPrimaryAccountHolder(c)
	if err != nil {
		return err
	}

	guestUser, err := h.persister.GetUserPersister().Get(uuid.FromStringOrNil(request.GuestUserId))
	if err != nil || guestUser == nil {
		return dto.NewHTTPError(http.StatusNotFound).SetInternal(fmt.Errorf("unable to find user id %s", request.GuestUserId))
	}

	grant, err := h.persister.GetAccountAccessGrantPersister().Get(uuid.FromStringOrNil(request.GrantId))
	if err != nil || grant == nil {
		return dto.NewHTTPError(http.StatusNotFound).SetInternal(fmt.Errorf("unable to find a grant with ID %s", request.GrantId))
	}

	var options *protocol.CredentialAssertion
	var sessionData *webauthn.SessionData

	webauthnUser, err := h.getWebauthnUser(h.persister.GetConnection(), uuid.FromStringOrNil(sessionToken.Subject()))
	if err != nil || webauthnUser == nil {
		return dto.NewHTTPError(http.StatusNotFound).SetInternal(fmt.Errorf("an error occurred fetching webauthn user for user id %s: %w", sessionToken.Subject(), err))
	}

	if webauthnUser == nil {
		return dto.NewHTTPError(http.StatusBadRequest, "user not found")
	}

	if len(webauthnUser.WebAuthnCredentials()) > 0 {
		options, sessionData, err = h.webauthn.BeginLogin(webauthnUser, webauthn.WithUserVerification(protocol.VerificationRequired))
		if err != nil {
			return fmt.Errorf("failed to create webauthn assertion options: %w", err)
		}
	}

	if options == nil && sessionData == nil {
		var err error
		options, sessionData, err = h.webauthn.BeginDiscoverableLogin(webauthn.WithUserVerification(protocol.VerificationRequired))
		if err != nil {
			return fmt.Errorf("failed to create webauthn assertion options for discoverable login: %w", err)
		}
	}

	err = h.persister.GetWebauthnSessionDataPersister().Create(*intern.WebauthnSessionDataToModel(sessionData, models.WebauthnOperationAuthentication))
	if err != nil {
		return fmt.Errorf("failed to store webauthn assertion session data: %w", err)
	}

	// Remove all transports, because of a bug in android and windows where the internal authenticator gets triggered,
	// when the transports array contains the type 'internal' although the credential is not available on the device.
	for i := range options.Response.AllowedCredentials {
		options.Response.AllowedCredentials[i].Transport = nil
	}

	grantAttestationObject := GrantAttestationObject{
		AccountAccessGrantId: grant.ID,
		GuestUserId:          guestUser.ID,
		CreatedAt:            grant.CreatedAt,
		ExpireByLogins:       grant.ExpireByLogins,
		ExpireByTime:         grant.ExpireByTime,
	}
	if grant.ExpireByTime {
		grantAttestationObject.MinutesAllowed = int(grant.MinutesAllowed.Int32)
	}
	if grant.ExpireByLogins {
		grantAttestationObject.LoginsAllowed = int(grant.LoginsAllowed.Int32)
	}

	return c.JSON(http.StatusOK, BeginCreateAccountWithGrantResponse{Options: options, Grant: grantAttestationObject})
}

type FinishCreateAccountWithGrantRequest struct {
	GuestUserId      string `param:"guestUserId" validate:"required,uuid4"`
	GrantId          string `param:"grantId" validate:"required,uuid4"`
	GrantAttestation string `param:"grantAttestation" validate:"required"`
}

func (h *AccountSharingHandler) FinishCreateAccountWithGrant(c echo.Context) error {
	startTime := time.Now().UTC()

	var bodyBytes []byte
	if c.Request().Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request().Body)
	}
	var body FinishCreateAccountWithGrantRequest
	err := json.Unmarshal(bodyBytes, &body)
	if err != nil {
		return dto.NewHTTPError(http.StatusBadRequest)
	}
	c.Validate(body)

	sessionToken, err := h.validateTokenForPrimaryAccountHolder(c)
	if err != nil {
		return err
	}

	// Because request body cannot be read more than once, we have to reset the request back to its original state
	// https://medium.com/@xoen/golang-read-from-an-io-readwriter-without-loosing-its-content-2c6911805361
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	request, err := protocol.ParseCredentialRequestResponse(c.Request())
	if err != nil {
		return dto.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	err, _, webauthnuser := h.validateWebauthnRequest(request)
	if err != nil {
		return err
	}

	grant, err := h.persister.GetAccountAccessGrantPersister().Get(uuid.FromStringOrNil(body.GrantId))
	if err != nil {
		return dto.NewHTTPError(http.StatusNotFound).SetInternal(fmt.Errorf("unable to find grant id: %s", body.GrantId))
	}

	if webauthnuser.UserId.String() != sessionToken.Subject() {
		return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(fmt.Errorf("webauthn user ID %s does not match session token user ID %s", webauthnuser.UserId, sessionToken.Subject()))
	}

	if grant.UserId.String() != sessionToken.Subject() {
		return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(fmt.Errorf("grant id %s does not belong to user ID %s", grant.ID, sessionToken.Subject()))
	}

	if !grant.IsActive {
		return dto.NewHTTPError(http.StatusNotFound).SetInternal(fmt.Errorf("grant id %s is no longer active", grant.ID))
	}

	if grant.CreatedAt.UTC().Add(time.Duration(grant.Ttl) * time.Second).Before(time.Now().UTC()) {
		return dto.NewHTTPError(http.StatusRequestTimeout).SetInternal(fmt.Errorf("grant id %s has expired", grant.ID))
	}

	guestUserId := uuid.FromStringOrNil(body.GuestUserId)
	primaryUserId := uuid.FromStringOrNil(sessionToken.Subject())

	existingUserGuestRelationships, err := h.persister.GetUserGuestRelationPersister().GetByGuestUserId(&guestUserId)
	if err != nil {
		fmt.Println("an error occurred fetching existing user guest relationships: ", err)
		return err
	}

	for _, relation := range existingUserGuestRelationships {
		if relation.ParentUserID == primaryUserId && relation.IsActive {
			return dto.NewHTTPError(http.StatusConflict).SetInternal(fmt.Errorf("an existing user guest relationship exists for this guest and primary pair: %s", relation.ID))
		}
	}

	relationId, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("unable to generate new UUID: %w", err)
	}

	grant.IsActive = false
	grant.UpdatedAt = startTime
	grant.ClaimedBy = &guestUserId
	grant.UserGuestRelationId = &relationId

	h.persister.GetAccountAccessGrantPersister().Update(*grant)

	hash := []byte(body.GrantAttestation)

	userGuestRelation := models.UserGuestRelation{
		ID:                      relationId,
		ParentUserID:            primaryUserId,
		GuestUserID:             guestUserId,
		ExpireByLogins:          grant.ExpireByLogins,
		LoginsAllowed:           grant.LoginsAllowed,
		ExpireByTime:            grant.ExpireByTime,
		MinutesAllowed:          grant.MinutesAllowed,
		CreatedAt:               startTime,
		UpdatedAt:               startTime,
		AssociatedAccessGrantId: grant.ID,
		IsActive:                true,
		GrantHash:               &hash,
	}

	h.persister.GetUserGuestRelationPersister().Create(userGuestRelation)

	return c.JSON(http.StatusOK, struct{}{})
}

func (*AccountSharingHandler) validateTokenForPrimaryAccountHolder(c echo.Context) (jwt.Token, error) {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return nil, dto.NewHTTPError(http.StatusUnauthorized)
	}
	surrogateId, err := jwt2.GetSurrogateKeyFromToken(sessionToken)
	if err != nil {
		return nil, dto.NewHTTPError(http.StatusUnauthorized).SetInternal(fmt.Errorf("unable to get surrogate ID from token: %w", err))
	}
	if sessionToken.Subject() != surrogateId {
		return nil, dto.NewHTTPError(http.StatusForbidden).SetInternal(fmt.Errorf("call cannot be made by a guest user"))
	}
	return sessionToken, nil
}

func (h AccountSharingHandler) getWebauthnUser(connection *pop.Connection, userId uuid.UUID) (*intern.WebauthnUser, error) {
	user, err := h.persister.GetUserPersisterWithConnection(connection).Get(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, nil
	}

	credentials, err := h.persister.GetWebauthnCredentialPersisterWithConnection(connection).GetFromUser(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get webauthn credentials: %w", err)
	}

	return intern.NewWebauthnUser(*user, credentials), nil
}

func (h AccountSharingHandler) validateWebauthnRequest(request *protocol.ParsedCredentialAssertionData) (error, *webauthn.Credential, *intern.WebauthnUser) {
	var credential *webauthn.Credential
	var webauthnUser *intern.WebauthnUser
	err := h.persister.Transaction(func(tx *pop.Connection) error {
		sessionDataPersister := h.persister.GetWebauthnSessionDataPersisterWithConnection(tx)
		sessionData, err := sessionDataPersister.GetByChallenge(request.Response.CollectedClientData.Challenge)
		if err != nil {
			return fmt.Errorf("failed to get webauthn assertion session data: %w", err)
		}

		if sessionData != nil && sessionData.Operation != models.WebauthnOperationAuthentication {
			sessionData = nil
		}

		if sessionData == nil {
			return dto.NewHTTPError(http.StatusUnauthorized, "Stored challenge and received challenge do not match").SetInternal(errors.New("sessionData not found"))
		}

		model := intern.WebauthnSessionDataFromModel(sessionData)

		if sessionData.UserId.IsNil() {
			// Discoverable Login
			userId, err := uuid.FromBytes(request.Response.UserHandle)
			if err != nil {
				return dto.NewHTTPError(http.StatusBadRequest, "failed to parse userHandle as uuid").SetInternal(err)
			}
			webauthnUser, err = h.getWebauthnUser(tx, userId)
			if err != nil {
				return fmt.Errorf("failed to get user: %w", err)
			}

			if webauthnUser == nil {
				return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(errors.New("user not found"))
			}

			credential, err = h.webauthn.ValidateDiscoverableLogin(func(rawID, userHandle []byte) (user webauthn.User, err error) {
				return webauthnUser, nil
			}, *model, request)
			if err != nil {
				return dto.NewHTTPError(http.StatusUnauthorized, "failed to validate assertion").SetInternal(err)
			}
		} else {
			// non discoverable Login
			webauthnUser, err = h.getWebauthnUser(tx, sessionData.UserId)
			if err != nil {
				return fmt.Errorf("failed to get user: %w", err)
			}
			if webauthnUser == nil {
				return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(errors.New("user not found"))
			}
			credential, err = h.webauthn.ValidateLogin(webauthnUser, *model, request)
			if err != nil {
				return dto.NewHTTPError(http.StatusUnauthorized, "failed to validate assertion").SetInternal(err)
			}
		}

		err = sessionDataPersister.Delete(*sessionData)
		if err != nil {
			return fmt.Errorf("failed to delete assertion session data: %w", err)
		}
		return nil
	})
	return err, credential, webauthnUser
}
