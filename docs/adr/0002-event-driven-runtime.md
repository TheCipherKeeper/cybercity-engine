# ADR-0002: Event-driven runtime with tick loop

## Status

Accepted

## Context

The engine must maintain a consistent runtime state while serving multiple
concurrent sources: player commands, scenario events, real service heartbeats,
simulated service responses, and scheduled background processes.

A request-response model would force every component to know every other
component, creating tight coupling and making provenance hard. An event-driven
model decouples producers from consumers and gives every state change a
recorded cause.

## Decision

The engine uses an **event-driven runtime** with a single tick loop:

1. External inputs are converted to events and placed in a queue or event bus.
2. The tick loop drains the queue, processes events, updates state, generates
   child events, and propagates them through the topology graph.
3. Background processes run once per tick and may emit events.
4. State changes are broadcast to subscribers (UI, scenario manager, audit log).

Only the engine's tick loop mutates `WorldState`.

## Consequences

### Positive

- Clear causality: every mutation is preceded by an event.
- Loose coupling between UI, scenarios, and real services.
- Easy to replay, test, and debug.
- Natural fit for Kafka/Redpanda event streaming.
- Deterministic with fixed seed and fixed event order.

### Negative

- Higher latency than direct mutation (acceptable for training use case).
- Need careful ordering to avoid race conditions.
- Event graph can grow large; needs retention strategy.

## Event lifecycle

```text
Producer → queue/bus → Engine._process_event() → handler → StateManager
                                            ↓
                                        EventRouter
                                            ↓
                                      child events
                                            ↓
                                       next tick
```

## Event ordering guarantees

Within a single tick:

1. All queued commands are drained.
2. Events are processed in FIFO order from the pending list.
3. A child event generated during processing is appended to the pending list
   and processed in the same tick if possible.

Across ticks, causality is preserved because child events carry
`parent_event_ids` and `correlation_id`.

## Background processes

Background processes are functions that read `WorldState` and optionally emit
events. They run after queued events are processed. Examples:

- `DegradationProcess` — reduces service level while a service is degraded.
- `RecoveryProcess` — advances recovery progress.
- `BackupPowerFuelProcess` — consumes fuel while backup power is active.

They are **not** nodes in the topology graph.

## Alternatives considered

- **Actor model (one actor per service)**: more complex for our scale, harder
  to debug causality.
- **Pure event sourcing with no tick**: makes background processes and
  scheduling awkward.
- **Request-response with shared mutable state**: rejected due to tight
  coupling and poor observability.

## Related

- ADR-0001: two-graph architecture.
- `docs/ARCHITECTURE.md` system context diagram.
