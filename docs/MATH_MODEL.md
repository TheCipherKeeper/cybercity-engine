# CyberCity Engine — Математическая модель

> Этот документ — черновик математической основы движка. Он фиксирует
> формальные определения, от которых вытекают архитектура, API и реализация.

## 1. Город как взвешенный ориентированный мультиграф

```
G = (V, E, φ)
```

- **V** — множество сервисов (vertices). Элементы: v₁, v₂, …, vₙ.
- **E ⊆ V × V × K** — множество направленных рёбер (links). Каждое ребро
  e = (u, v, k), где u ∈ V — источник, v ∈ V — цель, k ∈ K — тип связи.
- **φ: V → A** — статические атрибуты сервиса.

### 1.1. Статические атрибуты сервиса

Для каждого v ∈ V:

```
φ(v) = (kind, exposure, auth, data_class, criticality, software, ports, is_decoy)
```

где:

| Поле | Домен |
|------|-------|
| kind | {web, api, db, dns, ntp, backup, log, pos, scada, hmi, cctv, identity, ...} |
| exposure | {public, intranet, ot, mgmt} |
| auth | {none, local, sso, mfa, certificate} |
| data_class | {public, internal, confidential, restricted, pii, phi, pci} |
| criticality | {low, medium, high, critical} |
| software | (vendor, product, version, cve_id) |
| ports | список tcp/udp портов |
| is_decoy | {0, 1} |

### 1.2. Атрибуты рёбер

Для каждого e = (u, v, k) ∈ E:

```
ψ(e) = (kind, protocol, encryption, weight, decay, delay)
```

где:

| Поле | Домен | Смысл |
|------|-------|-------|
| kind | K | тип связи |
| protocol | строка или None | транспортный протокол |
| encryption | {none, tls, mtls, ipsec, sso-trust} | защита канала |
| weight | [0, 1] | базовая сила передачи влияния |
| decay | [0, 1] | затухание влияния при проходе |
| delay | ℕ₀ | задержка в тиках |

## 2. Runtime-состояние сервиса

Для каждого v ∈ V в дискретный момент времени t ∈ ℕ₀:

```
S_v(t) = (status_v(t), health_v(t), active_events_v(t), flags_v(t))
```

где:

| Поле | Домен | Смысл |
|------|-------|-------|
| status | {up, degraded, down, maintenance} | дискретный статус |
| health | [0, 1] | уровень работоспособности |
| active_events | 2^I_v | множество подтверждённых impact-событий |
| flags | 2^F | набор scenario-флагов |

Начальное состояние:

```
S_v(0) = (up, 1, ∅, ∅)
```

## 3. Impact-события

### 3.1. Определение

**Impact-событие** — это конкретное значимое событие, которое может случиться
с сервисом. В отличие от абстрактной "компрометации", оно:

- имеет однозначную семантику;
- может быть подтверждено наблюдением, сценарием или инструктором;
- имеет заранее определённые эффекты на сервис и его соседей.

### 3.2. Множество событий per service kind

Для каждого kind сервиса определён свой набор возможных событий:

```
I_v = impact_events(kind(v))
```

### 3.3. Базовый каталог событий

#### Web / API

| Событие | Severity | Кто подтверждает | Эффект |
|---------|----------|------------------|--------|
| `service_down` | 0.9 | агент health check | status=down, health=0 |
| `defaced` | 0.6 | агент/инструктор | reputation flag |
| `malware_served` | 0.7 | агент/сканер | downstream infection risk |
| `credentials_leaked` | 0.8 | сценарий/инструктор | downstream auth events |
| `data_exposed` | 0.9 | инструктор | compliance flag |

#### DB

| Событие | Severity | Кто подтверждает | Эффект |
|---------|----------|------------------|--------|
| `service_down` | 0.9 | агент health check | status=down, dependents degraded |
| `data_encrypted` | 0.95 | сценарий/инструктор | status=degraded, health-=0.4 |
| `data_exfiltrated` | 0.85 | инструктор/логи | compliance flag |
| `backup_failed` | 0.7 | агент health check | recovery blocked |
| `privilege_escalated` | 0.8 | инструктор | admin control |

#### Identity / SSO

| Событие | Severity | Кто подтверждает | Эффект |
|---------|----------|------------------|--------|
| `service_down` | 0.9 | агент health check | all auth-dependent degraded |
| `tokens_compromised` | 0.95 | сценарий/инструктор | downstream takeover |
| `admin_access_gained` | 0.9 | инструктор | domain-wide events |
| `brute_force_detected` | 0.5 | агент/логи | alert, potential lockout |

#### OT / SCADA / HMI

