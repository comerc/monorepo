// Package grpc реализует gRPC-транспорт auth-сервиса.
//
// Использование:
//
//	authpb.RegisterAuthServiceServer(server, grpc.New(service))
//
// Конфигурация:
//
//	GRPC_* — параметры сервера задаются в internal/config.
//
// Ограничения:
//
//   - Транспорт не содержит бизнес-логики.
//   - ValidateToken используется внутренними сервисами.
package grpc
