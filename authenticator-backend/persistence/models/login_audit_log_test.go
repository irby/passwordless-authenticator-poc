package models

import (
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_IsSurrogateAndGuestPopulated_WhenBothAreNil_ReturnsTrue(t *testing.T) {
	log := LoginAuditLog{
		SurrogateUserId:     nil,
		UserGuestRelationId: nil,
	}
	result := IsSurrogateAndGuestPopulated(&log)
	assert.True(t, result())
}

func Test_IsSurrogateAndGuestPopulated_WhenBothAreNotNil_ReturnsTrue(t *testing.T) {
	uuid1, _ := uuid.NewV4()
	uuid2, _ := uuid.NewV4()
	log := LoginAuditLog{
		SurrogateUserId:     &uuid1,
		UserGuestRelationId: &uuid2,
	}
	result := IsSurrogateAndGuestPopulated(&log)
	assert.True(t, result())
}

func Test_IsSurrogateAndGuestPopulated_WhenSurrogateIsNotNilButUserGuestRelationIsNil_ReturnsFalse(t *testing.T) {
	uuid1, _ := uuid.NewV4()
	log := LoginAuditLog{
		SurrogateUserId:     &uuid1,
		UserGuestRelationId: nil,
	}
	result := IsSurrogateAndGuestPopulated(&log)
	assert.False(t, result())
}

func Test_IsSurrogateAndGuestPopulated_WhenSurrogateIsNilButUserGuestRelationIsNotNil_ReturnsFalse(t *testing.T) {
	uuid1, _ := uuid.NewV4()
	log := LoginAuditLog{
		SurrogateUserId:     nil,
		UserGuestRelationId: &uuid1,
	}
	result := IsSurrogateAndGuestPopulated(&log)
	assert.False(t, result())
}
