package handler

import (
	"errors"
	"fmt"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/config"
	jwt2 "github.com/teamhanko/hanko/backend/crypto/jwt"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/dto/intern"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"net/http"
	"strings"
	"time"
)

type UserHandler struct {
	persister      persistence.Persister
	sessionManager session.Manager
	webauthn       *webauthn.WebAuthn
}

func NewUserHandler(cfg *config.Config, persister persistence.Persister, sessionManager session.Manager) *UserHandler {
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
	return &UserHandler{persister: persister, sessionManager: sessionManager, webauthn: wa}
}

type UserCreateBody struct {
	Email string `json:"email" validate:"required,email"`
}

func (h *UserHandler) Create(c echo.Context) error {
	var body UserCreateBody
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	body.Email = strings.ToLower(body.Email)

	return h.persister.Transaction(func(tx *pop.Connection) error {
		user, err := h.persister.GetUserPersisterWithConnection(tx).GetByEmail(body.Email)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if user != nil {
			return dto.NewHTTPError(http.StatusConflict).SetInternal(errors.New(fmt.Sprintf("user with email %s already exists", user.Email)))
		}

		newUser := models.NewUser(body.Email)
		err = h.persister.GetUserPersisterWithConnection(tx).Create(newUser)
		if err != nil {
			return fmt.Errorf("failed to store user: %w", err)
		}

		return c.JSON(http.StatusOK, newUser)
	})
}

func (h *UserHandler) Get(c echo.Context) error {
	userId := c.Param("id")

	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("missing or malformed jwt")
	}

	if sessionToken.Subject() != userId {
		return dto.NewHTTPError(http.StatusForbidden).SetInternal(errors.New(fmt.Sprintf("user %s tried to get user %s", sessionToken.Subject(), userId)))
	}

	user, err := h.persister.GetUserPersister().Get(uuid.FromStringOrNil(userId))
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return dto.NewHTTPError(http.StatusNotFound).SetInternal(errors.New("user not found"))
	}

	return c.JSON(http.StatusOK, user)
}

type UserGetByEmailBody struct {
	Email string `json:"email" validate:"required,email"`
}

func (h *UserHandler) GetUserIdByEmail(c echo.Context) error {
	var request UserGetByEmailBody
	if err := (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(request); err != nil {
		return dto.ToHttpError(err)
	}

	user, err := h.persister.GetUserPersister().GetByEmail(strings.ToLower(request.Email))
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return dto.NewHTTPError(http.StatusNotFound).SetInternal(errors.New("user not found"))
	}

	return c.JSON(http.StatusOK, struct {
		UserId                string `json:"id"`
		Verified              bool   `json:"verified"`
		HasWebauthnCredential bool   `json:"has_webauthn_credential"`
	}{
		UserId:                user.ID.String(),
		Verified:              user.Verified,
		HasWebauthnCredential: len(user.WebauthnCredentials) > 0,
	})
}

type GetUserGuestRelationDto struct {
	GrantId         uuid.UUID `json:"relationId"`
	GuestUserId     uuid.UUID `json:"guestUserId"`
	GuestUserEmail  string    `json:"guestUserEmail"`
	ParentUserId    uuid.UUID `json:"parentUserId"`
	ParentUserEmail string    `json:"parentUserEmail"`
	CreatedAt       time.Time `json:"createdAt"`
	IsActive        bool      `json:"isActive"`
}

func (h *UserHandler) GetUserGuestRelationsAsGuest(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("missing or malformed jwt")
	}

	surrogateId, err := jwt2.GetSurrogateKeyFromToken(sessionToken)
	if err != nil {
		return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(fmt.Errorf("unable to get surrogate ID from token: %w", err))
	}

	uuid := uuid.FromStringOrNil(surrogateId)

	guestGrants, err := h.persister.GetUserGuestRelationPersister().GetByGuestUserId(&uuid)
	if err != nil {
		return dto.NewHTTPError(http.StatusInternalServerError).SetInternal(errors.New("could not get guest grants"))
	}

	result := []GetUserGuestRelationDto{}

	userPersister := h.persister.GetUserPersister()

	guestUser, _ := userPersister.Get(uuid)

	for i := 0; i < len(guestGrants); i++ {
		grant := guestGrants[i]
		parentUser, _ := userPersister.Get(grant.ParentUserID)
		intermediate := GetUserGuestRelationDto{
			GrantId:         grant.ID,
			GuestUserId:     grant.GuestUserID,
			GuestUserEmail:  guestUser.Email,
			ParentUserId:    grant.ParentUserID,
			ParentUserEmail: parentUser.Email,
			CreatedAt:       grant.CreatedAt,
			IsActive:        grant.IsActive,
		}
		result = append(result, intermediate)
	}

	return c.JSON(http.StatusOK, result)

	return nil
}

