package utils

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPasswordBcrypt(password string) (string, error) {
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedPasswordBytes), nil
}

// GetUserIDFromCtx retrieves the user ID from the context.
func GetUserIDFromCtx(ctx context.Context) (uint64, error) {
	user, err := GetUserFromCtx(ctx)
	if err != nil {
		return 0, err
	}

	// Convert string ID to uint64 if needed
	// Note: This assumes the ID is a numeric string
	// You might want to adjust this based on your actual ID format
	var id uint64
	_, err = fmt.Sscanf(user.ID, "%d", &id)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID format: %v", err)
	}
	return id, nil
}
