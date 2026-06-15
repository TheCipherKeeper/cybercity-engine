"""Tests for core engine models."""

from __future__ import annotations

from cybercity_engine.models import (
    EventNode,
    EventType,
    ServiceState,
    TopologyEdge,
    TopologyGraph,
    TopologyNode,
)


def test_topology_node_defaults() -> None:
    node = TopologyNode(
        id="bank-web",
        org_id="bank",
        name="Bank portal",
        kind="web",
        exposure="public",
        host="portal.bank.corp",
    )
    assert node.auth == "local"
    assert node.criticality == "medium"
    assert node.is_decoy is False


def test_topology_graph_neighbors() -> None:
    graph = TopologyGraph(
        schema_version="3.0.0",
        source_version="0.1.0",
        services={
            "a": TopologyNode(
                id="a", org_id="x", name="A", kind="web", exposure="public", host="a.corp"
            ),
            "b": TopologyNode(
                id="b", org_id="x", name="B", kind="db", exposure="intranet", host="b.corp"
            ),
        },
        edges=[TopologyEdge(source="a", target="b", kind="api-call")],
    )
    neighbors = graph.neighbors("a")
    assert len(neighbors) == 1
    assert neighbors[0].target == "b"


def test_event_spawn_child() -> None:
    parent = EventNode(
        tick=1,
        source_type="player",
        source_id="p1",
        event_type=EventType.SCAN,
        target_id="bank-web",
        payload={"ports": [443]},
    )
    child = parent.spawn_child(
        event_type=EventType.PROPAGATION,
        source_type="engine",
        source_id="bank-web",
        target_id="bank-log",
        payload={"noise_level": 0.8},
        tick=2,
    )
    assert child.parent_event_ids == [parent.event_id]
    assert child.correlation_id == parent.correlation_id
    assert child.tick == 2
    assert child.event_type == EventType.PROPAGATION


def test_service_state_health_no_hard_bounds() -> None:
    """Model does not clamp health; StateManager clamps during runtime updates."""
    svc = ServiceState(service_id="x", health=2.0)
    assert svc.health == 2.0
