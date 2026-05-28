# Матрица тестирования — monorepo

## Контекст

Монорепозиторий содержит identity-срез из `user`, `auth`, `profile`, локального `gateway` и React frontend. Основной внешний интерфейс: federated GraphQL через `gateway`; внутренние интерфейсы: GraphQL у `auth`/`profile`, gRPC у `user`/`auth`/`profile`. API BDD поднимает PostgreSQL, Redis, RabbitMQ, Cosmo Router и SMTP capture через testcontainers; Browser BDD поднимает live backend stack, Vite preview и проверяет пользовательские сценарии через Playwright без API-моков.

**Архитектура:**
- `backend/auth` — passwordless-аутентификация по email-коду, JWT, cooldown и logout.
- `backend/user` — создание и чтение пользователя по email/ID.
- `backend/profile` — профиль пользователя и nickname через проверку auth-токена.
- `backend/gateway` — конфигурация federated GraphQL router.
- `frontend` — React UI для login, profile, nickname availability и logout.

**Структура тестов:**
- `backend/**/*_test.go` — юнит-тесты рядом с исходниками, `t.Parallel()`.
- `features/NN_epic/*.feature` — Gherkin-сценарии.
- `backend/test/bdd/` — API BDD runner через godog.
- `frontend/test/bdd/` — Browser BDD runner через playwright-bdd.
- `backend/test/smoke/*_test.go` — смоук-тесты, build tag `smoke`.

---

## Покрытие по пакетам

| Пакет | Unit-тесты | BDD/Smoke покрытие |
|-------|-----------|--------------------|
| `backend/auth/internal/service` | Да | API BDD identity |
| `backend/profile/internal/service` | Да | API BDD identity |
| `backend/auth/internal/transport/http` | Нет | API BDD identity |
| `backend/profile/internal/transport/http` | Нет | API BDD identity |
| `backend/user` | Нет | API BDD identity |
| `frontend/src` | Нет | Browser BDD identity |
| `backend/test/smoke` | Нет | Smoke-слой зарезервирован |

---

## Unit-тесты (internal/)

### Auth service (`backend/auth/internal/service`)

Файл: `backend/auth/internal/service/service_test.go` есть.

| ID | Тест | Статус |
|----|------|--------|
| AUTH-SVC-U-001 | `TestGenerateCodeReturnsFiveDigits` | Активен |
| AUTH-SVC-U-002 | `TestServiceRequestCodeSendsEmail` | Активен |
| AUTH-SVC-U-003 | `TestServiceRequestCodeReturnsDeliveryErrorWithoutCooldown` | Активен |
| AUTH-SVC-U-004 | `TestServiceLoginByCodeIssuesValidToken` | Активен |
| AUTH-SVC-U-005 | `TestServiceLogoutRevokesToken` | Активен |
| AUTH-SVC-U-006 | `TestServiceLogoutEverywhereRevokesCurrentToken` | Активен |
| AUTH-SVC-U-007 | `TestServiceLoginByCodeRejectsWrongCode` | Активен |
| AUTH-SVC-U-008 | `TestServiceRequestCodeRespectsCooldown` | Активен |
| AUTH-SVC-U-009 | `TestServiceLoginByCodeConsumesAllCodes` | Активен |
| AUTH-SVC-U-010 | `TestServiceLoginByCodeResetsEmailCodeCooldown` | Активен |
| AUTH-SVC-U-011 | `TestServiceLoginByCodeKeepsCodesWhenSessionCannotBeSaved` | Активен |
| AUTH-SVC-U-012 | `TestServiceRequestCodeRejectsDisplayNameEmail` | Активен |

---

### Profile service (`backend/profile/internal/service`)

Файл: `backend/profile/internal/service/service_test.go` есть.

| ID | Тест | Статус |
|----|------|--------|
| PROFILE-SVC-U-001 | `TestServiceSetNicknameTrimsOuterSpaces` | Активен |
| PROFILE-SVC-U-002 | `TestServiceSetNicknameRejectsTooShortNickname` | Активен |
| PROFILE-SVC-U-003 | `TestServiceIsNicknameAvailable` | Активен |

