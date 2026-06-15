# CyberCity — Engine

[![Part of CyberCity](https://img.shields.io/badge/CyberCity-composition-blueviolet)](https://github.com/TheCipherKeeper/cybercity)
[![Python](https://img.shields.io/badge/Python-3.12-3776AB?logo=python)](https://python.org)
[![License: MIT](https://img.shields.io/badge/code-MIT-green)](LICENSE)

Event-driven runtime engine for the CyberCity digital twin.

This is the **Python reference implementation** of the engine. It is intentionally
developed in Python for rapid iteration and concept validation, with a planned
future port to Go for production-grade performance.

## Architecture

The engine is built around **two graphs**:

1. **Topology Graph** — static blueprint of the city loaded from `cybercity-data`:
   - nodes: services (`bank-web`, `hospital-db`, ...)
   - edges: declared links (`api-call`, `auth`, `db-read`, `backup-of`, ...)
   - also: inferred edges (same network, same org, exposure chain)

2. **Event Graph** — dynamic causal graph of everything that happens:
   - nodes: events (scan, compromise, state change, player action)
   - edges: `caused_by`, `propagated_to`, `triggered_rule`
   - enables attack provenance, replay, explainability

Events flow through **Redpanda/Kafka** and are processed by the engine tick loop.
Runtime state is snapshotted to **PostgreSQL**.

## Documentation

| Document | Purpose |
|----------|---------|
| [`docs/VISION.md`](docs/VISION.md) | Why the project exists and what it wants to be. |
| [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | High-level architecture and system context. |
| [`docs/DATA_FLOW.md`](docs/DATA_FLOW.md) | How events move through the system. |
| [`docs/MODELS.md`](docs/MODELS.md) | Data model reference. |
| [`docs/API.md`](docs/API.md) | HTTP and WebSocket protocol. |
| [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md) | Local dev, home lab, production sketch. |
| [`docs/DEVELOPMENT.md`](docs/DEVELOPMENT.md) | How to work on the engine. |
| [`docs/adr/`](docs/adr/) | Architecture decision records. |

## Quick start (local Docker Compose)

```bash
# 1. Start dependencies
uv run docker compose up -d postgres redpanda minio

# 2. Build or copy a city artifact from cybercity-data, then run engine
uv run cybercity-engine --engine-zip /path/to/engine.zip
```

See [`docs/DEVELOPMENT.md`](docs/DEVELOPMENT.md) for the full workflow.

## Status

Core architecture, models, bootstrap, engine loop, API skeleton, and
documentation are in place. Next: PostgreSQL persistence, Redpanda integration,
background processes, and the first scenario.

## License

- Code: [MIT](LICENSE)
- Documentation: [CC BY 4.0](LICENSE-DOCS)
