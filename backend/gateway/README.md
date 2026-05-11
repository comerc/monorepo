# Gateway

Пакет содержит минимальную конфигурацию Cosmo Router для локальной GraphQL-федерации.

Локальный `router.json` собирается из схем auth/profile:

```bash
task compose
```

После этого Cosmo Router можно запускать с `config.yaml`; он читает `router.json` как static execution config и прокидывает `Authorization` в subgraph-сервисы.

