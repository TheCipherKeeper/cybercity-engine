# CyberCity Engine — Модели данных

Этот документ описывает модели данных движка: от статического blueprint до
динамического runtime и событий.

## Топологический граф

Загружается из артефактов `cybercity-data`. Иммутабелен во время симуляции.

### TopologyNode

Представляет сервис в городе.

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | `str` | Уникальный kebab-case идентификатор. |
| `org_id` | `str` | Организация-владелец. |
| `name` | `str` | Человекочитаемое имя. |
| `kind` | `str` | Тип сервиса: web, api, db, dns, scada и т.д. |
| `exposure` | `Literal["public","intranet","ot","mgmt"]` | Уровень сетевой экспозиции. |
| `host` | `str` | FQDN для DNS. |
| `network_id` | `str \| None` | Логическое сетевое размещение. |
| `bind_ip` | `str \| None` | Выделенный IP-адрес. |
| `auth` | `str` | Метод аутентификации. |
| `data_classification` | `str` | Метка чувствительности данных. |
| `criticality` | `Literal["critical","high","medium","low"]` | Бизнес-критичность. |
| `ports` | `list[str]` | Открытые порты, например `tcp/443`. |
| `is_decoy` | `bool` | Является ли сервис decoy. |
| `decoy_kind` | `str \| None` | Тип fingerprint decoy. |
| `software` | `dict[str, Any]` | Вендор/продукт/версия/CVE. |
| `os_hint` | `str \| None` | Подсказка об ОС. |

### TopologyEdge

Представляет отношение между двумя сервисами.

| Поле | Тип | Описание |
|------|-----|----------|
| `source` | `str` | Исходный сервис. |
| `target` | `str` | Целевой сервис. |
| `kind` | `str` | Тип связи: api-call, auth, db-read, db-write и т.д. |
| `protocol` | `str \| None` | Например `tcp/443`. |
| `encryption` | `str \| None` | Например `tls`, `mtls`. |
| `inferred` | `bool` | True, если не из `links`, а выведено. |

### TopologyGraph

Контейнер узлов и рёбер.

| Поле | Тип | Описание |
|------|-----|----------|
| `schema_version` | `str` | Версия схемы данных. |
| `source_version` | `str` | Версия артефакта города. |
| `services` | `dict[str, TopologyNode]` | Сервисы. |
| `edges` | `list[TopologyEdge]` | Все рёбра. |

## Runtime-состояние

Изменяемое состояние, владельцем которого является `StateManager`.

### ServiceState

| Поле | Тип | Описание |
|------|-----|----------|
| `service_id` | `str` | Ссылка на topology-узел. |
| `status` | `ServiceStatus` | up, down, compromised, maintenance. |
| `health` | `float` | Индикатор здоровья 0.0–1.0. |
| `compromise_vector` | `str \| None` | Как сервис был скомпрометирован. |
| `last_heartbeat` | `datetime \| None` | Последний heartbeat real-сервиса. |
| `seen_by` | `list[str]` | Наблюдатели (сканеры). |
| `flags` | `dict[str, Any]` | Сценарий-специфичные флаги. |
| `variables` | `dict[str, Any]` | Локальные переменные процессов. |

### PlayerState

| Поле | Тип | Описание |
|------|-----|----------|
| `player_id` | `str` | Уникальный id игрока. |
| `name` | `str` | Отображаемое имя. |
| `org_id` | `str \| None` | Назначенная стартовая организация. |
| `score` | `int` | Текущий счёт. |
| `flags` | `list[str]` | Захваченные флаги. |
| `status` | `Literal["active","idle","banned"]` | Статус игрока. |

### ScenarioState

| Поле | Тип | Описание |
|------|-----|----------|
| `scenario_id` | `str` | Идентификатор сценария. |
| `name` | `str` | Отображаемое имя. |
| `status` | `Literal["running","paused","stopped"]` | Жизненный цикл. |
| `started_at` | `datetime` | Время старта. |
| `ended_at` | `datetime \| None` | Время окончания. |
| `config` | `dict[str, Any]` | Параметры сценария. |

