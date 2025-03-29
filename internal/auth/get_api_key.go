package auth

import (
	"errors"
	"net/http"
	"strings"
)

var ErrInvalidAuthHeader = errors.New("invalid authorization header format")

// GetAPIKey extracts the API key from the "Authorization" header.
func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", ErrInvalidAuthHeader
	}

	// Ensure header follows format: "Authorization: ApiKey THE_KEY_HERE"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "ApiKey" {
		return "", ErrInvalidAuthHeader
	}

	return parts[1], nil // Return the extracted API key
}