func (h *UserHandler) GetUserGuestRelationsAsAccountHolder(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("missing or malformed jwt")
	}

	surrogateId, err := jwt2.GetSurrogateKeyFromToken(sessionToken)
	if err != nil {
		return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(fmt.Errorf("unable to get surrogate ID from token: %w", err))
	}

	uuid := uuid.FromStringOrNil(surrogateId)

	parentGrants, err := h.persister.GetUserGuestRelationPersister().GetByParentUserId(&uuid)
	if err != nil {
		return dto.NewHTTPError(http.StatusInternalServerError).SetInternal(errors.New("could not get parent grants"))
	}

	result := []GetUserGuestRelationDto{}

	userPersister := h.persister.GetUserPersister()

	parentUser, _ := userPersister.Get(uuid)

	for i := 0; i < len(parentGrants); i++ {
		grant := parentGrants[i]
		guestUser, _ := userPersister.Get(grant.GuestUserID)
		intermediate := GetUserGuestRelationDto{
			GrantId:         grant.ID,
			GuestUserId:     grant.GuestUserID,
			GuestUserEmail:  guestUser.Email,
			ParentUserId:    grant.ParentUserID,
			ParentUserEmail: parentUser.Email,
			CreatedAt:       grant.CreatedAt,
			IsActive:        grant.IsActive,
		}
		result = append(result, intermediate)
	}

	return c.JSON(http.StatusOK, result)

	return nil
}

func (h *UserHandler) GetUserGuestRelationsOverview(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("missing or malformed jwt")
	}

	surrogateId, err := jwt2.GetSurrogateKeyFromToken(sessionToken)
	if err != nil {
		return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(fmt.Errorf("unable to get surrogate ID from token: %w", err))
	}

	uuId := uuid.FromStringOrNil(surrogateId)
	actingUserId := uuid.FromStringOrNil(sessionToken.Subject())

	if uuId != actingUserId {
		return c.JSON(http.StatusOK, struct {
			HasGuestGrants  bool `json:"hasGuestGrants"`
			HasParentGrants bool `json:"hasParentGrants"`
		}{
			HasGuestGrants:  false,
			HasParentGrants: false,
		})
	}

	guestGrants, err := h.persister.GetUserGuestRelationPersister().GetByGuestUserId(&uuId)
	if err != nil {
		return dto.NewHTTPError(http.StatusInternalServerError).SetInternal(errors.New("could not get guest grants"))
	}

	parentGrants, err := h.persister.GetUserGuestRelationPersister().GetByParentUserId(&uuId)
	if err != nil {
		return dto.NewHTTPError(http.StatusInternalServerError).SetInternal(errors.New("could not get parent grants"))
	}

	return c.JSON(http.StatusOK, struct {
		HasGuestGrants  bool `json:"hasGuestGrants"`
		HasParentGrants bool `json:"hasParentGrants"`
	}{
		HasGuestGrants:  len(guestGrants) > 0,
		HasParentGrants: len(parentGrants) > 0,
	})

	return nil
}

type UserGuestRelationRequest struct {
	RelationId string `json:"relationId" validate:"required"`
}

