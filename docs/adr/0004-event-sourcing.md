# 0004 — event-sourced state

Status: accepted
Date: 2026-06-02

## Context

The city's current state (which traffic lights are green, which
patients are critical, which alarms are firing) must be derivable
from a single source. Multiple sources drift; one source is auditable.

## Decision

The city is modeled as an **append-only event log**. State is a
projection of the log. Projections are derived materialised views,
not authoritative state.

## Consequences

- Replay from log = reproducible state.
- New projections (new dashboards) are added without changing writers.
- Storage is append-only; this shapes internal/events.
