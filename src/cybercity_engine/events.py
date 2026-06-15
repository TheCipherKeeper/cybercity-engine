"""Event graph store.

Maintains the dynamic causal graph of runtime events. Events are persisted to
PostgreSQL for audit and replay, but the active window is kept in memory for
fast routing decisions.
"""

from __future__ import annotations

from collections import deque
from typing import Any

from .models import EventEdge, EventNode

__all__ = ["EventGraph"]


class EventGraph:
    """In-memory event graph with a bounded recent window."""

    def __init__(self, max_recent: int = 10_000) -> None:
        self._nodes: dict[str, EventNode] = {}
        self._edges: list[EventEdge] = []
        self._recent: deque[str] = deque(maxlen=max_recent)

    def add(self, event: EventNode) -> EventNode:
        """Store an event and link it to its parents."""
        self._nodes[event.event_id] = event
        self._recent.append(event.event_id)
        for parent_id in event.parent_event_ids:
            self._edges.append(
                EventEdge(
                    source_event_id=parent_id,
                    target_event_id=event.event_id,
                    kind="caused_by",
                )
            )
        return event

    def link(
        self,
        source_event_id: str,
        target_event_id: str,
        kind: Any,
    ) -> None:
        """Record a non-causal relationship between two events."""
        self._edges.append(
            EventEdge(
                source_event_id=source_event_id,
                target_event_id=target_event_id,
                kind=kind,
            )
        )

    def get(self, event_id: str) -> EventNode | None:
        return self._nodes.get(event_id)

    def recent(self, limit: int = 100) -> list[EventNode]:
        ids = list(self._recent)[-limit:]
        return [self._nodes[i] for i in ids if i in self._nodes]

    def lineage(self, event_id: str) -> list[EventNode]:
        """Return all ancestor events for a given event id."""
        ancestors: list[EventNode] = []
        seen: set[str] = set()
        stack = [event_id]
        while stack:
            current_id = stack.pop()
            if current_id in seen:
                continue
            seen.add(current_id)
            current = self._nodes.get(current_id)
            if current is None:
                continue
            ancestors.append(current)
            stack.extend(current.parent_event_ids)
        return list(reversed(ancestors))

    def __len__(self) -> int:
        return len(self._nodes)
