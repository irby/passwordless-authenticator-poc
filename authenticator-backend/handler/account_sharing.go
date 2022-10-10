package handler

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/mail"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"net/http"
	"strconv"
	"time"
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

const TimeToLiveMinutes = 15 // TODO: make into a config value

func NewAccountSharingHandler(cfg *config.Config, persister persistence.Persister, sessionManager session.Manager, mailer mail.Mailer) (*AccountSharingHandler, error) {
	renderer, err := mail.NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("failed to create new renderer: %w", err)
	}
	return &AccountSharingHandler{
		mailer:          mailer,
		renderer:        renderer,
		nanoidGenerator: crypto.NewNanoidGenerator(),
		persister:       persister,
		emailConfig:     cfg.Passcode.Email, // TODO: Separate out into its own config value
		serviceConfig:   cfg.Service,
		sessionManager:  sessionManager,
		cfg:             cfg,
	}, nil
}

type AccountShareRequest struct {
	Email           string `json:"email" validate:"required,email"`
	ExpireByTime    bool   `json:"expireByMinutes"`
	LifetimeMinutes int32  `json:"expireTimeMinutes"`
	ExpireByLogins  bool   `json:"expireByLogins"`
	LoginsAllowed   int32  `json:"loginsAllowed"`
}

func (h *AccountSharingHandler) BeginShare(c echo.Context) error {

	// Parse and validate request
	var request UserGetByEmailBody
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
		ID:        grantId,
		UserId:    uId,
		Ttl:       60 * TimeToLiveMinutes,
		Token:     string(hashedAccessToken),
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = h.persister.GetAccountAccessGrantPersister().Create(accessGrantModel)
	if err != nil {
		return fmt.Errorf("failed to create access grant: %w", err)
	}

	lang := c.Request().Header.Get("Accept-Language")

	data := map[string]interface{}{}
	str1, err := h.renderer.Render("accountShareSenderMail", lang, data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	messageToUser := gomail.NewMessage(gomail.SetEncoding(gomail.Base64))
	messageToUser.SetAddressHeader("To", user.Email, "")
	messageToUser.SetAddressHeader("From", "no-reply@hanko.io", "Hanko")
	messageToUser.SetHeader("Subject", "Access request provisioned for your account")
	messageToUser.SetBody("text/plain", str1)

	data = map[string]interface{}{
		"BaseUrl": "http://localhost:4200/#",
		"GrantId": grantId.String(),
		"Token":   accessToken,
		"TTL":     strconv.Itoa(TimeToLiveMinutes),
	}
	str2, err := h.renderer.Render("accountShareReceiverMail", lang, data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	fmt.Println("Receiver email body", str2)

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

	return c.JSON(http.StatusOK, map[string]string{})
}

func (h *AccountSharingHandler) GetAccountShareGrantWithToken(c echo.Context) error {
	grantId := c.Param("id")
	token := c.QueryParam("token")

	if grantId == "" || token == "" {
		return dto.NewHTTPError(http.StatusBadRequest, "id and token are both required")
	}

	startTime := time.Now().UTC()

	grantUid, err := uuid.FromString(grantId)
	if err != nil {
		return dto.NewHTTPError(http.StatusBadRequest, "failed to parse id as uuid").SetInternal(err)
	}

	var businessError error
	transactionError := h.persister.Transaction(func(tx *pop.Connection) error {
		grantPersister := h.persister.GetAccountAccessGrantPersister()
		grant, err := grantPersister.Get(grantUid)
		if err != nil {
			businessError = dto.NewHTTPError(http.StatusNotFound, "grant not found")
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
			return nil
		}

		return c.JSON(http.StatusOK, nil)
	})

	if businessError != nil {
		return businessError
	}

	return transactionError
}

func (h *AccountSharingHandler) CreateAccountWithGrant(grantId uuid.UUID, primaryUserId uuid.UUID, guestUserId uuid.UUID) error {
	currentTime := time.Now().UTC()
	grant, err := h.persister.GetAccountAccessGrantPersister().Get(grantId)

	if err != nil {
		fmt.Println("Unable to find grant: ", err)
		return err
	}

	if primaryUserId != grant.UserId {
		return errors.New("primary user ID does not match grant's user ID")
	}

	if guestUserId == primaryUserId {
		return errors.New("guest ID cannot equal primary user ID")
	}

	if currentTime.After(grant.CreatedAt.Add(time.Duration(grant.Ttl) * time.Second)) {
		return errors.New("grant has expired")
	}

	grant.IsActive = false
	grant.UpdatedAt = currentTime

	return nil
}
