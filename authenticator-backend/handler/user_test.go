package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestUserHandler_Create(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		func() models.User {
			return models.User{
				ID:        userId,
				Email:     "john.doe@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	body := UserCreateBody{Email: "jane.doe@example.com"}
	bodyJson, err := json.Marshal(body)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(users, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(&defaultConfig, p, sessionManager{})

	if assert.NoError(t, handler.Create(c)) {
		user := models.User{}
		err := json.Unmarshal(rec.Body.Bytes(), &user)
		assert.NoError(t, err)
		assert.False(t, user.ID.IsNil())
		assert.Equal(t, body.Email, user.Email)
	}
}

func TestUserHandler_Create_CaseInsensitive(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		func() models.User {
			return models.User{
				ID:        userId,
				Email:     "john.doe@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	body := UserCreateBody{Email: "JANE.DOE@EXAMPLE.COM"}
	bodyJson, err := json.Marshal(body)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(users, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(&defaultConfig, p, sessionManager{})

	if assert.NoError(t, handler.Create(c)) {
		user := models.User{}
		err := json.Unmarshal(rec.Body.Bytes(), &user)
		assert.NoError(t, err)
		assert.False(t, user.ID.IsNil())
		assert.Equal(t, strings.ToLower(body.Email), user.Email)
	}
}

func TestUserHandler_Create_UserExists(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		func() models.User {
			return models.User{
				ID:        userId,
				Email:     "john.doe@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	body := UserCreateBody{Email: "john.doe@example.com"}
	bodyJson, err := json.Marshal(body)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(users, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(&defaultConfig, p, sessionManager{})

	err = handler.Create(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusConflict, httpError.Code)
	}
}

func TestUserHandler_Create_UserExists_CaseInsensitive(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		func() models.User {
			return models.User{
				ID:        userId,
				Email:     "john.doe@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	body := UserCreateBody{Email: "JOHN.DOE@EXAMPLE.COM"}
	bodyJson, err := json.Marshal(body)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(users, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(&defaultConfig, p, sessionManager{})

	err = handler.Create(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusConflict, httpError.Code)
	}
}

func TestUserHandler_Create_InvalidEmail(t *testing.T) {
	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"email": 123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(&defaultConfig, p, sessionManager{})

	err := handler.Create(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	}
}

func TestUserHandler_Create_EmailMissing(t *testing.T) {
	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"bogus": 123}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(&defaultConfig, p, sessionManager{})

	err := handler.Create(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	}
}

func TestUserHandler_Get(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		func() models.User {
			return models.User{
				ID:        userId,
				Email:     "john.doe@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(userId.String())

	token := jwt.New()
	err := token.Set(jwt.SubjectKey, userId.String())
	require.NoError(t, err)
	c.Set("session", token)

	p := test.NewPersister(users, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(&defaultConfig, p, sessionManager{})

	if assert.NoError(t, handler.Get(c)) {
		assert.Equal(t, rec.Code, http.StatusOK)
		user := models.User{}
		err := json.Unmarshal(rec.Body.Bytes(), &user)
		assert.NoError(t, err)
		assert.Equal(t, userId, user.ID)
		assert.Equal(t, len(user.WebauthnCredentials), 0)
	}
}

func TestUserHandler_GetUserWithWebAuthnCredential(t *testing.T) {
	userId, _ := uuid.NewV4()
	aaguid, _ := uuid.FromString("adce0002-35bc-c60a-648b-0b25f1f05503")
	users := []models.User{
		{
			ID:        userId,
			Email:     "john.doe@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			WebauthnCredentials: []models.WebauthnCredential{
				{
					ID:              "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
					UserId:          userId,
					PublicKey:       "pQECAyYgASFYIPG9WtGAri-mevonFPH4p-lI3JBS29zjuvKvJmaP4_mRIlggOjHw31sdAGvE35vmRep-aPcbAAlbuc0KHxQ9u6zcHog",
					AttestationType: "none",
					AAGUID:          aaguid,
					SignCount:       1650958750,
					CreatedAt:       time.Time{},
					UpdatedAt:       time.Time{},
				},
			},
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(userId.String())

	token := jwt.New()
	err := token.Set(jwt.SubjectKey, userId.String())
	require.NoError(t, err)
	c.Set("session", token)

	p := test.NewPersister(users, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(&defaultConfig, p, sessionManager{})

	if assert.NoError(t, handler.Get(c)) {
		assert.Equal(t, rec.Code, http.StatusOK)
		user := models.User{}
		err := json.Unmarshal(rec.Body.Bytes(), &user)
		require.NoError(t, err)
		assert.Equal(t, userId, user.ID)
		assert.Equal(t, len(user.WebauthnCredentials), 1)
	}
}

func TestUserHandler_Get_InvalidUserId(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/invalidUserId", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	token := jwt.New()
	err := token.Set(jwt.SubjectKey, "completelyDifferentUserId")
	require.NoError(t, err)
	c.Set("session", token)

	p := test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(&defaultConfig, p, sessionManager{})

	err = handler.Get(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusForbidden, httpError.Code)
	}
}

func TestUserHandler_GetUserIdByEmail_InvalidEmail(t *testing.T) {
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{"email": "123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(&defaultConfig, p, sessionManager{})

	err := handler.GetUserIdByEmail(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	}
}

func TestUserHandler_GetUserIdByEmail_InvalidJson(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`"email": "123}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(&defaultConfig, p, sessionManager{})

	assert.Error(t, handler.GetUserIdByEmail(c))
}

func TestUserHandler_GetUserIdByEmail_UserNotFound(t *testing.T) {
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{"email": "unknownAddress@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(&defaultConfig, p, sessionManager{})

	err := handler.GetUserIdByEmail(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusNotFound, httpError.Code)
	}
}

func TestUserHandler_GetUserIdByEmail(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		{
			ID:        userId,
			Email:     "john.doe@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Verified:  true,
		},
	}
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{"email": "john.doe@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(users, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(&defaultConfig, p, sessionManager{})

	if assert.NoError(t, handler.GetUserIdByEmail(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		response := struct {
			UserId   string `json:"id"`
			Verified bool   `json:"verified"`
		}{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, users[0].ID.String(), response.UserId)
		assert.Equal(t, users[0].Verified, response.Verified)
	}
}

func TestUserHandler_GetUserIdByEmail_CaseInsensitive(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		{
			ID:        userId,
			Email:     "john.doe@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Verified:  true,
		},
	}
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{"email": "JOHN.DOE@EXAMPLE.COM"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(users, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(&defaultConfig, p, sessionManager{})

	if assert.NoError(t, handler.GetUserIdByEmail(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		response := struct {
			UserId   string `json:"id"`
			Verified bool   `json:"verified"`
		}{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, users[0].ID.String(), response.UserId)
		assert.Equal(t, users[0].Verified, response.Verified)
	}
}

func TestUserHandler_Me(t *testing.T) {
	userId := users[0].ID

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	token := jwt.New()
	err := token.Set(jwt.SubjectKey, userId.String())
	token.Set("surr", userId.String())
	require.NoError(t, err)
	c.Set("session", token)

	p := test.NewPersister(users, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(&defaultConfig, p, sessionManager{})

	if assert.NoError(t, handler.Me(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		response := struct {
			UserId string `json:"id"`
		}{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, userId.String(), response.UserId)
	}
}

func TestUserHandler_Logout(t *testing.T) {
	userId, _ := uuid.NewV4()

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	token := jwt.New()
	err := token.Set(jwt.SubjectKey, userId.String())
	require.NoError(t, err)
	c.Set("session", token)

	handler := generateUserHandler()

	if assert.NoError(t, handler.Logout(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		cookie := rec.Header().Get("Set-Cookie")
		assert.NotEmpty(t, cookie)

		split := strings.Split(cookie, ";")
		assert.Equal(t, "Max-Age=0", strings.TrimSpace(split[1]))
	}
}

func TestUserHandler_BeginLoginAsGuest_WhenRequestIsValid_GeneratesChallengeToSign(t *testing.T) {
	h := generateUserHandler()
	user1 := generateUser(t)
	user2 := generateUser(t)
	grant1 := models.UserGuestRelation{
		ID:           generateUuid(t),
		ParentUserID: user1.ID,
		GuestUserID:  user2.ID,
		IsActive:     true,
		CreatedAt:    time.Now().UTC(),
	}
	h.persister.GetUserPersister().Create(user1)
	h.persister.GetUserPersister().Create(user2)
	h.persister.GetUserGuestRelationPersister().Create(grant1)

	body := fmt.Sprintf(`{"relationId": "%s"}`, grant1.ID)

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/initialize-login-as-guest"), strings.NewReader(body))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, user1.ID, user1.ID, 60))

	if assert.NoError(t, h.BeginLoginAsGuest(c)) {
		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		assertionOptions := protocol.CredentialAssertion{}
		err := json.Unmarshal(rec.Body.Bytes(), &assertionOptions)
		assert.NoError(t, err)
		assert.NotEmpty(t, assertionOptions.Response.Challenge)
		assert.Equal(t, assertionOptions.Response.UserVerification, protocol.VerificationRequired)
		assert.Equal(t, defaultConfig.Webauthn.RelyingParty.Id, assertionOptions.Response.RelyingPartyID)
	}
}

func TestUserHandler_BeginLoginAsGuest_Errors_WhenGuestIsCurrentSession(t *testing.T) {
	h := generateUserHandler()
	user1 := generateUser(t)
	user2 := generateUser(t)
	grant1 := models.UserGuestRelation{
		ID:           generateUuid(t),
		ParentUserID: user1.ID,
		GuestUserID:  user2.ID,
		IsActive:     true,
		CreatedAt:    time.Now().UTC(),
	}
	h.persister.GetUserPersister().Create(user1)
	h.persister.GetUserPersister().Create(user2)
	h.persister.GetUserGuestRelationPersister().Create(grant1)

	body := fmt.Sprintf(`{"relationId": "%s"}`, grant1.ID)

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/initialize-login-as-guest"), strings.NewReader(body))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, user1.ID, user2.ID, 60))

	err := h.BeginLoginAsGuest(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, dto.ToHttpError(err).Code)
}

//func TestUserHandler_BeginLoginAsGuest_Errors_WhenGrantIsInactive(t *testing.T) {
//	h := generateUserHandler()
//	user1 := generateUser(t)
//	user2 := generateUser(t)
//	grant1 := models.UserGuestRelation{
//		ID:           generateUuid(t),
//		ParentUserID: user1.ID,
//		GuestUserID:  user2.ID,
//		IsActive:     false,
//		CreatedAt:    time.Now().UTC(),
//	}
//	h.persister.GetUserPersister().Create(user1)
//	h.persister.GetUserPersister().Create(user2)
//	h.persister.GetUserGuestRelationPersister().Create(grant1)
//
//	body := fmt.Sprintf(`{"relationId": "%s"}`, grant1.ID)
//
//	e := echo.New()
//	e.Validator = dto.NewCustomValidator()
//	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/initialize-login-as-guest"), strings.NewReader(body))
//	rec := httptest.NewRecorder()
//	c := e.NewContext(req, rec)
//	c.Set("session", generateJwt(t, user1.ID, user1.ID, 60))
//
//	err := h.BeginLoginAsGuest(c)
//	assert.Error(t, err)
//	assert.Equal(t, http.StatusNotFound, dto.ToHttpError(err).Code)
//}
//
//func TestUserHandler_BeginLoginAsGuest_Errors_WhenGrantIsExpiredByTime(t *testing.T) {
//	h := generateUserHandler()
//	user1 := generateUser(t)
//	user2 := generateUser(t)
//	grant1 := models.UserGuestRelation{
//		ID:             generateUuid(t),
//		ParentUserID:   user1.ID,
//		GuestUserID:    user2.ID,
//		IsActive:       true,
//		CreatedAt:      time.Now().UTC().Add(time.Duration(-15) * time.Minute),
//		ExpireByTime:   true,
//		MinutesAllowed: sql.NullInt32{Valid: true, Int32: 10},
//	}
//	h.persister.GetUserPersister().Create(user1)
//	h.persister.GetUserPersister().Create(user2)
//	h.persister.GetUserGuestRelationPersister().Create(grant1)
//
//	body := fmt.Sprintf(`{"relationId": "%s"}`, grant1.ID)
//
//	e := echo.New()
//	e.Validator = dto.NewCustomValidator()
//	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/initialize-login-as-guest"), strings.NewReader(body))
//	rec := httptest.NewRecorder()
//	c := e.NewContext(req, rec)
//	c.Set("session", generateJwt(t, user1.ID, user1.ID, 60))
//
//	err := h.BeginLoginAsGuest(c)
//	assert.Error(t, err)
//	assert.Equal(t, http.StatusNotFound, dto.ToHttpError(err).Code)
//}
//
//func TestUserHandler_BeginLoginAsGuest_Errors_WhenGrantIsExpiredByLogins(t *testing.T) {
//	h := generateUserHandler()
//	user1 := generateUser(t)
//	user2 := generateUser(t)
//	grant1 := models.UserGuestRelation{
//		ID:             generateUuid(t),
//		ParentUserID:   user1.ID,
//		GuestUserID:    user2.ID,
//		IsActive:       true,
//		CreatedAt:      time.Now().UTC(),
//		ExpireByLogins: true,
//		LoginsAllowed:  1,
//	}
//	h.persister.GetUserPersister().Create(user1)
//	h.persister.GetUserPersister().Create(user2)
//	h.persister.GetUserGuestRelationPersister().Create(grant1)
//
//	body := fmt.Sprintf(`{"relationId": "%s"}`, grant1.ID)
//
//	e := echo.New()
//	e.Validator = dto.NewCustomValidator()
//	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/initialize-login-as-guest"), strings.NewReader(body))
//	rec := httptest.NewRecorder()
//	c := e.NewContext(req, rec)
//	c.Set("session", generateJwt(t, user1.ID, user1.ID, 60))
//
//	err := h.BeginLoginAsGuest(c)
//	assert.Error(t, err)
//	assert.Equal(t, http.StatusNotFound, dto.ToHttpError(err).Code)
//}

func generateUserHandler() *UserHandler {
	p := test.NewPersister(users, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(&defaultConfig, p, sessionManager{})
	return handler
}

func generateUser(t *testing.T) models.User {
	uId := generateUuid(t)
	return models.User{
		ID:       uId,
		IsActive: true,
		Email:    fmt.Sprintf("test-%s@example.com", uId),
	}
}
