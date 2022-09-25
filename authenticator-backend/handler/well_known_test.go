package handler

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/config"
	hankoJwk "github.com/teamhanko/hanko/backend/crypto/jwk"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"net/http/httptest"
	"testing"
)

type faultyJwkManager struct {
}

func (f faultyJwkManager) GenerateKey() (jwk.Key, error) {
	panic("implement me")
}

func (f faultyJwkManager) GetPublicKeys() (jwk.Set, error) {
	return nil, errors.New("No Public Keys!")
}

func (f faultyJwkManager) GetSigningKey() (jwk.Key, error) {
	panic("implement me")
}

func TestSomethingWrongWithKeys(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	jwkMan := faultyJwkManager{}
	cfg := config.Config{Password: config.Password{Enabled: true}}
	h, err := NewWellKnownHandler(cfg, jwkMan)
	assert.NoError(t, err)

	err = h.GetPublicKeys(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusInternalServerError, httpError.Code)
		assert.Equal(t, "No Public Keys!", httpError.Internal.Error())
	}
}

func TestGetPublicKeys(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	jwkMan, err := hankoJwk.NewDefaultManager([]string{"superRandomAndSecure"}, test.NewJwkPersister(nil))
	assert.NoError(t, err)
	cfg := config.Config{Password: config.Password{Enabled: true}}
	h, err := NewWellKnownHandler(cfg, jwkMan)
	assert.NoError(t, err)

	if assert.NoError(t, h.GetPublicKeys(c)) {
		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		set := jwk.NewSet()
		err = json.Unmarshal(rec.Body.Bytes(), set)
		assert.Equal(t, 1, set.Len())
		assert.NoError(t, err)
	}
}
