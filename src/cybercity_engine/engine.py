"""Main engine loop and event processor.

The engine is the single authority over ``WorldState``. It consumes events
from the command queue and Redpanda, applies them through the state manager,
propagates them via the event router, and broadcasts resulting updates.
"""

from __future__ import annotations

import asyncio
from typing import Any

from .config import EngineConfig
from .events import EventGraph
from .models import EventNode, EventType, ServiceStatus, TopologyGraph
from .router import EventRouter
from .state import StateManager

__all__ = ["Engine"]


class Engine:
    """CyberCity simulation engine."""

    _handlers: dict[EventType, Any] = {}

    def __init__(
        self,
        topology: TopologyGraph,
        config: EngineConfig,
        router: EventRouter | None = None,
    ) -> None:
        self.topology = topology
        self.config = config
        self.state = StateManager(topology)
        self.router = router or EventRouter()
        self.event_graph = EventGraph()
        self._running = False
        self._command_queue: asyncio.Queue[EventNode] = asyncio.Queue()
        self._pending: list[EventNode] = []

    # ──────────────────────────────────────────────────────────────────
    # Public control
    # ──────────────────────────────────────────────────────────────────
    async def start(self) -> None:
        self._running = True
        while self._running:
            await self._tick()
            await asyncio.sleep(self.config.tick_ms / 1000.0)

    def stop(self) -> None:
        self._running = False

    async def submit_command(self, command: EventNode) -> None:
        """Enqueue a player/system command for processing."""
        await self._command_queue.put(command)

    # ──────────────────────────────────────────────────────────────────
    # Tick processing
    # ──────────────────────────────────────────────────────────────────
    async def _tick(self) -> None:
        self.state.increment_tick()

        # 1. Drain command queue
        commands = self._drain_commands()
        self._pending.extend(commands)

        # 2. Process pending events (FIFO for simplicity; later priority queue)
        while self._pending:
            event = self._pending.pop(0)
            await self._process_event(event)

        # 3. Background effects / recovery (placeholder)
        # TODO: background processes

    def _drain_commands(self) -> list[EventNode]:
        commands: list[EventNode] = []
        while not self._command_queue.empty():
            try:
                commands.append(self._command_queue.get_nowait())
            except asyncio.QueueEmpty:
                break
        return commands

    async def _process_event(self, event: EventNode) -> None:
        """Process a single event and propagate its effects."""
        self.event_graph.add(event)

        handler = self._handlers.get(event.event_type)
        if handler is None:
            event.status = "suppressed"
            return

        result = handler(self, event)
        if result is not None:
            child_events, state_changes = result
            for child in child_events:
                self.event_graph.add(child)
                self.event_graph.link(
                    event.event_id,
                    child.event_id,
                    "propagated_to",
                )
                self._pending.append(child)
            for change in state_changes:
                self.event_graph.add(change)
                self.event_graph.link(
                    event.event_id,
                    change.event_id,
                    "triggered_rule",
                )
                # State changes can also propagate through the graph.
                if change.target_id:
                    source_node = self.topology.services.get(change.target_id)
                    if source_node:
                        propagated = self.router.propagate(change, source_node, self.topology)
                        self._pending.extend(propagated)

        event.status = "processed"

    # ──────────────────────────────────────────────────────────────────
    # Event handlers
    # ──────────────────────────────────────────────────────────────────
    def _handle_heartbeat(self, event: EventNode) -> tuple[list[EventNode], list[EventNode]]:
        target_id = event.target_id
        if target_id is None:
            return [], []
        self.state.record_heartbeat(target_id, event.timestamp)
        return [], []

    def _handle_scan(self, event: EventNode) -> tuple[list[EventNode], list[EventNode]]:
        target_id = event.target_id
        if target_id is None:
            return [], []
        observer_id = event.source_id
        self.state.mark_seen_by(target_id, observer_id)
        source_node = self.topology.services.get(target_id)
        if source_node is None:
            return [], []
        propagated = self.router.propagate(event, source_node, self.topology)
        return propagated, []

    def _handle_attack(self, event: EventNode) -> tuple[list[EventNode], list[EventNode]]:
        target_id = event.target_id
        if target_id is None:
            return [], []
        success = event.payload.get("success", False)
        if not success:
            return [], []
        state_change = self.state.set_service_status(
            target_id, ServiceStatus.COMPROMISED, reason_event=event
        )
        return [], [state_change] if state_change else []

    def _handle_command(self, event: EventNode) -> tuple[list[EventNode], list[EventNode]]:
        action = event.payload.get("action")
        target_id = event.target_id
        if action == "ENABLE_BACKUP_POWER" and target_id:
            self.state.set_flag(target_id, "using_backup_power", True)
            change = self.state.set_service_status(
                target_id, ServiceStatus.UP, reason_event=event
            )
            return [], [change] if change else []
        if action == "ISOLATE_SERVICE" and target_id:
            change = self.state.set_service_status(
                target_id, ServiceStatus.MAINTENANCE, reason_event=event
            )
            return [], [change] if change else []
        return [], []

    def _handle_compromise(self, event: EventNode) -> tuple[list[EventNode], list[EventNode]]:
        target_id = event.target_id
        if target_id is None:
            return [], []
        change = self.state.set_service_status(
            target_id, ServiceStatus.COMPROMISED, reason_event=event
        )
        return [], [change] if change else []


# Bind handlers after class definition to avoid forward-reference issues.
Engine._handlers = {
    EventType.HEARTBEAT: Engine._handle_heartbeat,
    EventType.SCAN: Engine._handle_scan,
    EventType.ATTACK: Engine._handle_attack,
    EventType.COMPROMISE: Engine._handle_compromise,
    EventType.COMMAND: Engine._handle_command,
}
