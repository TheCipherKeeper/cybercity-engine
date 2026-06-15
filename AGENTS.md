# AGENTS.md — правила для AI-агентов и контрибьюторов CyberCity Engine

## Иерархия документов (от старшего к младшему)

1. `docs/adr/` — действующие архитектурные решения. ADR со статусом
   `superseded` не имеют силы.
2. `AGENTS.md` (этот файл) — операционные правила работы в репозитории.
3. `README.md` — краткое описание и quick start.
4. Всё остальное (код, тесты, compose) — реализация принятых решений.

Если документы противоречат друг другу, побеждает старший. Любое расхождение —
повод создать новый ADR.

## Ключевые принципы

- **События — единственный источник истины.** Состояние = проекция потока
  событий.
- **Два графа.** Топологический граф загружается из `cybercity-data`.
  Событийный граф строится в runtime.
- **Engine — единственный мутатор состояния.** API и WebSocket только
  валидируют вход и порождают события.
- **Go-first.** Движок реализован на Go (см. ADR-0006).
- **Onion / ports-and-adapters.** Domain (`internal/domain/`) не импортирует
  adapters, config или application. Все внешние зависимости проходят через
  ports (`internal/domain/ports/`).
- **LLM — помощник, не хозяин.** LLM пишет код и YAML, валидаторы и тесты
  решают.

## Правила для AI-агентов

### Что агенту МОЖНО

- Писать Go-код в `cmd/` и `internal/`.
- Редактировать `go.mod` зависимости с обоснованием.
- Создавать новые ADR, если меняется архитектурное решение.
- Запускать `go test ./...`, `go vet ./...`, `go build ./cmd/cybercity-engine`.
- Обновлять `README.md`, `AGENTS.md` при изменении структуры.

### Чего агенту НЕЛЬЗЯ

- Редактировать ADR без явного указания или создания нового ADR.
- Добавлять зависимости без обоснования в ADR или комментарии.
- Делать коммиты, пуши, PR — это делает человек.
- Писать «защитный» код в обход валидаторов или типов.

## Структура репозитория

```
cybercity-engine/
├── README.md                         # overview
├── AGENTS.md                         # этот файл
├── go.mod                            # Go module + deps
├── go.sum                            # pinned deps
├── compose.yaml                      # local dev dependencies
├── cmd/cybercity-engine/             # CLI entry point
│   └── main.go
├── internal/                         # implementation
│   ├── application/                  # composition root
│   │   └── runtime.go
│   ├── config/                       # env + flags
│   ├── domain/                       # чистая логика (ядро)
│   │   ├── models/                   # topology + event + state models
│   │   ├── state/                    # StateManager
│   │   ├── router/                   # propagation rules
│   │   ├── engine/                   # tick loop + handlers
│   │   └── ports/                    # interfaces (EventStore, Bus, ...)
│   └── adapters/                     # concrete implementations
│       ├── api/                      # HTTP + WebSocket server
│       ├── loader/                   # topology loader
│       ├── memory/                   # in-memory store / bus / broadcaster
│       └── (postgres, redpanda — future)
├── ..._test.go                       # unit tests inside packages
└── docs/                             # документация
    ├── VISION.md
    ├── ARCHITECTURE.md
    ├── DATA_FLOW.md
    ├── MODELS.md
    ├── API.md
    ├── DEPLOYMENT.md
    ├── DEVELOPMENT.md
    └── adr/
```

## Рабочий цикл

1. Прочитать соответствующий ADR и текущий код.
2. Внести изменения.
3. Запустить `go test ./...`, `go vet ./...`, `go build ./cmd/cybercity-engine`.
4. Показать результат пользователю. Не коммитить.

## Язык документации

Вся документация и ADR ведутся на русском языке. README может содержать
английские бейджи и ссылки, но основной текст — русский.
