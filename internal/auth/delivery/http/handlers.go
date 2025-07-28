package http

import (
	"net/http"

	"video-call/config"
	"video-call/internal/auth"
	"video-call/internal/models"
	"video-call/pkg/logger"
	"video-call/pkg/response"
	"video-call/pkg/utils"

	"github.com/gin-gonic/gin"
)

// For testing purposes
var GenerateJWTTokenFunc = utils.GenerateJWTToken

// News handlers
type handlers struct {
	cfg     *config.Config
	usecase auth.UseCase
	logger  logger.Logger
}

// NewNewsHandlers News handlers constructor
func NewHandlers(cfg *config.Config, usecase auth.UseCase, logger logger.Logger) auth.Handlers {
	return &handlers{cfg: cfg, usecase: usecase, logger: logger}
}

// GetUserByID godoc
// @Summary      Get user by ID
// @Description  Get details of a user by their ID
// @Tags         auth
// @Param        userId   path      string  true  "User ID (UUID)"
// @Success      200      {object}  UserResponse
// @Failure      400,404  {object}  response.Response
// @Router       /auth/user/{userId} [get]
func (h *handlers) GetUserByID(c *gin.Context) {
	userId := c.Param("userId")

	user, err := h.usecase.GetUserByID(c.Request.Context(), userId)
	if err != nil {
		response.WithMappedError(c, err, auth.MapError)
		return
	}

	responseUser := FromUserModel(user)
	response.WithOK(c, responseUser)

}

// Login godoc
// @Summary      User login
// @Description  Authenticate user and return JWT and refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        loginRequest  body      LoginRequest  true  "Login credentials"
// @Success      200           {object}  AuthSuccessResponse
// @Failure      400,401       {object}  response.Response
// @Router       /auth/login [post]
func (h *handlers) Login(c *gin.Context) {

	var loginRequest LoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		response.WithMappedError(c, err, auth.MapError)
		return
	}

	loginAttempt := &models.User{
		Email:    loginRequest.Email,
		Password: loginRequest.Password,
	}
	user, err := h.usecase.Login(c.Request.Context(), loginAttempt)
	if err != nil {
		response.WithMappedError(c, err, auth.MapError)
		return
	}

	tokenString, expiredAt, err := utils.GenerateJWTToken(user, h.cfg)
	if err != nil {
		response.WithMappedError(c, err, auth.MapError)
		return
	}

	refreshToken, refreshTokenExpiresAt, err := h.usecase.GenerateRefreshToken(c.Request.Context(), user.ID)
	if err != nil {
		response.WithMappedError(c, err, auth.MapError)
		return
	}

	response := AuthSuccessResponse{
		Token:                 tokenString,
		ExpiresAt:             FormatTime(expiredAt),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: FormatTime(refreshTokenExpiresAt),
		User:                  FromUserModel(user),
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken godoc
// @Summary      Refresh JWT token
// @Description  Issue a new JWT and refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        refreshTokenRequest  body      RefreshTokenRequest  true  "Refresh token payload"
// @Success      200                 {object}  RefreshTokenResponse
// @Failure      400,401             {object}  response.Response
// @Router       /auth/refresh [post]
func (h *handlers) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WithMappedError(c, err, auth.MapError)
		return
	}

	rt, err := h.usecase.ValidateRefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil || rt.Revoked {
		response.WithMappedError(c, auth.ErrInvalidToken, auth.MapError)
		return
	}

	user, err := h.usecase.GetUserByID(c.Request.Context(), rt.UserID)
	if err != nil {
		response.WithMappedError(c, err, auth.MapError)
		return
	}

	// // Optionally revoke the used refresh token (rotation)
	// h.usecase.RevokeRefreshToken(c.Request.Context(), req.RefreshToken)

	tokenString, expiredAt, err := utils.GenerateJWTToken(user, h.cfg)
	if err != nil {
		response.WithMappedError(c, err, auth.MapError)
		return
	}

	newRefreshToken, refreshExp, err := h.usecase.GenerateRefreshToken(c.Request.Context(), user.ID)
	if err != nil {
		response.WithMappedError(c, err, auth.MapError)
		return
	}

	response := RefreshTokenResponse{
		Token:        tokenString,
		ExpiresAt:    FormatTime(expiredAt),
		RefreshToken: newRefreshToken,
		RefreshExp:   FormatTime(refreshExp),
	}
	c.JSON(http.StatusOK, response)
}

// Register godoc
// @Summary      Register new user
// @Description  Create a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        registerRequest  body      RegisterRequest  true  "Registration info"
// @Success      200              {object}  UserResponse
// @Failure      400,409          {object}  response.Response
// @Router       /auth/register [post]
func (h *handlers) Register(c *gin.Context) {
	var registerRequest RegisterRequest
	if err := c.ShouldBindJSON(&registerRequest); err != nil {
		response.WithMappedError(c, err, auth.MapError)
		return
	}

	newUser := &models.User{
		Username: registerRequest.Username,
		Email:    registerRequest.Email,
		Password: registerRequest.Password,
		Role:     models.UserRole(registerRequest.Role),
	}
	user, err := h.usecase.Register(c.Request.Context(), newUser)
	if err != nil {
		response.WithMappedError(c, err, auth.MapError)
		return
	}

	responseUser := FromUserModel(user)
	response.WithOK(c, responseUser)
}

func (h *handlers) GetUsers(c *gin.Context) {
	users, err := h.usecase.GetUsers(c.Request.Context())
	if err != nil {
		response.WithMappedError(c, err, auth.MapError)
		return
	}
	responseUsers := FromUserModelList(users)
	response.WithOK(c, responseUsers)
}
