# CyberCity Engine — Architecture

## TL;DR

`cybercity-engine` is an **event-driven runtime** for the CyberCity digital twin.
It loads a static topology graph from `cybercity-data`, maintains a dynamic
runtime state, and processes a stream of events through a graph-aware router.
Everything that changes in the city happens through an event; every event is
linked to its causes, forming a causal graph.

## System context

```text
┌─────────────────────────────────────────────────────────────────────┐
│                         External users                               │
│   Players │ Instructors │ Read-only visitors │ Scenario authors     │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        CyberCity Platform                          │
│                                                                      │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────────┐  │
│  │     UI      │    │   Engine    │    │    Scenario Manager     │  │
│  │  (React/    │◄──►│  (Python)   │◄──►│      (Python)           │  │
│  │  WebSocket) │    │             │    │                         │  │
│  └──────┬──────┘    └──────┬──────┘    └─────────────────────────┘  │
│         │                   │                                        │
│         │                   ▼                                        │
│         │          ┌─────────────────┐                               │
│         │          │ Redpanda/Kafka  │                               │
│         │          │  (event bus)    │                               │
│         │          └─────────────────┘                               │
│         │                   │                                        │
│         │     ┌─────────────┼─────────────┐                        │
│         │     ▼             ▼             ▼                          │
│  ┌──────▼─────┐   ┌────────▼────────┐   ┌──────────────┐          │
│  │ PostgreSQL │   │  Real services  │   │ Simulated    │          │
│  │  (state)   │   │  (VM / pod)     │   │ services     │          │
│  └────────────┘   └─────────────────┘   └──────────────┘          │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Infrastructure layer                            │
│              Proxmox + Kubernetes + Cilium + Multus                │
└─────────────────────────────────────────────────────────────────────┘
```

## Core responsibilities

| Component | Responsibility |
|-----------|---------------|
| **cybercity-data** | Declarative city model, validation, artifact generation. |
| **cybercity-engine** | Runtime state, event processing, propagation, snapshots. |
| **cybercity-ui** | Visualization, player input, real-time updates. |
| **Scenario Manager** | Starts/pauses/stops scenarios, injects events, scoring. |
| **Redpanda** | Event bus between engine, UI, scenario manager, real services. |
| **PostgreSQL** | Snapshots and audit log of the event graph. |
| **MinIO / S3** | `engine.zip` artifacts and replay dumps. |

## The two graphs

### 1. Topology graph

Loaded from `cybercity-data` artifacts.

```text
Nodes: services
  id, org_id, kind, exposure, host, network_id, bind_ip
  auth, data_classification, criticality, ports, software

Edges: links
  declared: api-call, auth, db-read, db-write, backup-of,
            log-sink, trusts, vendor-vpn, dns-query, ntp-query

  inferred: same_network, same_org, exposure_chain
```

The topology graph is **immutable during a simulation**. It is replaced only
when a new city artifact is loaded.

### 2. Event graph

Built at runtime.

```text
Nodes: events
  event_id, parent_event_ids, correlation_id
  tick, timestamp, source_type, source_id
  event_type, target_id, payload, status

Edges:
  caused_by       ─ event B was caused by event A
  propagated_to   ─ event B reached a neighbour because of event A
  triggered_rule  ─ event B was created by a propagation rule
  response_to     ─ event B is a deliberate response to event A
```

The event graph is **append-only**. Events are never deleted; they may be
summarized or archived to cold storage.

### Link between graphs

```text
Topology              Event
   │                    │
   │◄── target_id ──────│  "this event happened to bank-web"
   │                    │
   │── neighbors() ────►│  "where can this event go next?"
   │                    │
   │◄── state change ───┤  "bank-web is now compromised"
```

## Event flow

```text
1. Source produces an event
      player scan → bank-web

2. Engine receives event via queue or Redpanda

3. Event is added to the event graph

4. Handler updates runtime state
      bank-web.seen_by += player-1

5. Router decides propagation
      log-sink edge + noisy scan → alert event

6. Child events are enqueued and processed

7. State changes emit STATE_CHANGE events
      bank-web.status: up → compromised

8. State changes may propagate again
      compromised bank-web affects bank-db via db-read

9. Snapshot + broadcast to UI
```

## Engine internals

