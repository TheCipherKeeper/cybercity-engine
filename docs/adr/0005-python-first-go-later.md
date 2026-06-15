# ADR-0005: Python-first reference implementation

## Status

Accepted

## Context

The engine needs to be built quickly to validate architecture, data models,
propagation rules, and UX. At the same time, the final platform may require
the concurrency and performance of a systems language.

## Decision

Implement the **Python reference version first**.

Goals of the Python version:

1. Validate the two-graph architecture.
2. Establish event schemas and propagation rules.
3. Build tests, ADRs, and documentation.
4. Provide a public demo and learning platform.

When the architecture is stable and the demo is working, a **Go port** may be
developed for production performance. The Python version remains the reference
for semantics and tests.

## Rationale

- Python ecosystem is excellent for rapid prototyping: FastAPI, Pydantic,
  NetworkX, pytest, asyncio.
- Cybersecurity domain often uses Python for tools and automation.
- Tests and ADRs written in Python transfer directly to Go design.
- Performance is not the bottleneck at the home-lab / demo stage.

## When to port to Go

Consider a Go implementation when:

- Single Python process cannot keep up with event throughput.
- Horizontal scaling becomes necessary.
- The public demo needs lower latency or higher player count.
- The architecture is proven stable for at least one full scenario.

## Boundaries

The Python engine must keep the core logic separate from framework details so
that a future Go port can reuse the same design:

- Models are pure data.
- Handlers are pure functions over state and events.
- Router rules are pure functions.
- Event store is an interface.
- Persistence is an interface.

## Consequences

### Positive

- Fast iteration and clear code.
- Strong test coverage before committing to Go.
- Lower barrier for contributors and reviewers.

### Negative

- Python's GIL limits CPU parallelism.
- Higher memory use than Go.
- Some production features (true zero-allocation hot path) may require Go.

## Related

- `README.md` — status and language choice.
- ADR-0002 — event-driven runtime.
