# ADR-0004: Redpanda как event bus

## Status

Accepted

## Context

Движку нужен messaging-слой для соединения:

- engine (event processor);
- UI (real-time updates и команды игрока);
- scenario manager (жизненный цикл сценариев);
- агентов real-сервисов (heartbeats, ответы);
- audit и replay consumers.

Kafka — естественный выбор для event streaming, но эксплуатация
ZooKeeper или KRaft-кластера в home lab добавляет сложности, которые на
этом этапе не дают ценности.

## Decision

Использовать **Redpanda** как основной event bus.

Redpanda совместима с Kafka API, не требует ZooKeeper и работает как один
бинарник или контейнер. Она поддерживает те же topics, producers, consumers и
API, которые будут использоваться в будущем Kafka deployment.

## Развёртывание

Для локальной разработки и home lab:

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

В production можно позже переключиться на multi-node Redpanda-кластер или на
Apache Kafka без изменения кода приложения.

## Topics

| Topic | Назначение | Retention |
|-------|-----------|-----------|
| `city.commands` | Команды игрока и инструктора | 7 days |
| `city.events` | Все runtime-события | 14 days |
| `city.state.changes` | Только изменения состояния | 30 days |
| `city.service.heartbeat` | Heartbeats real-сервисов | 3 days |
| `city.audit` | Immutable audit stream | 90 days |

## Client library

Использовать `aiokafka` для async producer/consumer. Она совместима с
Redpanda и останется совместимой при миграции на Kafka.

## Consequences

### Positive

- Проще в эксплуатации, чем Kafka в малых деплоях.
- Kafka-compatible API, лёгкая будущая миграция.
- Достаточная throughput и latency для масштаба cyber range.
- Single-node режим подходит для home lab.

### Negative

- Менее зрелая экосистема, чем у Kafka.
- Некоторые продвинутые Kafka-фичи не нужны и так.
- Сообщество иногда скептично к non-Kafka выбору.

## Alternatives considered

- **Apache Kafka**: больше operational overhead, тот же API.
- **NATS JetStream**: легче, но не Kafka-compatible.
- **Redis Streams**: проще, но менее durable и нет replay-сегментации.
- **RabbitMQ**: не log-based event stream.

## Related

- ADR-0002: событийный runtime.
- [`compose.yaml`](../../compose.yaml) — локальная поднималка.
