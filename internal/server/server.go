package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ductongnguyen/vivy-chat/config"
	"github.com/ductongnguyen/vivy-chat/pkg/cache/redis"
	"github.com/ductongnguyen/vivy-chat/pkg/logger"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

// Server struct
type Server struct {
	gin    *gin.Engine
	grpc   *grpc.Server
	cfg    *config.Config
	db     *gorm.DB
	redis  redis.Client
	logger logger.Logger
	// hub    *websocket.Hub // Đã chuyển sang RedisHub, không cần trường này nữa
}

// NewServer New Server constructor
func NewServer(cfg *config.Config, db *gorm.DB, redis redis.Client, opts ...Option) *Server {
	s := &Server{
		gin:   gin.New(),
		grpc:  grpc.NewServer(),
		cfg:   cfg,
		db:    db,
		redis: redis,
	}

	// Custom options
	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) Run() error {
	// Đã refactor sang RedisHub, không còn chạy hub cũ nữa

	if err := s.MapHandlers(); err != nil {
		return err
	}

	ctx := context.Background()
	go func() {
		s.logger.Infof(ctx, "Server is listening on PORT: %s", s.cfg.Server.Port)
		ln, _ := net.Listen("tcp", ":"+s.cfg.Server.Port)
		err := s.gin.RunListener(ln)
		if err != nil {
			s.logger.Fatalf(ctx, "Error starting Server: ", err)
		}
	}()

	go func() {
		l, err := net.Listen("tcp", ":"+s.cfg.Server.GrpcPort)
		if err != nil {
			s.logger.Fatalf(ctx, "Failed to listen for gRPC: %v", err)
		}
		s.logger.Infof(ctx, "gRPC server is listening on PORT: %s", s.cfg.Server.GrpcPort)
		if err := s.grpc.Serve(l); err != nil {
			s.logger.Fatalf(ctx, "Failed to serve gRPC: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	s.logger.Info(ctx, "Server Exited Properly")
	return nil
}
