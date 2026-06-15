# CyberCity Engine — Data Flow

This document describes how information moves through the CyberCity engine. It
is organized by use case, from high-level player action down to storage and
broadcast.

## 1. Player scans a service

```text
Player UI
    │ POST /commands { action: "SCAN", target: "bank-web", player_id: "p1" }
    ▼
FastAPI /commands endpoint
    │ converts command to EventNode(COMMAND)
    ▼
Engine.submit_command()
    │ queues event
    ▼
Tick loop
    │ drains command queue
    ▼
Engine._process_event(COMMAND)
    │ handler converts COMMAND → SCAN event
    ▼
Engine._process_event(SCAN)
    │ handler marks bank-web seen_by p1
    │ asks EventRouter to propagate
    ▼
EventRouter
    │ finds log-sink edge bank-web → bank-log
    │ noise_level 0.9 >= 0.5
    │ creates child PROPAGATION event
    ▼
Engine._process_event(PROPAGATION)
    │ no handler yet; status = suppressed
    ▼
EventGraph
    │ records SCAN, PROPAGATION, and propagated_to edge
    ▼
UI broadcast
    │ STATE_UPDATE for bank-web.seen_by
    │ EVENT_LOG for scan and alert
```

## 2. Player attacks and compromises a service

```text
Player UI
    │ POST /commands { action: "ATTACK", target: "bank-web", vector: "sqli" }
    ▼
Engine
    │ converts to ATTACK event with success: true
    ▼
Engine._process_event(ATTACK)
    │ handler calls StateManager.set_service_status(COMPROMISED)
    ▼
StateManager
    │ changes bank-web.status: up → compromised
    │ emits STATE_CHANGE event
    ▼
Engine._process_event(STATE_CHANGE)
    │ links to parent ATTACK
    │ asks EventRouter to propagate
    ▼
EventRouter._state_change_propagation_rule
    │ finds db-read edge bank-web → bank-db
    │ new_status == compromised
    │ emits BACKGROUND_EFFECT dependency_impact
    ▼
Engine._process_event(BACKGROUND_EFFECT)
    │ future handler may reduce bank-db health
    ▼
UI broadcast
    │ bank-web status change
    │ bank-db impact event
```

## 3. Real service heartbeat

```text
cybercity-agent on bank-web VM
    │ sends HEARTBEAT event every 10s
    ▼
Redpanda city.service.heartbeat topic
    ▼
Engine consumer
    │ receives HEARTBEAT
    ▼
Engine._process_event(HEARTBEAT)
    │ updates bank-web.last_heartbeat
    ▼
Health checker (background process)
    │ every N ticks checks last_heartbeat
    │ if missing for threshold → emit STATE_CHANGE down
```

## 4. Scenario manager starts a scenario

```text
Scenario Manager
    │ emits SCENARIO_START event
    ▼
Engine._process_event(SCENARIO_START)
    │ sets active_scenario in WorldState
    │ emits initial incident event(s)
    ▼
Engine._process_event(RESOURCE_IMPACT / STATE_CHANGE)
    │ cascades through topology
    ▼
Scenario Manager listens to city.events
    │ updates scoring, checks win/lose conditions
    │ emits FLAG_CAPTURED if player reaches objective
```

## 5. Snapshot and recovery

```text
Tick loop
    │ every snapshot_interval_ticks
    ▼
StateManager serializes WorldState
    ▼
PostgreSQL snapshots table
    │ INSERT (tick, state_json)
    ▼
On engine restart
    │ load latest snapshot
    │ resume from tick
    │ replay recent events from Redpanda if needed
```

## 6. Audit and replay

```text
city.audit topic
    │ receives every processed event
    ▼
PostgreSQL events table
    │ INSERT (tick, source, target, type, payload, correlation_id)
    ▼
Replay tool (future)
    │ reads events in tick order
    │ reconstructs WorldState deterministically
```

## Event schemas

See `MODELS.md` for full schema. Minimal example:

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

## Message ordering

- Commands from a single player are processed in submission order.
- Events within one tick are processed FIFO.
- Child events generated during a tick are appended to the pending list and
  processed before the next tick begins.
- Background processes run after queued events.

## Durability

- Events are written to `city.events` and `city.audit` as soon as they are
  processed.
- Snapshots are written periodically.
- In-memory event graph keeps only a recent window; full history is in
  PostgreSQL.

## Related

- `docs/ARCHITECTURE.md` — high-level architecture.
- `docs/MODELS.md` — data models.
- `docs/API.md` — HTTP/WebSocket protocol.
