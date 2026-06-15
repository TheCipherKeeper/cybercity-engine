# CyberCity Engine — Руководство разработчика

## Быстрый старт

```bash
cd /path/to/cybercity-engine

# Запуск зависимостей
docker compose up -d postgres redpanda minio

# Загрузка зависимостей и запуск тестов
go mod tidy
go test ./...

# Запуск движка локально
go run ./cmd/cybercity-engine --engine-zip /path/to/engine.zip
```

## Структура проекта

```
cybercity-engine/
├── go.mod                       # Go module + deps
├── go.sum                       # pinned deps
├── compose.yaml                 # локальные зависимости
├── README.md                    # обзор
├── AGENTS.md                    # правила для AI-агентов
├── cmd/cybercity-engine/        # CLI entry point
│   └── main.go
├── internal/                    # исходники
│   ├── application/             # composition root
│   │   └── runtime.go
│   ├── config/                  # env + flags
│   ├── domain/                  # ядро (чистая логика)
│   │   ├── models/              # Topology*, Event*, WorldState
│   │   ├── state/               # StateManager
│   │   ├── router/              # EventRouter / propagation rules
│   │   ├── engine/              # tick loop + handlers
│   │   └── ports/               # интерфейсы (EventStore, Bus, ...)
│   └── adapters/                # конкретные реализации портов
│       ├── api/                 # HTTP + WebSocket
│       ├── loader/              # topology loader
│       ├── memory/              # in-memory store / bus / broadcaster
│       └── (postgres, redpanda — будущее)
├── internal/*/*_test.go         # unit tests
└── docs/                        # документация
    ├── VISION.md
    ├── ARCHITECTURE.md
    ├── DATA_FLOW.md
    ├── MODELS.md
    ├── API.md
    ├── DEPLOYMENT.md
    └── adr/
```

## Работа над движком

### Добавление нового типа события

1. Добавить константу в `EventType` в `internal/domain/models/models.go`.
2. Добавить handler в `internal/domain/engine/handlers.go` и зарегистрировать
   в `defaultHandlers()`.
3. Добавить тесты в `internal/domain/engine/engine_test.go`.
4. Обновить `docs/MODELS.md`.

### Добавление propagation-правила

1. Написать чистую функцию в `internal/domain/router/router.go`.
2. Зарегистрировать в `defaultRules()`.
3. Добавить тесты на условия propagation.

### Изменение API

1. Эндпоинты остаются тонкими: преобразовать вход в событие и поставить в
   очередь.
2. Никогда не мутировать `WorldState` напрямую из `adapters/api`.
3. Документировать изменения в `docs/API.md`.

### Добавление нового backend (Postgres, Redpanda)

1. Добавить новый пакет в `internal/adapters/<name>/`.
2. Реализовать соответствующий порт из `internal/domain/ports/`.
3. В `internal/application/runtime.go` выбрать адаптер на основе config.
4. domain engine при этом не меняется.

## Тестирование

```bash
# Все тесты
go test ./...

# С подробным выводом
go test -v ./...

# Конкретный пакет
go test ./internal/engine -v
```

## Линтинг и проверки

```bash
go vet ./...
go build ./cmd/cybercity-engine
```

## Стиль коммитов

Использовать conventional commits:

```text
feat: add background degradation process
fix: clamp health values in StateManager
docs: update API.md with scenario endpoints
refactor: extract event store interface
```

Breaking changes должны включать `BREAKING CHANGE:` в тело.

## Процесс ADR

Если изменение затрагивает архитектурное решение:

1. Написать или обновить ADR в `docs/adr/`.
2. Сослаться на него из `docs/ARCHITECTURE.md`.
3. Старые ADR помечать `superseded`, а не удалять.

## Полезные команды

```bash
# Посмотреть структуру артефакта города
go run ./cmd/cybercity-engine --engine-zip engine.zip 2>&1 | head

# Подключиться к локальному PostgreSQL
psql postgresql://engine:engine@localhost:5432/cybercity

# Redpanda admin UI
open http://localhost:9644
```

## Troubleshooting

### Тесты падают на подключении к БД

Убедитесь, что PostgreSQL запущен:

```bash
docker compose up -d postgres
```

### Redpanda недоступна

Дождитесь прохождения health check:

```bash
docker compose logs -f redpanda
```

### Ошибки компиляции после изменений моделей

Запускать `go vet ./...` и `go build ./cmd/cybercity-engine`. Go статически
проверит типы и неиспользуемые поля.

## Связанные документы

- [`AGENTS.md`](../AGENTS.md) — правила для AI-агентов.
- [`docs/ARCHITECTURE.md`](ARCHITECTURE.md) — высокоуровневый дизайн.
- [`docs/DEPLOYMENT.md`](DEPLOYMENT.md) — как запускать в lab или production.
