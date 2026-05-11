// Package config описывает конфигурацию auth-сервиса из переменных окружения.
//
// Использование:
//
//	var cfg config.Config
//	if err := aenv.InitConfig(&cfg); err != nil {
//	    return fmt.Errorf("failed to load config: %w", err)
//	}
//
// Конфигурация:
//
//	ENVIRONMENT     — окружение приложения (default: development)
//	AUTH_CODE_TTL   — время жизни email-кода (default: 10m)
//	AUTH_JWT_SECRET — HMAC-секрет JWT (default: development-secret, production обязан переопределить)
//	USER_GRPC_ADDR  — адрес user-сервиса (default: localhost:50051)
//	SMTP_HOST       — SMTP-хост Mailpit или провайдера (default: localhost)
//	SMTP_PORT       — SMTP-порт Mailpit или провайдера (default: 1025)
//	SMTP_USER       — SMTP-пользователь (default: "")
//	SMTP_PASSWORD   — SMTP-пароль (default: "")
//	SMTP_FROM       — адрес отправителя (default: auth@example.com)
//	SMTP_TLS        — включить TLS/STARTTLS (default: false)
//	SMTP_INSECURE   — пропустить проверку сертификата (default: true)
//	SMTP_MAX_RETRIES — число попыток отправки (default: 1)
//	POSTGRES_*      — параметры подключения к PostgreSQL из adapters/db/pg/pgx
//	REDIS_*         — параметры подключения к Redis из adapters/kv/redis
//	RABBITMQ_URL    — строка подключения к RabbitMQ (required)
//	WEBSERVER_*     — параметры HTTP-сервера из adapters/httpserver/std
//	GRPC_*          — параметры gRPC-сервера из adapters/grpc/std
//	LOG_*           — параметры логирования из platform/monitoring
//
// Ограничения:
//
//   - Config передаётся по значению.
//   - Production не должен использовать AUTH_JWT_SECRET по умолчанию.
package config
