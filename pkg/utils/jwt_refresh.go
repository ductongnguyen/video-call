package utils

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

// GenerateRefreshToken generates a secure random string for refresh tokens
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// RefreshTokenDuration is the default duration for refresh tokens (e.g., 7 days)
const RefreshTokenDuration = 7 * 24 * time.Hour