| Событие | Severity | Кто подтверждает | Эффект |
|---------|----------|------------------|--------|
| `service_down` | 0.9 | агент health check | physical process affected |
| `control_lost` | 0.95 | сценарий/инструктор | status=degraded, safety flag |
| `sensor_spoofed` | 0.7 | агент/инструктор | wrong downstream decisions |
| `ot_isolation_broken` | 0.85 | агент/network monitor | high severity alert |

#### DNS / NTP

| Событие | Severity | Кто подтверждает | Эффект |
|---------|----------|------------------|--------|
| `service_down` | 0.8 | агент health check | resolution failures |
| `dns_hijacked` | 0.85 | агент/резолвинг | redirection, phishing |
| `ntp_manipulated` | 0.75 | агент/time check | auth/cert failures |

#### Backup / Log

| Событие | Severity | Кто подтверждает | Эффект |
|---------|----------|------------------|--------|
| `service_down` | 0.7 | агент health check | backup unavailable |
| `logs_tampered` | 0.8 | агент/инструктор | audit loss |
| `backup_encrypted` | 0.9 | сценарий/инструктор | recovery impossible |

#### POS

| Событие | Severity | Кто подтверждает | Эффект |
|---------|----------|------------------|--------|
| `service_down` | 0.8 | агент health check | payments blocked |
| `card_data_stolen` | 0.95 | инструктор | PCI flag, financial loss |
| `terminal_compromised` | 0.85 | агент/инструктор | payment fraud |

### 3.4. Структура impact-события

```
i = (id, kind, severity, confirmed_by, effects)
```

где:

| Поле | Смысл |
|------|-------|
| id | Уникальный идентификатор события на сервисе |
| kind | Тип события, например `data_encrypted` |
| severity | [0, 1] — степень влияния |
| confirmed_by | {agent, scenario, instructor, engine} |
| effects | Список эффектов на self и соседей |

## 4. Эффекты события

### 4.1. Структура эффекта

```
eff = (target, field, operation, value, condition)
```

где:

| Поле | Домен | Смысл |
|------|-------|-------|
| target | {self, neighbor} | на кого действует |
| field | {status, health, flag} | что меняется |
| operation | {set, add, subtract, multiply} | как меняется |
| value | depends on field | величина изменения |
| condition | optional | условие на рёбра или статус |

### 4.2. Пример эффектов

```
event: data_encrypted on bank-db
severity: 0.95
effects:
  - target: self
    field: status
    operation: set
    value: degraded

  - target: self
    field: health
    operation: subtract
    value: 0.4

  - target: neighbor
    condition: edge_kind == db-read
    field: status
    operation: set
    value: degraded

  - target: neighbor
    condition: edge_kind == db-read
    field: health
    operation: subtract
    value: 0.2
```

### 4.3. Применение эффектов

```
apply(eff, v, t) =
  if eff.target == self:
    S_v(t+1) = mutate(S_v(t), eff.field, eff.operation, eff.value)
  if eff.target == neighbor:
    for e=(v,u,k) in out_edges(v):
      if eff.condition is None or satisfies(k, eff.condition):
        S_u(t+delay(e)) = mutate(S_u(...), eff.field, eff.operation, eff.value * weight(e) * decay(e))
```

## 5. Подтверждение событий

### 5.1. Источники подтверждения

| Источник | Что может подтвердить | Как |
|----------|----------------------|-----|
| **Agent** | Простые факты: service_down, file_changed, process_detected, dns_hijacked | Агент на real VM наблюдает и шлёт событие |
| **Scenario** | Заранее определённые события в сценарии | Scenario manager инжектирует по таймеру/условию |
| **Instructor** | Достижение игрока | Человек в UI нажимает "подтвердить" |
| **Engine emulator** | События для simulated/decoy сервисов | Движок вычисляет по правилам |

### 5.2. Протокол агента

Агент на real VM отправляет факты:

```json
{
  "event_type": "IMPACT_EVENT_CONFIRM",
  "target_id": "bank-db",
  "source_type": "service",
  "source_id": "bank-db-agent",
  "payload": {
    "kind": "service_down",
    "facts": [
      {"type": "process", "name": "mysqld", "running": false},
      {"type": "port", "port": "tcp/3306", "open": false}
    ]
  }
}
```

Движок по `kind` находит impact-событие и применяет его эффекты.

## 6. Защищённость и сила атаки

### 6.1. Защита сервиса

Защита влияет на то, как easily simulated/сценарий-события могут произойти:

```
defense(v) = α · auth_score(auth(v))
           + β · exposure_score(exposure(v))
           + γ · software_score(software(v))
           + δ · criticality_bonus(criticality(v))
```

где α + β + γ + δ = 1.

### 6.2. Сила атаки

