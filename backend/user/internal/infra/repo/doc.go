// Package repo реализует PostgreSQL-хранилище пользователей.
//
// Использование:
//
//	repository := repo.New(pool)
//	if err := repository.Start(ctx); err != nil {
//	    return err
//	}
//
// Конфигурация:
//
//	POSTGRES_* — параметры подключения задаются в internal/config.
//
// Ограничения:
//
//   - Конструктор не выполняет I/O.
//   - Start должен быть вызван до первого бизнес-метода.
//   - Потокобезопасность обеспечивается pgxpool.Pool.
package repo
