package grpc

// import (
// 	"context"

// 	"video-call/config"
// 	"video-call/internal/auth"
// 	"video-call/internal/models"
// 	"video-call/pkg/logger"
// 	"video-call/pkg/utils"
// 	proto "video-call/proto/v1/auth"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// 	"google.golang.org/protobuf/types/known/timestamppb"
// )

// type AuthServiceHandler struct {
// 	proto.UnimplementedAuthServiceServer
// 	cfg     *config.Config
// 	usecase auth.UseCase
// 	logger  logger.Logger
// }

// func NewAuthServiceHandler(cfg *config.Config, usecase auth.UseCase, logger logger.Logger) *AuthServiceHandler {
// 	return &AuthServiceHandler{usecase: usecase, cfg: cfg, logger: logger}
// }

// func (h *AuthServiceHandler) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
// 	user, err := h.usecase.Login(ctx, &models.User{
// 		Email:    req.GetEmail(),
// 		Password: req.GetPassword(),
// 	})
// 	if err != nil {
// 		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials: %v", err)
// 	}

// 	accessToken, _, err := utils.GenerateJWTToken(user, h.cfg)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "failed to generate access token: %v", err)
// 	}

// 	refreshToken, _, err := h.usecase.GenerateRefreshToken(ctx, user.ID)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "failed to generate refresh token: %v", err)
// 	}

// 	return &proto.LoginResponse{
// 		AccessToken:  accessToken,
// 		RefreshToken: refreshToken,
// 	}, nil
// }

// func (h *AuthServiceHandler) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
// 	user, err := h.usecase.Register(ctx, &models.User{
// 		Username: req.GetUsername(),
// 		Email:    req.GetEmail(),
// 		Password: req.GetPassword(),
// 		Role:     models.RoleUser, // Default to user role
// 	})
// 	if err != nil {
// 		return nil, status.Errorf(codes.InvalidArgument, "failed to register user: %v", err)
// 	}

// 	accessToken, _, err := utils.GenerateJWTToken(user, h.cfg)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "failed to generate access token: %v", err)
// 	}

// 	refreshToken, _, err := h.usecase.GenerateRefreshToken(ctx, user.ID)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "failed to generate refresh token: %v", err)
// 	}

// 	return &proto.RegisterResponse{
// 		AccessToken:  accessToken,
// 		RefreshToken: refreshToken,
// 	}, nil
// }

// func (h *AuthServiceHandler) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
// 	claims, err := utils.ValidateJWTToken(req.GetToken(), h.cfg)
// 	if err != nil {
// 		return &proto.ValidateTokenResponse{Valid: false, Error: err.Error()}, nil
// 	}

// 	return &proto.ValidateTokenResponse{
// 		Valid:    true,
// 		UserId:   int64(claims.Id),
// 		Username: claims.Username,
// 	}, nil
// }

// func (h *AuthServiceHandler) RefreshToken(ctx context.Context, req *proto.RefreshTokenRequest) (*proto.RefreshTokenResponse, error) {
// 	refreshToken, err := h.usecase.ValidateRefreshToken(ctx, req.GetRefreshToken())
// 	if err != nil {
// 		return nil, status.Errorf(codes.Unauthenticated, "invalid refresh token: %v", err)
// 	}

// 	user, err := h.usecase.GetUserByID(ctx, refreshToken.UserID)
// 	if err != nil {
// 		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
// 	}

// 	accessToken, _, err := utils.GenerateJWTToken(user, h.cfg)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "failed to generate access token: %v", err)
// 	}

// 	newRefreshToken, _, err := h.usecase.GenerateRefreshToken(ctx, user.ID)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "failed to generate refresh token: %v", err)
// 	}

// 	return &proto.RefreshTokenResponse{
// 		AccessToken:  accessToken,
// 		RefreshToken: newRefreshToken,
// 	}, nil
// }

// func (h *AuthServiceHandler) GetCurrentUser(ctx context.Context, req *proto.GetCurrentUserRequest) (*proto.GetCurrentUserResponse, error) {
// 	claims, err := utils.ValidateJWTToken(req.GetToken(), h.cfg)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
// 	}

// 	user, err := h.usecase.GetUserByID(ctx, claims.Id)
// 	if err != nil {
// 		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
// 	}

// 	return &proto.GetCurrentUserResponse{
// 		UserId:    int64(user.ID),
// 		Username:  user.Username,
// 		Email:     user.Email,
// 		Role:      string(user.Role),
// 		CreatedAt: timestamppb.New(user.CreatedAt),
// 		UpdatedAt: timestamppb.New(user.UpdatedAt),
// 	}, nil
// }
