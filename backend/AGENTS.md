# Backend AGENTS.md

Project-specific backend invariants. Общие правила репозитория остаются в корневом `AGENTS.md`.

## Identity-модули

- `user` — связующее звено между `auth` и `profile`: оба модуля могут зависеть от `user`, но `auth` и `profile` не импортируют и не вызывают друг друга напрямую.
- `auth` владеет email-кодами, сессиями, JWT и отзывом токенов.
- `profile` владеет профилем и nickname; пользовательский контекст получает из доверенного `X-User-ID` от gateway, а данные пользователя — через `user`.
