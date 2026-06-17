# ADR-0003: Гибридное исполнение — real VM + simulated pods

## Status

Superseded — см.
[`cybercity/adr/0004-runtime-kind-vm-container-lite.md`](https://github.com/TheCipherKeeper/cybercity/blob/main/adr/0004-runtime-kind-vm-container-lite.md).
Режимы `{real, simulated, decoy}` заменены на `runtime_kind {vm, container, lite}`
(deployment-time) + флаг назначения `honeypot` (бывший `decoy`); движок —
регистратор, не симулятор (класса «engine-synthesized service events» больше нет).
Тело ADR сохранено как история.

## Context

Киберполигону нужны и реализм, и масштаб. Реальные виртуальные машины дают
игрокам hands-on опыт с настоящими ОС, сервисами и уязвимостями. Но
запускать VM на каждый сервис в городе из 300+ сервисов непрактично в home
lab и расточительно даже в production.

Движок должен поддерживать гибридную модель, где часть сервисов — реальные
workloads, а часть — симулированные движком.

## Decision

У каждого сервиса есть **runtime mode**:

- `real` — внешняя VM или контейнер; отвечает на события через наблюдатель.
- `simulated` — лёгкий ответ, генерируемый эмулятором движка.
- `decoy` — симулированный honeypot с deliberate fingerprint.

Mode — это deployment-time concern, не часть канонической city data model.
Топологический граф описывает *что* существует; deployment-конфигурация
решает, *как* каждый сервис исполняется.

## Обнаружение и health

Реальные сервисы наблюдаются **out-of-band** через `cybercity-collector`
(на гипервизоре/K8s-узле, read-only), а не in-guest агентом. Это часть
доверительной границы — см.
[`cybercity/adr/0003-collector-rust-out-of-band.md`](https://github.com/TheCipherKeeper/cybercity/blob/main/adr/0003-collector-rust-out-of-band.md)
и
[`cybercity/adr/0002-trust-boundary.md`](https://github.com/TheCipherKeeper/cybercity/blob/main/adr/0002-trust-boundary.md).

1. `service_id` и настройки берутся из топологии/метаданных.
2. Коллектор шлёт подписанные (Ed25519) `heartbeat`/telemetry в event bus.
3. Движок слушает события, адресованные `service_id`.
4. Эффекты scan/attack на real-сервисе наблюдаются коллектором и идут в движок
   как доверенный поток.

Если реальный сервис перестаёт наблюдаться, движок помечает его как `down`.
In-guest телеметрия (если используется) — best-effort, не источник для scoring.

## Simulated-сервисы

Эмулятор движка генерирует правдоподобные ответы на основе:

- `kind` сервиса (web, db, dns и т.д.);
- `ports` и `software` из топологии;
- `decoy` fingerprint, если присутствует;
- текущего runtime-состояния (up, down, compromised).

Simulated-сервисы дешёвы и масштабируются горизонтально в Kubernetes.

## Decoys

Decoys всегда simulated. Они существуют, чтобы привлекать сканирования и
атаки, собирать threat intelligence и замедлять атакующих. Эмулятор может
генерировать разные персоны: default creds, known CVEs, realistic banners.

## Consequences

### Positive

- Реализм там, где это важно; масштаб там, где не важен.
- Home lab становится feasible с горсткой VM.
- Учебные сценарии могут смешивать real exploitation с simulated
  city-эффектами.
- Легко апгрейдить сервис с simulated до real без изменения топологии.

### Negative

- Два execution path, которые нужно поддерживать.
- Коллектор нужно размещать на каждом хосте с real-сервисами (через
  `cybercity-manage`).
- Сеть должна одинаково достигать и VM, и pods.

## Пример mapping

```yaml
# deployment/home-lab/service-mapping.yaml
runtime_mode:
  bank-web: real
  bank-db: real
  bank-log: simulated
  hospital-pacs: real
  power-scada: real
  decoy-printer-01: decoy
  # всё остальное по умолчанию simulated
```

## Alternatives considered

- **Всё real**: слишком дорого, слишком много maintenance.
- **Всё simulated**: не хватает hands-on ценности для игроков.
- **Real per organization**: проще, но всё ещё расточительно для маленьких org.

## Related

- ADR-0001: топологический граф не кодирует runtime mode.
- [`cybercity/adr/0003-collector-rust-out-of-band.md`](https://github.com/TheCipherKeeper/cybercity/blob/main/adr/0003-collector-rust-out-of-band.md) — коллектор out-of-band.
- [`cybercity/adr/0002-trust-boundary.md`](https://github.com/TheCipherKeeper/cybercity/blob/main/adr/0002-trust-boundary.md) — доверительная граница.
- [`docs/ARCHITECTURE.md`](../ARCHITECTURE.md) — режимы исполнения.