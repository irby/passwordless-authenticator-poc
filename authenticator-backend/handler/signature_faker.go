package handler

import (
	"encoding/base64"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/crypto/ecdsa"
	"github.com/teamhanko/hanko/backend/dto"
	"net/http"
)

type SignChallengeRequest struct {
	Email     string `json:"email" validate:"required"`
	Challenge string `json:"challenge" validate:"required"`
}

type SignChallengeResponse struct {
	Signature string `json:"signature"`
}

func SignChallengeAsUser(c echo.Context) error {
	var body SignChallengeRequest
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}
	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}
	result, err := ecdsa.SignChallengeForUser(body.Email, body.Challenge)
	if err != nil {
		return dto.ToHttpError(err)
	}
	data := base64.URLEncoding.EncodeToString(result)
	response := SignChallengeResponse{
		Signature: data,
	}
	return c.JSON(http.StatusOK, response)
}
