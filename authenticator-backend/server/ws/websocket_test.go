package ws

import (
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	jwt2 "github.com/teamhanko/hanko/backend/crypto/jwt"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/test/mocks"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebSocketHandler_getSessionTokenFromContext_NoError_WhenSurrogateIsEqualToSubject(t *testing.T) {
	handler := generateWebsocketHandler(t)
	c := generateContext()
	uId1 := generateUuid()
	assignSessionToken(t, c, &uId1, &uId1)

	token, err := handler.getSessionTokenFromContext(c)

	assert.NoError(t, err)
	assert.NotNil(t, token)
}

func TestWebSocketHandler_getSessionTokenFromContext_Errors_WhenSurrogateIsNotEqualToSubject(t *testing.T) {
	handler := generateWebsocketHandler(t)
	c := generateContext()
	uId1 := generateUuid()
	uId2 := generateUuid()
	assignSessionToken(t, c, &uId1, &uId2)

	token, err := handler.getSessionTokenFromContext(c)

	assert.Error(t, err)
	assert.Nil(t, token)

	httpError := dto.ToHttpError(err)
	assert.Equal(t, http.StatusForbidden, httpError.Code)
}

func TestWebSocketHandler_getSessionTokenFromContext_Errors_WhenSessionTokenIsInvalid(t *testing.T) {
	handler := generateWebsocketHandler(t)
	c := generateContext()
	assignSessionToken(t, c, nil, nil)

	token, err := handler.getSessionTokenFromContext(c)

	assert.Error(t, err)
	assert.Nil(t, token)

	httpError := dto.ToHttpError(err)
	assert.Equal(t, http.StatusInternalServerError, httpError.Code)
}

/* == Private == */
func generateWebsocketHandler(t *testing.T) *WebsocketHandler {
	config := mocks.GenerateMockConfig()
	handler, err := NewWebsocketHandler(&config, nil, nil, nil)
	assert.NoError(t, err)
	return handler
}

func generateContext() echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodConnect, "/ws", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	return c
}

func assignSessionToken(t *testing.T, c echo.Context, subjectId *uuid.UUID, surrogateId *uuid.UUID) {
	token := jwt.New()

	if subjectId != nil {
		err := token.Set(jwt.SubjectKey, subjectId.String())
		assert.NoError(t, err)
	}

	if surrogateId != nil {
		err := token.Set(jwt2.SurrogateKey, surrogateId.String())
		assert.NoError(t, err)
	}

	c.Set("session", token)
}

func generateUuid() uuid.UUID {
	uId, _ := uuid.NewV4()
	return uId
}
