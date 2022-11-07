package handler

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
)

func TestSignatureFakerHandler_WhenSigningGrantAttestation_WhenExpireByTimeIsTrue_ReturnsSignature(t *testing.T) {
	handler := createSignatureFaker()
	relation := models.UserGuestRelation{
		ID:                      generateUuid(t),
		AssociatedAccessGrantId: generateUuid(t),
		ParentUserID:            generateUuid(t),
		GuestUserID:             generateUuid(t),
		ExpireByLogins:          false,
		ExpireByTime:            true,
		MinutesAllowed:          sql.NullInt32{Valid: true, Int32: 60},
		LoginsAllowed:           sql.NullInt32{Valid: false, Int32: 0},
		CreatedAt:               time.Now().UTC(),
	}
	handler.persister.GetUserGuestRelationPersister().Create(relation)
	handler.persister.GetUserPersister().Create(models.User{
		ID:       relation.ParentUserID,
		IsActive: true,
	})
	handler.persister.GetUserPersister().Create(models.User{
		ID:       relation.GuestUserID,
		IsActive: true,
	})
	handler.persister.GetAccountAccessGrantPersister().Create(models.AccountAccessGrant{
		ID:             relation.AssociatedAccessGrantId,
		IsActive:       false,
		UserId:         relation.ParentUserID,
		ExpireByLogins: relation.ExpireByLogins,
		ExpireByTime:   relation.ExpireByTime,
		MinutesAllowed: relation.MinutesAllowed,
		LoginsAllowed:  relation.LoginsAllowed,
	})

	grant := GrantAttestationObject{
		AccountAccessGrantId: relation.AssociatedAccessGrantId,
		GuestUserId:          relation.GuestUserID,
		ExpireByTime:         relation.ExpireByTime,
		ExpireByLogins:       relation.ExpireByLogins,
		MinutesAllowed:       int(relation.MinutesAllowed.Int32),
		LoginsAllowed:        int(relation.LoginsAllowed.Int32),
		CreatedAt:            relation.CreatedAt,
	}

	conv := convertToBase64(grant, t)

	request := fmt.Sprintf(`{"userId":%q, "challenge": %q}`, relation.ParentUserID, conv)

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/initialize-login-as-guest", strings.NewReader(request))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", generateJwt(t, relation.ParentUserID, relation.ParentUserID, 60))

	if assert.NoError(t, handler.SignChallengeAsUser(c)) {
		assert.Equal(t, rec.Code, http.StatusOK)
		response := SignChallengeResponse{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotEmpty(t, response.Signature)
	}
}

func Test_shortId_IsValid(t *testing.T) {
	id := GenerateShortID()
	assert.Equal(t, 27, len(id))
}

func convertToBase64(obj GrantAttestationObject, t *testing.T) string {
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	err := json.NewEncoder(encoder).Encode(obj)
	assert.NoError(t, err)
	encoder.Close()
	return buf.String()
}

func createSignatureFaker() *SignatureFakerHandler {
	p := test.NewPersister(users, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewSignatureFakerHandler(p)
	return handler
}
