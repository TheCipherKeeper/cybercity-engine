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

```text
┌─────────────────────────────────────────────────────────────┐
│                     Topology Graph                          │
│          (loaded from cybercity-data engine.zip)              │
│              services + links + networks                      │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     │ static blueprint
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                      Engine Runtime                           │
│                                                               │
│  ┌─────────────┐      ┌─────────────────┐      ┌──────────┐ │
│  │  API / WS   │◄────►│  Event Processor│◄────►│  Router  │ │
│  │  (FastAPI)  │      │                 │      │          │ │
│  └──────┬──────┘      └────────┬────────┘      └────┬─────┘ │
│         │                       │                   │       │
│         ▼                       ▼                   ▼       │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                    State Manager                         ││
│  │  services: {id → ServiceState}                           ││
│  │  players:  {id → PlayerState}                            ││
│  │  scenario: ScenarioState | None                          ││
│  └─────────────────────────────────────────────────────────┘│
│                                                               │
└────────────────────┬──────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                     Event Graph / Stream                    │
│              (Redpanda + PostgreSQL snapshots)                │
└─────────────────────────────────────────────────────────────┘
```

## Quick start (local Docker Compose)

```bash
# 1. Start dependencies
uv run docker compose up -d postgres redpanda minio

# 2. Run engine
uv run cybercity-engine --config envs/local.yaml

# 3. Or run in development mode with hot reload
uv run uvicorn cybercity_engine.api:app --reload
```

## CLI

```bash
cybercity-engine --config envs/local.yaml
```

## Status

Early development. Core models and bootstrap are being established.

## License

- Code: [MIT](LICENSE)
- Documentation: [CC BY 4.0](LICENSE-DOCS)