func (h *UserHandler) BeginLoginAsGuest(c echo.Context) error {
	sessionToken, err := h.parseAndValidateToken(c, true)
	if err != nil {
		return err
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

func (h *UserHandler) FinishLoginAsGuest(c echo.Context) error {
	return nil
}

func (h *UserHandler) InitiateLoginAsGuest(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("missing or malformed jwt")
	}

	surrogateId, err := jwt2.GetSurrogateKeyFromToken(sessionToken)
	if err != nil {
		return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(fmt.Errorf("unable to get surrogate ID from token: %w", err))
	}

	var body UserGuestRelationRequest
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	relation, err := h.persister.GetUserGuestRelationPersister().Get(uuid.FromStringOrNil(body.RelationId))
	if err != nil {
		return dto.ToHttpError(err)
	}

	if relation == nil {
		return dto.NewHTTPError(http.StatusNotFound).SetInternal(errors.New("user guest relation not found"))
	}

	// Check to verify guest user ID matches the ID coming over on request
	if relation.GuestUserID.String() != surrogateId {
		return dto.NewHTTPError(http.StatusForbidden).SetInternal(errors.New(fmt.Sprintf("User ID %s does not have access to assume guest relation ID %s", surrogateId, relation.ID)))
	}

	if relation.ExpireByTime && time.Now().UTC().After(relation.CreatedAt.UTC().Add(time.Duration(relation.MinutesAllowed.Int32)*time.Minute)) {
		relation.IsActive = false
		relation.UpdatedAt = time.Now().UTC()

		_ = h.persister.GetUserGuestRelationPersister().Update(*relation)

		return dto.NewHTTPError(http.StatusForbidden).SetInternal(errors.New(fmt.Sprintf("Access on relation ID %s has expired", relation.ID)))
	}

	if relation.ExpireByLogins {
		models, err := h.persister.GetLoginAuditLogPersister().GetByGuestUserIdAndGrantId(relation.GuestUserID, relation.ID)
		if err != nil {
			return dto.NewHTTPError(http.StatusInternalServerError).SetInternal(fmt.Errorf("an error occurred while fetching login audit records: %w", err))
		}
		if int32(len(models)) >= relation.LoginsAllowed.Int32 {
			relation.IsActive = false
			relation.UpdatedAt = time.Now().UTC()

			_ = h.persister.GetUserGuestRelationPersister().Update(*relation)

			return dto.NewHTTPError(http.StatusForbidden).SetInternal(fmt.Errorf("access on relation ID %s has expired", relation.ID))
		}
	}

	token, err := h.sessionManager.GenerateJWT(relation.ParentUserID, relation.GuestUserID, relation.ID)
	if err != nil {
		return fmt.Errorf("failed to generate jwt: %w", err)
	}

	cookie, err := h.sessionManager.GenerateCookie(token)
	if err != nil {
		return fmt.Errorf("failed to create session token: %w", err)
	}

	c.SetCookie(cookie)

	log := models.LoginAuditLog{
		UserId:              relation.ParentUserID,
		SurrogateUserId:     &relation.GuestUserID,
		UserGuestRelationId: &relation.ID,
		ClientIpAddress:     c.Request().RemoteAddr,
		ClientUserAgent:     c.Request().UserAgent(),
		LoginMethod:         dto.LoginMethodToValue(dto.Webauthn),
	}
	err = h.persister.GetLoginAuditLogPersister().Create(log)
	if err != nil {
		return dto.NewHTTPError(http.StatusInternalServerError, "An error occurred generating login audit record", err.Error())
	}

	return c.JSON(http.StatusOK, struct{}{})

	return nil
}

func (h *UserHandler) LogoutAsGuest(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("missing or malformed jwt")
	}

	surrogateId, err := jwt2.GetSurrogateKeyFromToken(sessionToken)
	if err != nil {
		return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(fmt.Errorf("unable to get surrogate ID from token: %w", err))
	}

	if sessionToken.Subject() == surrogateId {
		return dto.NewHTTPError(http.StatusForbidden)
	}

	uId := uuid.FromStringOrNil(surrogateId)

	token, err := h.sessionManager.GenerateJWT(uId, uId, uuid.Nil)
	if err != nil {
		return fmt.Errorf("failed to generate jwt: %w", err)
	}

	cookie, err := h.sessionManager.GenerateCookie(token)
	if err != nil {
		return fmt.Errorf("failed to create session token: %w", err)
	}

	c.SetCookie(cookie)

	log := models.LoginAuditLog{
		UserId:          uuid.FromStringOrNil(surrogateId),
		ClientIpAddress: c.Request().RemoteAddr,
		ClientUserAgent: c.Request().UserAgent(),
		LoginMethod:     dto.LoginMethodToValue(dto.LogoutAsGuest),
	}
	err = h.persister.GetLoginAuditLogPersister().Create(log)
	if err != nil {
		return dto.NewHTTPError(http.StatusInternalServerError, "An error occurred generating login audit record", err.Error())
	}

	return c.JSON(http.StatusOK, struct{}{})
}

