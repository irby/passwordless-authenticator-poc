package intern

import (
	"encoding/base64"
	"fmt"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

func WebauthnSessionDataFromModel(data *models.WebauthnSessionData) *webauthn.SessionData {
	var allowedCredentials [][]byte
	fmt.Println("Webauthn session data", data)
	for _, credential := range data.AllowedCredentials {
		credentialId, err := base64.RawURLEncoding.DecodeString(credential.CredentialId)
		fmt.Println("credentialId from webauthn session data", credentialId)
		if err != nil {
			continue
		}
		allowedCredentials = append(allowedCredentials, credentialId)
	}
	var userId []byte = nil
	if !data.UserId.IsNil() {
		userId = data.UserId.Bytes()
	}
	return &webauthn.SessionData{
		Challenge:            data.Challenge,
		UserID:               userId,
		AllowedCredentialIDs: allowedCredentials,
		UserVerification:     protocol.UserVerificationRequirement(data.UserVerification),
	}
}

func WebauthnSessionDataToModel(data *webauthn.SessionData, operation models.Operation) *models.WebauthnSessionData {
	id, _ := uuid.NewV4()
	userId, _ := uuid.FromBytes(data.UserID)
	now := time.Now()

	var allowedCredentials []models.WebauthnSessionDataAllowedCredential
	for _, credentialID := range data.AllowedCredentialIDs {
		aId, _ := uuid.NewV4()
		allowedCredential := models.WebauthnSessionDataAllowedCredential{
			ID:                    aId,
			CredentialId:          base64.RawURLEncoding.EncodeToString(credentialID),
			WebauthnSessionDataID: id,
			CreatedAt:             now,
			UpdatedAt:             now,
		}

		allowedCredentials = append(allowedCredentials, allowedCredential)
	}

	return &models.WebauthnSessionData{
		ID:                 id,
		Challenge:          data.Challenge,
		UserId:             userId,
		UserVerification:   string(data.UserVerification),
		CreatedAt:          now,
		UpdatedAt:          now,
		Operation:          operation,
		AllowedCredentials: allowedCredentials,
	}
}
