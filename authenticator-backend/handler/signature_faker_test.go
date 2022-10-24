package handler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_shortId_IsValid(t *testing.T) {
	id := GenerateShortID()
	assert.Equal(t, 27, len(id))
}
