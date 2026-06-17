# CyberCity — Engine

[![Part of CyberCity](https://img.shields.io/badge/CyberCity-composition-blueviolet)](https://github.com/TheCipherKeeper/cybercity)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/code-MIT-green)](LICENSE)

Событийный runtime-движок для цифрового двойника CyberCity.

Реализован на **Go**. Движок загружает статический топологический граф из
`cybercity-data`, ведёт runtime-состояние города и обрабатывает поток событий
через graph-aware router. См. ADR-0006 для обоснования выбора языка.

> Канон композиции (6 репозиториев, контракты, доверительная граница) —
> [`cybercity/COMPOSITION.md`](https://github.com/TheCipherKeeper/cybercity/blob/main/COMPOSITION.md).

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
docker compose up -d postgres redpanda minio

# 2. Сборка или копирование артефакта города из cybercity-data, затем запуск движка
go run ./cmd/cybercity-engine --engine-zip /path/to/engine.zip
```

Подробнее — в [`docs/DEVELOPMENT.md`](docs/DEVELOPMENT.md).

## Статус

Базовая архитектура, модели, bootstrap, engine-loop, API-скелет и документация
готовы. Дальше: persistence в PostgreSQL, интеграция с Redpanda,
background-процессы и первый сценарий.

## Лицензия

- Код: [MIT](LICENSE)
- Документация: [CC BY 4.0](LICENSE-DOCS)
