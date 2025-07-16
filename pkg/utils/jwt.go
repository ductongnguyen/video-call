package utils

import (
	"time"

	"github.com/ductongnguyen/vivy-chat/config"
	"github.com/ductongnguyen/vivy-chat/internal/models"
	"github.com/golang-jwt/jwt"
)

// JWT Claims struct
type Claims struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	jwt.StandardClaims
}

// Generate new JWT Token
func GenerateJWTToken(user *models.User, config *config.Config) (string, time.Time, error) {
	// Register the JWT claims, which includes the username and expiry time
	expiredAt := time.Now().Add(time.Minute * 60)
	claims := &Claims{
		Id:       user.ID,
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 60).Unix(),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Register the JWT string
	tokenString, err := token.SignedString([]byte(config.Server.JwtSecretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiredAt, nil
}

// ValidateJWTToken validates the JWT token
func ValidateJWTToken(tokenString string, config *config.Config) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Server.JwtSecretKey), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}
