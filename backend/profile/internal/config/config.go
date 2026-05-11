package config

import (
	apgx "github.com/pure-golang/adapters/db/pg/pgx"
	agrpc "github.com/pure-golang/adapters/grpc/std"
	ahttp "github.com/pure-golang/adapters/httpserver/std"
	aredis "github.com/pure-golang/adapters/kv/redis"
	armq "github.com/pure-golang/adapters/queue/rabbitmq"
	"github.com/pure-golang/platform/monitoring"
)

// Config описывает конфигурацию profile-сервиса из переменных окружения.
type Config struct {
	Environment string `envconfig:"ENVIRONMENT" default:"development"`
	AuthGRPC    string `envconfig:"AUTH_GRPC_ADDR" default:"localhost:50052"`
	Monitoring  monitoring.Config
	HTTPServer  ahttp.Config
	GRPCServer  agrpc.Config
	Database    apgx.Config
	Redis       aredis.Config
	RabbitMQ    armq.Config
}
