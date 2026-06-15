# CyberCity Engine — Data Models

This document describes the data models used by the engine, moving from static
blueprint to dynamic runtime to events.

## Topology graph

Loaded from `cybercity-data` artifacts. Immutable during a simulation.

### TopologyNode

Represents a service in the city.

| Field | Type | Description |
|-------|------|-------------|
| `id` | `str` | Kebab-case unique identifier. |
| `org_id` | `str` | Owning organization. |
| `name` | `str` | Human-readable name. |
| `kind` | `str` | Service kind: web, api, db, dns, scada, etc. |
| `exposure` | `Literal["public","intranet","ot","mgmt"]` | Network exposure level. |
| `host` | `str` | FQDN for DNS. |
| `network_id` | `str \| None` | Logical network placement. |
| `bind_ip` | `str \| None` | Allocated IP address. |
| `auth` | `str` | Authentication method. |
| `data_classification` | `str` | Sensitivity label. |
| `criticality` | `Literal["critical","high","medium","low"]` | Business impact. |
| `ports` | `list[str]` | Exposed ports, e.g. `tcp/443`. |
| `is_decoy` | `bool` | Whether this is a decoy service. |
| `decoy_kind` | `str \| None` | Decoy fingerprint type. |
| `software` | `dict[str, Any]` | Vendor/product/version/CVE. |
| `os_hint` | `str \| None` | Operating system hint. |

### TopologyEdge

Represents a relationship between two services.

| Field | Type | Description |
|-------|------|-------------|
| `source` | `str` | Source service id. |
| `target` | `str` | Target service id. |
| `kind` | `str` | Link kind: api-call, auth, db-read, db-write, etc. |
| `protocol` | `str \| None` | e.g. `tcp/443`. |
| `encryption` | `str \| None` | e.g. `tls`, `mtls`. |
| `inferred` | `bool` | True if not from `links` but derived. |

### TopologyGraph

Container for nodes and edges.

| Field | Type | Description |
|-------|------|-------------|
| `schema_version` | `str` | Version of the data schema. |
| `source_version` | `str` | Version of the city artifact. |
| `services` | `dict[str, TopologyNode]` | Service nodes. |
| `edges` | `list[TopologyEdge]` | All edges. |

## Runtime state

Mutable state owned by `StateManager`.

### ServiceState

| Field | Type | Description |
|-------|------|-------------|
| `service_id` | `str` | Reference to topology node. |
| `status` | `ServiceStatus` | up, down, compromised, maintenance. |
| `health` | `float` | 0.0–1.0 health indicator. |
| `compromise_vector` | `str \| None` | How the service was compromised. |
| `last_heartbeat` | `datetime \| None` | Last real service heartbeat. |
| `seen_by` | `list[str]` | Observers (scanners). |
| `flags` | `dict[str, Any]` | Scenario-specific flags. |
| `variables` | `dict[str, Any]` | Process-local variables. |

### PlayerState

| Field | Type | Description |
|-------|------|-------------|
| `player_id` | `str` | Unique player id. |
| `name` | `str` | Display name. |
| `org_id` | `str \| None` | Starting org assignment. |
| `score` | `int` | Current score. |
| `flags` | `list[str]` | Captured flags. |
| `status` | `Literal["active","idle","banned"]` | Player status. |

### ScenarioState

| Field | Type | Description |
|-------|------|-------------|
| `scenario_id` | `str` | Scenario identifier. |
| `name` | `str` | Display name. |
| `status` | `Literal["running","paused","stopped"]` | Lifecycle. |
| `started_at` | `datetime` | Start time. |
| `ended_at` | `datetime \| None` | End time. |
| `config` | `dict[str, Any]` | Scenario parameters. |

### WorldState

Complete snapshot of the runtime world.

| Field | Type | Description |
|-------|------|-------------|
| `tick` | `int` | Current simulation tick. |
| `started_at` | `datetime` | Engine start time. |
| `services` | `dict[str, ServiceState]` | Per-service runtime state. |
| `players` | `dict[str, PlayerState]` | Players. |
| `active_scenario` | `ScenarioState \| None` | Running scenario. |
| `variables` | `dict[str, Any]` | Global variables. |

## Event graph

Append-only causal graph.

### EventNode

| Field | Type | Description |
|-------|------|-------------|
| `event_id` | `str` | UUID of the event. |
| `parent_event_ids` | `list[str]` | Immediate causal parents. |
| `correlation_id` | `str` | Scenario/incident grouping id. |
| `tick` | `int` | Tick at which event was generated. |
| `timestamp` | `datetime` | Wall-clock timestamp. |
| `source_type` | `Literal[...]` | engine, service, scenario, player, system, background. |
| `source_id` | `str` | Id of the source. |
| `event_type` | `EventType` | Type of event. |
| `target_id` | `str \| None` | Target topology node id. |
| `payload` | `dict[str, Any]` | Event-specific data. |
| `status` | `Literal["pending","processed","failed","suppressed"]` | Processing status. |

### EventType

| Value | Meaning |
|-------|---------|
| `HEARTBEAT` | Real service liveness ping. |
| `SCAN` | Network/service scan. |
| `ATTACK` | Offensive action against a service. |
| `COMPROMISE` | Service confirmed compromised. |
| `RESTORE` | Restoration action. |
| `STATE_CHANGE` | Runtime state changed. |
| `COMMAND` | Player/instructor command. |
| `SCENARIO_START` | Scenario begins. |
| `SCENARIO_STOP` | Scenario ends. |
| `FLAG_CAPTURED` | Player achieved objective. |
| `BACKGROUND_EFFECT` | Emitted by background processes. |
| `PROPAGATION` | Event propagated through topology. |

### EventEdge

| Field | Type | Description |
|-------|------|-------------|
| `source_event_id` | `str` | Parent event. |
| `target_event_id` | `str` | Child event. |
| `kind` | `Literal["caused_by","propagated_to","triggered_rule","response_to"]` | Relationship. |

## Configuration

### EngineConfig

| Field | Default | Description |
|-------|---------|-------------|
| `app_name` | `cybercity-engine` | Application name. |
| `debug` | `False` | Debug mode. |
| `tick_ms` | `1000` | Tick interval. |
| `engine_zip_url` | local MinIO | Source of topology artifact. |
| `kafka_bootstrap_servers` | `localhost:9092` | Redpanda/Kafka address. |
| `database_url` | local PostgreSQL | Snapshot/audit DB. |
| `snapshot_interval_ticks` | `10` | Snapshot frequency. |
| `host` / `port` | `0.0.0.0:8000` | API bind. |

## Serialization

- Models use Pydantic v2.
- API returns JSON via `model_dump(mode="json")`.
- PostgreSQL stores snapshots and events as JSONB.
- Redpanda messages are JSON by default; Avro may be added later.

## Related

- `docs/ARCHITECTURE.md` — how models fit together.
- `docs/DATA_FLOW.md` — how events move through models.
