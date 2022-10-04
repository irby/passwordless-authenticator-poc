package handler

import (
	_ "github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/test"
	"testing"
)

func TestNewAccountSharingHandler(t *testing.T) {
	handler, err := NewAccountSharingHandler(&config.Config{}, test.NewPersister(nil, nil, nil, nil, nil, nil, nil), sessionManager{}, mailer{})
	assert.NoError(t, err)
	assert.NotEmpty(t, handler)
}
