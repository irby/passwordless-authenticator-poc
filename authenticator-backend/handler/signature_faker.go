package handler

import (
	ecdsa2 "crypto/ecdsa"
	"encoding/base64"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/crypto/ecdsa"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"math/rand"
	"net/http"
	"time"
)

type SignatureFakerHandler struct {
	persister persistence.Persister
}

func NewSignatureFakerHandler(persister persistence.Persister) *SignatureFakerHandler {
	return &SignatureFakerHandler{persister: persister}
}

type SignChallengeRequest struct {
	UserId    string `json:"userId" validate:"required"`
	Challenge string `json:"challenge" validate:"required"`
}

type SignChallengeResponse struct {
	ID                string `json:"id"`
	Signature         string `json:"signature"`
	ClientDataJson    string `json:"clientDataJson"`
	AuthenticatorData string `json:"authenticatorData"`
	UserHandle        string `json:"userHandle"`
}

func (h *SignatureFakerHandler) SignChallengeAsUser(c echo.Context) error {
	var body SignChallengeRequest
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}
	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	key, credential, err := h.getWebauthnCredentialForUser(body.UserId)
	if err != nil {
		return dto.ToHttpError(err)
	}

	result, err := ecdsa.SignChallengeForUser(key.PrivateKey, body.Challenge)
	if err != nil {
		return dto.ToHttpError(err)
	}

	authData, _ := ecdsa.GetAuthenticatorData()
	clientDataJson, _ := ecdsa.GetClientData(body.Challenge)

	response := SignChallengeResponse{
		ID:                credential.ID,
		Signature:         base64.RawURLEncoding.EncodeToString(result),
		AuthenticatorData: base64.RawURLEncoding.EncodeToString(authData),
		ClientDataJson:    base64.RawURLEncoding.EncodeToString(clientDataJson),
		UserHandle:        base64.RawURLEncoding.EncodeToString(credential.UserId.Bytes()),
	}

	return c.JSON(http.StatusOK, response)
}

func (h *SignatureFakerHandler) getWebauthnCredentialForUser(userId string) (*models.WebauthnCredentialsPrivateKey, *models.WebauthnCredential, error) {
	user, err := h.persister.GetUserPersister().Get(uuid.FromStringOrNil(userId))
	if err != nil {
		return nil, nil, err
	}

	credentials, err := h.persister.GetWebauthnCredentialPersister().GetFromUser(user.ID)
	if err != nil {
		return nil, nil, err
	}

	var credential *models.WebauthnCredential

	for _, cred := range credentials {
		if cred.AAGUID != uuid.Nil {
			credential = &cred
		}
	}

	var key *ecdsa2.PrivateKey

	if credential == nil {
		key, err = ecdsa.GeneratePrivateKey()
		if err != nil {
			return nil, nil, err
		}

		pub, err := ecdsa.GenerateEC2PublicKeyDataFromPrivateKey(*key)
		if err != nil {
			return nil, nil, err
		}

		newCred, err := h.createWebauthnCredentialForUser(user, pub)
		if err != nil {
			return nil, nil, err
		}
		credential = newCred
	}

	privateKey, err := h.persister.GetWebauthnCredentialsPrivateKeyPersister().Get(credential.ID)
	if err != nil {
		return nil, nil, err
	}
	if privateKey == nil {
		if key == nil {
			return nil, nil, errors.New("a private key must be generated prior to creating a new private key")
		}
		privateKey, err = h.createWebauthnCredentialsPrivateKeyForUser(credential, key)
		if err != nil {
			return nil, nil, err
		}
	}

	return privateKey, credential, nil
}

func (h *SignatureFakerHandler) createWebauthnCredentialForUser(user *models.User, pubKey string) (*models.WebauthnCredential, error) {
	uId, _ := uuid.NewV4()
	credential := models.WebauthnCredential{
		ID:              GenerateShortID(),
		UserId:          user.ID,
		PublicKey:       pubKey,
		AttestationType: "none",
		AAGUID:          uId,
		SignCount:       0,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	err := h.persister.GetWebauthnCredentialPersister().Create(credential)
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

func (h *SignatureFakerHandler) createWebauthnCredentialsPrivateKeyForUser(credential *models.WebauthnCredential, privateKey *ecdsa2.PrivateKey) (*models.WebauthnCredentialsPrivateKey, error) {
	key := models.WebauthnCredentialsPrivateKey{
		ID:         credential.ID,
		PrivateKey: privateKey.D.String(),
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	err := h.persister.GetWebauthnCredentialsPrivateKeyPersister().Create(key)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func GenerateShortID() string {
	var upperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	var id string

	firstChar := upperChars[rand.Intn(len(upperChars))]
	id += string(firstChar)
	id += "-"
	for i := 0; i < 10; i++ {
		id += string(chars[rand.Intn(len(chars))])
	}
	id += "-"
	for i := 0; i < 14; i++ {
		id += string(chars[rand.Intn(len(chars))])
	}

	return id
}
