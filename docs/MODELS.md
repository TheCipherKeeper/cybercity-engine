# CyberCity Engine — Модели данных

Этот документ описывает модели данных движка: от статического blueprint до
динамического runtime и событий. Типы указаны в Go-названиях (см. ADR-0006);
перечисления (`Exposure`, `Criticality`, `ServiceStatus`, `PlayerStatus`,
`ScenarioStatus`, `SourceType`, `EventType`, `EventStatus`, `EdgeKind`) —
в `internal/domain/models/models.go`.

## Топологический граф

Загружается из артефактов `cybercity-data`. Иммутабелен во время симуляции.

### TopologyNode

Представляет сервис в городе.

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | `string` | Уникальный kebab-case идентификатор. |
| `org_id` | `string` | Организация-владелец. |
| `name` | `string` | Человекочитаемое имя. |
| `kind` | `string` | Тип сервиса: web, api, db, dns, scada и т.д. |
| `exposure` | `Exposure` | Уровень сетевой экспозиции: public, intranet, ot, mgmt. |
| `host` | `string` | FQDN для DNS. |
| `network_id` | `*string` | Логическое сетевое размещение (nullable). |
| `bind_ip` | `*string` | Выделенный IP-адрес (nullable). |
| `auth` | `string` | Метод аутентификации. |
| `data_classification` | `string` | Метка чувствительности данных. |
| `criticality` | `Criticality` | Бизнес-критичность: critical, high, medium, low. |
| `ports` | `[]string` | Открытые порты, например `tcp/443`. |
| `is_decoy` | `bool` | Является ли сервис decoy. |
| `decoy_kind` | `*string` | Тип fingerprint decoy (nullable). |
| `software` | `map[string]any` | Вендор/продукт/версия/CVE. |
| `os_hint` | `*string` | Подсказка об ОС (nullable). |

### TopologyEdge

Представляет отношение между двумя сервисами.

| Поле | Тип | Описание |
|------|-----|----------|
| `source` | `string` | Исходный сервис. |
| `target` | `string` | Целевой сервис. |
| `kind` | `string` | Тип связи: api-call, auth, db-read, db-write и т.д. |
| `protocol` | `*string` | Например `tcp/443` (nullable). |
| `encryption` | `*string` | Например `tls`, `mtls` (nullable). |
| `inferred` | `bool` | `true`, если не из `links`, а выведено. |

### TopologyGraph

Контейнер узлов и рёбер.

| Поле | Тип | Описание |
|------|-----|----------|
| `schema_version` | `string` | Версия схемы данных. |
| `source_version` | `string` | Версия артефакта города. |
| `services` | `map[string]TopologyNode` | Сервисы. |
| `edges` | `[]TopologyEdge` | Все рёбра. |

## Runtime-состояние

Изменяемое состояние, владельцем которого является `StateManager`.

### ServiceState

| Поле | Тип | Описание |
|------|-----|----------|
| `service_id` | `string` | Ссылка на topology-узел. |
| `status` | `ServiceStatus` | up, down, compromised, maintenance. |
| `health` | `float64` | Индикатор здоровья 0.0–1.0. |
| `compromise_vector` | `*string` | Как сервис был скомпрометирован (nullable). |
| `last_heartbeat` | `*time.Time` | Последний heartbeat real-сервиса (nullable). |
| `seen_by` | `[]string` | Наблюдатели (сканеры). |
| `flags` | `map[string]any` | Сценарий-специфичные флаги. |
| `variables` | `map[string]any` | Локальные переменные процессов. |

### PlayerState

| Поле | Тип | Описание |
|------|-----|----------|
| `player_id` | `string` | Уникальный id игрока. |
| `name` | `string` | Отображаемое имя. |
| `org_id` | `*string` | Назначенная стартовая организация (nullable). |
| `score` | `int` | Текущий счёт. |
| `flags` | `[]string` | Захваченные флаги. |
| `status` | `PlayerStatus` | Статус игрока: active, idle, banned. |

### ScenarioState

| Поле | Тип | Описание |
|------|-----|----------|
| `scenario_id` | `string` | Идентификатор сценария. |
| `name` | `string` | Отображаемое имя. |
| `status` | `ScenarioStatus` | Жизненный цикл: running, paused, stopped. |
| `started_at` | `time.Time` | Время старта. |
| `ended_at` | `*time.Time` | Время окончания (nullable). |
| `config` | `map[string]any` | Параметры сценария. |

### WorldState

Полный снапшот runtime-мира.

| Поле | Тип | Описание |
|------|-----|----------|
| `tick` | `int` | Текущий tick симуляции. |
| `started_at` | `time.Time` | Время старта движка. |
| `services` | `map[string]ServiceState` | Состояние по сервисам. |
| `players` | `map[string]PlayerState` | Игроки. |
| `active_scenario` | `*ScenarioState` | Запущенный сценарий (nullable). |
| `variables` | `map[string]any` | Глобальные переменные. |

## Событийный граф

Append-only causal graph.

### EventNode

| Поле | Тип | Описание |
|------|-----|----------|
| `event_id` | `string` | UUID события. |
| `parent_event_ids` | `[]string` | Непосредственные causal-родители. |
| `correlation_id` | `string` | Группировка по сценарию/инциденту. |
| `tick` | `int` | Tick, в котором событие сгенерировано. |
| `timestamp` | `time.Time` | Wall-clock timestamp. |
| `source_type` | `SourceType` | engine, service, scenario, player, system, background, collector. |
| `source_id` | `string` | Id источника. |
| `event_type` | `EventType` | Тип события. |
| `target_id` | `*string` | Целевой topology-узел (nullable). |
| `payload` | `map[string]any` | Событие-специфичные данные. |
| `status` | `EventStatus` | Статус обработки: pending, processed, failed, suppressed. |

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
| `source_event_id` | `string` | Родительское событие. |
| `target_event_id` | `string` | Дочернее событие. |
| `kind` | `EdgeKind` | Отношение: caused_by, propagated_to, triggered_rule, response_to. |

## Конфигурация

### EngineConfig

| Поле | По умолчанию | Описание |
|------|--------------|----------|
| `app_name` | `cybercity-engine` | Имя приложения. |
| `debug` | `false` | Режим отладки. |
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
- Формат event-envelope на границах репозиториев — в
  [`cybercity/CONVENTIONS.md`](https://github.com/TheCipherKeeper/cybercity/blob/main/CONVENTIONS.md).

## Связанные документы

- [`docs/ARCHITECTURE.md`](ARCHITECTURE.md) — как модели связаны между собой.
- [`docs/DATA_FLOW.md`](DATA_FLOW.md) — как события движутся через модели.