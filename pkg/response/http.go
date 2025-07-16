package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Common error variables
var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrInvalidUserID   = errors.New("invalid user ID")
	ErrInvalidRequest  = errors.New("invalid request")
	ErrInternalServer  = errors.New("internal server error")
	ErrNotFound        = errors.New("not found")
	ErrForbidden       = errors.New("forbidden")
)

const (
	CodeOK = 0
)

const (
	MessageOK = "Success"
)

type Response struct {
	Message string `json:"message,omitempty"`
	Result  any    `json:"result,omitempty"`
}

func WithOK(c *gin.Context, data any) {
	WithCode(c, http.StatusOK, data)
}

func WithNoContent(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

func WithCode(c *gin.Context, code int, data any) {
	c.JSON(code, Response{
		Message: MessageOK,
		Result:  data,
	})
}

func WithError(c *gin.Context, err error) {
	statusCode := http.StatusInternalServerError
	message := "Internal server error"

	// Map common error messages to appropriate status codes
	switch {
	case errors.Is(err, ErrUnauthorized):
		statusCode = http.StatusUnauthorized
		message = "Unauthorized"
	case errors.Is(err, ErrNotFound):
		statusCode = http.StatusNotFound
		message = "Not found"
	case errors.Is(err, ErrForbidden):
		statusCode = http.StatusForbidden
		message = "Forbidden"
	case errors.Is(err, ErrInvalidRequest):
		statusCode = http.StatusBadRequest
		message = "Invalid request"
	}

	c.JSON(statusCode, Response{
		Message: message,
		Result:  nil,
	})
}

func WithErrorCode(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Message: message,
	})
}

// WithData sends a JSON response with status code and data
func WithData(c *gin.Context, statusCode int, data any) {
	c.JSON(statusCode, Response{
		Message: MessageOK,
		Result:  data,
	})
}

// You call this in handlers for domain-layer errors
func WithMappedError(c *gin.Context, err error, mapFunc func(error) (int, string)) {
	code, msg := mapFunc(err)
	WithErrorCode(c, code, msg)
}
