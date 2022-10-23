package handler

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
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
	req := httptest.NewRequest(http.MethodGet, "/users?"+q.Encode(), nil)
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

func createAdmin() (models.User, persistence.Persister) {
	userId := uuid.FromStringOrNil("6bc3a580-d922-42f3-9032-a4faf8faef5e")
	user := models.User{
		ID:       userId,
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
