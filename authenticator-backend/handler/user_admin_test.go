package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
)

func TestUserHandlerAdmin_Delete(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		{
			ID:        userId,
			Email:     "john.doe@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(userId.String())

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)
	addUsers(users, persister)

	handler := NewUserHandlerAdmin(persister)

	if assert.NoError(t, handler.Delete(c)) {
		assert.Equal(t, http.StatusNoContent, rec.Code)
	}
}

func TestUserHandlerAdmin_Delete_InvalidUserId(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues("invalidId")

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	handler := NewUserHandlerAdmin(persister)

	err := handler.Delete(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	}
}

func TestUserHandlerAdmin_Delete_UnknownUserId(t *testing.T) {
	userId, _ := uuid.NewV4()
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(userId.String())

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	handler := NewUserHandlerAdmin(persister)

	err := handler.Delete(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusNotFound, httpError.Code)
	}
}

func TestUserHandlerAdmin_ToggleIsActiveForUser_UserIsNotAdmin_Error(t *testing.T) {
	userId, _ := uuid.NewV4()
	e := echo.New()

	q := make(url.Values)
	q.Set("per_page", "invalidPerPageValue")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(userId.String())

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	adminUser.IsAdmin = false
	persister.GetUserPersister().Update(adminUser)
	persister.GetUserPersister().Create(models.User{
		ID:       userId,
		Email:    "testy@example.com",
		IsActive: true,
	})

	handler := NewUserHandlerAdmin(persister)
	err := handler.ToggleIsActiveForUser(c)
	assert.Error(t, err)
}

func TestUserHandlerAdmin_ToggleIsActiveForUser_UserIsAdmin_UserIsInactiveIfActiveBefore(t *testing.T) {
	userId, _ := uuid.NewV4()
	e := echo.New()

	q := make(url.Values)
	q.Set("per_page", "invalidPerPageValue")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(userId.String())

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	persister.GetUserPersister().Create(models.User{
		ID:       userId,
		Email:    "testy@example.com",
		IsActive: true,
	})

	handler := NewUserHandlerAdmin(persister)
	err := handler.ToggleIsActiveForUser(c)
	assert.NoError(t, err)
	u, err := persister.GetUserPersister().Get(userId)
	assert.Equal(t, false, u.IsActive)
}

func TestUserHandlerAdmin_ToggleIsActiveForUser_UserIsAdmin_UserIsActiveIfInactiveBefore(t *testing.T) {
	userId, _ := uuid.NewV4()
	e := echo.New()

	q := make(url.Values)
	q.Set("per_page", "invalidPerPageValue")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(userId.String())

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	persister.GetUserPersister().Create(models.User{
		ID:       userId,
		Email:    "testy@example.com",
		IsActive: false,
	})

	handler := NewUserHandlerAdmin(persister)
	err := handler.ToggleIsActiveForUser(c)
	assert.NoError(t, err)
	u, err := persister.GetUserPersister().Get(userId)
	assert.Equal(t, true, u.IsActive)
}

func TestUserHandlerAdmin_ToggleIsActiveForUser_UserIsAdmin_UserCannotToggleThemselves(t *testing.T) {
	e := echo.New()

	q := make(url.Values)
	q.Set("per_page", "invalidPerPageValue")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(adminUser.ID.String())

	handler := NewUserHandlerAdmin(persister)
	err := handler.ToggleIsActiveForUser(c)
	assert.Error(t, err)
	httpError := dto.ToHttpError(err)
	assert.Equal(t, http.StatusConflict, httpError.Code)
}

func TestUserHandlerAdmin_DeactivateGrantsForUser_UserIsNotAdmin_Error(t *testing.T) {
	userId, _ := uuid.NewV4()
	e := echo.New()

	q := make(url.Values)
	q.Set("per_page", "invalidPerPageValue")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(userId.String())

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	adminUser.IsAdmin = false
	persister.GetUserPersister().Update(adminUser)
	persister.GetUserPersister().Create(models.User{
		ID:       userId,
		Email:    "testy@example.com",
		IsActive: true,
	})
	grantPersister := persister.GetUserGuestRelationPersister()
	for i := 0; i < 4; i++ {
		uId, _ := uuid.NewV4()
		grantPersister.Create(models.UserGuestRelation{
			ID:           uId,
			ParentUserID: userId,
			IsActive:     true,
		})
	}

	handler := NewUserHandlerAdmin(persister)
	err := handler.DeactivateGrantsForUser(c)
	assert.Error(t, err)
}

