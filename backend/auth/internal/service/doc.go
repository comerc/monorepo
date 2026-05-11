// Package service реализует passwordless-аутентификацию по email-коду.
//
// Использование:
//
//	svc := service.New(cfg, codeStore, userClient, mailSender)
//
// Конфигурация:
//
//	AUTH_CODE_TTL   — время жизни одноразового кода.
//	AUTH_JWT_SECRET — HMAC-секрет JWT.
//	SMTP_*          — настройки отправки email.
//
// Ограничения:
//
//   - Конструктор не выполняет I/O.
//   - JWT не имеет срока действия и должен отзываться через logout.
package service
