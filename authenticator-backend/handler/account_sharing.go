package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/teamhanko/hanko/backend/dto/intern"
	"net/http"
	"strconv"
	"time"

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
	wa, err := webauthn.New(&webauthn.Config{
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
			businessError = dto.NewHTTPError(http.StatusRequestTimeout, "grant request timed out").SetInternal(errors.New(fmt.Sprintf("createdAt: %s -> lastVerificationTime: %s", grant.CreatedAt, expirationTime)))
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

	return c.JSON(http.StatusOK, options)
}

func (h *AccountSharingHandler) FinishCreateAccountWithGrant(c echo.Context) error {
	return nil
	//var request BeginCreateAccountWithGrantRequest
	//if err := c.Bind(&request); err != nil {
	//	return dto.ToHttpError(err)
	//}
	//if err := c.Validate(request); err != nil {
	//	return dto.ToHttpError(err)
	//}
	//sessionToken, err := h.validateTokenForPrimaryAccountHolder(c)
	//if err != nil {
	//	return err
	//}
	//
	//var options *protocol.CredentialAssertion
	//var sessionData *webauthn.SessionData
	//
	//webauthnUser, err := h.getWebauthnUser(h.persister.GetConnection(), uuid.FromStringOrNil(sessionToken.Subject()))
	//
	//if webauthnUser == nil {
	//	return dto.NewHTTPError(http.StatusBadRequest, "user not found")
	//}
	//
	//if len(webauthnUser.WebAuthnCredentials()) > 0 {
	//	options, sessionData, err = h.webauthn.BeginLogin(webauthnUser, webauthn.WithUserVerification(protocol.VerificationRequired))
	//	if err != nil {
	//		return fmt.Errorf("failed to create webauthn assertion options: %w", err)
	//	}
	//}
	//
	//if options == nil && sessionData == nil {
	//	var err error
	//	options, sessionData, err = h.webauthn.BeginDiscoverableLogin(webauthn.WithUserVerification(protocol.VerificationRequired))
	//	if err != nil {
	//		return fmt.Errorf("failed to create webauthn assertion options for discoverable login: %w", err)
	//	}
	//}
	//
	//err = h.persister.GetWebauthnSessionDataPersister().Create(*intern.WebauthnSessionDataToModel(sessionData, models.WebauthnOperationAuthentication))
	//if err != nil {
	//	return fmt.Errorf("failed to store webauthn assertion session data: %w", err)
	//}
	//
	//// Remove all transports, because of a bug in android and windows where the internal authenticator gets triggered,
	//// when the transports array contains the type 'internal' although the credential is not available on the device.
	//for i := range options.Response.AllowedCredentials {
	//	options.Response.AllowedCredentials[i].Transport = nil
	//}
	//
	//return c.JSON(http.StatusOK, options)
}

func (h *AccountSharingHandler) CreateAccountWithGrant(grantId uuid.UUID, primaryUserId uuid.UUID, guestUserId uuid.UUID) error {
	startTime := time.Now().UTC()
	grant, err := h.persister.GetAccountAccessGrantPersister().Get(grantId)

	if err != nil || grant == nil {
		return fmt.Errorf("unable to find grant: %w", err)
	}

	if primaryUserId != grant.UserId {
		return errors.New("primary user ID does not match grant's user ID")
	}

	if !grant.IsActive {
		return errors.New("grant is no longer active")
	}

	if guestUserId == primaryUserId {
		return errors.New("guest ID cannot equal primary user ID")
	}

	expirationTime := grant.CreatedAt.Add(time.Duration(grant.Ttl) * time.Second)
	if expirationTime.Before(startTime) {
		return errors.New("grant has expired")
	}

	existingUserGuestRelationships, err := h.persister.GetUserGuestRelationPersister().GetByGuestUserId(&guestUserId)
	if err != nil {
		fmt.Println("an error occurred fetching existing user guest relationships: ", err)
		return err
	}

	for _, relation := range existingUserGuestRelationships {
		if relation.ParentUserID == primaryUserId && relation.IsActive {
			return fmt.Errorf("an existing user guest relationship exists for this guest and primary pair: %s", relation.ID)
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
	}

	h.persister.GetUserGuestRelationPersister().Create(userGuestRelation)

	return nil
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
