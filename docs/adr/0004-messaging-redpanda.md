# ADR-0004: Redpanda as event bus

## Status

Accepted

## Context

The engine needs a messaging layer to connect:

- engine (event processor);
- UI (real-time updates and player commands);
- scenario manager (scenario lifecycle);
- real service agents (heartbeats, responses);
- audit and replay consumers.

Kafka is the natural choice for event streaming, but operating a ZooKeeper or
KRaft cluster in a home lab adds complexity that does not add value at this
stage.

## Decision

Use **Redpanda** as the primary event bus.

Redpanda is Kafka-compatible, requires no ZooKeeper, and runs as a single
binary or container. It supports the same topics, producers, consumers, and
APIs that a future Kafka deployment would use.

## Deployment

For local development and home lab:

```yaml
# compose.yaml
redpanda:
  image: docker.redpanda.com/redpandadata/redpanda:v24.1.1
  command:
    - redpanda start
    - --smp 1
    - --memory 1G
    - --overprovisioned
    - --node-id 0
    - --kafka-addr PLAINTEXT://0.0.0.0:9092
    - --advertise-kafka-addr redpanda:9092
```

Production may later switch to a multi-node Redpanda cluster or to Apache
Kafka without changing the application code.

## Topics

| Topic | Purpose | Retention |
|-------|---------|-----------|
| `city.commands` | Player and instructor commands | 7 days |
| `city.events` | All runtime events | 14 days |
| `city.state.changes` | State changes only | 30 days |
| `city.service.heartbeat` | Real service heartbeats | 3 days |
| `city.audit` | Immutable audit stream | 90 days |

## Client library

Use `aiokafka` for async producer/consumer. It is compatible with Redpanda
and will remain compatible if we migrate to Kafka.

## Consequences

### Positive

- Simpler operations than Kafka in small deployments.
- Kafka-compatible API, easy future migration.
- Good enough throughput and latency for cyber-range scale.
- Single-node mode fits home lab.

### Negative

- Less mature ecosystem than Kafka.
- Some advanced Kafka features not needed anyway.
- Community sometimes skeptical of non-Kafka choice.

## Alternatives considered

- **Apache Kafka**: more operational overhead, same API.
- **NATS JetStream**: lighter, but not Kafka-compatible.
- **Redis Streams**: simpler, but less durable and no replay segmentation.
- **RabbitMQ**: not a log-based event stream.

## Related

- ADR-0002: event-driven runtime.
- `compose.yaml` local setup.
