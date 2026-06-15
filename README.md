# CyberCity — Engine

[![Part of CyberCity](https://img.shields.io/badge/CyberCity-composition-blueviolet)](https://github.com/TheCipherKeeper/cybercity)
[![Python](https://img.shields.io/badge/Python-3.12-3776AB?logo=python)](https://python.org)
[![License: MIT](https://img.shields.io/badge/code-MIT-green)](LICENSE)

Событийный runtime-движок для цифрового двойника CyberCity.

Это **Python reference-реализация** движка. Она намеренно сделана на Python для
быстрой итерации и валидации концепции, с планируемым в будущем переносом на
Go для production-grade производительности.

## Архитектура

Движок построен вокруг **двух графов**:

1. **Топологический граф** — статический blueprint города, загружаемый из
   `cybercity-data`:
   - узлы: сервисы (`bank-web`, `hospital-db`, ...)
   - рёбра: декларированные связи (`api-call`, `auth`, `db-read`, `backup-of`, ...)
   - также: inferred-рёбра (same network, same org, exposure chain)

2. **Событийный граф** — динамический causal-граф всего, что происходит:
   - узлы: события (scan, compromise, state change, player action)
   - рёбра: `caused_by`, `propagated_to`, `triggered_rule`
   - даёт attack provenance, replay, explainability

События идут через **Redpanda/Kafka** и обрабатываются tick-циклом движка.
Runtime-состояние сохраняется в **PostgreSQL**.

## Документация

| Документ | Назначение |
|----------|-----------|
| [`docs/VISION.md`](docs/VISION.md) | Зачем существует проект и к чему стремится. |
| [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | Высокоуровневая архитектура и системный контекст. |
| [`docs/DATA_FLOW.md`](docs/DATA_FLOW.md) | Как события движутся через систему. |
| [`docs/MODELS.md`](docs/MODELS.md) | Справочник по моделям данных. |
| [`docs/API.md`](docs/API.md) | Протокол HTTP и WebSocket. |
| [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md) | Локальная разработка, home lab, production-набросок. |
| [`docs/DEVELOPMENT.md`](docs/DEVELOPMENT.md) | Как работать над движком. |
| [`docs/adr/`](docs/adr/) | Architecture decision records. |

## Быстрый старт (local Docker Compose)

```bash
# 1. Запуск зависимостей
uv run docker compose up -d postgres redpanda minio

# 2. Сборка или копирование артефакта города из cybercity-data, затем запуск движка
uv run cybercity-engine --engine-zip /path/to/engine.zip
```

Подробнее — в [`docs/DEVELOPMENT.md`](docs/DEVELOPMENT.md).

## Статус

Базовая архитектура, модели, bootstrap, engine-loop, API-скелет и документация
готовы. Дальше: persistence в PostgreSQL, интеграция с Redpanda,
background-процессы и первый сценарий.

## Лицензия

- Код: [MIT](LICENSE)
- Документация: [CC BY 4.0](LICENSE-DOCS)
