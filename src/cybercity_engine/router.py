"""Event router — propagates events through the topology graph.

Given an event and a topology node, the router decides which neighbouring
services should receive derived events and what those events look like.
"""

from __future__ import annotations

from collections.abc import Callable

from .models import EventNode, EventType, TopologyEdge, TopologyGraph, TopologyNode

__all__ = ["EventRouter"]


PropagationRule = Callable[[EventNode, TopologyNode, TopologyEdge, TopologyGraph], EventNode | None]


class EventRouter:
    """Graph-aware event propagation engine.

    Rules are pure functions that inspect an event, the source node and an
    outgoing edge, then optionally return a child event for the neighbour.
    """

    def __init__(self, rules: list[PropagationRule] | None = None) -> None:
        self.rules = rules or self._default_rules()

    def propagate(
        self,
        event: EventNode,
        source_node: TopologyNode,
        topology: TopologyGraph,
    ) -> list[EventNode]:
        """Return child events for all affected neighbours."""
        children: list[EventNode] = []
        for edge in topology.neighbors(source_node.id):
            target = topology.services.get(edge.target)
            if target is None:
                continue
            for rule in self.rules:
                child = rule(event, source_node, edge, topology)
                if child is not None:
                    children.append(child)
        return children

    @staticmethod
    def _default_rules() -> list[PropagationRule]:
        return [
            _scan_alert_rule,
            _compromise_propagation_rule,
            _state_change_propagation_rule,
        ]


def _scan_alert_rule(
    event: EventNode, source: TopologyNode, edge: TopologyEdge, topology: TopologyGraph
) -> EventNode | None:
    """If a noisy scan happens on a service, alert its log-sink / IDS neighbours."""
    if event.event_type != EventType.SCAN:
        return None
    if edge.kind not in {"log-sink", "trusts"}:
        return None
    noisy = event.payload.get("noise_level", 0.0)
    if noisy < 0.5:
        return None
    return event.spawn_child(
        event_type=EventType.PROPAGATION,
        source_type="engine",
        source_id=source.id,
        target_id=edge.target,
        payload={
            "kind": "scan_alert",
            "original_event_id": event.event_id,
            "noise_level": noisy,
            "via": edge.kind,
        },
    )


def _compromise_propagation_rule(
    event: EventNode, source: TopologyNode, edge: TopologyEdge, topology: TopologyGraph
) -> EventNode | None:
    """A compromised service may propagate to trusted neighbours."""
    if event.event_type != EventType.COMPROMISE:
        return None
    if edge.kind not in {"trusts", "auth", "vendor-vpn"}:
        return None
    severity = event.payload.get("severity", 0.5)
    if severity < 0.3:
        return None
    return event.spawn_child(
        event_type=EventType.ATTACK,
        source_type="engine",
        source_id=source.id,
        target_id=edge.target,
        payload={
            "kind": "lateral_movement",
            "original_event_id": event.event_id,
            "severity": severity * 0.8,
            "via": edge.kind,
        },
    )


def _state_change_propagation_rule(
    event: EventNode, source: TopologyNode, edge: TopologyEdge, topology: TopologyGraph
) -> EventNode | None:
    """When a service goes down, send a resource-impact event to dependents."""
    if event.event_type != EventType.STATE_CHANGE:
        return None
    if event.payload.get("field") != "status":
        return None
    new_status = event.payload.get("new_value")
    if new_status not in {"down", "compromised"}:
        return None
    if edge.kind not in {"api-call", "db-read", "db-write", "backup-of", "log-sink"}:
        return None
    return event.spawn_child(
        event_type=EventType.BACKGROUND_EFFECT,
        source_type="engine",
        source_id=source.id,
        target_id=edge.target,
        payload={
            "kind": "dependency_impact",
            "original_event_id": event.event_id,
            "status": new_status,
            "via": edge.kind,
        },
    )