func (h *UserHandler) RemoveAccessToRelation(c echo.Context) error {
	relationId := c.Param("id")

	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("missing or malformed jwt")
	}

	surrogateId, err := jwt2.GetSurrogateKeyFromToken(sessionToken)
	if err != nil {
		return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(fmt.Errorf("unable to get surrogate ID from token: %w", err))
	}

	relation, err := h.persister.GetUserGuestRelationPersister().Get(uuid.FromStringOrNil(relationId))
	if err != nil {
		return dto.ToHttpError(err)
	}

	if relation == nil {
		return dto.NewHTTPError(http.StatusNotFound).SetInternal(errors.New("user guest relation not found"))
	}

	// Check to verify parent user ID matches the ID coming over on request
	if relation.ParentUserID.String() != surrogateId {
		return dto.NewHTTPError(http.StatusForbidden).SetInternal(errors.New(fmt.Sprintf("User ID %s does not have access to assume guest relation ID %s", surrogateId, relation.ID)))
	}

	relation.IsActive = false
	relation.UpdatedAt = time.Now().UTC()

	err = h.persister.GetUserGuestRelationPersister().Update(*relation)
	if err != nil {
		return dto.NewHTTPError(http.StatusInternalServerError).SetInternal(errors.New(fmt.Sprintf("An error occurred while updating the relation ID %s", relation.ID)))
	}

	return c.JSON(http.StatusOK, struct{}{})

	return nil
}

type MeResponseDto struct {
	Id              uuid.UUID `json:"id"`
	Email           string    `json:"email"`
	IsAccountHolder bool      `json:"isAccountHolder"`
	IsAdmin         bool      `json:"isAdmin"`
}

func (h *UserHandler) Me(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("failed to cast session object")
	}

	user, err := h.persister.GetUserPersister().Get(uuid.FromStringOrNil(sessionToken.Subject()))
	if err != nil {
		return dto.NewHTTPError(http.StatusNotFound).SetInternal(fmt.Errorf("user id %s could not found: %w", sessionToken.Subject(), err))
	}

	surrogateId, _ := jwt2.GetSurrogateKeyFromToken(sessionToken)
	isAdmin := false
	if sessionToken.Subject() == surrogateId {
		isAdmin = user.IsAdmin
	}

	dto := MeResponseDto{
		Email:           user.Email,
		Id:              user.ID,
		IsAccountHolder: sessionToken.Subject() == surrogateId,
		IsAdmin:         isAdmin,
	}

	return c.JSON(http.StatusOK, dto)
}

func (h *UserHandler) Logout(c echo.Context) error {
	_, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("failed to cast session object")
	}

	cookie, err := h.sessionManager.DeleteCookie()
	if err != nil {
		return fmt.Errorf("failed to create session token: %w", err)
	}

	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, map[string]string{})
}

func (h *UserHandler) parseAndValidateToken(c echo.Context, shouldBeAccountHolder bool) (jwt.Token, error) {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok || sessionToken == nil {
		return nil, dto.NewHTTPError(http.StatusUnauthorized).SetInternal(fmt.Errorf("invalid or missing session token"))
	}
	surrogateId, err := jwt2.GetSurrogateKeyFromToken(sessionToken)
	if err != nil || surrogateId == "" {
		return nil, dto.NewHTTPError(http.StatusUnauthorized).SetInternal(fmt.Errorf("invalid or missing surrogate id"))
	}
	if shouldBeAccountHolder && sessionToken.Subject() != surrogateId {
		return nil, dto.NewHTTPError(http.StatusForbidden).SetInternal(fmt.Errorf("should be called by primary account holder"))
	}
	return sessionToken, nil
}

func (h *UserHandler) getWebauthnUser(connection *pop.Connection, userId uuid.UUID) (*intern.WebauthnUser, error) {
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