```
attack_strength(a, v, t) = base(vector(a))
                         × tool_level(a)
                         × reach(source(a), v, t)
                         × fatigue_penalty(attacker_id(a), t)
```

### 6.3. Условие успеха для simulated-сервисов

Для simulated сервисов движок решает, произошло ли событие:

```
p_success(a, v, kind) = σ(λ · (attack_strength(a, v, t) - defense(v) + severity_bonus(kind)))
```

где `severity_bonus(kind)` — насколько серьёзное событие легче/труднее вызвать.

## 7. Динамика состояния

### 7.1. Переход статуса

```
if health_v(t) < θ_down:
    status_v(t) = down
elif health_v(t) < θ_degraded:
    status_v(t) = degraded
elif active_events_v(t) contains any event with severity ≥ 0.8:
    status_v(t) = degraded
elif active_events_v(t) contains any event with severity ≥ 0.95:
    status_v(t) = down
else:
    status_v(t) = up
```

| Порог | Значение по умолчанию |
|-------|-----------------------|
| θ_down | 0.10 |
| θ_degraded | 0.50 |

### 7.2. Восстановление

Восстановление удаляет события поэтапно:

```
REBOOT:       remove service_down
PATCH:        remove malware_served, defaced
RESTORE_DATA: remove data_encrypted, data_exfiltrated, backup_encrypted
REVOKE_CREDS: remove credentials_leaked, tokens_compromised, admin_access_gained
ISOLATE:      stop propagation, status = maintenance
REBUILD:      remove all events, status = up, health = 1.0
```

## 8. Распространение по графу

Когда сервис v получает impact-событие i, его эффекты на соседей
распространяются по исходящим рёбрам:

```
for eff in i.effects:
  if eff.target == neighbor:
    for e=(v,u,k) in out_edges(v):
      if satisfies(k, eff.condition):
        S_u(t + delay(e)) += eff scaled by weight(e) · decay(e)
```

## 9. Decoy-специфика

Decoy-сервисы всегда обрабатываются эмулятором движка. Они:

- легко отвечают на scan/attack;
- генерируют `noise_alert` события при любом контакте;
- не дают реальных downstream-эффектов;
- могут имитировать события `defaced`, `service_down` и т.п., но эти события
  не влияют на соседей.

## 10. Глобальные метрики

### 10.1. Устойчивость города

```
city_resilience(t) = Σ_v criticality_value(v) · health_v(t)
                     / Σ_v criticality_value(v)
```

### 10.2. Attack surface

```
attack_surface(t) = |{v ∈ V : exposure(v) = public ∧ status_v(t) ≠ down}|
```

### 10.3. Число инцидентов

```
incident_count(t) = |{v ∈ V : status_v(t) ∈ {degraded, down}}|
```

### 10.4. Очки игрока

```
score(attacker, t) = Σ_{v ∈ V} Σ_{i ∈ active_events_v(t)} severity(i) · criticality_value(v)
                     · 1{attacker triggered i}
```

## 11. События как дискретные изменения

```
e = (t, event_id, parent_ids, correlation_id, source_type, source_id,
     event_type, target_id, payload, ΔS)
```

где ΔS включает добавление/удаление impact-событий и изменение health.

## 12. Параметры баланса

| Параметр | Значение по умолчанию | Описание |
|----------|----------------------|----------|
| α | 0.40 | вес auth в защите |
| β | 0.25 | вес exposure в защите |
| γ | 0.20 | вес software в защите |
| δ | 0.15 | вес criticality bonus |
| λ | 5.0 | крутизна сигмоида |
| θ_down | 0.10 | порог статуса down |
| θ_degraded | 0.50 | порог статуса degraded |

## 13. Связь с архитектурой

| Математика | Компонент |
|------------|-----------|
| G = (V, E, φ) | TopologyGraph |
| S_v(t) | ServiceState |
| I_v — impact events | ImpactEventCatalog |
| i = (kind, severity, effects) | ImpactEventDefinition |
| confirmed_by | ConfirmationSource |
| apply(eff, v) | StateManager |
| propagation по рёбрам | EventRouter / PropagationModel |
| defense/attack_strength | DefenseCalculator / AttackCalculator |
| simulated events | Engine Emulator |
| city_resilience | MetricsEngine |
| score | ScoringEngine |

## 14. Открытые вопросы

1. Формат конфига impact-событий: YAML рядом с `cybercity-data` или JSON в
   `engine.zip`?
2. Как инструктор подтверждает события: отдельный UI или API?
3. Нужен ли автоматический scorer для перевода agent facts в impact-события?
4. Как decoy-события влияют на scoring?
5. Как учитывать encryption рёбер при propagation?
