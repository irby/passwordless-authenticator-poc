package handler

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	jwt2 "github.com/teamhanko/hanko/backend/crypto/jwt"
	"github.com/teamhanko/hanko/backend/dto"
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
}

func NewUserHandler(persister persistence.Persister, sessionManager session.Manager) *UserHandler {
	return &UserHandler{persister: persister, sessionManager: sessionManager}
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

	if relation.ExpireByTime && time.Now().UTC().Before(relation.CreatedAt.Add(time.Duration(relation.MinutesAllowed.Int32)*time.Minute)) {
		relation.IsActive = false
		relation.UpdatedAt = time.Now().UTC()

		_ = h.persister.GetUserGuestRelationPersister().Update(*relation)

		return dto.NewHTTPError(http.StatusForbidden).SetInternal(errors.New(fmt.Sprintf("Access on relation ID %s has expired", relation.ID)))
	}

	// TODO: Check the expire by logins

	token, err := h.sessionManager.GenerateJWT(relation.ParentUserID, relation.GuestUserID, relation.ID)
	if err != nil {
		return fmt.Errorf("failed to generate jwt: %w", err)
	}

	cookie, err := h.sessionManager.GenerateCookie(token)
	if err != nil {
		return fmt.Errorf("failed to create session token: %w", err)
	}

	c.SetCookie(cookie)

	// TODO: Record login on relation

	return c.JSON(http.StatusOK, struct{}{})

	return nil
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

func (h *UserHandler) Me(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("failed to cast session object")
	}

	return c.JSON(http.StatusOK, map[string]string{"id": sessionToken.Subject()})
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
