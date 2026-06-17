# CyberCity Engine — Потоки данных

Этот документ описывает, как информация движется через CyberCity engine.
Организовано по сценариям использования — от high-level действия игрока до
хранения и broadcast.

## 1. Игрок сканирует сервис

```text
UI игрока
    │ POST /commands { action: "SCAN", target: "bank-web", player_id: "p1" }
    ▼
HTTP POST /commands endpoint
    │ преобразует команду в EventNode(COMMAND)
    ▼
Engine.submit_command()
    │ ставит событие в очередь
    ▼
Tick loop
    │ опустошает очередь команд
    ▼
Engine._process_event(COMMAND)
    │ handler превращает COMMAND → SCAN событие
    ▼
Engine._process_event(SCAN)
    │ handler отмечает bank-web.seen_by p1
    │ просит EventRouter распространить
    ▼
EventRouter
    │ находит log-sink ребро bank-web → bank-log
    │ noise_level 0.9 >= 0.5
    │ создаёт дочернее PROPAGATION событие
    ▼
Engine._process_event(PROPAGATION)
    │ handler пока нет; status = suppressed
    ▼
EventGraph
    │ записывает SCAN, PROPAGATION и propagated_to ребро
    ▼
UI broadcast
    │ STATE_UPDATE для bank-web.seen_by
    │ EVENT_LOG для scan и alert
```

## 2. Игрок атакует и компрометирует сервис

```text
UI игрока
    │ POST /commands { action: "ATTACK", target: "bank-web", vector: "sqli" }
    ▼
Engine
    │ преобразует в ATTACK событие с success: true
    ▼
Engine._process_event(ATTACK)
    │ handler вызывает StateManager.set_service_status(COMPROMISED)
    ▼
StateManager
    │ меняет bank-web.status: up → compromised
    │ эмитит STATE_CHANGE событие
    ▼
Engine._process_event(STATE_CHANGE)
    │ связывает с родительским ATTACK
    │ просит EventRouter распространить
    ▼
EventRouter._state_change_propagation_rule
    │ находит db-read ребро bank-web → bank-db
    │ new_status == compromised
    │ эмитит BACKGROUND_EFFECT dependency_impact
    ▼
Engine._process_event(BACKGROUND_EFFECT)
    │ будущий handler может снизить health bank-db
    ▼
UI broadcast
    │ изменение статуса bank-web
    │ impact-событие bank-db
```

## 3. Heartbeat реального сервиса

```text
cybercity-collector (out-of-band наблюдатель, на гипервизоре/K8s-узле)
    │ наблюдает VM bank-web, шлёт подписанный (Ed25519) HEARTBEAT каждые 10s
    ▼
Redpanda topic city.service.heartbeat
    ▼
Consumer движка
    │ получает HEARTBEAT
    ▼
Engine._process_event(HEARTBEAT)
    │ обновляет bank-web.last_heartbeat
    ▼
Health checker (background-процесс)
    │ каждые N тиков проверяет last_heartbeat
    │ если пропущен порог → эмитит STATE_CHANGE down
```

## 4. Scenario manager запускает сценарий

```text
Scenario Manager
    │ эмитит SCENARIO_START событие
    ▼
Engine._process_event(SCENARIO_START)
    │ устанавливает active_scenario в WorldState
    │ эмитит начальные incident-события
    ▼
Engine._process_event(RESOURCE_IMPACT / STATE_CHANGE)
    │ каскадирует по топологии
    ▼
Scenario Manager слушает city.events
    │ обновляет scoring, проверяет win/lose условия
    │ эмитит FLAG_CAPTURED, если игрок достиг цели
```

## 5. Снапшот и восстановление

```text
Tick loop
    │ каждые snapshot_interval_ticks
    ▼
StateManager сериализует WorldState
    ▼
Таблица PostgreSQL snapshots
    │ INSERT (tick, state_json)
    ▼
При перезапуске движка
    │ загрузить последний снапшот
    │ возобновить с нужного tick
    │ при необходимости replay недавних событий из Redpanda
```

## 6. Аудит и replay

```text
city.audit topic
    │ получает каждое обработанное событие
    ▼
Таблица PostgreSQL events
    │ INSERT (tick, source, target, type, payload, correlation_id)
    ▼
Replay tool (будущее)
    │ читает события в порядке tick
    │ детерминированно восстанавливает WorldState
```

## Схема события

Полная схема — в `MODELS.md`. Минимальный пример:

```json
{
  "event_id": "evt-uuid",
  "parent_event_ids": ["parent-uuid"],
  "correlation_id": "incident-uuid",
  "tick": 42,
  "timestamp": "2026-06-15T12:00:00Z",
  "source_type": "player",
  "source_id": "p1",
  "event_type": "SCAN",
  "target_id": "bank-web",
  "payload": { "noise_level": 0.9, "ports": ["tcp/443"] },
  "status": "processed"
}
```

## Упорядочивание сообщений

- Команды от одного игрока обрабатываются в порядке отправки.
- События внутри одного tick обрабатываются FIFO.
- Дочерние события, сгенерированные во время tick, добавляются в pending-список
  и обрабатываются до начала следующего tick.
- Background-процессы запускаются после queued-событий.

## Долговечность

- События пишутся в `city.events` и `city.audit` сразу после обработки.
- Снапшоты сохраняются периодически.
- In-memory event graph хранит только недавнее окно; полная история — в
  PostgreSQL.

## Связанные документы

- [`docs/ARCHITECTURE.md`](ARCHITECTURE.md) — высокоуровневая архитектура.
- [`docs/MODELS.md`](MODELS.md) — модели данных.
- [`docs/API.md`](API.md) — HTTP/WebSocket протокол.