---

## Feature inventory (features/)

### Identity email login

Feature: `features/01_identity/01_email_login.feature`

| ID | Сценарий | Execution tags | API BDD | Browser BDD |
|----|----------|----------------|---------|-------------|
| IDENTITY-F-001 | `01_request_email_code` | `@api` | Активен | Нет |
| IDENTITY-F-002 | `01A_request_email_code_rejects_empty_email` | `@api` | Активен | Нет |
| IDENTITY-F-003 | `01B_request_email_code_rejects_invalid_email` | `@api` | Активен | Нет |
| IDENTITY-F-004 | `01C_request_email_code_normalizes_email` | `@api` | Активен | Нет |
| IDENTITY-F-005 | `01D_request_email_code_respects_cooldown` | `@api` | Активен | Нет |
| IDENTITY-F-006 | `01E_request_email_code_escalates_cooldown` | `@api` | Активен | Нет |
| IDENTITY-F-007 | `02_login_with_email_code` | `@api` | Активен | Нет |
| IDENTITY-F-008 | `02A_login_rejects_wrong_email_code` | `@api` | Активен | Нет |
| IDENTITY-F-009 | `02B_login_rejects_expired_email_code` | `@api` | Активен | Нет |
| IDENTITY-F-010 | `02C_login_accepts_any_active_requested_code` | `@api` | Активен | Нет |
| IDENTITY-F-011 | `02D_login_consumes_all_active_email_codes` | `@api` | Активен | Нет |
| IDENTITY-F-012 | `02E_login_rejects_reused_email_code` | `@api` | Активен | Нет |
| IDENTITY-F-013 | `02F_login_resets_email_code_cooldown` | `@api` | Активен | Нет |
| IDENTITY-F-014 | `03_logout_revokes_token` | `@api` | Активен | Нет |
| IDENTITY-F-015 | `03A_logout_current_session_keeps_other_sessions` | `@api` | Активен | Нет |
| IDENTITY-F-016 | `03B_logout_everywhere_revokes_all_sessions` | `@browser` | Нет | Активен |
| IDENTITY-F-017 | `04_login_with_email_code_in_browser` | `@browser` | Нет | Активен |
| IDENTITY-F-018 | `04A_browser_login_shows_invalid_code_error` | `@browser` | Нет | Активен |
| IDENTITY-F-019 | `04B_browser_login_shows_email_validation_error` | `@browser` | Нет | Активен |
| IDENTITY-F-020 | `04C_browser_login_hides_email_after_code_request` | `@browser` | Нет | Активен |
| IDENTITY-F-021 | `04D_browser_resend_code_uses_countdown` | `@browser` | Нет | Активен |
| IDENTITY-F-022 | `04F_browser_login_validates_email_after_submit` | `@browser` | Нет | Активен |
| IDENTITY-F-023 | `04G_browser_login_requires_email_after_submit` | `@browser` | Нет | Активен |

### Identity profile

Feature: `features/01_identity/02_profile.feature`

| ID | Сценарий | Execution tags | API BDD | Browser BDD |
|----|----------|----------------|---------|-------------|
| IDENTITY-F-025 | `01_set_unique_nickname` | `@browser` | Нет | Активен |
| IDENTITY-F-026 | `01A_set_nickname_rejects_taken_nickname` | `@browser` | Нет | Активен |
| IDENTITY-F-027 | `01B_set_nickname_trims_outer_spaces` | `@api` | Активен | Нет |
| IDENTITY-F-028 | `01C_set_nickname_rejects_too_short_nickname` | `@api` | Активен | Нет |
| IDENTITY-F-029 | `01D_set_nickname_allows_case_sensitive_variants` | `@api` | Активен | Нет |
| IDENTITY-F-030 | `01E_change_existing_nickname` | `@api` | Активен | Нет |
| IDENTITY-F-031 | `01F_profile_requires_authenticated_user` | `@api` | Активен | Нет |
| IDENTITY-F-032 | `01G_browser_profile_without_login_opens_login_form` | `@browser` | Нет | Активен |
| IDENTITY-F-033 | `01H_browser_set_nickname_shows_short_error` | `@browser` | Нет | Активен |
| IDENTITY-F-039 | `01I_browser_stale_session_redirects_to_login` | `@browser` | Нет | Активен |
| IDENTITY-F-034 | `02_show_saved_nickname_in_header` | `@browser` | Нет | Активен |
| IDENTITY-F-035 | `03_check_nickname_availability_after_debounce` | `@browser` | Нет | Активен |
| IDENTITY-F-036 | `03A_check_nickname_availability_not_started_for_two_chars` | `@browser` | Нет | Активен |
| IDENTITY-F-037 | `03B_check_nickname_availability_shows_taken` | `@browser` | Нет | Активен |
| IDENTITY-F-038 | `03C_check_nickname_availability_shows_free` | `@browser` | Нет | Активен |

