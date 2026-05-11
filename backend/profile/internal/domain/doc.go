// Package domain описывает бизнес-сущности profile-сервиса.
//
// Использование:
//
//	profile := domain.Profile{UserID: userID, Nickname: nickname}
//
// Конфигурация:
//
//	Переменные окружения отсутствуют.
//
// Ограничения:
//
//   - Пакет не выполняет I/O.
//   - Nickname является необязательным, но уникальным при заполнении.
package domain
