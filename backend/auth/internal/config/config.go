package config

import (
	"time"

	apgx "github.com/pure-golang/adapters/db/pg/pgx"
	agrpc "github.com/pure-golang/adapters/grpc/std"
	ahttp "github.com/pure-golang/adapters/httpserver/std"
	aredis "github.com/pure-golang/adapters/kv/redis"
	armq "github.com/pure-golang/adapters/queue/rabbitmq"
	"github.com/pure-golang/platform/monitoring"
)

// Config описывает конфигурацию auth-сервиса из переменных окружения.
type Config struct {
	Environment string        `envconfig:"ENVIRONMENT" default:"development"`
	CodeTTL     time.Duration `envconfig:"AUTH_CODE_TTL" default:"5m"`
	JWTSecret   string        `envconfig:"AUTH_JWT_SECRET" default:"development-secret"`
	UserGRPC    string        `envconfig:"USER_GRPC_ADDR" default:"localhost:50051"`
	Monitoring  monitoring.Config
	HTTPServer  ahttp.Config
	GRPCServer  agrpc.Config
	Database    apgx.Config
	Redis       aredis.Config
	RabbitMQ    armq.Config
	Mail        MailConfig
}

// MailConfig описывает SMTP-настройки для adapters/mail.
type MailConfig struct {
	Host       string `envconfig:"SMTP_HOST" default:"localhost"`
	Port       int    `envconfig:"SMTP_PORT" default:"1025"`
	Username   string `envconfig:"SMTP_USER"`
	Password   string `envconfig:"SMTP_PASSWORD"` //nolint:gosec // Поле описывает env-ключ пароля SMTP.
	From       string `envconfig:"SMTP_FROM" default:"auth@example.com"`
	TLS        bool   `envconfig:"SMTP_TLS" default:"false"`
	Insecure   bool   `envconfig:"SMTP_INSECURE" default:"true"`
	MaxRetries int    `envconfig:"SMTP_MAX_RETRIES" default:"1"`
}
