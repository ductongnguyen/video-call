package repository

import (
	"context"
	"errors"

	"video-call/internal/auth"
	"video-call/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	pkgErrors "video-call/pkg/errors"
)

// News Repository
type repo struct {
	db *gorm.DB
}

// News repository constructor
func NewRepository(db *gorm.DB) auth.Repository {
	return &repo{db: db}
}

// GetUserByID implements auth.Repository.
func (r *repo) GetUserByID(ctx context.Context, userId string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", userId).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Login implements auth.Repository.
func (r *repo) Login(ctx context.Context, user *models.User) (*models.User, error) {
	var u models.User

	if err := r.db.WithContext(ctx).Where("email = ?", user.Email).First(&u).Error; err != nil {
		return nil, err
	}

	// Compare hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(user.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &u, nil
}

// Register implements auth.Repository.
func (r *repo) Register(ctx context.Context, user *models.User) (*models.User, error) {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByUsername implements auth.Repository.
func (r *repo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgErrors.NotFound // Return nil if no record is found
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail implements auth.Repository.
func (r *repo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgErrors.NotFound // Return nil if no record is found
		}
		return nil, err
	}
	return &user, nil
}

// CreateRefreshToken implements auth.Repository.
func (r *repo) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return err
	}
	return nil
}

// GetRefreshTokenByToken implements auth.Repository.
func (r *repo) GetRefreshTokenByToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&rt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgErrors.NotFound // Return nil if no record is found
		}
		return nil, err
	}
	return &rt, nil
}

// RevokeRefreshToken implements auth.Repository.
func (r *repo) RevokeRefreshToken(ctx context.Context, token string) error {
	if err := r.db.WithContext(ctx).Where("token = ?", token).Delete(&models.RefreshToken{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *repo) GetUsers(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
