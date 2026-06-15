# CyberCity Engine — Development Guide

## Quick start

```bash
cd /path/to/cybercity-engine

# Start dependencies
docker compose up -d postgres redpanda minio

# Install dependencies
uv sync

# Run tests
uv run pytest -q

# Run linters
uv run ruff check
uv run mypy --strict src/cybercity_engine

# Run engine locally
uv run cybercity-engine --engine-zip /path/to/engine.zip
```

## Project layout

```
cybercity-engine/
├── pyproject.toml              # package + deps + tool config
├── compose.yaml                # local dependencies
├── README.md                   # overview
├── AGENTS.md                   # rules for AI agents
├── src/cybercity_engine/       # engine source
│   ├── models.py               # topology + event + state models
│   ├── bootstrap.py            # load topology from engine.zip
│   ├── state.py                # StateManager
│   ├── events.py               # EventGraph
│   ├── router.py               # EventRouter
│   ├── engine.py               # main tick loop + handlers
│   ├── api.py                  # FastAPI + WebSocket
│   ├── config.py               # Pydantic settings
│   └── __main__.py             # CLI entry point
├── tests/                      # pytest suite
└── docs/                       # documentation
    ├── VISION.md
    ├── ARCHITECTURE.md
    ├── DATA_FLOW.md
    ├── MODELS.md
    ├── API.md
    ├── DEPLOYMENT.md
    └── adr/
```

## Working on the engine

### Adding a new event type

1. Add the value to `EventType` in `src/cybercity_engine/models.py`.
2. Add a handler in `src/cybercity_engine/engine.py` and register it in
   `Engine._handlers`.
3. Add tests in `tests/test_engine.py`.
4. Update `docs/MODELS.md`.

### Adding a propagation rule

1. Write a pure function in `src/cybercity_engine/router.py`.
2. Register it in `EventRouter._default_rules()`.
3. Add tests that verify propagation conditions.

### Modifying the API

1. Keep endpoints thin: convert input to events and enqueue them.
2. Never mutate `WorldState` directly from `api.py`.
3. Document changes in `docs/API.md`.

## Testing

```bash
# Run all tests
uv run pytest -q

# With coverage
uv run pytest -q --cov=src/cybercity_engine --cov-report=term-missing

# Specific test
uv run pytest tests/test_engine.py -q
```

## Linting and type checking

```bash
uv run ruff check
uv run ruff check --fix    # auto-fix where possible
uv run mypy --strict src/cybercity_engine
```

## Commit style

Use conventional commits:

```text
feat: add background degradation process
fix: clamp health values in StateManager
docs: update API.md with scenario endpoints
refactor: extract event store interface
```

Breaking changes must include `BREAKING CHANGE:` in the body.

## ADR process

If your change alters an architectural decision:

1. Write or update an ADR under `docs/adr/`.
2. Reference it from `docs/ARCHITECTURE.md`.
3. Mark old ADRs as `superseded` rather than deleting them.

## Useful commands

```bash
# Inspect city artifact
uv run python -c "from cybercity_engine.bootstrap import load_topology; t = load_topology('engine.zip'); print(len(t.services), len(t.edges))"

# Connect to local PostgreSQL
psql postgresql://engine:engine@localhost:5432/cybercity

# Redpanda admin UI
open http://localhost:9644
```

## Troubleshooting

### Tests fail with database connection

Make sure PostgreSQL is running:

```bash
docker compose up -d postgres
```

### Redpanda not reachable

Wait for the health check to pass:

```bash
docker compose logs -f redpanda
```

### mypy errors after model changes

Run with `--show-error-codes` and check for `Any` leakage, missing return
annotations, and literal enum usage.

## Related

- `AGENTS.md` — rules for AI agents.
- `docs/ARCHITECTURE.md` — high-level design.
- `docs/DEPLOYMENT.md` — how to run in lab or production.
