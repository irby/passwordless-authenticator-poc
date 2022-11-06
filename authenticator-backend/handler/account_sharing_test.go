package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	jwt2 "github.com/teamhanko/hanko/backend/crypto/jwt"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_AccountSharingHandler_GetAccountShareWithToken_Errors_IfGrantUidCannotBeParsed(t *testing.T) {
	handler := generateHandler(t)

	err := handler.GetAccountShareGrantWithToken("hellothisisnotaguid", "booooooooooooooo")
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_GetAccountShareWithToken_Errors_IfGrantCantBeFound(t *testing.T) {
	handler := generateHandler(t)
	grantId := generateUuid(t)

	err := handler.GetAccountShareGrantWithToken(grantId.String(), "booooooooooooooo")
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_GetAccountShareWithToken_Errors_IfGrantIsNotActive(t *testing.T) {
	handler := generateHandler(t)
	grantId := generateUuid(t)

	handler.persister.GetAccountAccessGrantPersister().
		Create(models.AccountAccessGrant{
			ID:       grantId,
			IsActive: false,
		})

	err := handler.GetAccountShareGrantWithToken(grantId.String(), "booooooooooooooo")
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_GetAccountShareWithToken_Errors_IfGrantIsExpired(t *testing.T) {
	handler := generateHandler(t)

	grantId := generateUuid(t)

	handler.persister.GetAccountAccessGrantPersister().
		Create(models.AccountAccessGrant{
			ID:        grantId,
			IsActive:  true,
			CreatedAt: time.Now().UTC().Add(time.Duration(-15) * time.Minute),
			Ttl:       10,
		})

	err := handler.GetAccountShareGrantWithToken(grantId.String(), "booooooooooooooo")
	assert.Error(t, err)
	assert.Equal(t, http.StatusRequestTimeout, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_GetAccountShareWithToken_Errors_IfTokenIsInvalid(t *testing.T) {
	handler := generateHandler(t)

	grantId := generateUuid(t)

	token := "thisisatesttoken"

	handler.persister.GetAccountAccessGrantPersister().
		Create(models.AccountAccessGrant{
			ID:        grantId,
			IsActive:  true,
			CreatedAt: time.Now().UTC(),
			Ttl:       10000,
			Token:     generateHash(t, token),
		})

	err := handler.GetAccountShareGrantWithToken(grantId.String(), "booooooooooooooo")
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_GetAccountShareWithToken_DoesNotError_IfGrantIsActiveAndTokenIsCorrect(t *testing.T) {
	handler := generateHandler(t)

	grantId := generateUuid(t)

	token := "thisisatesttoken"

	handler.persister.GetAccountAccessGrantPersister().
		Create(models.AccountAccessGrant{
			ID:        grantId,
			IsActive:  true,
			CreatedAt: time.Now().UTC(),
			Ttl:       10000,
			Token:     generateHash(t, token),
		})

	err := handler.GetAccountShareGrantWithToken(grantId.String(), token)
	assert.NoError(t, err)
}

func Test_AccountSharingHandler_BeginCreateAccountWithGrant_WhenRequestIsValid_CreatesWebauthnToken(t *testing.T) {
	handler := generateHandler(t)

	primaryUser := models.User{
		ID:       generateUuid(t),
		Email:    "hello@example.com",
		IsActive: true,
	}
	guestUser := models.User{
		ID:       generateUuid(t),
		Email:    "world@example.com",
		IsActive: true,
	}
	grant := models.AccountAccessGrant{
		ID:        generateUuid(t),
		UserId:    primaryUser.ID,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		Ttl:       TimeToLiveMinutes,
	}

	handler.persister.GetUserPersister().Create(primaryUser)
	handler.persister.GetUserPersister().Create(guestUser)
	handler.persister.GetAccountAccessGrantPersister().Create(grant)

	body := fmt.Sprintf(`{
"guestUserId": "%s",
"grantId": "%s"
	}`, guestUser.ID.String(), grant.ID.String())

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/begin-create-account-with-grant", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, primaryUser.ID, primaryUser.ID, 60))

	if assert.NoError(t, handler.BeginCreateAccountWithGrant(c)) {
		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		assertionOptions := protocol.CredentialAssertion{}
		err := json.Unmarshal(rec.Body.Bytes(), &assertionOptions)
		assert.NoError(t, err)
		assert.NotEmpty(t, assertionOptions.Response.Challenge)
		assert.Equal(t, assertionOptions.Response.UserVerification, protocol.VerificationRequired)
		assert.Equal(t, defaultConfig.Webauthn.RelyingParty.Id, assertionOptions.Response.RelyingPartyID)
	}
}

func Test_AccountSharingHandler_BeginCreateAccountWithGrant_Errors_WhenCalledByAGuestUser(t *testing.T) {
	handler := generateHandler(t)

	primaryUser := models.User{
		ID:       generateUuid(t),
		Email:    "hello@example.com",
		IsActive: true,
	}
	guestUser := models.User{
		ID:       generateUuid(t),
		Email:    "world@example.com",
		IsActive: true,
	}
	grant := models.AccountAccessGrant{
		ID:        generateUuid(t),
		UserId:    primaryUser.ID,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		Ttl:       TimeToLiveMinutes,
	}

	handler.persister.GetUserPersister().Create(primaryUser)
	handler.persister.GetUserPersister().Create(guestUser)
	handler.persister.GetAccountAccessGrantPersister().Create(grant)

	body := fmt.Sprintf(`{
"guestUserId": "%s",
"grantId": "%s"
	}`, guestUser.ID.String(), grant.ID.String())

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/begin-create-account-with-grant", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, primaryUser.ID, generateUuid(t), 60))

	err := handler.BeginCreateAccountWithGrant(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_BeginCreateAccountWithGrant_Errors_WhenGuestUserIdCannotBeFound(t *testing.T) {
	handler := generateHandler(t)

	primaryUser := models.User{
		ID:       generateUuid(t),
		Email:    "hello@example.com",
		IsActive: true,
	}
	guestUser := models.User{
		ID:       generateUuid(t),
		Email:    "world@example.com",
		IsActive: true,
	}
	grant := models.AccountAccessGrant{
		ID:        generateUuid(t),
		UserId:    primaryUser.ID,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		Ttl:       TimeToLiveMinutes,
	}

	handler.persister.GetUserPersister().Create(primaryUser)
	handler.persister.GetAccountAccessGrantPersister().Create(grant)

	body := fmt.Sprintf(`{
"guestUserId": "%s",
"grantId": "%s"
	}`, guestUser.ID.String(), grant.ID.String())

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/begin-create-account-with-grant", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, primaryUser.ID, primaryUser.ID, 60))

	err := handler.BeginCreateAccountWithGrant(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_BeginCreateAccountWithGrant_Errors_WhenGrantIdCannotBeFound(t *testing.T) {
	handler := generateHandler(t)

	primaryUser := models.User{
		ID:       generateUuid(t),
		Email:    "hello@example.com",
		IsActive: true,
	}
	guestUser := models.User{
		ID:       generateUuid(t),
		Email:    "world@example.com",
		IsActive: true,
	}
	grant := models.AccountAccessGrant{
		ID:        generateUuid(t),
		UserId:    primaryUser.ID,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		Ttl:       TimeToLiveMinutes,
	}

	handler.persister.GetUserPersister().Create(primaryUser)
	handler.persister.GetUserPersister().Create(guestUser)

	body := fmt.Sprintf(`{
"guestUserId": "%s",
"grantId": "%s"
	}`, guestUser.ID.String(), grant.ID.String())

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/begin-create-account-with-grant", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, primaryUser.ID, primaryUser.ID, 60))

	err := handler.BeginCreateAccountWithGrant(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_FinishCreateAccountWithGrant_WhenRequestIsValid_CreatesAccount(t *testing.T) {
	handler := generateHandler(t)

	primaryUser := models.User{
		ID:       generateUuid(t),
		Email:    "hello@example.com",
		IsActive: true,
	}
	guestUser := models.User{
		ID:       generateUuid(t),
		Email:    "world@example.com",
		IsActive: true,
	}
	grant := models.AccountAccessGrant{
		ID:        generateUuid(t),
		UserId:    primaryUser.ID,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		Ttl:       TimeToLiveMinutes,
	}

	handler.persister.GetUserPersister().Create(primaryUser)
	handler.persister.GetUserPersister().Create(guestUser)
	handler.persister.GetAccountAccessGrantPersister().Create(grant)
	handler.generateCredentialsAndSessionDataForUserId(primaryUser.ID)

	formattedBody := fmt.Sprintf(signedRequestBody, base64.RawURLEncoding.EncodeToString(primaryUser.ID.Bytes()), guestUser.ID, grant.ID, "")

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/finalize-create-account-with-grant?guestUserId=%s&grantId=%s", guestUser.ID, grant.ID), strings.NewReader(formattedBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, primaryUser.ID, primaryUser.ID, 60))

	if assert.NoError(t, handler.FinishCreateAccountWithGrant(c)) {
		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		relation, err := handler.persister.GetUserGuestRelationPersister().GetByGuestUserId(&guestUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(relation))
		assert.Equal(t, primaryUser.ID, relation[0].ParentUserID)
		assert.Equal(t, guestUser.ID, relation[0].GuestUserID)
		// TODO: Test attestation
	}
}

func Test_AccountSharingHandler_FinishCreateAccountWithGrant_Errors_WhenWebauthnCredentialsDoNotExistForUser(t *testing.T) {
	handler := generateHandler(t)

	primaryUser := models.User{
		ID:       generateUuid(t),
		Email:    "hello@example.com",
		IsActive: true,
	}
	guestUser := models.User{
		ID:       generateUuid(t),
		Email:    "world@example.com",
		IsActive: true,
	}
	grant := models.AccountAccessGrant{
		ID:        generateUuid(t),
		UserId:    primaryUser.ID,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		Ttl:       TimeToLiveMinutes,
	}

	handler.persister.GetUserPersister().Create(primaryUser)
	handler.persister.GetUserPersister().Create(guestUser)
	handler.persister.GetAccountAccessGrantPersister().Create(grant)
	handler.generateCredentialsAndSessionDataForUserId(generateUuid(t))

	formattedBody := fmt.Sprintf(signedRequestBody, base64.RawURLEncoding.EncodeToString(primaryUser.ID.Bytes()), guestUser.ID, grant.ID, "")

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/finalize-create-account-with-grant?guestUserId=%s&grantId=%s", guestUser.ID, grant.ID), strings.NewReader(formattedBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, primaryUser.ID, primaryUser.ID, 60))

	err := handler.FinishCreateAccountWithGrant(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_FinishCreateAccountWithGrant_Errors_WhenRequestIsMissingField(t *testing.T) {
	handler := generateHandler(t)

	primaryUser := models.User{
		ID:       generateUuid(t),
		Email:    "hello@example.com",
		IsActive: true,
	}
	guestUser := models.User{
		ID:       generateUuid(t),
		Email:    "world@example.com",
		IsActive: true,
	}
	grant := models.AccountAccessGrant{
		ID:        generateUuid(t),
		UserId:    primaryUser.ID,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		Ttl:       TimeToLiveMinutes,
	}

	handler.persister.GetUserPersister().Create(primaryUser)
	handler.persister.GetUserPersister().Create(guestUser)
	handler.persister.GetAccountAccessGrantPersister().Create(grant)
	handler.generateCredentialsAndSessionDataForUserId(primaryUser.ID)

	// Missing authenticator data
	body := fmt.Sprintf(`{
"id": "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
"rawId": "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
"type": "public-key",
"response": {
"clientDataJSON": "eyJ0eXBlIjoid2ViYXV0aG4uZ2V0IiwiY2hhbGxlbmdlIjoiZ0tKS21oOTB2T3BZTzU1b0hwcWFIWF9vTUNxNG9UWnQtRDBiNnRlSXpyRSIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0",
"signature": "MEYCIQDi2vYVspG6pf38I4GyQCPOojGbvX4nwSPXCi0hm80twAIhAO3EWjhAnj0UpjU_l0AH5sEh3zq4LDvkvo3AUqaqfGYD",
"userHandle": "%s"
}
}`, base64.RawURLEncoding.EncodeToString(primaryUser.ID.Bytes()))

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/finalize-create-account-with-grant?guestUserId=%s&grantId=%s", guestUser.ID, grant.ID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, primaryUser.ID, primaryUser.ID, 60))

	err := handler.FinishCreateAccountWithGrant(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_FinishCreateAccountWithGrant_Errors_WhenRequestSignatureIsInvalid(t *testing.T) {
	handler := generateHandler(t)

	primaryUser := models.User{
		ID:       generateUuid(t),
		Email:    "hello@example.com",
		IsActive: true,
	}
	guestUser := models.User{
		ID:       generateUuid(t),
		Email:    "world@example.com",
		IsActive: true,
	}
	grant := models.AccountAccessGrant{
		ID:        generateUuid(t),
		UserId:    primaryUser.ID,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		Ttl:       TimeToLiveMinutes,
	}

	handler.persister.GetUserPersister().Create(primaryUser)
	handler.persister.GetUserPersister().Create(guestUser)
	handler.persister.GetAccountAccessGrantPersister().Create(grant)
	handler.generateCredentialsAndSessionDataForUserId(primaryUser.ID)

	// Signature is invalid
	body := fmt.Sprintf(`{
"id": "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
"rawId": "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
"type": "public-key",
"response": {
"authenticatorData": "SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2MFYmezOw",
"clientDataJSON": "eyJ0eXBlIjoid2ViYXV0aG4uZ2V0IiwiY2hhbGxlbmdlIjoiZ0tKS21oOTB2T3BZTzU1b0hwcWFIWF9vTUNxNG9UWnQtRDBiNnRlSXpyRSIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0",
"signature": "MEYCIQDi2vYVspG6pHHHI4GyQCPOojGbvX4nwSPXCi0hm80twAIhAO3EWjhAnj0UpjU_l0AH5sEh3zq4LDvkvo3AUqaqfGYD",
"userHandle": "%s"
}
}`, base64.RawURLEncoding.EncodeToString(primaryUser.ID.Bytes()))

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/finalize-create-account-with-grant?guestUserId=%s&grantId=%s", guestUser.ID, grant.ID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, primaryUser.ID, primaryUser.ID, 60))

	err := handler.FinishCreateAccountWithGrant(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_FinishCreateAccountWithGrant_Errors_WhenCalledByAGuestUser(t *testing.T) {
	handler := generateHandler(t)

	primaryUser := models.User{
		ID:       generateUuid(t),
		Email:    "hello@example.com",
		IsActive: true,
	}
	guestUser := models.User{
		ID:       generateUuid(t),
		Email:    "world@example.com",
		IsActive: true,
	}
	grant := models.AccountAccessGrant{
		ID:        generateUuid(t),
		UserId:    primaryUser.ID,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		Ttl:       TimeToLiveMinutes,
	}

	handler.persister.GetUserPersister().Create(primaryUser)
	handler.persister.GetUserPersister().Create(guestUser)
	handler.persister.GetAccountAccessGrantPersister().Create(grant)
	handler.generateCredentialsAndSessionDataForUserId(primaryUser.ID)

	formattedBody := fmt.Sprintf(signedRequestBody, base64.RawURLEncoding.EncodeToString(primaryUser.ID.Bytes()), guestUser.ID, grant.ID, "")

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/finalize-create-account-with-grant?guestUserId=%s&grantId=%s", guestUser.ID, grant.ID), strings.NewReader(formattedBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, primaryUser.ID, generateUuid(t), 60))

	err := handler.FinishCreateAccountWithGrant(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_FinishCreateAccountWithGrant_Errors_WhenGuestUserAlreadyHasAnActiveGrant(t *testing.T) {
	handler := generateHandler(t)

	primaryUser := models.User{
		ID:       generateUuid(t),
		Email:    "hello@example.com",
		IsActive: true,
	}
	guestUser := models.User{
		ID:       generateUuid(t),
		Email:    "world@example.com",
		IsActive: true,
	}
	grant := models.AccountAccessGrant{
		ID:        generateUuid(t),
		UserId:    primaryUser.ID,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		Ttl:       TimeToLiveMinutes,
	}

	handler.persister.GetUserPersister().Create(primaryUser)
	handler.persister.GetUserPersister().Create(guestUser)
	handler.persister.GetAccountAccessGrantPersister().Create(grant)
	handler.persister.GetUserGuestRelationPersister().Create(models.UserGuestRelation{
		ID:           generateUuid(t),
		ParentUserID: primaryUser.ID,
		GuestUserID:  guestUser.ID,
		IsActive:     true,
	})
	handler.generateCredentialsAndSessionDataForUserId(primaryUser.ID)

	formattedBody := fmt.Sprintf(signedRequestBody, base64.RawURLEncoding.EncodeToString(primaryUser.ID.Bytes()), guestUser.ID, grant.ID, "")

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/finalize-create-account-with-grant?guestUserId=%s&grantId=%s", guestUser.ID, grant.ID), strings.NewReader(formattedBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, primaryUser.ID, primaryUser.ID, 60))

	err := handler.FinishCreateAccountWithGrant(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusConflict, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_FinishCreateAccountWithGrant_Errors_WhenGrantDoesNotBelongToUser(t *testing.T) {
	handler := generateHandler(t)

	primaryUser := models.User{
		ID:       generateUuid(t),
		Email:    "hello@example.com",
		IsActive: true,
	}
	guestUser := models.User{
		ID:       generateUuid(t),
		Email:    "world@example.com",
		IsActive: true,
	}
	grant := models.AccountAccessGrant{
		ID:        generateUuid(t),
		UserId:    generateUuid(t),
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		Ttl:       TimeToLiveMinutes,
	}

	handler.persister.GetUserPersister().Create(primaryUser)
	handler.persister.GetUserPersister().Create(guestUser)
	handler.persister.GetAccountAccessGrantPersister().Create(grant)
	handler.generateCredentialsAndSessionDataForUserId(primaryUser.ID)

	formattedBody := fmt.Sprintf(signedRequestBody, base64.RawURLEncoding.EncodeToString(primaryUser.ID.Bytes()), guestUser.ID, grant.ID, "")

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/finalize-create-account-with-grant?guestUserId=%s&grantId=%s", guestUser.ID, grant.ID), strings.NewReader(formattedBody))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, primaryUser.ID, primaryUser.ID, 60))

	err := handler.FinishCreateAccountWithGrant(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_FinishCreateAccountWithGrant_Errors_WhenGrantIsExpired(t *testing.T) {
	handler := generateHandler(t)

	primaryUser := models.User{
		ID:       generateUuid(t),
		Email:    "hello@example.com",
		IsActive: true,
	}
	guestUser := models.User{
		ID:       generateUuid(t),
		Email:    "world@example.com",
		IsActive: true,
	}
	grant := models.AccountAccessGrant{
		ID:        generateUuid(t),
		UserId:    primaryUser.ID,
		IsActive:  true,
		CreatedAt: time.Now().UTC().Add(time.Duration(-20) * time.Minute),
		Ttl:       TimeToLiveMinutes,
	}

	handler.persister.GetUserPersister().Create(primaryUser)
	handler.persister.GetUserPersister().Create(guestUser)
	handler.persister.GetAccountAccessGrantPersister().Create(grant)
	handler.generateCredentialsAndSessionDataForUserId(primaryUser.ID)

	formattedBody := fmt.Sprintf(signedRequestBody, base64.RawURLEncoding.EncodeToString(primaryUser.ID.Bytes()), guestUser.ID, grant.ID, "")

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/finalize-create-account-with-grant?guestUserId=%s&grantId=%s", guestUser.ID, grant.ID), strings.NewReader(formattedBody))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, primaryUser.ID, primaryUser.ID, 60))

	err := handler.FinishCreateAccountWithGrant(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusRequestTimeout, dto.ToHttpError(err).Code)
}

func generateHandler(t *testing.T) *AccountSharingHandler {
	handler, err := NewAccountSharingHandler(&defaultConfig, test.NewPersister(users, nil, nil, nil, nil, nil, nil, nil, nil), sessionManager{}, mailer{})
	assert.NoError(t, err)
	assert.NotEmpty(t, handler)
	return handler
}

func generateUuid(t *testing.T) uuid.UUID {
	uId, err := uuid.NewV4()
	assert.NoError(t, err)
	return uId
}

func generateHash(t *testing.T, plaintext string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	assert.NoError(t, err)
	return string(hash)
}

func generateJwt(t *testing.T, subjectUserId uuid.UUID, surrogateUserId uuid.UUID, sessionLengthMinutes int) jwt.Token {
	token := jwt.New()
	assert.NoError(t, token.Set(jwt.SubjectKey, subjectUserId.String()))
	assert.NoError(t, token.Set(jwt2.SurrogateKey, surrogateUserId.String()))
	assert.NoError(t, token.Set(jwt.ExpirationKey, time.Now().UTC().Add(time.Duration(sessionLengthMinutes)*time.Minute)))
	return token
}

func (h *AccountSharingHandler) generateCredentialsAndSessionDataForUserId(userId uuid.UUID) {
	credentials := []models.WebauthnCredential{
		func() models.WebauthnCredential {
			aaguid, _ := uuid.FromString("adce0002-35bc-c60a-648b-0b25f1f05503")
			return models.WebauthnCredential{
				ID:              "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
				UserId:          userId,
				PublicKey:       "pQECAyYgASFYIPG9WtGAri-mevonFPH4p-lI3JBS29zjuvKvJmaP4_mRIlggOjHw31sdAGvE35vmRep-aPcbAAlbuc0KHxQ9u6zcHog",
				AttestationType: "none",
				AAGUID:          aaguid,
				SignCount:       1650958750,
				CreatedAt:       time.Time{},
				UpdatedAt:       time.Time{},
			}
		}(),
		func() models.WebauthnCredential {
			aaguid, _ := uuid.FromString("adce0002-35bc-c60a-648b-0b25f1f05503")
			return models.WebauthnCredential{
				ID:              "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjK",
				UserId:          userId,
				PublicKey:       "pQECAyYgASFYIPG9WtGAri-mevonFPH4p-lI3JBS29zjuvKvJmaP4_mRIlggOjHw31sdAGvE35vmRep-aPcbAAlbuc0KHxQ9u6zcHoj",
				AttestationType: "none",
				AAGUID:          aaguid,
				SignCount:       1650958750,
				CreatedAt:       time.Time{},
				UpdatedAt:       time.Time{},
			}
		}(),
	}

	sessionData := []models.WebauthnSessionData{
		func() models.WebauthnSessionData {
			id, _ := uuid.NewV4()
			return models.WebauthnSessionData{
				ID:                 id,
				Challenge:          "tOrNDCD2xQf4zFjEjwxaP8fOErP3zz08rMoTlJGtnKU",
				UserId:             userId,
				UserVerification:   string(protocol.VerificationRequired),
				CreatedAt:          time.Time{},
				UpdatedAt:          time.Time{},
				Operation:          models.WebauthnOperationRegistration,
				AllowedCredentials: nil,
			}
		}(),
		func() models.WebauthnSessionData {
			id, _ := uuid.NewV4()
			return models.WebauthnSessionData{
				ID:                 id,
				Challenge:          "gKJKmh90vOpYO55oHpqaHX_oMCq4oTZt-D0b6teIzrE",
				UserId:             uuid.UUID{},
				UserVerification:   string(protocol.VerificationRequired),
				CreatedAt:          time.Time{},
				UpdatedAt:          time.Time{},
				Operation:          models.WebauthnOperationAuthentication,
				AllowedCredentials: nil,
			}
		}(),
	}

	for _, credential := range credentials {
		h.persister.GetWebauthnCredentialPersister().Create(credential)
	}
	for _, session := range sessionData {
		h.persister.GetWebauthnSessionDataPersister().Create(session)
	}
}

var signedRequestBody = `{
"id": "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
"rawId": "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
"type": "public-key",
"response": {
"authenticatorData": "SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2MFYmezOw",
"clientDataJSON": "eyJ0eXBlIjoid2ViYXV0aG4uZ2V0IiwiY2hhbGxlbmdlIjoiZ0tKS21oOTB2T3BZTzU1b0hwcWFIWF9vTUNxNG9UWnQtRDBiNnRlSXpyRSIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0",
"signature": "MEYCIQDi2vYVspG6pf38I4GyQCPOojGbvX4nwSPXCi0hm80twAIhAO3EWjhAnj0UpjU_l0AH5sEh3zq4LDvkvo3AUqaqfGYD",
"userHandle": "%s"
},
"guestUserId": "%s",
"grantId": "%s",
"grantAttestation": "%s"
}`
