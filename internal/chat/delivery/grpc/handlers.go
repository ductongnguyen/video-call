package grpc

import (
	"github.com/ductongnguyen/vivy-chat/config"
	"github.com/ductongnguyen/vivy-chat/internal/chat"
	"github.com/ductongnguyen/vivy-chat/pkg/logger"
	proto "github.com/ductongnguyen/vivy-chat/proto/v1/chat"
)

type ConversationServiceHandler struct {
	proto.UnimplementedConversationServiceServer
	cfg     *config.Config
	usecase chat.UseCase
	logger  logger.Logger
}

func NewConversationServiceHandler(cfg *config.Config, usecase chat.UseCase, logger logger.Logger) *ConversationServiceHandler {
	return &ConversationServiceHandler{usecase: usecase, cfg: cfg, logger: logger}
}
