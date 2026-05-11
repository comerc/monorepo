# BDD и E2E feature-файлы

Короткое целевое соглашение для BDD/E2E-сценариев.

## Layout

| Тип проекта | Feature-файлы | API BDD | Browser BDD |
|---|---|---|---|
| Monorepo | `features/NN_epic/NN_story.feature` | `backend/test/bdd/` | `frontend/test/bdd/` |
| Backend-only | `features/NN_epic/NN_story.feature` | `test/bdd/` | Не применимо |
| Frontend-only | `features/NN_epic/NN_story.feature` | Не применимо | `test/bdd/` |

`features/**/*.feature` — общий каталог сценариев. Реализация сценариев живёт отдельно у каждого раннера.

Внешняя ссылка на сценарий строится как:

```text
features/NN_epic/NN_user_story.feature#NN[A-Z]_case
```

`NN[A-Z]_case` означает имя `Scenario`: две цифры, опциональная заглавная буква для edge/corner/negative case базового use case и slug в `snake_case`. Регекс: `^\d{2}[A-Z]?_[a-z0-9_]+$`.

Буквенный суффикс используется только для вариантов базового сценария с тем же номером: `01_success`, `01A_validation_error`, `01B_expired_code`.

Пример: `features/01_identity/01_email_login.feature#02A_expired_code`.

## Execution tags

Execution tags нужны только в monorepo или другом multi-channel проекте, где один каталог `features/` читают несколько runner-ов.

В monorepo каждый `Scenario` должен иметь хотя бы один тег канала исполнения:

- `@api` — сценарий исполняется backend BDD runner-ом через публичный API. Текущий API-канал: GraphQL supergraph `/graphql`.
- `@browser` — сценарий исполняется frontend BDD runner-ом через браузер. Текущий runner: `playwright-bdd` поверх Playwright Test.
- `@mobile` — зарезервировано для будущего mobile E2E runner-а.

Сценарий может иметь несколько тегов, если один business intent должен быть проверен через несколько каналов:

```gherkin
@api @browser
Scenario: 01_successful_email_login
```

В backend-only проекте теги каналов не нужны: все `features/**/*.feature` исполняет корневой `test/bdd/` API runner.

В frontend-only проекте теги каналов не нужны: все `features/**/*.feature` исполняет корневой `test/bdd/` browser runner.

## Границы

- `@api` не означает «backend-only»; это проверка поведения через API.
- `@browser` не означает «UI-only»; это проверка поведения через браузер.
- В monorepo UI-технические сценарии получают только `@browser`.
- В monorepo API-специфичные сценарии получают только `@api`.

## Антипаттерны

- `.feature` вне `features/NN_epic/`.
- Runner-owned feature files: `.feature` внутри `backend/test/bdd/`, `frontend/test/bdd/` или корневого `test/bdd/`.
- Дополнительный `features/` внутри runner-а: `frontend/test/bdd/features/`.
