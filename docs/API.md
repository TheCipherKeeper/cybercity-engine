# CyberCity Engine — API

The engine exposes a **thin HTTP and WebSocket API**. It does not contain
business logic; it converts external input into events and forwards them to the
engine. All state changes happen inside the engine tick loop.

## Base URL

```text
http://engine-host:8000
```

## Health

### `GET /health`

Returns engine status.

```json
{
  "status": "ok",
  "tick": 42,
  "services": 263
}
```

## State

### `GET /state`

Returns the full current `WorldState`.

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

## Topology

### `GET /topology`

Returns the static `TopologyGraph`.

```json
{
  "schema_version": "3.0.0",
  "source_version": "0.4.0",
  "services": { ... },
  "edges": [ ... ]
}
```

## Events

### `GET /events/recent?limit=100`

Returns recent events from the in-memory event graph.

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

## Commands

### `POST /commands`

Submit a player or instructor command.

Request:

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

Response:

```json
{
  "status": "queued",
  "event_id": "evt-uuid"
}
```

Supported actions are handled by the engine, not the API. The API only
validates that required fields are present.

## WebSocket `/ws`

Connect for real-time updates.

### On connect

Server sends a snapshot:

```json
{
  "type": "SNAPSHOT",
  "data": { /* WorldState */ }
}
```

### Client to server: command

```json
{
  "player_id": "p1",
  "action": "SCAN",
  "target": "bank-web",
  "params": { "noise_level": 0.9 }
}
```

### Server to client: command result

```json
{
  "type": "COMMAND_RESULT",
  "status": "ACCEPTED",
  "event_id": "evt-uuid"
}
```

### Server to client: state update

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

### Server to client: event log

```json
{
  "type": "EVENT_LOG",
  "tick": 43,
  "event": { /* EventNode */ }
}
```

### Server to client: simulation status

```json
{
  "type": "SIMULATION_STATUS",
  "tick": 43,
  "status": "RUNNING",
  "speed": 1.0
}
```

## Error responses

The API returns HTTP 422 for invalid command payloads. Unknown commands or
targets are accepted as events; the engine may later mark them as suppressed
or rejected via the event graph.

WebSocket errors are sent as:

```json
{
  "type": "ERROR",
  "message": "Invalid message format"
}
```

## Future additions

- `POST /scenarios/{id}/start`
- `POST /scenarios/{id}/pause`
- `POST /scenarios/{id}/stop`
- `GET /events/{event_id}/lineage`
- `GET /replay?from_tick=0&to_tick=100`

## Related

- `docs/ARCHITECTURE.md` — system context.
- `docs/MODELS.md` — data schemas.
- `docs/DATA_FLOW.md` — how commands become events.
