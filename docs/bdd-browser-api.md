# Browser BDD и API-контракт

Короткое целевое соглашение: browser BDD остаётся быстрым UX-слоем с моками GraphQL, но эти моки должны проверяться через общий API-контракт, а не жить как независимая правда.

## Цель

- `@api` сценарии проверяют реальный backend через публичный API.
- `@browser` сценарии проверяют UX через Playwright и GraphQL mocks.
- Browser mocks используют те же контрактные примеры ответов, которые проверяются API BDD.

Так мы не дублируем live-browser сценарии и уменьшаем риск fake green, когда UI тесты зелёные, а frontend и backend уже разъехались.

## Что сделать

1. Завести каталог контрактных примеров, например `features/contracts/graphql/`.
2. Для ключевых операций описать success/error examples:
   - `RequestEmailCode`
   - `LoginWithEmailCode`
   - `MyProfile`
   - `SetNickname`
   - `NicknameAvailability`
   - `Logout`
3. В API BDD добавить assertions, что реальный backend возвращает ответы, совместимые с этими примерами.
4. В browser mocks заменить handwritten payloads на загрузку тех же примеров.
5. Типизировать mock responses через frontend generated GraphQL types.
6. Добавить unit/contract test для `frontend/test/bdd/support/graphql.ts`: все mock operations должны иметь контрактный пример и валидный response shape.
7. В `docs/TEST-MATRIX.md` явно отметить слой: `Browser mock contract`.

## Что не делать

- Не заменять все `@browser` mocks на live backend: это сделает UX-тесты медленнее и хрупче.
- Не писать второй набор live-browser сценариев, если он повторяет `@api`.
- Не проверять browser mock внутренним state-only Then: mock должен отдавать контрактный response, а сценарий должен наблюдать UI/network-visible effect.

## Минимальный первый шаг

Начать с `NicknameAvailability` и `SetNickname`: вынести их success/error examples в общий JSON, подключить эти examples в browser mock и добавить API BDD-проверку against examples.
