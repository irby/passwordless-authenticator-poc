package handler

import (
	"encoding/base64"
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
	Email     string `json:"email" validate:"required"`
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

	credential, err := h.getWebauthnCredentialForUser(body.Email)
	if err != nil {
		return dto.ToHttpError(err)
	}

	result, err := ecdsa.SignChallengeForUser(body.Email, body.Challenge)
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

func (h *SignatureFakerHandler) getWebauthnCredentialForUser(email string) (*models.WebauthnCredential, error) {
	user, err := h.persister.GetUserPersister().GetByEmail(email)
	if err != nil {
		return nil, err
	}
	credential, err := h.persister.GetWebauthnCredentialPersister().GetFromUser(user.ID)
	if err != nil {
		return nil, err
	}

	if len(credential) > 0 {
		return &credential[0], nil
	}
	newCred, err := h.createWebauthnCredentialForUser(user)
	return newCred, nil
}

func (h *SignatureFakerHandler) createWebauthnCredentialForUser(user *models.User) (*models.WebauthnCredential, error) {
	credential := models.WebauthnCredential{
		ID:              GenerateShortID(),
		UserId:          user.ID,
		PublicKey:       "",
		AttestationType: "none",
		AAGUID:          uuid.Nil,
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
