# CyberCity Engine — Архитектура

## TL;DR

`cybercity-engine` — это **событийный runtime** для цифрового двойника
CyberCity. Он загружает статический топологический граф из `cybercity-data`,
поддерживает динамическое runtime-состояние и обрабатывает поток событий
через graph-aware router. Всё, что изменяется в городе, происходит через
событие; каждое событие связано с причинами, формируя causal graph.

## Системный контекст

```text
┌─────────────────────────────────────────────────────────────────────┐
│                         Внешние пользователи                           │
│   Игроки │ Инструкторы │ Read-only посетители │ Авторы сценариев    │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Платформа CyberCity                         │
│                                                                      │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────────┐  │
│  │     UI      │    │   Engine    │    │    Scenario Manager     │  │
│  │  (React/    │◄──►│    (Go)     │◄──►│      (Python)           │  │
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
│                    Инфраструктурный слой                           │
│              Proxmox + Kubernetes + Cilium + Multus                │
└─────────────────────────────────────────────────────────────────────┘
```

## Основные ответственности

| Компонент | Ответственность |
|-----------|-----------------|
| **cybercity-data** | Декларативная модель города, валидация, генерация артефактов. |
| **cybercity-engine** | Runtime-состояние, обработка событий, propagation, снапшоты. |
| **cybercity-ui** | Визуализация, ввод игрока, real-time обновления. |
| **Scenario Manager** | Старт/пауза/стоп сценариев, инжекция событий, подсчёт очков. |
| **Redpanda** | Event bus между движком, UI, scenario manager, реальными сервисами. |
| **PostgreSQL** | Снапшоты и audit log событийного графа. |
| **MinIO / S3** | Артефакты `engine.zip` и replay-дампы. |

## Два графа

### 1. Топологический граф

Загружается из артефактов `cybercity-data`.

```text
Узлы: сервисы
  id, org_id, kind, exposure, host, network_id, bind_ip
  auth, data_classification, criticality, ports, software

Рёбра: связи
  декларированные: api-call, auth, db-read, db-write, backup-of,
                   log-sink, trusts, vendor-vpn, dns-query, ntp-query

  inferred: same_network, same_org, exposure_chain
```

Топологический граф **иммутабелен во время симуляции**. Заменяется только при
загрузке нового артефакта города.

### 2. Событийный граф

Строится во время работы.

```text
Узлы: события
  event_id, parent_event_ids, correlation_id
  tick, timestamp, source_type, source_id
  event_type, target_id, payload, status

Рёбра:
  caused_by       ─ событие B вызвано событием A
  propagated_to   ─ событие B дошло до соседа благодаря событию A
  triggered_rule  ─ событие B создано правилом propagation
  response_to     ─ событие B — намеренный ответ на событие A
```

Событийный граф **append-only**. События никогда не удаляются; могут только
суммироваться или уходить в cold storage.

### Связь между графами

```text
Топология              Событие
   │                    │
   │◄── target_id ─────│  "это событие случилось с bank-web"
   │                    │
   │── neighbors() ───►│  "куда это событие может пойти дальше?"
   │                    │
   │◄── state change ───┤  "bank-web теперь compromised"
```

## Поток событий

```text
1. Источник производит событие
      player scan → bank-web

2. Движок получает событие через queue или Redpanda

3. Событие добавляется в событийный граф

4. Handler обновляет runtime-состояние
      bank-web.seen_by += player-1

5. Router решает propagation
      log-sink edge + noisy scan → alert event

6. Дочерние события ставятся в очередь и обрабатываются

7. Изменения состояния эмитят STATE_CHANGE события
      bank-web.status: up → compromised

8. Изменения состояния могут снова распространяться
      compromised bank-web влияет на bank-db через db-read

9. Снапшот + broadcast в UI
```

## Внутреннее устройство движка (Onion / Ports-and-Adapters)

```text
         Adapters (infra)
┌─────────────────────────────────────────┐
│  HTTP/WS (api) │ Kafka (bus) │ PostgreSQL│
│  TopologyLoader│ ServiceAgent│ Snapshot  │
└─────────────────────────────────────────┘
                   │
                   ▼
         Application (wiring)
┌─────────────────────────────────────────┐
│  NewRuntime: config → ports → engine    │
└─────────────────────────────────────────┘
                   │
                   ▼
         Domain (core) — pure logic
┌─────────────────────────────────────────┐
│              Engine                     │
│                                          │
│  ┌─────────────┐    ┌─────────────────┐ │
│  │ Event       │◄──►│ Event Processor │ │
│  │ Processor   │    │                 │ │
│  └──────┬──────┘    └────────┬────────┘ │
│         │                     │           │
│         ▼                     ▼           │
│  ┌─────────────────────────────────────┐ │
│  │          StateManager               │ │
│  │   services, players, scenario      │ │
│  └─────────────────────────────────────┘ │
│         │                     │           │
│         ▼                     ▼           │
│  ┌─────────────┐    ┌─────────────────┐  │
│  │ EventStore  │    │  EventRouter    │  │
│  │  (port)     │    │  (port/impl)    │  │
│  └─────────────┘    └─────────────────┘  │
│                                          │
└─────────────────────────────────────────┘
```

