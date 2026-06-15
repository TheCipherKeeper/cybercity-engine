# CyberCity Engine — Vision

## What is CyberCity?

CyberCity is a **digital twin of a city's IT/OT infrastructure** designed as a
platform for cybersecurity training, research, and public demonstration.

It models a fictional city as a directed graph:

- **Organizations** represent real-world entities: banks, hospitals, power
  utilities, government offices, schools, telecoms.
- **Services** are the nodes: web portals, databases, SCADA gateways, DNS,
  backups, workstations.
- **Links** are the edges: authentication flows, API calls, database reads,
  backup streams, vendor VPNs.

The goal is not to build yet another vulnerable-by-design CTF image, but a
**living, observable, explainable city** where security incidents propagate
through dependencies, where players can trace attack paths, and where every
state change has a recorded cause.

## Why this project exists

1. **Public footprint for a senior engineer.** The project demonstrates
   architecture, distributed systems, platform engineering, security domain
   knowledge, and operational discipline in a single coherent system.

2. **Reference implementation of an event-driven digital twin.** Many security
   platforms need causal provenance, replay, and graph-based propagation. This
   engine explores those ideas in a concrete, testable form.

3. **Playground for hybrid cyber ranges.** It bridges simulated services and
   real virtual machines through a common event bus, making it possible to
   start small and grow into a full lab.

## Guiding principles

### Data as code

The city is declared in YAML under `cybercity-data`. It is versioned,
validated, and rendered into artifacts. The engine consumes those artifacts,
never hard-codes topology.

### Events as the single source of truth

Runtime state is a projection of the event stream. If you know all events, you
can reconstruct the entire history of the city. This enables:

- audit and compliance;
- replay and what-if analysis;
- explainable state transitions;
- debugging and scenario review.

### Two graphs, one runtime

- **Topology graph** answers: *what is connected to what?*
- **Event graph** answers: *what happened and why?*

They are separate but linked. Topology provides the rails; events describe the
traffic on those rails.

### Engine as the sole mutator

Only the engine changes the world state. UI, scenario manager, agents, and real
services send events; the engine validates and applies them. This boundary keeps
consistency simple and makes the system observable.

### Hybrid by design

A service can be:

- **simulated** — lightweight, engine-managed decoy;
- **real** — a VM or container running real software;
- **decoy** — a deliberately attractive fake target.

The engine treats all three uniformly at the event level. The difference is in
who answers the event: the engine itself or an external agent.

### Python first, Go later

This engine is intentionally written in Python for rapid iteration and clarity.
When the architecture stabilizes, a production-grade Go port may follow. The
Python version remains the **reference implementation** for concepts, tests, and
ADRs.

### Secure by default

- Network segmentation is explicit.
- Secrets and credentials are never committed.
- Public access is read-only and tunnelled.
- OT/ICS segments are isolated from management and public networks.

## Target audiences

| Audience | What they get |
|----------|---------------|
| **Recruiters / hiring managers** | A coherent, documented, tested system with clear architecture and public demo. |
| **Security engineers** | A model for attack surface, propagation, incident response training. |
| **Platform engineers** | An example of event-driven runtime, GitOps, observability, hybrid execution. |
| **Students / learners** | A way to see how infrastructure dependencies create cascading risk. |
| **Future contributors** | Clear ADRs, boundaries, and extension points. |

## Success criteria

The project is successful when:

1. A visitor can open a public URL and see a live, interactive city graph.
2. A player can trigger an event and watch it propagate through dependencies.
3. Every state change in the city is explainable via the event graph.
4. The repository has clear architecture, ADRs, tests, and CI/CD.
5. The system can run on a home lab and scale conceptually to a production
   cyber range.

## Non-goals

- Full physical realism of water, power, or transport.
- Replacing commercial cyber-range platforms.
- Real-world traffic simulation at ISP scale.
- Game engine-quality 3D visualization.

## Related repositories

| Repository | Role |
|------------|------|
| `cybercity` | Landing page and public showcase. |
| `cybercity-data` | Canonical declarative city model, validator, builder. |
| `cybercity-engine` | **This repository** — event-driven runtime. |
| `cybercity-ui` | Web frontend for visualization and interaction. |
| `cybercity-agents` | LLM-assisted content generators (future). |
| `cybercity-blueprints` | Reusable org/service templates (future). |

## License

- Code: MIT
- Documentation: CC BY 4.0
