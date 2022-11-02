package handler

import (
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"testing"
	"time"
)

func Test_AccountSharingHandler_GetAccountShareWithToken_Errors_IfSubjectIdDoesNotEqualSurrogateId(t *testing.T) {
	handler := generateHandler(t)

	subjectId := generateUuid(t)
	surrogateId := generateUuid(t)

	err := handler.GetAccountShareGrantWithToken("", "", subjectId.String(), surrogateId.String())
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_GetAccountShareWithToken_Errors_IfGrantUidCannotBeParsed(t *testing.T) {
	handler := generateHandler(t)

	subjectId := generateUuid(t)

	err := handler.GetAccountShareGrantWithToken("hellothisisnotaguid", "booooooooooooooo", subjectId.String(), subjectId.String())
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_GetAccountShareWithToken_Errors_IfGrantCantBeFound(t *testing.T) {
	handler := generateHandler(t)

	subjectId := generateUuid(t)
	grantId := generateUuid(t)

	err := handler.GetAccountShareGrantWithToken(grantId.String(), "booooooooooooooo", subjectId.String(), subjectId.String())
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_GetAccountShareWithToken_Errors_IfGrantIsNotActive(t *testing.T) {
	handler := generateHandler(t)

	subjectId := generateUuid(t)
	grantId := generateUuid(t)

	handler.persister.GetAccountAccessGrantPersister().
		Create(models.AccountAccessGrant{
			ID:       grantId,
			IsActive: false,
		})

	err := handler.GetAccountShareGrantWithToken(grantId.String(), "booooooooooooooo", subjectId.String(), subjectId.String())
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_GetAccountShareWithToken_Errors_IfGrantIsExpired(t *testing.T) {
	handler := generateHandler(t)

	subjectId := generateUuid(t)
	grantId := generateUuid(t)

	handler.persister.GetAccountAccessGrantPersister().
		Create(models.AccountAccessGrant{
			ID:        grantId,
			IsActive:  true,
			CreatedAt: time.Now().UTC().Add(time.Duration(-15) * time.Minute),
			Ttl:       10,
		})

	err := handler.GetAccountShareGrantWithToken(grantId.String(), "booooooooooooooo", subjectId.String(), subjectId.String())
	assert.Error(t, err)
	assert.Equal(t, http.StatusRequestTimeout, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_GetAccountShareWithToken_Errors_IfTokenIsInvalid(t *testing.T) {
	handler := generateHandler(t)

	subjectId := generateUuid(t)
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

	err := handler.GetAccountShareGrantWithToken(grantId.String(), "booooooooooooooo", subjectId.String(), subjectId.String())
	assert.Error(t, err)
	assert.Equal(t, http.StatusNotFound, dto.ToHttpError(err).Code)
}

func Test_AccountSharingHandler_GetAccountShareWithToken_DoesNotError_IfGrantIsActiveAndTokenIsCorrect(t *testing.T) {
	handler := generateHandler(t)

	subjectId := generateUuid(t)
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

	err := handler.GetAccountShareGrantWithToken(grantId.String(), token, subjectId.String(), subjectId.String())
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

func generateHandler(t *testing.T) *AccountSharingHandler {
	handler, err := NewAccountSharingHandler(&config.Config{}, test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil), sessionManager{}, mailer{})
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