---

## API BDD runner (`backend/test/bdd/`)

Требует: Docker, PostgreSQL, Redis, RabbitMQ, Cosmo Router и SMTP capture. Канонический запуск из monorepo: `task backend:test-bdd`.

Steps: `backend/test/bdd/steps_01_identity.go`

| ID | Тест | Статус |
|----|------|--------|
| API-BDD-IDENTITY-001 | `01_email_login.feature` сценарии с `@api` | 15 активных |
| API-BDD-IDENTITY-002 | `02_profile.feature` сценарии с `@api` | 5 активных |

---

## Browser BDD runner (`frontend/test/bdd/`)

Требует: Docker, live backend stack, Mailpit, установленный Chromium для Playwright и Vite preview. Канонический запуск из monorepo: `task frontend:test-bdd`; локально из `frontend/`: `npm run test:bdd`.

Browser BDD ходит в live federated GraphQL endpoint `http://127.0.0.1:3002/graphql`. API route mocks для `.feature`-сценариев запрещены guard-скриптом `frontend/test/bdd/support/assert-no-api-mocks.mjs`.

Steps: `frontend/test/bdd/steps_01_identity.ts`

| ID | Тест | Статус |
|----|------|--------|
| BROWSER-BDD-IDENTITY-001 | `01_email_login.feature` сценарии с `@browser` | 8 активных |
| BROWSER-BDD-IDENTITY-002 | `02_profile.feature` сценарии с `@browser` | 10 активных |

---

## Смоук-тесты (backend/test/smoke/)

Требуют: собранный стек. Build tag: `smoke`.

Файл: `backend/test/smoke/doc.go` резервирует слой, `*_test.go` файлов нет.

| ID | Тест | Статус |
|----|------|--------|
| SMOKE-001 | Liveness/readiness собранного стека | Не реализовано |

---

## Сводная таблица

| Пакет | Unit | API BDD | Browser BDD | Smoke |
|-------|------|---------|-------------|-------|
| `backend/auth/internal/service` | 12 | BDD identity | 0 | 0 |
| `backend/profile/internal/service` | 3 | BDD identity | 0 | 0 |
| `backend/auth/internal/transport/http` | 0 | BDD identity | 0 | 0 |
| `backend/profile/internal/transport/http` | 0 | BDD identity | 0 | 0 |
| `backend/user` | 0 | BDD identity | 0 | 0 |
| `frontend/src` | 0 | 0 | BDD identity | 0 |
| `backend/test/smoke` | 0 | 0 | 0 | 0 |
| **Итого** | **15** | **20 API BDD-сценария** | **18 Browser BDD-сценариев** | **0** |

---

## Покрытие кода

Команда `task cover` в текущем Taskfile не определена. Общее покрытие: **не снято**.

| Пакет | Покрытие | Примечание |
|-------|----------|------------|
| `backend/auth/internal/service` | Не снято | Проверено через `task backend:test-short` |
| `backend/profile/internal/service` | Не снято | Проверено через `task backend:test-short` |
| `frontend` | Не снято | Browser BDD проверяет сценарии, code coverage не собирает |
