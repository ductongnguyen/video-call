// Package auth provides authentication domain errors and utilities for consistent error handling.
package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	// errUserAlreadyExists is returned when a user with the given credentials already exists.
	errUserAlreadyExists = "user already exists"
	// errUserNotFound is returned when a user is not found.
	errUserNotFound = "user not found"
	// errInvalidCredentials is returned when login credentials are invalid.
	errInvalidCredentials = "invalid credentials"
	// errInvalidToken is returned when a token is invalid or expired.
	errInvalidToken = "invalid token"
	// errFailedToCheckUsername is returned when a username check fails.
	errFailedToCheckUsername = "failed to check username"
	// errFailedToCheckEmail is returned when an email check fails.
	errFailedToCheckEmail = "failed to check email"
	// errFailedToHashPassword is returned when password hashing fails.
	errFailedToHashPassword = "failed to hash password"
	// errFailedToRegisterUser is returned when user registration fails.
	errFailedToRegisterUser = "failed to register user"
)

var (
	// ErrUserAlreadyExists indicates that the user already exists.
	ErrUserAlreadyExists = errors.New(errUserAlreadyExists)
	// ErrUserNotFound indicates that the user was not found.
	ErrUserNotFound = errors.New(errUserNotFound)
	// ErrInvalidCredentials indicates invalid login credentials.
	ErrInvalidCredentials = errors.New(errInvalidCredentials)
	// ErrInvalidToken indicates an invalid or expired token.
	ErrInvalidToken = errors.New(errInvalidToken)
	// ErrFailedToCheckUsername indicates a failure to check username.
	ErrFailedToCheckUsername = errors.New(errFailedToCheckUsername)
	// ErrFailedToCheckEmail indicates a failure to check email.
	ErrFailedToCheckEmail = errors.New(errFailedToCheckEmail)
	// ErrFailedToHashPassword indicates a failure to hash password.
	ErrFailedToHashPassword = errors.New(errFailedToHashPassword)
	// ErrFailedToRegisterUser indicates a failure to register user.
	ErrFailedToRegisterUser = errors.New(errFailedToRegisterUser)
)

// MapError maps an authentication error to an HTTP status code and message.
// It provides a unified way to translate auth errors to HTTP responses.
func MapError(err error) (status int, message string) {
	// Handle JSON binding/unmarshal errors as 400 Bad Request
	switch err.(type) {
	case *json.UnmarshalTypeError, *json.SyntaxError:
		return http.StatusBadRequest, "Invalid request format"
	case *strconv.NumError:
		return http.StatusBadRequest, "Invalid parameter format"
	}
	if ginErr, ok := err.(*gin.Error); ok && ginErr.Type == gin.ErrorTypeBind {
		return http.StatusBadRequest, "Invalid request format"
	}

	switch {
	case errors.Is(err, ErrUserAlreadyExists):
		return http.StatusConflict, errUserAlreadyExists
	case errors.Is(err, ErrUserNotFound):
		return http.StatusNotFound, errUserNotFound
	case errors.Is(err, ErrInvalidCredentials):
		return http.StatusUnauthorized, errInvalidCredentials
	case errors.Is(err, ErrInvalidToken):
		return http.StatusUnauthorized, errInvalidToken
	case errors.Is(err, ErrFailedToCheckUsername):
		return http.StatusInternalServerError, errFailedToCheckUsername
	case errors.Is(err, ErrFailedToCheckEmail):
		return http.StatusInternalServerError, errFailedToCheckEmail
	case errors.Is(err, ErrFailedToHashPassword):
		return http.StatusInternalServerError, errFailedToHashPassword
	case errors.Is(err, ErrFailedToRegisterUser):
		return http.StatusInternalServerError, errFailedToRegisterUser
	default:
		return http.StatusInternalServerError, "Internal server error"
	}
}
