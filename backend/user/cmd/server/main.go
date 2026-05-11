package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	apgx "github.com/pure-golang/adapters/db/pg/pgx"
	aenv "github.com/pure-golang/adapters/env"
	agrpc "github.com/pure-golang/adapters/grpc/std"
	aredis "github.com/pure-golang/adapters/kv/redis"
	armq "github.com/pure-golang/adapters/queue/rabbitmq"
	"github.com/pure-golang/platform/monitoring"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/pure-golang/monorepo/backend/user/internal/config"
	"github.com/pure-golang/monorepo/backend/user/internal/infra/repo"
	"github.com/pure-golang/monorepo/backend/user/internal/service"
	tgrpc "github.com/pure-golang/monorepo/backend/user/internal/transport/grpc"
	userpb "github.com/pure-golang/monorepo/backend/user/pkg/grpc"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var cfg config.Config
	if err := aenv.InitConfig(&cfg); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	closeMonitoring := monitoring.InitDefault(cfg.Monitoring)
	logger := slog.Default().With("module", "main")
	defer func() {
		if err := closeMonitoring(); err != nil {
			logger.Error("Failed to close monitoring", slog.Any("err", err))
		}
	}()

	logger.Info("Connecting dependencies")

	db, err := apgx.NewDefault(cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database", slog.Any("err", err))
		}
	}()

	rdb, err := aredis.NewDefault(cfg.Redis)
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	defer func() {
		if err := rdb.Close(); err != nil {
			logger.Error("Failed to close redis", slog.Any("err", err))
		}
	}()

	rmqDialer := armq.NewDefaultDialer(cfg.RabbitMQ.URL)
	defer func() {
		if err := rmqDialer.Close(); err != nil {
			logger.Error("Failed to close RabbitMQ dialer", slog.Any("err", err))
		}
	}()
	if err := armq.NewConnector(rmqDialer).ConnectWithRetry(ctx); err != nil {
		return fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	userRepo := repo.New(db.Pool)
	if err := userRepo.Start(ctx); err != nil {
		return fmt.Errorf("failed to start user repo: %w", err)
	}
	userService := service.New(userRepo)

	grpcServer := agrpc.NewDefault(cfg.GRPCServer, func(s *grpc.Server) {
		userpb.RegisterUserServiceServer(s, tgrpc.New(userService))
	})
	go func() {
		grpcServer.Run()
		cancel()
	}()

	logger.Info("All dependencies initialized")
	<-ctx.Done()
	logger.Info("Shutting down server")

	g := new(errgroup.Group)
	g.Go(grpcServer.Shutdown)
	return g.Wait()
}
