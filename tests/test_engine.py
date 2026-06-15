"""Tests for the engine event loop."""

from __future__ import annotations

import pytest

from cybercity_engine.config import EngineConfig
from cybercity_engine.engine import Engine
from cybercity_engine.models import (
    EventNode,
    EventType,
    TopologyEdge,
    TopologyGraph,
    TopologyNode,
)


def _minimal_topology() -> TopologyGraph:
    return TopologyGraph(
        schema_version="3.0.0",
        source_version="0.1.0",
        services={
            "bank-web": TopologyNode(
                id="bank-web",
                org_id="bank",
                name="Bank portal",
                kind="web",
                exposure="public",
                host="portal.bank.corp",
            ),
            "bank-db": TopologyNode(
                id="bank-db",
                org_id="bank",
                name="Bank DB",
                kind="db",
                exposure="intranet",
                host="db.bank.corp",
            ),
            "bank-log": TopologyNode(
                id="bank-log",
                org_id="bank",
                name="Bank log sink",
                kind="log",
                exposure="intranet",
                host="log.bank.corp",
            ),
        },
        edges=[
            TopologyEdge(source="bank-web", target="bank-db", kind="db-read"),
            TopologyEdge(source="bank-web", target="bank-log", kind="log-sink"),
        ],
    )


def _make_engine() -> Engine:
    return Engine(_minimal_topology(), EngineConfig(tick_ms=0))


@pytest.mark.asyncio
async def test_heartbeat_updates_service() -> None:
    engine = _make_engine()
    event = EventNode(
        source_type="service",
        source_id="bank-web",
        event_type=EventType.HEARTBEAT,
        target_id="bank-web",
    )
    await engine._process_event(event)
    assert engine.state.world.services["bank-web"].last_heartbeat is not None
    assert event.status == "processed"


@pytest.mark.asyncio
async def test_scan_propagates_to_log_sink() -> None:
    engine = _make_engine()
    event = EventNode(
        source_type="player",
        source_id="p1",
        event_type=EventType.SCAN,
        target_id="bank-web",
        payload={"noise_level": 0.9},
    )
    await engine._process_event(event)
    recent = engine.event_graph.recent(10)
    kinds = {e.event_type for e in recent}
    assert EventType.SCAN in kinds


@pytest.mark.asyncio
async def test_attack_compromises_service() -> None:
    engine = _make_engine()
    event = EventNode(
        source_type="player",
        source_id="p1",
        event_type=EventType.ATTACK,
        target_id="bank-web",
        payload={"success": True, "vector": "sqli"},
    )
    await engine._process_event(event)
    assert engine.state.world.services["bank-web"].status == "compromised"


@pytest.mark.asyncio
async def test_command_isolates_service() -> None:
    engine = _make_engine()
    event = EventNode(
        source_type="player",
        source_id="p1",
        event_type=EventType.COMMAND,
        target_id="bank-web",
        payload={"action": "ISOLATE_SERVICE"},
    )
    await engine._process_event(event)
    assert engine.state.world.services["bank-web"].status == "maintenance"
