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
- **Python-first.** Этот репозиторий — reference implementation на Python для
  скорости итераций. Production port на Go запланирован позже.
- **LLM — помощник, не хозяин.** LLM пишет код и YAML, валидаторы и тесты
  решают.

## Правила для AI-агентов

### Что агенту МОЖНО

- Писать Python-код в `src/cybercity_engine/` и `tests/`.
- Редактировать `pyproject.toml` зависимости с обоснованием.
- Создавать новые ADR, если меняется архитектурное решение.
- Запускать `pytest`, `ruff`, `mypy`.
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
├── pyproject.toml                    # Python package + deps
├── compose.yaml                      # local dev dependencies
├── src/cybercity_engine/
│   ├── __init__.py
│   ├── __main__.py                   # CLI entry point
│   ├── api.py                        # FastAPI + WebSocket
│   ├── bootstrap.py                  # load topology from engine.zip
│   ├── config.py                     # Pydantic settings
│   ├── engine.py                     # main simulation loop
│   ├── events.py                     # event graph store
│   ├── models.py                     # topology + event models
│   ├── router.py                     # event propagation rules
│   └── state.py                      # runtime state manager
├── tests/                            # pytest suite
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
3. Запустить `pytest`, `ruff check`, `mypy --strict src/cybercity_engine`.
4. Показать результат пользователю. Не коммитить.

## Язык документации

Вся документация и ADR ведутся на русском языке. README может содержать
английские бейджи и ссылки, но основной текст — русский.
