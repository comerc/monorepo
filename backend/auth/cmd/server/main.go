package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	apgx "github.com/pure-golang/adapters/db/pg/pgx"
	aenv "github.com/pure-golang/adapters/env"
	agrpc "github.com/pure-golang/adapters/grpc/std"
	amiddleware "github.com/pure-golang/adapters/httpserver/middleware"
	ahttp "github.com/pure-golang/adapters/httpserver/std"
	aredis "github.com/pure-golang/adapters/kv/redis"
	asmtp "github.com/pure-golang/adapters/mail/smtp"
	armq "github.com/pure-golang/adapters/queue/rabbitmq"
	"github.com/pure-golang/platform/monitoring"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/pure-golang/monorepo/backend/auth/internal/config"
	"github.com/pure-golang/monorepo/backend/auth/internal/infra/kv"
	"github.com/pure-golang/monorepo/backend/auth/internal/infra/user"
	"github.com/pure-golang/monorepo/backend/auth/internal/service"
	tgrpc "github.com/pure-golang/monorepo/backend/auth/internal/transport/grpc"
	thttp "github.com/pure-golang/monorepo/backend/auth/internal/transport/http"
	"github.com/pure-golang/monorepo/backend/auth/internal/transport/http/resolvers"
	authpb "github.com/pure-golang/monorepo/backend/auth/pkg/grpc"
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

	userClient, err := user.New(cfg.UserGRPC)
	if err != nil {
		return fmt.Errorf("failed to create user grpc client: %w", err)
	}
	defer func() {
		if err := userClient.Close(); err != nil {
			logger.Error("Failed to close user grpc client", slog.Any("err", err))
		}
	}()

	mailSender := asmtp.NewSender(asmtp.Config{
		Host:       cfg.Mail.Host,
		Port:       cfg.Mail.Port,
		Username:   cfg.Mail.Username,
		Password:   cfg.Mail.Password,
		From:       cfg.Mail.From,
		TLS:        cfg.Mail.TLS,
		Insecure:   cfg.Mail.Insecure,
		MaxRetries: cfg.Mail.MaxRetries,
	})
	defer func() {
		if err := mailSender.Close(); err != nil {
			logger.Error("Failed to close mail sender", slog.Any("err", err))
		}
	}()

	authService := service.New(service.Config{
		CodeTTL:   cfg.CodeTTL,
		JWTSecret: cfg.JWTSecret,
		MailFrom:  cfg.Mail.From,
	}, kv.New(rdb), userClient, mailSender)

	mux := http.NewServeMux()
	thttp.New(resolvers.New(authService)).EnrichRoutes(mux)
	handlerWithMiddleware := amiddleware.Chain(
		mux,
		amiddleware.Monitoring(),
		amiddleware.Recovery,
	)

	httpServer := ahttp.NewDefault(cfg.HTTPServer, handlerWithMiddleware)
	go func() {
		httpServer.Run()
		cancel()
	}()

	grpcServer := agrpc.NewDefault(cfg.GRPCServer, func(s *grpc.Server) {
		authpb.RegisterAuthServiceServer(s, tgrpc.New(authService))
	})
	go func() {
		grpcServer.Run()
		cancel()
	}()

	logger.Info("All dependencies initialized")
	<-ctx.Done()
	logger.Info("Shutting down server")

	g := new(errgroup.Group)
	g.Go(httpServer.Shutdown)
	g.Go(grpcServer.Shutdown)
	return g.Wait()
}
