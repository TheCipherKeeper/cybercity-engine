# ADR-0001: Two-graph runtime architecture

## Status

Accepted

## Context

The CyberCity engine must model a city-wide IT/OT infrastructure and run
cybersecurity scenarios on top of it. Two distinct responsibilities emerged:

1. **Static deployment blueprint** — what exists, how it is connected, how it
   should be deployed.
2. **Dynamic causal history** — what happened, why it happened, how influence
   propagated.

A single data structure for both concerns became confusing: the topology graph
is mostly immutable, while the event graph is append-only and temporal.

## Decision

Use **two separate but linked graphs** at runtime:

- **Topology Graph** — loaded from `cybercity-data` artifacts (`engine.zip`).
  Nodes are services. Edges are declared links plus inferred adjacency (same
  network, same org, exposure reachability). This graph answers "what is
  connected to what".

- **Event Graph** — built at runtime. Nodes are events. Edges are causal or
  propagation relationships (`caused_by`, `propagated_to`, `triggered_rule`).
  This graph answers "what happened and why".

Events reference topology nodes via `target_id`. Topology nodes influence event
routing through the `EventRouter`.

## Consequences

### Positive

- Clear separation of static and dynamic concerns.
- Topology graph can be reloaded without losing runtime provenance.
- Attack paths and incident timelines become first-class data.
- Replay and audit are straightforward.

### Negative

- More complex state to keep consistent.
- Event graph can grow quickly; needs retention/summarization strategy.
- Need to maintain mapping between service identity in topology and service
  identity in events.

## Alternatives considered

- **Single graph with mutable attributes**: rejected because it mixes
  blueprint and history, making provenance hard to reconstruct.
- **Event sourcing only with no topology**: rejected because deployment and
  routing need a stable structural model.

## Related

- `cybercity-data` provides the canonical topology.
- `EventRouter` consumes topology edges to decide propagation.
- `docs/ARCHITECTURE.md` — high-level system architecture.
- `docs/MODELS.md` — detailed data models.
