"""Core data models for the CyberCity engine.

The engine operates on two linked structures:

1. Topology graph — static blueprint loaded from `cybercity-data`.
2. Event graph — dynamic causal graph of runtime events.
"""

from __future__ import annotations

from datetime import UTC, datetime
from enum import StrEnum
from typing import Any, Literal
from uuid import uuid4

from pydantic import BaseModel, ConfigDict, Field

__all__ = [
    "ServiceMode",
    "ServiceStatus",
    "TopologyNode",
    "TopologyEdge",
    "TopologyGraph",
    "EventType",
    "EventNode",
    "EventEdge",
    "ServiceState",
    "PlayerState",
    "ScenarioState",
    "WorldState",
]


class ServiceMode(StrEnum):
    """How a service is realized at runtime."""

    SIMULATED = "simulated"
    REAL = "real"
    DECOY = "decoy"


class ServiceStatus(StrEnum):
    """Runtime status of a service."""

    UP = "up"
    DOWN = "down"
    COMPROMISED = "compromised"
    MAINTENANCE = "maintenance"


class TopologyNode(BaseModel):
    """Static service description from the city blueprint.

    Mirrors ``Service`` from ``cybercity-data`` but keeps only the fields the
    engine needs for routing and emulation.
    """

    model_config = ConfigDict(extra="ignore")

    id: str
    org_id: str
    name: str
    kind: str
    exposure: Literal["public", "intranet", "ot", "mgmt"]
    host: str
    network_id: str | None = None
    bind_ip: str | None = None
    auth: str = "local"
    data_classification: str = "internal"
    criticality: Literal["critical", "high", "medium", "low"] = "medium"
    ports: list[str] = Field(default_factory=list)
    is_decoy: bool = False
    decoy_kind: str | None = None
    software: dict[str, Any] = Field(default_factory=dict)
    os_hint: str | None = None


class TopologyEdge(BaseModel):
    """Static or inferred relationship between two services.

    Declared edges come from ``cybercity-data`` links. Inferred edges are
    added at runtime (same network, same org, exposure reachability).
    """

    source: str
    target: str
    kind: str
    protocol: str | None = None
    encryption: str | None = None
    inferred: bool = False


class TopologyGraph(BaseModel):
    """Immutable blueprint loaded from ``cybercity-data``."""

    schema_version: str
    source_version: str
    services: dict[str, TopologyNode] = Field(default_factory=dict)
    edges: list[TopologyEdge] = Field(default_factory=list)

    def neighbors(self, service_id: str) -> list[TopologyEdge]:
        """Return all outgoing edges from a service."""
        return [e for e in self.edges if e.source == service_id]


class EventType(StrEnum):
    """Types of runtime events."""

    HEARTBEAT = "heartbeat"
    SCAN = "scan"
    ATTACK = "attack"
    COMPROMISE = "compromise"
    RESTORE = "restore"
    STATE_CHANGE = "state_change"
    COMMAND = "command"
    SCENARIO_START = "scenario_start"
    SCENARIO_STOP = "scenario_stop"
    FLAG_CAPTURED = "flag_captured"
    BACKGROUND_EFFECT = "background_effect"
    PROPAGATION = "propagation"


class EventNode(BaseModel):
    """A single event in the dynamic event graph.

    Events are immutable. Every event may have one or more causal parents,
    forming a provenance graph that can be replayed and inspected.
    """

    event_id: str = Field(default_factory=lambda: str(uuid4()))
    parent_event_ids: list[str] = Field(default_factory=list)
    correlation_id: str = Field(default_factory=lambda: str(uuid4()))

    tick: int = 0
    timestamp: datetime = Field(default_factory=lambda: datetime.now(UTC))

    source_type: Literal["engine", "service", "scenario", "player", "system", "background"]
    source_id: str
    event_type: EventType
    target_id: str | None = None

    payload: dict[str, Any] = Field(default_factory=dict)
    status: Literal["pending", "processed", "failed", "suppressed"] = "pending"

    def spawn_child(
        self,
        event_type: EventType,
        source_type: Literal["engine", "service", "scenario", "player", "system", "background"],
        source_id: str,
        target_id: str | None = None,
        payload: dict[str, Any] | None = None,
        tick: int | None = None,
    ) -> EventNode:
        """Create a child event caused by this event."""
        return EventNode(
            parent_event_ids=[self.event_id, *self.parent_event_ids],
            correlation_id=self.correlation_id,
            tick=tick if tick is not None else self.tick,
            source_type=source_type,
            source_id=source_id,
            event_type=event_type,
            target_id=target_id,
            payload=payload or {},
        )


class EventEdge(BaseModel):
    """Causal or propagation relationship between two events."""

    source_event_id: str
    target_event_id: str
    kind: Literal["caused_by", "propagated_to", "triggered_rule", "response_to"]


class ServiceState(BaseModel):
    """Mutable runtime state attached to a topology node."""

    service_id: str
    status: ServiceStatus = ServiceStatus.UP
    health: float = Field(default=1.0)
    compromise_vector: str | None = None
    last_heartbeat: datetime | None = None
    seen_by: list[str] = Field(default_factory=list)
    flags: dict[str, Any] = Field(default_factory=dict)
    variables: dict[str, Any] = Field(default_factory=dict)


class PlayerState(BaseModel):
    """Player in an exercise."""

    player_id: str
    name: str
    org_id: str | None = None
    score: int = 0
    flags: list[str] = Field(default_factory=list)
    status: Literal["active", "idle", "banned"] = "idle"


class ScenarioState(BaseModel):
    """Active training scenario."""

    scenario_id: str
    name: str
    status: Literal["running", "paused", "stopped"] = "running"
    started_at: datetime = Field(default_factory=lambda: datetime.now(UTC))
    ended_at: datetime | None = None
    config: dict[str, Any] = Field(default_factory=dict)


class WorldState(BaseModel):
    """Complete runtime snapshot of the simulation."""

    tick: int = 0
    started_at: datetime = Field(default_factory=lambda: datetime.now(UTC))
    services: dict[str, ServiceState] = Field(default_factory=dict)
    players: dict[str, PlayerState] = Field(default_factory=dict)
    active_scenario: ScenarioState | None = None
    variables: dict[str, Any] = Field(default_factory=dict)

    def service(self, service_id: str) -> ServiceState | None:
        return self.services.get(service_id)
