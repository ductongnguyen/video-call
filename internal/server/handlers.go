package server

import (
	"context"

	conversationHttp "video-call/internal/chat/delivery/http"
	conversationWs "video-call/internal/chat/delivery/ws"

	conversationRepository "video-call/internal/chat/repository"
	conversationUseCase "video-call/internal/chat/usecase"
	apiMiddlewares "video-call/internal/middleware"

	"video-call/pkg/metric"
	"video-call/pkg/websocket"

	"github.com/gin-contrib/requestid"
	redis "github.com/redis/go-redis/v9"

	authHttp "video-call/internal/auth/delivery/http"
	authRepository "video-call/internal/auth/repository"
	authUseCase "video-call/internal/auth/usecase"

	"github.com/gin-contrib/cors"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	signalingHttp "video-call/internal/signaling/delivery/http"
	signalingWs "video-call/internal/signaling/delivery/ws"
	signalingRepo "video-call/internal/signaling/repository"
	signalingUC "video-call/internal/signaling/usecase"
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

	conversationRepo := conversationRepository.NewRepository(s.db)
	conversationRedisRepo := conversationRepository.NewRedisRepo(s.redis)
	authRepo := authRepository.NewRepository(s.db)
	authRedisRepo := authRepository.NewRedisRepo(s.redis)

	conversationUC := conversationUseCase.NewUseCase(s.cfg, conversationRepo, conversationRedisRepo, s.logger)
	authUC := authUseCase.NewUseCase(s.cfg, authRepo, authRedisRepo, s.logger)

	authHandlers := authHttp.NewHandlers(s.cfg, authUC, s.logger)

	callRepo := signalingRepo.NewPostgresRepository(s.db)
	callUC := signalingUC.NewUseCase(s.cfg, callRepo, s.logger)
	wsNotificationHandler := signalingWs.NewWsNotificationHandler(callUC)
	callREST := signalingHttp.NewHandler(callUC, wsNotificationHandler, s.logger)

	redisClient := redis.NewClient(&redis.Options{
		Addr: s.cfg.Redis.Standalone.RedisAddr,
	})

	redisHub := websocket.NewRedisHub(redisClient)

	messageWriter := conversationUseCase.NewMessageWriter(conversationUC, 4, 1000) // 4 worker, 1000 queue

	wsChatHandler := conversationWs.NewWsHandler(redisHub, conversationUC, messageWriter)

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

	// Map signaling routes
	signalingGroup := v1.Group("/signaling")
	signalingHttp.MapRoutes(signalingGroup, callREST, mw)

	// Đăng ký route signaling WebSocket
	v1.GET("/call/ws/notifications", wsNotificationHandler.ServeWsNotifications)

	return nil
}
