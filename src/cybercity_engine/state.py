"""Runtime state manager for the CyberCity engine.

The state manager owns the in-memory ``WorldState`` and persists periodic
snapshots to PostgreSQL. It is the only component allowed to mutate runtime
state directly.
"""

from __future__ import annotations

from datetime import UTC, datetime
from typing import Any

from .models import (
    EventNode,
    EventType,
    PlayerState,
    ScenarioState,
    ServiceState,
    ServiceStatus,
    TopologyGraph,
    WorldState,
)

__all__ = ["StateManager"]


class StateManager:
    """Owns the mutable runtime world state."""

    def __init__(self, topology: TopologyGraph) -> None:
        self.topology = topology
        self.world = WorldState()
        self._init_services()

    def _init_services(self) -> None:
        for svc_id in self.topology.services:
            self.world.services[svc_id] = ServiceState(service_id=svc_id)

    # ──────────────────────────────────────────────────────────────────
    # Service state mutations
    # ──────────────────────────────────────────────────────────────────
    def set_service_status(
        self, service_id: str, status: ServiceStatus, reason_event: EventNode | None = None
    ) -> EventNode | None:
        svc = self._require_service(service_id)
        if svc.status == status:
            return None

        old_status = svc.status
        svc.status = status
        svc.last_heartbeat = datetime.now(UTC)

        return EventNode(
            tick=self.world.tick,
            source_type="engine",
            source_id="state-manager",
            event_type=EventType.STATE_CHANGE,
            target_id=service_id,
            payload={
                "entity": "service",
                "field": "status",
                "old_value": old_status,
                "new_value": status,
                "parent_event_id": reason_event.event_id if reason_event else None,
            },
        )

    def record_heartbeat(self, service_id: str, timestamp: datetime | None = None) -> None:
        svc = self._require_service(service_id)
        svc.last_heartbeat = timestamp or datetime.now(UTC)

    def mark_seen_by(self, service_id: str, observer_id: str) -> None:
        svc = self._require_service(service_id)
        if observer_id not in svc.seen_by:
            svc.seen_by.append(observer_id)

    def set_health(self, service_id: str, health: float) -> None:
        svc = self._require_service(service_id)
        svc.health = max(0.0, min(1.0, health))

    def set_flag(self, service_id: str, key: str, value: Any) -> None:
        svc = self._require_service(service_id)
        svc.flags[key] = value

    # ──────────────────────────────────────────────────────────────────
    # Player / scenario state
    # ──────────────────────────────────────────────────────────────────
    def add_player(self, player: PlayerState) -> None:
        self.world.players[player.player_id] = player

    def set_scenario(self, scenario: ScenarioState | None) -> None:
        self.world.active_scenario = scenario

    # ──────────────────────────────────────────────────────────────────
    # Tick / snapshot helpers
    # ──────────────────────────────────────────────────────────────────
    def increment_tick(self) -> None:
        self.world.tick += 1

    # ──────────────────────────────────────────────────────────────────
    # Helpers
    # ──────────────────────────────────────────────────────────────────
    def _require_service(self, service_id: str) -> ServiceState:
        if service_id not in self.world.services:
            raise KeyError(f"unknown service {service_id!r}")
        return self.world.services[service_id]
