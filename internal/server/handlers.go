package server

import (
	"context"

	conversationGrpc "github.com/ductongnguyen/vivy-chat/internal/chat/delivery/grpc"
	conversationHttp "github.com/ductongnguyen/vivy-chat/internal/chat/delivery/http"
	conversationWs "github.com/ductongnguyen/vivy-chat/internal/chat/delivery/ws"

	conversationRepository "github.com/ductongnguyen/vivy-chat/internal/chat/repository"
	conversationUseCase "github.com/ductongnguyen/vivy-chat/internal/chat/usecase"
	apiMiddlewares "github.com/ductongnguyen/vivy-chat/internal/middleware"
	proto "github.com/ductongnguyen/vivy-chat/proto/v1/chat"

	"github.com/ductongnguyen/vivy-chat/pkg/metric"
	"github.com/ductongnguyen/vivy-chat/pkg/websocket"
	"github.com/gin-contrib/requestid"
	redis "github.com/redis/go-redis/v9"

	authHttp "github.com/ductongnguyen/vivy-chat/internal/auth/delivery/http"
	authRepository "github.com/ductongnguyen/vivy-chat/internal/auth/repository"
	authUseCase "github.com/ductongnguyen/vivy-chat/internal/auth/usecase"
	"github.com/gin-contrib/cors"

	// Swagger UI imports
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Map Server Handlers
func (s *Server) MapHandlers() error {
	ctx := context.Background()
	metrics, err := metric.CreateMetrics(s.cfg.Metrics.URL, s.cfg.Metrics.ServiceName)
	if err != nil {
		s.logger.Errorf(ctx, "CreateMetrics Error: %s", err)
	}
	s.logger.Info(
		ctx,
		"Metrics available URL: %s, ServiceName: %s",
		s.cfg.Metrics.URL,
		s.cfg.Metrics.ServiceName,
	)

	metrics.SetSkipPath([]string{"readiness"})

	// Init repositories
	conversationRepo := conversationRepository.NewRepository(s.db)
	conversationRedisRepo := conversationRepository.NewRedisRepo(s.redis)
	authRepo := authRepository.NewRepository(s.db)
	authRedisRepo := authRepository.NewRedisRepo(s.redis)

	// Init useCases
	conversationUC := conversationUseCase.NewUseCase(s.cfg, conversationRepo, conversationRedisRepo, s.logger)
	authUC := authUseCase.NewUseCase(s.cfg, authRepo, authRedisRepo, s.logger)

	// Init handlers
	conversationGrpcHandlers := conversationGrpc.NewConversationServiceHandler(s.cfg, conversationUC, s.logger)
	authHandlers := authHttp.NewHandlers(s.cfg, authUC, s.logger)

	// Register gRPC services
	proto.RegisterConversationServiceServer(s.grpc, conversationGrpcHandlers)

	// Khởi tạo Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: s.cfg.Redis.Standalone.RedisAddr,
		// Password: "", // Nếu cần, hãy thêm trường Password vào config
		// DB: 0,        // Nếu cần, hãy thêm trường DB vào config
	})

	// Khởi tạo RedisHub
	redisHub := websocket.NewRedisHub(redisClient)

	// Khởi tạo worker pool ghi DB
	messageWriter := conversationUseCase.NewMessageWriter(conversationUC, 4, 1000) // 4 worker, 1000 queue

	// Khởi tạo handler và truyền callback onMessage
	wsChatHandler := conversationWs.NewWsHandler(redisHub, conversationUC, messageWriter)

	// Initialize chat HTTP handlers
	chatHandlers := conversationHttp.NewHandler(conversationUC, s.logger)

	mw := apiMiddlewares.NewMiddlewareManager(s.cfg, []string{"*"}, s.logger)

	s.gin.Use(requestid.New())
	s.gin.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, 
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	s.gin.Use(mw.MetricsMiddleware(metrics))
	s.gin.Use(mw.LoggerMiddleware(s.logger))

	// Swagger docs endpoint
	s.gin.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := s.gin.Group("/api/v1")
	v1.GET("/ws", wsChatHandler.ServeWs)

	// Create auth group and map auth routes
	authGroup := v1.Group("/auth")
	authHttp.MapRoutes(authGroup, authHandlers, mw)

	// Map chat routes
	chatGroup := v1.Group("/chat")
	conversationHttp.MapRoutes(chatGroup, chatHandlers, mw)

	return nil
}
