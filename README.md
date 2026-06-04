# monorepo

Монорепозиторий продукта с backend-сервисами на Go и монолитным frontend на Vite + React + Ant Design + Tailwind CSS.

## Карта репозитория

```text
backend/
  go.work                 # Workspace backend-модулей.
  auth/                   # Auth subgraph: email-code login, JWT, logout.
  profile/                # Profile subgraph: профиль и nickname.
  user/                   # Внутренний user-сервис.
  gateway/                # Cosmo Router config и composed supergraph.
  test/bdd/               # API BDD runner на godog.
  test/smoke/             # Smoke tests.
features/
  01_identity/            # Общие Gherkin-сценарии для API и browser runner-ов.
frontend/
  src/                    # Монолитный React frontend.
  test/bdd/               # Browser BDD runner на Playwright.
```

## Go workspace

Backend живёт как набор самостоятельных Go-модулей, объединённых файлом `backend/go.work`.

Сейчас workspace включает:

- `auth`
- `profile`
- `user`
- `test/bdd`
- `test/smoke`

Локальные backend-команды запускаются из `backend/`. Если меняются зависимости Go-модулей, используй:

```bash
task backend:sync
```

Этот task делает `go mod download`, `go mod tidy` по модулям и затем `go work sync`. Для локальных workspace-модулей не добавляй ручные `replace` или фиктивные `require`: сначала проверь `backend/go.work`.

## Backend API

Внешний API собран через Cosmo Gateway:

- Gateway доступен как GraphQL endpoint: `http://localhost:3002/graphql`.
- `auth` и `profile` публикуют GraphQL subgraph-и.
- Cosmo Router compose config лежит в `backend/gateway/router-compose.yaml`.
- Скомпонованный router config генерируется в `backend/gateway/router.json`.

Subgraph-и:

- `auth`: `backend/auth/internal/transport/http/graphql/schema.graphqls`
- `profile`: `backend/profile/internal/transport/http/graphql/schema.graphqls`

GraphQL-контракт и generated-код живут в `internal/transport/http/graphql/`, а реализация резолверов должна оставаться тонкой и ходить в use case слой через частично применяемые интерфейсы. GraphQL layout описан в `x-gqlgen`, а правила consumer-side интерфейсов — в `x-unit-test-partial-interface`.

Внутреннее взаимодействие сервисов идёт через gRPC:

- `user` публикует `UserService` в `backend/user/pkg/grpc/user.proto`.
- `auth` и `profile` используют `USER_GRPC_ADDR`, чтобы обращаться к `user-service`.
- gRPC-контракты регенерируются командой:

```bash
task backend:proto
```

Локальный docker stack запускается из корня:

```bash
task docker-start
```

Перед стартом task пересобирает Cosmo router config через `task backend:gateway-compose`, затем поднимает Gateway, backend-сервисы и инфраструктуру.

## Замещаемые BDD-тесты

`.feature` файлы являются общей спецификацией продукта и лежат только в `features/NN_epic/`. Runner-ы не владеют спецификацией:

- API BDD: `backend/test/bdd/`, Go + godog, сценарии с тегом `@api`.
- Browser BDD: `frontend/test/bdd/`, Playwright + playwright-bdd, сценарии с тегом `@browser`.

Рабочая стратегия здесь: сначала реализуем сценарий через `@api`, чтобы быстро закрепить контракт, edge cases и server-side поведение. Когда пользовательский путь становится доступен через UI, happy-path или UI-значимый сценарий замещается сценарием `@browser`, а API-слой остаётся для контрактных, негативных и трудно наблюдаемых через браузер состояний.

Один business intent обычно исполняется одним каналом: либо `@api`, либо `@browser`. Двойной тег `@api @browser` считается исключением. Browser BDD должен ходить в реальный публичный API через Gateway, а не подменять продуктовый API моками.

Команды:

```bash
task backend:test-bdd
task frontend:test-bdd
```

Канонические правила layout, тегов и runner-ов находятся в `x-bdd-api` и `x-bdd-browser`.

## Unit dependencies

Для unit isolation зависимости объявляются у потребителя как частично применяемые интерфейсы: локальные, обычно неэкспортируемые и только с реально используемыми методами. Это позволяет сервисным пакетам зависеть от контракта, а не от конкретной реализации leaf-пакета.

Если зависимость естественно выражается функцией, используется callback alias вместо интерфейса. Типы из внешних библиотек можно использовать в сигнатурах напрямую как рабочий контракт; не нужно дублировать их в `internal/domain/` только ради тестов или моков.

Канонические правила этого паттерна находятся в `x-unit-test-partial-interface`.

## Reviewers

Проектные reviewer-роли лежат в `.agents/agents/`. Это тонкие проверяющие поверх skills: они не задают новые нормы, а сверяют реально затронутые файлы с каноническими владельцами правил.

- `dod-reviewer` — код, документация, конфигурация, logging, observability и общие DoD-проверки.
- `test-reviewer` — unit/integration/smoke test layer, mock workflow, testcontainers и test matrix.
- `bdd-reviewer` — только BDD feature/step артефакты, включая `@api`, `@browser` и проверку на fake green.

Перед завершением задачи запускай только reviewer-ов для реально затронутых областей. Если BDD-слой не менялся, `bdd-reviewer` не нужен.

## Frontend

Frontend является монолитным приложением на:

- Vite
- React
- Ant Design v6
- Tailwind CSS v4
- GraphQL codegen
- Playwright BDD

Приложение обращается к внешнему API через GraphQL Gateway, а не напрямую к внутренним gRPC-сервисам.

AntD CSS подключается через общий stylesheet frontend-а, Tailwind используется для локальной композиции в JSX, а scoped overrides внутренних `.ant-*` узлов оформляются через `@layer components`. Детали соглашений по AntD + Tailwind описаны в `x-antd`.

## Основные task-и

```bash
task docker-start        # Start local stack.
task docker-stop         # Stop local stack.
task backend:sync        # Sync backend Go modules and go.work.
task backend:proto       # Regenerate gRPC contracts.
task backend:test-bdd    # Run API BDD.
task backend:test-smoke  # Run smoke tests.
task frontend:run-dev    # Run Vite dev server.
task frontend:test-bdd   # Run browser BDD.
```

## Источники соглашений

README даёт навигационную картину. Детальные правила не дублируются здесь, чтобы у каждого повторяемого правила был один владелец:

- `x-gqlgen` — GraphQL/gqlgen layout, federation, generated vs resolvers.
- `x-unit-test-partial-interface` — частично применяемые интерфейсы, consumer-side contracts, внешние типы в сигнатурах.
- `x-bdd-api` — общий `features/`, API BDD, `@api`, godog runner.
- `x-bdd-browser` — browser BDD, `@browser`, Playwright runner.
- `x-antd` — Ant Design + Tailwind CSS во frontend.