func TestUserHandlerAdmin_DeactivateGrantsForUser_UserIsAdmin_DeactivatesGrants(t *testing.T) {
	userId, _ := uuid.NewV4()
	e := echo.New()

	q := make(url.Values)
	q.Set("per_page", "invalidPerPageValue")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(userId.String())

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	persister.GetUserPersister().Create(models.User{
		ID:       userId,
		Email:    "testy@example.com",
		IsActive: true,
	})
	grantPersister := persister.GetUserGuestRelationPersister()
	for i := 0; i < 4; i++ {
		uId, _ := uuid.NewV4()
		grantPersister.Create(models.UserGuestRelation{
			ID:           uId,
			ParentUserID: userId,
			IsActive:     true,
		})
	}

	handler := NewUserHandlerAdmin(persister)
	err := handler.DeactivateGrantsForUser(c)
	assert.NoError(t, err)

	grants, _ := grantPersister.GetByParentUserId(&userId)
	for _, grant := range grants {
		assert.False(t, grant.IsActive)
	}

}

func TestUserHandlerAdmin_Patch(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		{
			ID:        userId,
			Email:     "john.doe@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"email": "jane.doe@example.com", "verified": true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(userId.String())

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)
	addUsers(users, persister)

	handler := NewUserHandlerAdmin(persister)

	if assert.NoError(t, handler.Patch(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestUserHandlerAdmin_Patch_InvalidUserIdAndEmail(t *testing.T) {
	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"email": "invalidEmail"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues("invalidUserId")

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	handler := NewUserHandlerAdmin(persister)

	err := handler.Patch(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	}
}

func TestUserHandlerAdmin_Patch_EmailNotAvailable(t *testing.T) {
	users := []models.User{
		func() models.User {
			userId, _ := uuid.NewV4()
			return models.User{
				ID:        userId,
				Email:     "john.doe@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
		func() models.User {
			userId, _ := uuid.NewV4()
			return models.User{
				ID:        userId,
				Email:     "jane.doe@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"email": "jane.doe@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(users[0].ID.String())

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)
	addUsers(users, persister)

	handler := NewUserHandlerAdmin(persister)

	err := handler.Patch(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	}
}

func TestUserHandlerAdmin_Patch_UnknownUserId(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		{
			ID:        userId,
			Email:     "john.doe@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"email": "jane.doe@example.com", "verified": true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	unknownUserId, _ := uuid.NewV4()
	c.SetParamValues(unknownUserId.String())

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)
	addUsers(users, persister)

	handler := NewUserHandlerAdmin(persister)

	err := handler.Patch(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusNotFound, httpError.Code)
	}
}

func TestUserHandlerAdmin_Patch_InvalidJson(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		{
			ID:        userId,
			Email:     "john.doe@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`"email: "jane.doe@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	unknownUserId, _ := uuid.NewV4()
	c.SetParamValues(unknownUserId.String())

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	addUsers(users, persister)

	handler := NewUserHandlerAdmin(persister)

	err := handler.Patch(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	}
}

func TestUserHandlerAdmin_List(t *testing.T) {
	users := []models.User{
		func() models.User {
			userId, _ := uuid.NewV4()
			return models.User{
				ID:        userId,
				Email:     "john.doe@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
		func() models.User {
			userId, _ := uuid.NewV4()
			return models.User{
				ID:        userId,
				Email:     "jane.doe@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)
	addUsers(users, persister)

	handler := NewUserHandlerAdmin(persister)

	if assert.NoError(t, handler.List(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var users []models.User
		err := json.Unmarshal(rec.Body.Bytes(), &users)
		assert.NoError(t, err)
		assert.Equal(t, 2+1, len(users))
	}
}

func TestUserHandlerAdmin_List_Pagination(t *testing.T) {
	users := []models.User{
		func() models.User {
			userId, _ := uuid.NewV4()
			return models.User{
				ID:        userId,
				Email:     "john.doe@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
		func() models.User {
			userId, _ := uuid.NewV4()
			return models.User{
				ID:        userId,
				Email:     "jane.doe@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	e := echo.New()

	q := make(url.Values)
	q.Set("per_page", "1")
	q.Set("page", "2")
	req := httptest.NewRequest(http.MethodGet, "/users?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	addUsers(users, persister)

	handler := NewUserHandlerAdmin(persister)

	if assert.NoError(t, handler.List(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var got []models.User
		err := json.Unmarshal(rec.Body.Bytes(), &got)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(got))
	}
}

func TestUserHandlerAdmin_List_NoUsers(t *testing.T) {
	e := echo.New()

	q := make(url.Values)
	q.Set("per_page", "1")
	q.Set("page", "1")
	req := httptest.NewRequest(http.MethodGet, "/users?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	handler := NewUserHandlerAdmin(persister)

	if assert.NoError(t, handler.List(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var got []models.User
		err := json.Unmarshal(rec.Body.Bytes(), &got)
		assert.NoError(t, err)
		assert.Equal(t, 0+1, len(got))
	}
}

func TestUserHandlerAdmin_List_InvalidPaginationParam(t *testing.T) {
	e := echo.New()

	q := make(url.Values)
	q.Set("per_page", "invalidPerPageValue")
	req := httptest.NewRequest(http.MethodGet, "/logins?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	handler := NewUserHandlerAdmin(persister)

	err := handler.List(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	}
}

func TestGetLoginAuditRecordsByUserId_List(t *testing.T) {
	userId := uuid.FromStringOrNil("8846cc7e-c748-4a5d-bf8b-673284b01974")
	guestUserId := uuid.FromStringOrNil("40ac1f81-4d2d-4bf0-bc24-a0f6c067f171")
	audits := []models.LoginAuditLog{
		func() models.LoginAuditLog {
			return models.LoginAuditLog{
				ID:        uuid.FromStringOrNil("8b17f0e3-b498-4a82-8e35-909184c93288"),
				UserId:    userId,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
		func() models.LoginAuditLog {
			return models.LoginAuditLog{
				ID:              uuid.FromStringOrNil("ce761126-2a2e-4534-9793-0b022c4daa68"),
				UserId:          userId,
				SurrogateUserId: &guestUserId,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
		}(),
		func() models.LoginAuditLog {
			return models.LoginAuditLog{
				ID:              uuid.FromStringOrNil("621b7c7a-d738-4a88-96aa-08b72ad32ef8"),
				UserId:          guestUserId,
				SurrogateUserId: &guestUserId,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
		}(),
		func() models.LoginAuditLog {
			return models.LoginAuditLog{
				ID:              uuid.FromStringOrNil("a4135123-80d6-4bfd-93d2-1dd9df95e579"),
				UserId:          guestUserId,
				SurrogateUserId: &userId,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
		}(),
	}

	e := echo.New()

	body := fmt.Sprintf(`{"userId": "%s"}`, userId.String())
	req := httptest.NewRequest(http.MethodPost, "/logins", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	for _, audit := range audits {
		persister.GetLoginAuditLogPersister().Create(audit)
	}

	mainUser := models.User{
		ID:    userId,
		Email: "test@example.com",
	}
	guestUser := models.User{
		ID:    guestUserId,
		Email: "world@example.com",
	}
	persister.GetUserPersister().Create(mainUser)
	persister.GetUserPersister().Create(guestUser)

	handler := NewUserHandlerAdmin(persister)

	if assert.NoError(t, handler.GetLoginAuditRecordsForUser(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var got LoginAuditRecordResponseDto
		err := json.Unmarshal(rec.Body.Bytes(), &got)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(got.LoginsToAccount))
		assert.Equal(t, 1, len(got.LoginsAsGuest))
	}
}

func TestUserHandlerAdmin_GetGrantsForUser_WhenUserIsNotAdmin_Errors(t *testing.T) {
	userId := uuid.FromStringOrNil("8846cc7e-c748-4a5d-bf8b-673284b01974")
	grants := []models.UserGuestRelation{
		func() models.UserGuestRelation {
			guestUserId, _ := uuid.NewV4()
			return models.UserGuestRelation{
				ID:           uuid.FromStringOrNil("8b17f0e3-b498-4a82-8e35-909184c93288"),
				ParentUserID: userId,
				GuestUserID:  guestUserId,
				IsActive:     true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
		}(),
		func() models.UserGuestRelation {
			guestUserId, _ := uuid.NewV4()
			return models.UserGuestRelation{
				ID:           uuid.FromStringOrNil("9b17f0e3-b498-4a82-8e35-909184c93288"),
				ParentUserID: userId,
				GuestUserID:  guestUserId,
				IsActive:     true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
		}(),
		func() models.UserGuestRelation {
			guestUserId, _ := uuid.NewV4()
			return models.UserGuestRelation{
				ID:           uuid.FromStringOrNil("ab17f0e3-b498-4a82-8e35-909184c93288"),
				ParentUserID: userId,
				GuestUserID:  guestUserId,
				IsActive:     false,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
		}(),
	}

	e := echo.New()

	body := fmt.Sprintf(`{"userId": %q}`, userId.String())
	req := httptest.NewRequest(http.MethodPost, "/logins", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	adminUser.IsAdmin = false
	persister.GetUserPersister().Update(adminUser)

	for _, grant := range grants {
		persister.GetUserGuestRelationPersister().Create(grant)
		persister.GetUserPersister().Create(models.User{
			ID:       grant.GuestUserID,
			IsActive: true,
			Email:    fmt.Sprintf("user-%s@example.com", grant.GuestUserID),
		})
	}

	mainUser := models.User{
		ID:    userId,
		Email: "test@example.com",
	}
	persister.GetUserPersister().Create(mainUser)

	handler := NewUserHandlerAdmin(persister)

	assert.Error(t, handler.GetGrantsForUser(c))
}

func TestUserHandlerAdmin_GetGrantsForUser_WhenUserIsAdmin(t *testing.T) {
	userId := uuid.FromStringOrNil("8846cc7e-c748-4a5d-bf8b-673284b01974")
	grants := []models.UserGuestRelation{
		func() models.UserGuestRelation {
			guestUserId, _ := uuid.NewV4()
			return models.UserGuestRelation{
				ID:           uuid.FromStringOrNil("8b17f0e3-b498-4a82-8e35-909184c93288"),
				ParentUserID: userId,
				GuestUserID:  guestUserId,
				IsActive:     true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
		}(),
		func() models.UserGuestRelation {
			guestUserId, _ := uuid.NewV4()
			return models.UserGuestRelation{
				ID:           uuid.FromStringOrNil("9b17f0e3-b498-4a82-8e35-909184c93288"),
				ParentUserID: userId,
				GuestUserID:  guestUserId,
				IsActive:     true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
		}(),
		func() models.UserGuestRelation {
			guestUserId, _ := uuid.NewV4()
			return models.UserGuestRelation{
				ID:           uuid.FromStringOrNil("ab17f0e3-b498-4a82-8e35-909184c93288"),
				ParentUserID: userId,
				GuestUserID:  guestUserId,
				IsActive:     false,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
		}(),
	}

	e := echo.New()

	body := fmt.Sprintf(`{"userId": %q}`, userId.String())
	req := httptest.NewRequest(http.MethodPost, "/logins", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	for _, grant := range grants {
		persister.GetUserGuestRelationPersister().Create(grant)
		persister.GetUserPersister().Create(models.User{
			ID:       grant.GuestUserID,
			IsActive: true,
			Email:    fmt.Sprintf("user-%s@example.com", grant.GuestUserID),
		})
	}

	mainUser := models.User{
		ID:    userId,
		Email: "test@example.com",
	}
	persister.GetUserPersister().Create(mainUser)

	c.SetParamNames("id")
	c.SetParamValues(userId.String())

	handler := NewUserHandlerAdmin(persister)

	err := handler.GetGrantsForUser(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	var got GetGrantsForUserResponse
	err = json.Unmarshal(rec.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(got.Grants))
}

func Test_validateAdminPermission_WhenUserIsAdminAndActive_ReturnsNoError(t *testing.T) {
	e := echo.New()

	q := make(url.Values)
	q.Set("per_page", "invalidPerPageValue")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	handler := NewUserHandlerAdmin(persister)
	err, isSuccess := handler.validateAdminPermission(c)
	assert.NoError(t, err)
	assert.True(t, isSuccess)
}

func Test_validateAdminPermission_WhenUserIsNotAdminAndActive_ReturnsError(t *testing.T) {
	e := echo.New()

	q := make(url.Values)
	q.Set("per_page", "invalidPerPageValue")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	adminUser.IsAdmin = false
	persister.GetUserPersister().Update(adminUser)

	handler := NewUserHandlerAdmin(persister)
	err, isSuccess := handler.validateAdminPermission(c)
	assert.Error(t, err)
	assert.False(t, isSuccess)
}

func Test_validateAdminPermission_WhenUserIsAdminAndNotActive_ReturnsError(t *testing.T) {
	e := echo.New()

	q := make(url.Values)
	q.Set("per_page", "invalidPerPageValue")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	adminUser, persister := createAdmin()
	setSessionToken(t, c, adminUser)

	adminUser.IsActive = false
	persister.GetUserPersister().Update(adminUser)

	handler := NewUserHandlerAdmin(persister)
	err, isSuccess := handler.validateAdminPermission(c)
	assert.Error(t, err)
	assert.False(t, isSuccess)
}

func createAdmin() (models.User, persistence.Persister) {
	userId := uuid.FromStringOrNil("6bc3a580-d922-42f3-9032-a4faf8faef5e")
	user := models.User{
		ID:       userId,
		Email:    "admin@example.com",
		IsAdmin:  true,
		IsActive: true,
	}
	persister := test.NewPersister(append([]models.User{}, user), nil, nil, nil, nil, nil, nil, nil, nil)
	return user, persister
}

func setSessionToken(t *testing.T, c echo.Context, adminUser models.User) {
	token := jwt.New()
	err := token.Set(jwt.SubjectKey, adminUser.ID.String())
	require.NoError(t, err)
	err = token.Set("surr", adminUser.ID.String())
	require.NoError(t, err)
	c.Set("session", token)
}

func addUsers(users []models.User, persister persistence.Persister) {
	for _, user := range users {
		persister.GetUserPersister().Create(user)
	}
}
