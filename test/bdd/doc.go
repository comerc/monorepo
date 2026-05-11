//go:build bdd

// Package bdd запускает BDD-сценарии монорепозитория как black-box проверки.
//
// # Использование
//
// Пакет содержит общий runner для BDD-эпиков:
//
//	go test -tags bdd ./test/bdd/
//
// TestMain поднимает общий Stack один раз на весь BDD-прогон. Каждый эпик
// запускается отдельным Test* в этом пакете и использует feature-файлы из
// test/bdd/NN_*.
//
// # Конфигурация
//
// Для локальной диагностики доступны BDD_PATHS с перечнем feature-файлов или
// директорий через запятую и BDD_GODOG_FORMAT для выбора формата вывода godog.
//
// # Ограничения
//
// Stack поднимает инфраструктуру через testcontainers: PostgreSQL, Redis и
// RabbitMQ, а также локальный SMTP-capture и реальные backend-сервисы. Пакет
// собирается только с build tag bdd и не использует testing.Short.
package bdd
