# Матрица тестирования — monorepo

## Контекст

Монорепозиторий содержит identity-срез из `user`, `auth`, `profile` и локального `gateway`. Основной внешний интерфейс: federated GraphQL через `gateway`; внутренние интерфейсы: GraphQL у `auth`/`profile`, gRPC у `user`/`auth`/`profile`. Инфраструктурные зависимости BDD-слоя: PostgreSQL, Redis, RabbitMQ, Cosmo Router и тестовый SMTP capture внутри `test/bdd`.

**Архитектура:**
- `backend/auth` — passwordless-аутентификация по email-коду, JWT и logout.
- `backend/user` — создание и чтение пользователя по email/ID.
- `backend/profile` — профиль пользователя и nickname через проверку auth-токена.
- `backend/gateway` — конфигурация federated GraphQL router для auth/profile.

**Структура тестов:**
- `backend/**/*_test.go` — юнит-тесты рядом с исходниками.
- `test/integration/*_test.go` — интеграционные тесты, сейчас отсутствуют.
- `test/bdd/NN_epic/*.feature` + `test/bdd/*.go` — BDD-тесты через godog.
- `test/e2e/*_test.go` — e2e-тесты, сейчас отсутствуют.
- `test/smoke/*_test.go` — smoke-слой зарезервирован, тестов сейчас нет.

---

## Покрытие по пакетам

| Пакет | Unit-тесты | Интеграция |
|-------|-----------|------------|
| `backend/auth/internal/service` | Да | Через BDD identity |
| `backend/auth/internal/transport/http` | Нет | Через BDD identity |
| `backend/user/internal/service` | Нет | Через BDD identity |
| `backend/profile/internal/service` | Нет | Через BDD identity |
| `test/bdd/01_identity` | Не применимо | BDD |
| `test/smoke` | Нет | Smoke-слой зарезервирован |

---

## Unit-тесты (internal/)

### Auth service (`backend/auth/internal/service`)

Файл: `backend/auth/internal/service/service_test.go` есть.

| ID | Тест | Статус |
|----|------|--------|
| AUTH-SVC-U-001 | `TestServiceRequestCodeSendsEmail` | Активен |
| AUTH-SVC-U-002 | `TestServiceLoginByCodeIssuesValidToken` | Активен |
| AUTH-SVC-U-003 | `TestServiceLogoutRevokesToken` | Активен |
| AUTH-SVC-U-004 | `TestServiceLoginByCodeRejectsWrongCode` | Активен |

---

## Интеграционные тесты (test/integration/)

Требуют: реальные технические зависимости конкретного адаптера. Skip: `testing.Short()`.

Каталог `test/integration/` сейчас отсутствует.

| ID | Тест | Статус |
|----|------|--------|
| INT-001 | Интеграционные тесты адаптеров и репозиториев | Не реализовано |

---

## BDD-тесты (test/bdd/)

Требуют: Docker, PostgreSQL, Redis, RabbitMQ и Cosmo Router. Пакет выбирается через build tag `bdd`. Пользовательские шаги обращаются только к общему federation `/graphql` gateway.

### Identity (`test/bdd/01_identity`)

Feature: `test/bdd/01_identity/01_email_login.feature`  
Steps: `test/bdd/steps_01_identity.go`

| ID | Сценарий | Статус |
|----|----------|--------|
| BDD-IDENTITY-001 | `01_request_email_code` | Активен |
| BDD-IDENTITY-002 | `02_login_with_email_code` | Активен |
| BDD-IDENTITY-003 | `03_logout_revokes_token` | Активен |

Feature: `test/bdd/01_identity/02_profile.feature`  
Steps: `test/bdd/steps_01_identity.go`

| ID | Сценарий | Статус |
|----|----------|--------|
| BDD-IDENTITY-004 | `01_set_unique_nickname` | Активен |

---

## E2E-тесты (test/e2e/)

Требуют: staging-окружение. Build tag: `e2e`.

Каталог `test/e2e/` сейчас отсутствует.

| ID | Тест | Статус |
|----|------|--------|
| E2E-001 | Сквозные проверки staging-сервисов | Не реализовано |

---

## Смоук-тесты (test/smoke/)

Требуют: собранный стек. Build tag: `smoke`.

Файл: `test/smoke/doc.go` резервирует слой, `*_test.go` файлов нет.

| ID | Тест | Статус |
|----|------|--------|
| SMOKE-001 | Liveness/readiness собранного стека | Не реализовано |

---

## Сводная таблица

| Пакет | Unit | Integration/E2E | Smoke |
|-------|------|-----------------|-------|
| `backend/auth/internal/service` | 4 | BDD identity | 0 |
| `backend/auth/internal/transport/http` | 0 | BDD identity | 0 |
| `backend/user` | 0 | BDD identity | 0 |
| `backend/profile` | 0 | BDD identity | 0 |
| `test/smoke` | 0 | 0 | 0 |
| **Итого** | **4** | **4 BDD-сценария** | **0** |

---

## Покрытие кода

Общее покрытие: **не снято**. В текущем Taskfile отдельная coverage-команда не определена.

| Пакет | Покрытие | Примечание |
|-------|----------|------------|
| `backend/auth/internal/service` | Не снято | Coverage в текущей задаче не запускался |
| `backend/user` | Не снято | Coverage в текущей задаче не запускался |
| `backend/profile` | Не снято | Coverage в текущей задаче не запускался |
