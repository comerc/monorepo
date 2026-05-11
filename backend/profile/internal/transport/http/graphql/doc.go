// Package graphql содержит GraphQL-контракт profile-сервиса.
//
// Использование:
//
//	schema := graphql.NewExecutableSchema(graphql.Config{Resolvers: resolver})
//
// Конфигурация:
//
//	Переменные окружения отсутствуют.
//
// Ограничения:
//
//   - Вручную редактируется только schema.graphqls.
//   - Generated-файлы обновляются через gqlgen.
package graphql