### Domain

- **Чистая логика:** models, StateManager, EventRouter, Engine.
- **Нет зависимостей от HTTP, Kafka, PostgreSQL, env-переменных.**
- Все внешние действия проходят через интерфейсы-порты.

### Ports (interfaces)

- `EventStore` — хранение событийного графа.
- `SnapshotRepository` — снапшоты WorldState.
- `MessageBus` — pub/sub событий.
- `ServiceAgent` — управление real/decoy сервисами.
- `TopologyLoader` — загрузка topology-артефактов.
- `StateBroadcaster` — рассылка состояния подписчикам.

### Adapters

- `adapters/api` — HTTP + WebSocket на `net/http` + `gorilla/websocket`.
- `adapters/loader` — парсинг `engine.zip` / `engine.json`.
- `adapters/memory` — in-memory реализации всех портов для тестов и демо.
- Будущие: `adapters/postgres`, `adapters/redpanda`, `adapters/grpc-agent`.

### Application

- `application.NewRuntime(cfg)` — composition root. Создаёт адаптеры,
  маппит config в domain-config и собирает engine.

### StateManager

- Единственный владелец изменяемого `WorldState`.
- Применяет события и производит `STATE_CHANGE` события.
- Сохраняет снапшоты через репозиторий (PostgreSQL).

### EventGraph

- In-memory окно недавних событий.
- Автоматически строит causal-рёбра из `parent_event_ids`.
- Поддерживает lineage-запросы: «почему bank-web стал compromised?»

### EventRouter

- Чистые правила, которые анализируют событие + исходный узел +
  исходящее ребро.
- Решают, стоит ли и как распространять к соседям.
- Правила композитные и unit-testable.

## Режимы исполнения сервисов

| Режим | Кто отвечает на события | Когда используется |
|-------|-------------------------|--------------------|
| **simulated** | Эмулятор движка | Лёгкие сервисы, массовые decoys. |
| **real** | Внешний агент на VM/pod | High-value target для hands-on. |
| **decoy** | Эмулятор с fake fingerprint | Honeypots, threat intelligence. |

Движок обнаруживает real-сервисы через **heartbeat-события**, отправляемые
небольшим агентом на каждой real VM.

## Слои развёртывания

| Слой | Назначение | Примеры инструментов |
|------|------------|----------------------|
| **Management** | Админский доступ, CI/CD, мониторинг | Proxmox host, Terraform, Ansible |
| **Control** | Движок, БД, messaging, GitOps | K8s, Redpanda, PostgreSQL, ArgoCD |
| **City / Data** | Real VMs, simulated pods, player workstations | VMs, Multus, Cilium, VyOS |

## Observability

Движок наблюдаем по дизайну:

- **Метрики:** tick duration, queue depth, event throughput, health сервисов.
- **Логи:** structured JSON logs с correlation IDs.
- **Трейсы:** lineage событий через событийный граф.
- **Дашборды:** Grafana с city-level и per-service видами.

## Модель безопасности

- Сетевая сегментация явно задана в топологическом графе.
- Публичные сервисы достижимы только через declared exposure.
- OT-сегменты изолированы.
- Агенты real-сервисов аутентифицируются к event bus.
- Публичный UI read-only; действия игрока требуют аутентифицированной сессии.
- Секреты живут в Vault или cloud KMS, никогда в репозиториях.

## Целевые показатели масштабируемости

| Ресурс | Home lab | Production-набросок |
|--------|----------|---------------------|
| Сервисы | 300 | 1,000+ |
| Событий/сек | 100 | 10,000+ |
| Игроки | 10 | 100+ |
| Real VMs | 6–10 | 50–100 |
| Latency | <1s на tick | <100ms на событие |

## Точки расширения

Добавление нового поведения не требует изменения ядра движка:

- Новый тип события → добавить handler.
- Новое propagation-правило → добавить в `EventRouter`.
- Новый background-процесс → зарегистрировать в tick-цикле.
- Новый сценарий → scenario manager инжектирует события.
- Новая организация → добавить YAML в `cybercity-data`, перезагрузить артефакт.

## Дорожная карта к первой публичной демонстрации

1. **Core engine** ✅ — topology, event graph, router, state, API.
2. **Persistence** — PostgreSQL-снапшоты и audit.
3. **Messaging** — интеграция с Redpanda.
4. **Scenario manager** — первый скриптованный сценарий.
5. **UI** — интерактивный граф, event log, панель команд.
6. **Home lab deployment** — Proxmox + K8s.
7. **Public read-only demo** — Cloudflare tunnel.

## Связанные документы

- [`VISION.md`](VISION.md) — цель проекта и принципы.
- [`docs/adr/0001-two-graph-architecture.md`](adr/0001-two-graph-architecture.md) — ADR о двух графах.
- [`DATA_FLOW.md`](DATA_FLOW.md) — детальный поток событий.
- [`MODELS.md`](MODELS.md) — справочник моделей.
- [`API.md`](API.md) — протокол HTTP/WebSocket.
- [`DEPLOYMENT.md`](DEPLOYMENT.md) — руководство по развёртыванию.
