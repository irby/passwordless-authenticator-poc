package models

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_LoginsAndMinutesMutuallyExclusive_WhenExclusive_ReturnsTrue(t *testing.T) {
	grant := AccountAccessGrant{
		LoginsAllowed:  sql.NullInt32{Valid: true},
		MinutesAllowed: sql.NullInt32{Valid: false},
		ExpireByLogins: true,
		ExpireByTime:   false,
	}
	result := LoginsAndMinutesMutuallyExclusive(&grant)
	assert.True(t, result())
}

func Test_LoginsAndMinutesMutuallyExclusive_WhenExpireByNotExclusive_ReturnsFalse(t *testing.T) {
	grant := AccountAccessGrant{
		LoginsAllowed:  sql.NullInt32{Valid: true},
		MinutesAllowed: sql.NullInt32{Valid: false},
		ExpireByLogins: true,
		ExpireByTime:   true,
	}
	result := LoginsAndMinutesMutuallyExclusive(&grant)
	assert.False(t, result())
}

func Test_LoginsAndMinutesMutuallyExclusive_WhenAllowedValidNotExclusive_ReturnsFalse(t *testing.T) {
	grant := AccountAccessGrant{
		LoginsAllowed:  sql.NullInt32{Valid: true},
		MinutesAllowed: sql.NullInt32{Valid: true},
		ExpireByLogins: true,
		ExpireByTime:   false,
	}
	result := LoginsAndMinutesMutuallyExclusive(&grant)
	assert.False(t, result())
}

func Test_IsLoginsAllowedPopulated_WhenExpireByLoginsFalse_ReturnsTrue(t *testing.T) {
	grant := AccountAccessGrant{
		ExpireByLogins: false,
	}
	result := IsLoginsAllowedPopulated(&grant)
	assert.True(t, result())
}

func Test_IsLoginsAllowedPopulated_WhenExpireByLoginsTrueAndModelIsValid_ReturnsTrue(t *testing.T) {
	grant := AccountAccessGrant{
		LoginsAllowed:  sql.NullInt32{Int32: 1, Valid: true},
		ExpireByLogins: true,
	}
	result := IsLoginsAllowedPopulated(&grant)
	assert.True(t, result())
}

func Test_IsLoginsAllowedPopulated_WhenExpireByLoginsTrueAndInt32Is0_ReturnsFalse(t *testing.T) {
	grant := AccountAccessGrant{
		LoginsAllowed:  sql.NullInt32{Int32: 0, Valid: true},
		ExpireByLogins: true,
	}
	result := IsLoginsAllowedPopulated(&grant)
	assert.False(t, result())
}

func Test_IsLoginsAllowedPopulated_WhenExpireByLoginsTrueAndValidIsFalse_ReturnsFalse(t *testing.T) {
	grant := AccountAccessGrant{
		LoginsAllowed:  sql.NullInt32{Int32: 1, Valid: false},
		ExpireByLogins: true,
	}
	result := IsLoginsAllowedPopulated(&grant)
	assert.False(t, result())
}
