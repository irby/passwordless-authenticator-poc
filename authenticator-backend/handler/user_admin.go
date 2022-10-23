package handler

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	jwt2 "github.com/teamhanko/hanko/backend/crypto/jwt"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"net/http"
	"strings"
)

type UserHandlerAdmin struct {
	persister persistence.Persister
}

func NewUserHandlerAdmin(persister persistence.Persister) *UserHandlerAdmin {
	return &UserHandlerAdmin{persister: persister}
}

func (h *UserHandlerAdmin) Delete(c echo.Context) error {
	userId, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return dto.NewHTTPError(http.StatusBadRequest, "failed to parse userId as uuid").SetInternal(err)
	}

	err, isSuccess := h.validateAdminPermission(c)
	if !isSuccess {
		return err
	}

	p := h.persister.GetUserPersister()
	user, err := p.Get(userId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return dto.NewHTTPError(http.StatusNotFound, "user not found")
	}

	err = p.Delete(*user)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return c.JSON(http.StatusNoContent, nil)
}

type UserPatchRequest struct {
	UserId   string `param:"id" validate:"required,uuid4"`
	Email    string `json:"email" validate:"omitempty,email"`
	Verified *bool  `json:"verified"`
}

func (h *UserHandlerAdmin) Patch(c echo.Context) error {
	var patchRequest UserPatchRequest
	if err := c.Bind(&patchRequest); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(patchRequest); err != nil {
		return dto.ToHttpError(err)
	}

	err, isSuccess := h.validateAdminPermission(c)
	if !isSuccess {
		return err
	}

	patchRequest.Email = strings.ToLower(patchRequest.Email)

	p := h.persister.GetUserPersister()
	user, err := p.Get(uuid.FromStringOrNil(patchRequest.UserId))
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return dto.NewHTTPError(http.StatusNotFound, "user not found")
	}

	if patchRequest.Email != "" && patchRequest.Email != user.Email {
		maybeExistingUser, err := p.GetByEmail(patchRequest.Email)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if maybeExistingUser != nil {
			return dto.NewHTTPError(http.StatusBadRequest, "email address not available")
		}

		user.Email = patchRequest.Email
	}

	if patchRequest.Verified != nil {
		user.Verified = *patchRequest.Verified
	}

	err = p.Update(*user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return c.JSON(http.StatusOK, nil) // TODO: mabye we should return the user object???
}

type UserListRequest struct {
	PerPage int `query:"per_page"`
	Page    int `query:"page"`
}

func (h *UserHandlerAdmin) List(c echo.Context) error {
	// TODO: return 'X-Total-Count' header, which includes the all users count
	// TODO; return 'Link' header, which includes links to next, previous, current(???), first, last page (example https://docs.github.com/en/rest/guides/traversing-with-pagination)
	var request UserListRequest
	err := (&echo.DefaultBinder{}).BindQueryParams(c, &request)
	if err != nil {
		return dto.ToHttpError(err)
	}

	err, isSuccess := h.validateAdminPermission(c)
	if !isSuccess {
		return err
	}

	users, err := h.persister.GetUserPersister().List(request.Page, request.PerPage)
	if err != nil {
		return fmt.Errorf("failed to get lsist of users: %w", err)
	}

	return c.JSON(http.StatusOK, users)
}

func (h *UserHandlerAdmin) validateAdminPermission(c echo.Context) (error, bool) {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return dto.NewHTTPError(http.StatusUnauthorized), false
	}

	surrogateId, err := jwt2.GetSurrogateKeyFromToken(sessionToken)
	if err != nil {
		return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(fmt.Errorf("unable to get surrogate ID from token: %w", err)), false
	}

	if sessionToken.Subject() != surrogateId {
		return dto.NewHTTPError(http.StatusForbidden), false
	}

	currentUser, err := h.persister.GetUserPersister().Get(uuid.FromStringOrNil(sessionToken.Subject()))

	if !currentUser.IsAdmin || !currentUser.IsActive {
		return dto.NewHTTPError(http.StatusForbidden), false
	}

	return nil, true
}
