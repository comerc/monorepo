// Package http реализует HTTP-транспорт profile-сервиса.
//
// Использование:
//
//	transport := http.New(resolver)
//	transport.EnrichRoutes(mux)
//
// Конфигурация:
//
//	WEBSERVER_* — параметры HTTP-сервера задаются в internal/config.
//
// Ограничения:
//
//   - HTTP-транспорт не содержит бизнес-логики.
//   - Authorization header переносится в request context для резолверов.
package http
