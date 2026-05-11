// Package config описывает конфигурацию user-сервиса из переменных окружения.
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
//	ENVIRONMENT  — окружение приложения (default: development)
//	POSTGRES_*   — параметры подключения к PostgreSQL из adapters/db/pg/pgx
//	REDIS_*      — параметры подключения к Redis из adapters/kv/redis
//	RABBITMQ_URL — строка подключения к RabbitMQ (required)
//	GRPC_*       — параметры gRPC-сервера из adapters/grpc/std
//	LOG_*        — параметры логирования из platform/monitoring
//
// Ограничения:
//
//   - Config передаётся по значению.
//   - Подключения создаются в cmd/server, а не в конструкторе конфигурации.
package config
