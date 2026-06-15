"""FastAPI application and WebSocket endpoint.

The API is intentionally thin: it validates input, converts it to events, and
forwards them to the engine. No business logic lives here.
"""

from __future__ import annotations

from collections.abc import AsyncGenerator
from contextlib import asynccontextmanager
from typing import Any

from fastapi import FastAPI, WebSocket, WebSocketDisconnect

from .config import EngineConfig
from .engine import Engine
from .models import EventNode, EventType

__all__ = ["create_app"]


def create_app(engine: Engine, config: EngineConfig) -> FastAPI:
    """Create a FastAPI app bound to an engine instance."""

    @asynccontextmanager
    async def _lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
        task = None
        if config.tick_ms > 0:
            import asyncio

            task = asyncio.create_task(engine.start())
        yield
        engine.stop()
        if task is not None:
            task.cancel()
            try:
                await task
            except asyncio.CancelledError:
                pass

    app = FastAPI(title="CyberCity Engine", version="0.1.0", lifespan=_lifespan)

    @app.get("/health")
    async def health() -> dict[str, Any]:
        return {
            "status": "ok",
            "tick": engine.state.world.tick,
            "services": len(engine.state.world.services),
        }

    @app.get("/state")
    async def get_state() -> dict[str, Any]:
        return engine.state.world.model_dump(mode="json")

    @app.get("/topology")
    async def get_topology() -> dict[str, Any]:
        return engine.topology.model_dump(mode="json")

    @app.get("/events/recent")
    async def recent_events(limit: int = 100) -> list[dict[str, Any]]:
        return [e.model_dump(mode="json") for e in engine.event_graph.recent(limit)]

    @app.post("/commands")
    async def post_command(command: dict[str, Any]) -> dict[str, Any]:
        event = EventNode(
            source_type="player",
            source_id=command.get("player_id", "anonymous"),
            event_type=EventType.COMMAND,
            target_id=command.get("target"),
            payload={
                "action": command.get("action"),
                "params": command.get("params", {}),
            },
        )
        await engine.submit_command(event)
        return {"status": "queued", "event_id": event.event_id}

    @app.websocket("/ws")
    async def websocket(ws: WebSocket) -> None:
        await ws.accept()
        try:
            await ws.send_json(
                {
                    "type": "SNAPSHOT",
                    "data": engine.state.world.model_dump(mode="json"),
                }
            )
            while True:
                msg = await ws.receive_json()
                action = msg.get("action")
                target = msg.get("target")
                event = EventNode(
                    source_type="player",
                    source_id=msg.get("player_id", "anonymous"),
                    event_type=EventType.COMMAND,
                    target_id=target,
                    payload={
                        "action": action,
                        "params": msg.get("params", {}),
                    },
                )
                await engine.submit_command(event)
                await ws.send_json(
                    {
                        "type": "COMMAND_RESULT",
                        "status": "ACCEPTED",
                        "event_id": event.event_id,
                    }
                )
        except WebSocketDisconnect:
            pass

    return app
