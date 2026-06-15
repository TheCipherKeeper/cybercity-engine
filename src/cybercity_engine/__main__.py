"""CLI entry point for the CyberCity engine."""

from __future__ import annotations

import argparse
import logging
import sys

import uvicorn

from .bootstrap import load_topology
from .config import EngineConfig
from .engine import Engine

__all__ = ["main"]


def _setup_logging(level: int = logging.INFO) -> None:
    logging.basicConfig(
        level=level,
        format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
    )


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(prog="cybercity-engine")
    parser.add_argument(
        "--engine-zip",
        type=str,
        default=None,
        help="Path or URL to the engine.zip / engine.json artifact.",
    )
    parser.add_argument(
        "--host", type=str, default=None, help="API bind host (default 0.0.0.0)."
    )
    parser.add_argument(
        "--port", type=int, default=None, help="API bind port (default 8000)."
    )
    parser.add_argument(
        "--debug", action="store_true", help="Enable debug logging."
    )
    args = parser.parse_args(argv)

    _setup_logging(level=logging.DEBUG if args.debug else logging.INFO)

    config = EngineConfig()
    if args.host is not None:
        config.host = args.host
    if args.port is not None:
        config.port = args.port

    source = args.engine_zip or config.engine_zip_url
    topology = load_topology(source)

    engine = Engine(topology, config)

    from .api import create_app

    app = create_app(engine, config)

    uvicorn.run(
        app,
        host=config.host,
        port=config.port,
        log_level="debug" if config.debug else "info",
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
