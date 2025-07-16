package main

//go:generate protoc --proto_path=../../proto --go_out=../../proto --go-grpc_out=../../proto ../../proto/v1/chat/chat.proto

import (
	"context"
	"log"

	"github.com/ductongnguyen/vivy-chat/config"
	"github.com/ductongnguyen/vivy-chat/internal/server"
	"github.com/ductongnguyen/vivy-chat/pkg/cache/redis"
	"github.com/ductongnguyen/vivy-chat/pkg/database/postgres"
	"github.com/ductongnguyen/vivy-chat/pkg/logger"
	"github.com/joho/godotenv"

	_ "github.com/ductongnguyen/vivy-chat/docs" // Swagger docs import
)

func main() {
	log.Println("Starting api server")

	// Load .env file from the root, overriding any existing variables
	if err := godotenv.Overload(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("LoadConfig: %v", err)
	}

	ctx := context.Background()
	appLogger := logger.NewApiLogger(cfg)
	appLogger.InitLogger()
	appLogger.Infof(ctx, "AppVersion: %s, LogLevel: %s, Mode: %s", cfg.Server.AppVersion, cfg.Logger.Level, cfg.Server.Mode)

	// Repository
	postgresDB, err := postgres.New(&cfg.Postgres)
	if err != nil {
		appLogger.Fatalf(ctx, "PostgreSQL init: %s", err)
	}

	redisClient, err := redis.NewRedisClient(&cfg.Redis.Standalone)
	if err != nil {
		appLogger.Fatalf(ctx, "Redis init: %s", err)
	}

	s := server.NewServer(
		cfg,
		postgresDB,
		redisClient,
		server.Logger(appLogger),
	)
	if err = s.Run(); err != nil {
		log.Fatal(err)
	}
}
