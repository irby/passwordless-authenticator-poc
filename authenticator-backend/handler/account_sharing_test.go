package handler

import (
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

func Test_AccountSharingHandler_CreateAccountWithGrant_Errors_WhenPrimaryUserIdAndGuestUserIdAreTheSame(t *testing.T) {
	handler := generateHandler(t)

	primaryUserId := generateUuid(t)
	grantId := generateUuid(t)

	handler.persister.GetAccountAccessGrantPersister().
		Create(models.AccountAccessGrant{
			ID:       grantId,
			IsActive: true,
			UserId:   primaryUserId,
		})

	err := handler.CreateAccountWithGrant(grantId, primaryUserId, primaryUserId)
	assert.Error(t, err)
}

func Test_AccountSharingHandler_CreateAccountWithGrant_Errors_WhenGrantCannotBeFound(t *testing.T) {
	handler := generateHandler(t)

	primaryUserId := generateUuid(t)
	guestUserId := generateUuid(t)
	grantId := generateUuid(t)

	err := handler.CreateAccountWithGrant(grantId, primaryUserId, guestUserId)
	assert.Error(t, err)
}

func Test_AccountSharingHandler_CreateAccountWithGrant_Errors_WhenGrantIsInactive(t *testing.T) {
	handler := generateHandler(t)

	primaryUserId := generateUuid(t)
	guestUserId := generateUuid(t)
	grantId := generateUuid(t)

	handler.persister.GetAccountAccessGrantPersister().
		Create(models.AccountAccessGrant{
			ID:        grantId,
			IsActive:  false,
			UserId:    primaryUserId,
			CreatedAt: time.Now().UTC(),
			Ttl:       TimeToLiveMinutes,
		})

	err := handler.CreateAccountWithGrant(grantId, primaryUserId, guestUserId)
	assert.Error(t, err)
}

func Test_AccountSharingHandler_CreateAccountWithGrant_Errors_WhenGrantDoesNotBelongToPrimaryAccountId(t *testing.T) {
	handler := generateHandler(t)

	primaryUserId := generateUuid(t)
	guestUserId := generateUuid(t)
	grantId := generateUuid(t)

	handler.persister.GetAccountAccessGrantPersister().
		Create(models.AccountAccessGrant{
			ID:        grantId,
			IsActive:  true,
			CreatedAt: time.Now().UTC(),
			Ttl:       TimeToLiveMinutes,
			UserId:    generateUuid(t),
		})

	err := handler.CreateAccountWithGrant(grantId, primaryUserId, guestUserId)
	assert.Error(t, err)
}

func Test_AccountSharingHandler_CreateAccountWithGrant_Errors_WhenGrantIsExpired(t *testing.T) {
	handler := generateHandler(t)

	primaryUserId := generateUuid(t)
	guestUserId := generateUuid(t)
	grantId := generateUuid(t)

	handler.persister.GetAccountAccessGrantPersister().
		Create(models.AccountAccessGrant{
			ID:        grantId,
			IsActive:  true,
			CreatedAt: time.Now().UTC().Add(time.Duration(-15) * time.Minute),
			Ttl:       5,
			UserId:    primaryUserId,
		})

	err := handler.CreateAccountWithGrant(grantId, primaryUserId, guestUserId)
	assert.Error(t, err)
}

func Test_AccountSharingHandler_CreateAccountWithGrant_Errors_WhenGuestAlreadyHasAnActiveGrantForAccount(t *testing.T) {
	handler := generateHandler(t)

	primaryUserId := generateUuid(t)
	guestUserId := generateUuid(t)
	grantId := generateUuid(t)

	handler.persister.GetAccountAccessGrantPersister().
		Create(models.AccountAccessGrant{
			ID:        grantId,
			IsActive:  true,
			CreatedAt: time.Now().UTC(),
			Ttl:       TimeToLiveMinutes,
			UserId:    primaryUserId,
		})
	handler.persister.GetUserGuestRelationPersister().
		Create(models.UserGuestRelation{
			ID:           generateUuid(t),
			ParentUserID: primaryUserId,
			GuestUserID:  guestUserId,
			IsActive:     true,
		})

	err := handler.CreateAccountWithGrant(grantId, primaryUserId, guestUserId)
	assert.Error(t, err)
}

func Test_AccountSharingHandler_CreateAccountWithGrant_DoesNotError_WhenConditionsAreMet(t *testing.T) {
	handler := generateHandler(t)

	primaryUserId := generateUuid(t)
	guestUserId := generateUuid(t)
	grantId := generateUuid(t)

	handler.persister.GetAccountAccessGrantPersister().
		Create(models.AccountAccessGrant{
			ID:        grantId,
			IsActive:  true,
			CreatedAt: time.Now().UTC(),
			Ttl:       TimeToLiveMinutes,
			UserId:    primaryUserId,
		})

	err := handler.CreateAccountWithGrant(grantId, primaryUserId, guestUserId)
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

//func Test_AccountSharingHandler_FinishCreateAccountWithGrant_WhenRequestIsValid_CreatesAccount(t *testing.T) {
//	handler := generateHandler(t)
//
//	primaryUser := models.User{
//		ID:       generateUuid(t),
//		Email:    "hello@example.com",
//		IsActive: true,
//	}
//	guestUser := models.User{
//		ID:       generateUuid(t),
//		Email:    "world@example.com",
//		IsActive: true,
//	}
//	grant := models.AccountAccessGrant{
//		ID:        generateUuid(t),
//		UserId:    primaryUser.ID,
//		IsActive:  true,
//		CreatedAt: time.Now().UTC(),
//		Ttl:       TimeToLiveMinutes,
//	}
//
//	handler.persister.GetUserPersister().Create(primaryUser)
//	handler.persister.GetUserPersister().Create(guestUser)
//	handler.persister.GetAccountAccessGrantPersister().Create(grant)
//
//	body := fmt.Sprintf(`{
//"guestUserId": "%s",
//"grantId": "%s"
//	}`, guestUser.ID.String(), grant.ID.String())
//
//	e := echo.New()
//	e.Validator = dto.NewCustomValidator()
//	req := httptest.NewRequest(http.MethodPost, "/begin-create-account-with-grant", strings.NewReader(body))
//	req.Header.Set("Content-Type", "application/json")
//	rec := httptest.NewRecorder()
//	c := e.NewContext(req, rec)
//	c.Set("session", generateJwt(t, primaryUser.ID, primaryUser.ID, 60))
//
//	if assert.NoError(t, handler.BeginCreateAccountWithGrant(c)) {
//		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
//		assertionOptions := protocol.CredentialAssertion{}
//		err := json.Unmarshal(rec.Body.Bytes(), &assertionOptions)
//		assert.NoError(t, err)
//		assert.NotEmpty(t, assertionOptions.Response.Challenge)
//		assert.Equal(t, assertionOptions.Response.UserVerification, protocol.VerificationRequired)
//		assert.Equal(t, defaultConfig.Webauthn.RelyingParty.Id, assertionOptions.Response.RelyingPartyID)
//	}
//}

func generateHandler(t *testing.T) *AccountSharingHandler {
	handler, err := NewAccountSharingHandler(&defaultConfig, test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil), sessionManager{}, mailer{})
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
