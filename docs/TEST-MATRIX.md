# Матрица тестирования — monorepo

## Контекст

Монорепозиторий содержит identity-срез из `user`, `auth`, `profile`, локального `gateway` и React frontend. Основной внешний интерфейс: federated GraphQL через `gateway`; внутренние интерфейсы: GraphQL у `auth`/`profile`, gRPC у `user`/`auth`/`profile`. Инфраструктурные зависимости API BDD-слоя: PostgreSQL, Redis, RabbitMQ, Cosmo Router и тестовый SMTP capture внутри `backend/test/bdd`. Browser BDD-слой поднимает frontend через Vite preview и мокает GraphQL-ответы на уровне браузерных запросов.

**Архитектура:**
- `backend/auth` — passwordless-аутентификация по email-коду, JWT и logout.
- `backend/user` — создание и чтение пользователя по email/ID.
- `backend/profile` — профиль пользователя и nickname через проверку auth-токена.
- `backend/gateway` — конфигурация federated GraphQL router для auth/profile.
- `frontend` — React UI для email login, profile и logout.

**Структура тестов:**
- `backend/**/*_test.go` — юнит-тесты рядом с исходниками.
- `test/integration/*_test.go` — интеграционные тесты, сейчас отсутствуют.
- `features/NN_epic/*.feature` + `backend/test/bdd/*.go` — API BDD-тесты через godog.
- `features/NN_epic/*.feature` + `frontend/test/bdd/**/*.ts` — browser BDD-тесты через playwright-bdd.
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
| `frontend/src/pages` | Нет | Через Browser BDD identity |
| `features/01_identity` | Не применимо | API BDD, Browser BDD |
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

## BDD-тесты (features/ + backend/test/bdd/)

Требуют: Docker, PostgreSQL, Redis, RabbitMQ и Cosmo Router. Пакет выбирается через build tag `bdd`. Пользовательские шаги обращаются только к общему federation `/graphql` gateway.

### Identity (`features/01_identity`)

Feature: `features/01_identity/01_email_login.feature`  
Steps: `backend/test/bdd/steps_01_identity.go`

| ID | Сценарий | Статус |
|----|----------|--------|
| BDD-IDENTITY-001 | `01_request_email_code` | Активен |
| BDD-IDENTITY-002 | `02_login_with_email_code` | Активен |
| BDD-IDENTITY-003 | `03_logout_revokes_token` | Активен |

Feature: `features/01_identity/02_profile.feature`  
Steps: `backend/test/bdd/steps_01_identity.go`

| ID | Сценарий | Статус |
|----|----------|--------|
| BDD-IDENTITY-004 | `01_set_unique_nickname` | Активен, также покрыт Browser BDD |

---

## Browser BDD-тесты (features/ + frontend/test/bdd/)

Требуют: установленный Chromium для Playwright. Runner генерирует Playwright specs из сценариев с тегом `@browser`, поднимает frontend через Vite preview и проверяет пользовательские happy-path потоки в браузере.

### Identity (`features/01_identity`)

Feature: `features/01_identity/01_email_login.feature`  
Steps: `frontend/test/bdd/steps_01_identity.ts`

| ID | Сценарий | Статус |
|----|----------|--------|
| BROWSER-BDD-IDENTITY-001 | `04_login_with_email_code_in_browser` | Активен |

Feature: `features/01_identity/02_profile.feature`  
Steps: `frontend/test/bdd/steps_01_identity.ts`

| ID | Сценарий | Статус |
|----|----------|--------|
| BROWSER-BDD-IDENTITY-002 | `01_set_unique_nickname` | Активен |
| BROWSER-BDD-IDENTITY-003 | `02_show_saved_nickname_in_header` | Активен |

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

| Пакет | Unit | API BDD | Browser BDD | Smoke |
|-------|------|---------|-------------|-------|
| `backend/auth/internal/service` | 4 | BDD identity | 0 | 0 |
| `backend/auth/internal/transport/http` | 0 | BDD identity | 0 | 0 |
| `backend/user` | 0 | BDD identity | 0 | 0 |
| `backend/profile` | 0 | BDD identity | 0 | 0 |
| `frontend/src` | 0 | 0 | BDD identity | 0 |
| `test/smoke` | 0 | 0 | 0 | 0 |
| **Итого** | **4** | **4 API BDD-сценария** | **3 Browser BDD-сценария** | **0** |

---

## Покрытие кода

Общее покрытие: **не снято**. В текущем Taskfile отдельная coverage-команда не определена.

| Пакет | Покрытие | Примечание |
|-------|----------|------------|
| `backend/auth/internal/service` | Не снято | Coverage в текущей задаче не запускался |
| `backend/user` | Не снято | Coverage в текущей задаче не запускался |
| `backend/profile` | Не снято | Coverage в текущей задаче не запускался |
| `frontend` | Не снято | Browser BDD проверяет сценарии, code coverage не собирает |
