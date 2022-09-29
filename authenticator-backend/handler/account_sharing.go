package handler

import (
	"errors"
	"fmt"
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

	messageToUser := gomail.NewMessage()
	messageToUser.SetAddressHeader("To", user.Email, "")
	messageToUser.SetAddressHeader("From", "no-reply@hanko.io", "Hanko")
	messageToUser.SetHeader("Subject", "Access request provisioned for your account")
	messageToUser.SetBody("text/plain", "A request to access your account has been provisioned. If you initiated this request, please ignore this email. If you did not initiate this request, please contact us immediately.")

	messageToReceiver := gomail.NewMessage()
	messageToReceiver.SetAddressHeader("To", request.Email, "")
	messageToReceiver.SetAddressHeader("From", "no-reply@hanko.io", "Hanko")
	messageToReceiver.SetHeader("Subject", "You have been invited to access an account!")
	messageToReceiver.SetBody("text/html", "A user has invited to share their account with you. Please visit the link below to initiate sharing: <a href=\"http://localhost:8000/share/"+accessToken+"\">http://localhost:8000/share/"+accessToken+"</a>")

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
