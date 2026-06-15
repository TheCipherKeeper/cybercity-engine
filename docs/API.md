# CyberCity Engine — API

Движок предоставляет **тонкий HTTP и WebSocket API**. Он не содержит
бизнес-логики; он валидирует входные данные, преобразует их в события и
пересылает в движок. Все изменения состояния происходят внутри tick-цикла
движка.

## Базовый URL

```text
http://engine-host:8000
```

## Health

### `GET /health`

Возвращает статус движка.

```json
{
  "status": "ok",
  "tick": 42,
  "services": 263
}
```

## Состояние

### `GET /state`

Возвращает полный текущий `WorldState`.

```json
{
  "tick": 42,
  "started_at": "2026-06-15T12:00:00Z",
  "services": { ... },
  "players": { ... },
  "active_scenario": null,
  "variables": {}
}
```

## Топология

### `GET /topology`

Возвращает статический `TopologyGraph`.

```json
{
  "schema_version": "3.0.0",
  "source_version": "0.4.0",
  "services": { ... },
  "edges": [ ... ]
}
```

## События

### `GET /events/recent?limit=100`

Возвращает недавние события из in-memory событийного графа.

```json
[
  {
    "event_id": "evt-uuid",
    "event_type": "SCAN",
    "target_id": "bank-web",
    "source_id": "p1",
    "tick": 42,
    ...
  }
]
```

## Команды

### `POST /commands`

Отправляет команду игрока или инструктора.

Запрос:

```json
{
  "player_id": "p1",
  "action": "SCAN",
  "target": "bank-web",
  "params": {
    "noise_level": 0.9
  }
}
```

Ответ:

```json
{
  "status": "queued",
  "event_id": "evt-uuid"
}
```

Поддерживаемые действия обрабатываются движком, а не API. API только
проверяет наличие обязательных полей.

## WebSocket `/ws`

Подключение для real-time обновлений.

### При подключении

Сервер отправляет снапшот:

```json
{
  "type": "SNAPSHOT",
  "data": { /* WorldState */ }
}
```

### Клиент → сервер: команда

```json
{
  "player_id": "p1",
  "action": "SCAN",
  "target": "bank-web",
  "params": { "noise_level": 0.9 }
}
```

### Сервер → клиент: результат команды

```json
{
  "type": "COMMAND_RESULT",
  "status": "ACCEPTED",
  "event_id": "evt-uuid"
}
```

### Сервер → клиент: изменение состояния

```json
{
  "type": "STATE_UPDATE",
  "tick": 43,
  "changes": [
    {
      "entity": "service",
      "id": "bank-web",
      "field": "seen_by",
      "added": "p1"
    }
  ]
}
```

### Сервер → клиент: event log

```json
{
  "type": "EVENT_LOG",
  "tick": 43,
  "event": { /* EventNode */ }
}
```

### Сервер → клиент: статус симуляции

```json
{
  "type": "SIMULATION_STATUS",
  "tick": 43,
  "status": "RUNNING",
  "speed": 1.0
}
```

## Ответы с ошибками

API возвращает HTTP 422 для невалидных command payloads. Неизвестные команды
или target'ы принимаются как события; движок позже может пометить их как
suppressed или rejected через событийный граф.

WebSocket-ошибки отправляются как:

```json
{
  "type": "ERROR",
  "message": "Invalid message format"
}
```

## Будущие дополнения

- `POST /scenarios/{id}/start`
- `POST /scenarios/{id}/pause`
- `POST /scenarios/{id}/stop`
- `GET /events/{event_id}/lineage`
- `GET /replay?from_tick=0&to_tick=100`

## Связанные документы

- [`docs/ARCHITECTURE.md`](ARCHITECTURE.md) — системный контекст.
- [`docs/MODELS.md`](MODELS.md) — схемы данных.
- [`docs/DATA_FLOW.md`](DATA_FLOW.md) — как команды превращаются в события.
