package usecase

import (
	"context"
	"fmt"
	"time"

	"video-call/config"
	"video-call/internal/auth"
	"video-call/internal/models"
	"video-call/pkg/logger"
	"video-call/pkg/utils"

	pkgErrors "video-call/pkg/errors"
)

const (
	basePrefix    = "api-user:"
	cacheDuration = 3600
)

// News UseCase constructor
func NewUseCase(cfg *config.Config, repo auth.Repository, redisRepo auth.RedisRepository, logger logger.Logger) auth.UseCase {
	return &usecase{cfg: cfg, repo: repo, redisRepo: redisRepo, logger: logger}
}

// useCase

// GenerateRefreshToken generates and stores a new refresh token for a user
func (u *usecase) GenerateRefreshToken(ctx context.Context, userID string) (string, time.Time, error) {
	token, err := utils.GenerateRefreshToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	rt := &models.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	if err := u.repo.CreateRefreshToken(ctx, rt); err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

// ValidateRefreshToken checks if a refresh token is valid (not revoked/expired)
func (u *usecase) ValidateRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	return u.repo.GetRefreshTokenByToken(ctx, token)
}

// RevokeRefreshToken marks a refresh token as revoked
func (u *usecase) RevokeRefreshToken(ctx context.Context, token string) error {
	return u.repo.RevokeRefreshToken(ctx, token)
}

type usecase struct {
	cfg       *config.Config
	repo      auth.Repository
	redisRepo auth.RedisRepository
	logger    logger.Logger
}

// GetUserByID implements auth.UseCase.
func (u *usecase) GetUserByID(ctx context.Context, userId string) (*models.User, error) {
	cacheKey := fmt.Sprintf("%s%s", basePrefix, userId)

	user, err := u.redisRepo.GetUserByIDCtx(ctx, cacheKey)

	if err != nil {
		u.logger.Errorf(ctx, "Redis error fetching user %s (key: %s): %v. Attempting DB lookup.", userId, cacheKey, err)

		user, err = u.repo.GetUserByID(ctx, userId)
		if err != nil {
			return nil, auth.ErrUserNotFound
		}

		cacheSetErr := u.redisRepo.SetUserByIDCtx(ctx, cacheKey, user)
		if cacheSetErr != nil {
			u.logger.Errorf(ctx, "Failed to set user %s in cache (key: %s): %v", userId, cacheKey, cacheSetErr)
		}

		return user, nil
	}

	u.logger.Debugf(ctx, "User %s found in cache (key: %s).", userId, cacheKey)

	return user, nil
}

// Login implements auth.UseCase.
func (u *usecase) Login(ctx context.Context, user *models.User) (*models.User, error) {
	user, err := u.repo.Login(ctx, user)
	if err != nil {
		return nil, auth.ErrInvalidCredentials
	}

	return user, nil
}

// Register implements auth.UseCase.
func (u *usecase) Register(ctx context.Context, user *models.User) (*models.User, error) {
	// Check if username already exists
	if _, err := u.repo.GetUserByUsername(ctx, user.Username); err == nil {
		return nil, auth.ErrUserAlreadyExists
	} else if err != pkgErrors.NotFound {
		return nil, auth.ErrFailedToCheckUsername
	}

	// Check if email already exists
	if _, err := u.repo.GetUserByEmail(ctx, user.Email); err == nil {
		return nil, auth.ErrUserAlreadyExists
	} else if err != pkgErrors.NotFound {
		return nil, auth.ErrFailedToCheckEmail
	}

	// Hash the password BEFORE saving
	hashedPassword, err := utils.HashPasswordBcrypt(user.Password)
	if err != nil {
		return nil, auth.ErrFailedToHashPassword
	}
	user.Password = hashedPassword

	// Save user with hashed password
	user, err = u.repo.Register(ctx, user)
	if err != nil {
		return nil, auth.ErrFailedToRegisterUser
	}

	// Set user in cache (optional)
	cacheKey := fmt.Sprintf("%s%s", basePrefix, user.ID)
	if err := u.redisRepo.SetUserByIDCtx(ctx, cacheKey, user); err != nil {
		u.logger.Errorf(ctx, "Failed to set user %s in cache (key: %s): %v", user.ID, cacheKey, err)
	}

	return user, nil
}
