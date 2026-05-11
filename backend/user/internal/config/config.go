package config

import (
	apgx "github.com/pure-golang/adapters/db/pg/pgx"
	agrpc "github.com/pure-golang/adapters/grpc/std"
	aredis "github.com/pure-golang/adapters/kv/redis"
	armq "github.com/pure-golang/adapters/queue/rabbitmq"
	"github.com/pure-golang/platform/monitoring"
)

// Config описывает конфигурацию сервиса пользователей из переменных окружения.
type Config struct {
	Environment string `envconfig:"ENVIRONMENT" default:"development"`
	Monitoring  monitoring.Config
	GRPCServer  agrpc.Config
	Database    apgx.Config
	Redis       aredis.Config
	RabbitMQ    armq.Config
}
