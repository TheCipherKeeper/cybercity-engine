# ADR — cybercity-engine

Локальные архитектурные решения движка. Сквозные решения (затрагивающие
несколько репозиториев) — в
[`cybercity/adr/`](https://github.com/TheCipherKeeper/cybercity/blob/main/adr/).

| № | Решение | Статус |
|---|---------|--------|
| [0001](0001-two-graph-architecture.md) | Два графа: топологический + событийный | Accepted |
| [0002](0002-event-driven-runtime.md) | Event-driven runtime с одним tick-loop | Accepted |
| [0003](0003-hybrid-execution.md) | Гибридное исполнение: real VM + simulated + decoy | Accepted |
| [0004](0004-messaging-redpanda.md) | Redpanda как event bus | Amended (см. ADR-0006) |
| [0005](0005-python-first-go-later.md) | Python-first (отменён) | Superseded (см. ADR-0006) |
| [0006](0006-go-first.md) | Go-first реализация | Accepted |

Формат ADR — в
[`cybercity/CONVENTIONS.md`](https://github.com/TheCipherKeeper/cybercity/blob/main/CONVENTIONS.md).