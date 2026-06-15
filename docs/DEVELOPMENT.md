# CyberCity Engine — Руководство разработчика

## Быстрый старт

```bash
cd /path/to/cybercity-engine

# Запуск зависимостей
docker compose up -d postgres redpanda minio

# Установка зависимостей
uv sync

# Запуск тестов
uv run pytest -q

# Линтеры
uv run ruff check
uv run mypy --strict src/cybercity_engine

# Запуск движка локально
uv run cybercity-engine --engine-zip /path/to/engine.zip
```

## Структура проекта

```
cybercity-engine/
├── pyproject.toml              # пакет + deps + конфиг инструментов
├── compose.yaml                # локальные зависимости
├── README.md                   # обзор
├── AGENTS.md                   # правила для AI-агентов
├── src/cybercity_engine/       # исходники движка
│   ├── models.py               # topology + event + state модели
│   ├── bootstrap.py            # загрузка topology из engine.zip
│   ├── state.py                # StateManager
│   ├── events.py               # EventGraph
│   ├── router.py               # EventRouter
│   ├── engine.py               # главный tick loop + handlers
│   ├── api.py                  # FastAPI + WebSocket
│   ├── config.py               # Pydantic-настройки
│   └── __main__.py             # CLI entry point
├── tests/                      # pytest-suite
└── docs/                       # документация
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

1. Добавить значение в `EventType` в `src/cybercity_engine/models.py`.
2. Добавить handler в `src/cybercity_engine/engine.py` и зарегистрировать в
   `Engine._handlers`.
3. Добавить тесты в `tests/test_engine.py`.
4. Обновить `docs/MODELS.md`.

### Добавление propagation-правила

1. Написать чистую функцию в `src/cybercity_engine/router.py`.
2. Зарегистрировать в `EventRouter._default_rules()`.
3. Добавить тесты на условия propagation.

### Изменение API

1. Эндпоинты остаются тонкими: преобразовать вход в событие и поставить в
   очередь.
2. Никогда не мутировать `WorldState` напрямую из `api.py`.
3. Документировать изменения в `docs/API.md`.

## Тестирование

```bash
# Все тесты
uv run pytest -q

# С покрытием
uv run pytest -q --cov=src/cybercity_engine --cov-report=term-missing

# Конкретный тест
uv run pytest tests/test_engine.py -q
```

## Линтинг и типизация

```bash
uv run ruff check
uv run ruff check --fix    # автофикс где возможно
uv run mypy --strict src/cybercity_engine
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
uv run python -c "from cybercity_engine.bootstrap import load_topology; t = load_topology('engine.zip'); print(len(t.services), len(t.edges))"

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

### Ошибки mypy после изменений моделей

Запускать с `--show-error-codes` и проверять утечку `Any`, missing return
annotations и использование literal-enum.

## Связанные документы

- [`AGENTS.md`](../AGENTS.md) — правила для AI-агентов.
- [`docs/ARCHITECTURE.md`](ARCHITECTURE.md) — высокоуровневый дизайн.
- [`docs/DEPLOYMENT.md`](DEPLOYMENT.md) — как запускать в lab или production.
