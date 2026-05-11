// Package grpc реализует gRPC-транспорт user-сервиса.
//
// Использование:
//
//	userpb.RegisterUserServiceServer(server, grpc.New(service))
//
// Конфигурация:
//
//	GRPC_* — параметры сервера задаются в internal/config.
//
// Ограничения:
//
//   - Транспорт не содержит бизнес-логики.
//   - Ошибки домена маппятся в gRPC status codes.
package grpc
