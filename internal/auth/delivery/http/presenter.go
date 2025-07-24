package http

import (
	"time"

	"video-call/internal/models"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthSuccessResponse struct {
	Token                 string       `json:"token"`
	ExpiresAt             string       `json:"expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt string       `json:"refresh_token_expires_at"`
	User                  UserResponse `json:"user"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=admin user"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenResponse struct {
	Token        string `json:"token"`
	ExpiresAt    string `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
	RefreshExp   string `json:"refresh_expires_at"`
}

func FormatTime(time time.Time) string {
	return time.Format("2006-01-02 15:04:05")
}

func FromUserModel(user *models.User) UserResponse {
	if user == nil {
		return UserResponse{}
	}

	return UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      string(user.Role),
		CreatedAt: FormatTime(user.CreatedAt),
		UpdatedAt: FormatTime(user.UpdatedAt),
	}

}

func FromUserModelList(users []*models.User) []UserResponse {
	if users == nil {
		return []UserResponse{}
	}

	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = FromUserModel(user)
	}

	return userResponses
}
