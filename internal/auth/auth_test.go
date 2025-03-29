package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestJWT(t *testing.T) {
	// Test values
	tokenSecret := "supersecretkey"
	userID := uuid.New()
	expiration := time.Minute * 5

	// Generate JWT
	token, err := MakeJWT(userID, tokenSecret, expiration)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate JWT
	parsedUserID, err := ValidateJWT(token, tokenSecret)
	assert.NoError(t, err)
	assert.Equal(t, userID, parsedUserID)
}

func TestExpiredJWT(t *testing.T) {
	tokenSecret := "supersecretkey"
	userID := uuid.New()
	expiration := -time.Minute // Expired token

	// Generate expired JWT
	token, err := MakeJWT(userID, tokenSecret, expiration)
	assert.NoError(t, err)

	// Validate (should fail)
	_, err = ValidateJWT(token, tokenSecret)
	assert.Error(t, err)
}

func TestInvalidSecretJWT(t *testing.T) {
	tokenSecret := "supersecretkey"
	wrongSecret := "wrongkey"
	userID := uuid.New()
	expiration := time.Minute * 5

	// Generate JWT
	token, err := MakeJWT(userID, tokenSecret, expiration)
	assert.NoError(t, err)

	// Validate with wrong secret (should fail)
	_, err = ValidateJWT(token, wrongSecret)
	assert.Error(t, err)
}

func TestGetBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer my-token-value")

	token, err := GetBearerToken(headers)
	assert.NoError(t, err)
	assert.Equal(t, "my-token-value", token)
}

func TestGetBearerToken_MissingHeader(t *testing.T) {
	headers := http.Header{}

	token, err := GetBearerToken(headers)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestGetBearerToken_InvalidFormat(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "InvalidFormat")

	token, err := GetBearerToken(headers)
	assert.Error(t, err)
	assert.Empty(t, token)
}
