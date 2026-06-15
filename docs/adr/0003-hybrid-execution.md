# ADR-0003: Гибридное исполнение — real VM + simulated pods

## Status

Accepted

## Context

Киберполигону нужны и реализм, и масштаб. Реальные виртуальные машины дают
игрокам hands-on опыт с настоящими ОС, сервисами и уязвимостями. Но
запускать VM на каждый сервис в городе из 300+ сервисов непрактично в home
lab и расточительно даже в production.

Движок должен поддерживать гибридную модель, где часть сервисов — реальные
workloads, а часть — симулированные движком.

## Decision

У каждого сервиса есть **runtime mode**:

- `real` — внешняя VM или контейнер; отвечает на события через агент.
- `simulated` — лёгкий ответ, генерируемый эмулятором движка.
- `decoy` — симулированный honeypot с deliberate fingerprint.

Mode — это deployment-time concern, не часть канонической city data model.
Топологический граф описывает *что* существует; deployment-конфигурация
решает, *как* каждый сервис исполняется.

## Обнаружение и health

Реальные сервисы запускают небольшого `cybercity-agent`, который:

1. Читает `service_id` и настройки подключения из метаданных (cloud-init,
   env).
2. Шлёт периодические `heartbeat` события в event bus.
3. Слушает события, адресованные `service_id`.
4. Возвращает результаты scan/attack обратно в движок.

Если реальный сервис перестаёт слать heartbeats, движок помечает его как
`down`.

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
- Агенты real-сервисов нужно паковать и распространять.
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
- [`docs/ARCHITECTURE.md`](../ARCHITECTURE.md) — слои развёртывания.