```text
┌─────────────────────────────────────────┐
│              Engine                       │
│                                          │
│  ┌─────────────┐    ┌─────────────────┐  │
│  │ API / WS    │◄──►│ Event Processor │  │
│  │ (FastAPI)   │    │                 │  │
│  └──────┬──────┘    └────────┬────────┘  │
│         │                     │            │
│         ▼                     ▼            │
│  ┌─────────────────────────────────────┐  │
│  │          StateManager               │  │
│  │   services, players, scenario      │  │
│  └─────────────────────────────────────┘  │
│         │                     │            │
│         ▼                     ▼            │
│  ┌─────────────┐    ┌─────────────────┐   │
│  │ EventGraph  │    │  EventRouter    │   │
│  │  (causal)   │    │  (propagation)  │   │
│  └─────────────┘    └─────────────────┘   │
│                                          │
└─────────────────────────────────────────┘
```

### StateManager

- Single owner of mutable `WorldState`.
- Applies events and produces `STATE_CHANGE` events.
- Persists snapshots through a repository (PostgreSQL).

### EventGraph

- In-memory recent window.
- Builds causal edges automatically from `parent_event_ids`.
- Supports lineage queries: "why did bank-web become compromised?"

### EventRouter

- Pure rules that inspect an event + source node + outgoing edge.
- Decides whether and how to propagate to neighbours.
- Rules are composable and unit-testable.

## Service execution modes

| Mode | Who answers events | Use case |
|------|-------------------|----------|
| **simulated** | Engine emulator | Lightweight services, bulk decoys. |
| **real** | External agent on VM/pod | High-value targets for hands-on training. |
| **decoy** | Engine emulator with fake fingerprint | Honeypots, threat intelligence. |

The engine discovers real services through **heartbeat events** sent by a small
agent installed on each real VM.

## Deployment layers

| Layer | Purpose | Example tools |
|-------|---------|---------------|
| **Management** | Admin access, CI/CD, monitoring | Proxmox host, Terraform, Ansible |
| **Control** | Engine, database, messaging, GitOps | K8s, Redpanda, PostgreSQL, ArgoCD |
| **City / Data** | Real VMs, simulated pods, player workstations | VMs, Multus, Cilium, VyOS |

## Observability

The engine is observable by design:

- **Metrics:** tick duration, queue depth, event throughput, service health.
- **Logs:** structured JSON logs with correlation IDs.
- **Traces:** event lineage through the event graph.
- **Dashboards:** Grafana with city-level and per-service views.

## Security model

- Network segmentation is explicit in the topology graph.
- Public services are reachable only through declared exposure.
- OT segments are isolated.
- Real service agents authenticate to the event bus.
- Public UI is read-only; player actions require authenticated sessions.
- Secrets live in Vault or cloud KMS, never in repositories.

## Scalability targets

| Resource | Home lab | Production sketch |
|----------|----------|---------------------|
| Services | 300 | 1,000+ |
| Events/sec | 100 | 10,000+ |
| Players | 10 | 100+ |
| Real VMs | 6–10 | 50–100 |
| Latency | <1s per tick | <100ms per event |

## Extension points

Adding new behaviour does not require changing the engine core:

- New event type → add handler.
- New propagation rule → add to `EventRouter`.
- New background process → register in tick loop.
- New scenario → scenario manager injects events.
- New organization → add YAML in `cybercity-data`, reload artifact.

## Roadmap to first public demo

1. **Core engine** ✅ — topology, event graph, router, state, API.
2. **Persistence** — PostgreSQL snapshots and audit.
3. **Messaging** — Redpanda integration.
4. **Scenario manager** — first scripted scenario.
5. **UI** — interactive graph, event log, command panel.
6. **Home lab deployment** — Proxmox + K8s.
7. **Public read-only demo** — Cloudflare tunnel.

## Related documents

- [`VISION.md`](VISION.md) — project purpose and principles.
- [`docs/adr/0001-two-graph-architecture.md`](adr/0001-two-graph-architecture.md) — ADR on two graphs.
- [`DATA_FLOW.md`](DATA_FLOW.md) — detailed event flow (to be written).
- [`MODELS.md`](MODELS.md) — data model reference (to be written).
- [`API.md`](API.md) — HTTP/WebSocket protocol (to be written).
- [`DEPLOYMENT.md`](DEPLOYMENT.md) — deployment guide (to be written).