### WorldState

Полный снапшот runtime-мира.

| Поле | Тип | Описание |
|------|-----|----------|
| `tick` | `int` | Текущий tick симуляции. |
| `started_at` | `datetime` | Время старта движка. |
| `services` | `dict[str, ServiceState]` | Состояние по сервисам. |
| `players` | `dict[str, PlayerState]` | Игроки. |
| `active_scenario` | `ScenarioState \| None` | Запущенный сценарий. |
| `variables` | `dict[str, Any]` | Глобальные переменные. |

## Событийный граф

Append-only causal graph.

### EventNode

| Поле | Тип | Описание |
|------|-----|----------|
| `event_id` | `str` | UUID события. |
| `parent_event_ids` | `list[str]` | Непосредственные causal-родители. |
| `correlation_id` | `str` | Группировка по сценарию/инциденту. |
| `tick` | `int` | Tick, в котором событие сгенерировано. |
| `timestamp` | `datetime` | Wall-clock timestamp. |
| `source_type` | `Literal[...]` | engine, service, scenario, player, system, background. |
| `source_id` | `str` | Id источника. |
| `event_type` | `EventType` | Тип события. |
| `target_id` | `str \| None` | Целевой topology-узел. |
| `payload` | `dict[str, Any]` | Событие-специфичные данные. |
| `status` | `Literal["pending","processed","failed","suppressed"]` | Статус обработки. |

### EventType

| Значение | Значение |
|----------|----------|
| `HEARTBEAT` | Liveness-пинг real-сервиса. |
| `SCAN` | Сканирование сети/сервиса. |
| `ATTACK` | Атака на сервис. |
| `COMPROMISE` | Подтверждённая компрометация сервиса. |
| `RESTORE` | Восстановление. |
| `STATE_CHANGE` | Изменение runtime-состояния. |
| `COMMAND` | Команда игрока/инструктора. |
| `SCENARIO_START` | Сценарий начался. |
| `SCENARIO_STOP` | Сценарий закончился. |
| `FLAG_CAPTURED` | Игрок достиг цели. |
| `BACKGROUND_EFFECT` | Событие от background-процесса. |
| `PROPAGATION` | Событие, распространённое по топологии. |

### EventEdge

| Поле | Тип | Описание |
|------|-----|----------|
| `source_event_id` | `str` | Родительское событие. |
| `target_event_id` | `str` | Дочернее событие. |
| `kind` | `Literal["caused_by","propagated_to","triggered_rule","response_to"]` | Отношение. |

## Конфигурация

### EngineConfig

| Поле | По умолчанию | Описание |
|------|--------------|----------|
| `app_name` | `cybercity-engine` | Имя приложения. |
| `debug` | `False` | Режим отладки. |
| `tick_ms` | `1000` | Интервал tick. |
| `engine_zip_url` | local MinIO | Источник topology-артефакта. |
| `kafka_bootstrap_servers` | `localhost:9092` | Адрес Redpanda/Kafka. |
| `database_url` | local PostgreSQL | БД снапшотов/audit. |
| `snapshot_interval_ticks` | `10` | Частота снапшотов. |
| `host` / `port` | `0.0.0.0:8000` | Привязка API. |

## Сериализация

- Модели — чистые Go struct с JSON-тегами.
- API возвращает JSON через стандартный `encoding/json`.
- PostgreSQL хранит снапшоты и события как JSONB.
- Сообщения в Redpanda по умолчанию JSON; позже может быть Avro.

## Связанные документы

- [`docs/ARCHITECTURE.md`](ARCHITECTURE.md) — как модели связаны между собой.
- [`docs/DATA_FLOW.md`](DATA_FLOW.md) — как события движутся через модели.
