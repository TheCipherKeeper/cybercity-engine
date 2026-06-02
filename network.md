# City Network — Organizations, Services & Decoys

Декларативный реестр организаций, сервисов и decoy-хостов CyberCity. Источник
правды для `network.yml` (ADR-0007) и проекция K8s-манифестов.

## Обозначения

### Organizations & services

- `[*]` — **эталонная организация**, описывается руками как образец для LLM.
- `kind` сервиса — фиксированный словарь (ADR-0002):
  `web, api, pos, identity, db, file-share, rmm, vpn, ot, cctv, mail, dns, ntp, backup, log`.
- `exposure` — где виден сервис:
  - `public` — доступен из интернета;
  - `intranet` — только из corp/mgmt сегментов своего сценария;
  - `ot` — изолированный OT-сегмент;
  - `mgmt` — служебный сегмент (MSP, бэкапы, мониторинг).

### Decoy-хосты

- `subnet` — сегмент (`corp` | `ot` | `mgmt` | `public`); decoy всегда попадает
  в один сегмент с реальным сервисом того же типа.
- `os` — отпечаток ОС, отдаваемый при сканировании (`windows-10`, `linux-3.x`,
  `cisco-ios`, `printer`, `iot-camera`, и т.п.).
- `role` — функциональная роль (`workstation` | `printer` | `iot` | `server` |
  `router` | `camera` | `phone`).
- `ports[]` — открытые порты с баннерами; то, что увидит `nmap -sV`.

Decoy-хост — **не участник учения** (ADR-0003 → ADR-0007): не генерирует
событий, не учитывается в scoring, не является целью red team. Это декорация
масштаба, чтобы `nmap /24` выглядел реалистично.

---

## Organizations

### Government

#### `[*] city-hall` — City Hall (мэрия, крупная)

| id | name | kind | exposure |
|---|---|---|---|
| `cityhall-portal` | City Services Portal | web | public |
| `cityhall-ad` | Active Directory | identity | intranet |
| `civic-db` | Residents Registry | db | intranet |
| `civic-file-share` | Internal File Share | file-share | intranet |
| `cityhall-mail` | Internal Mail | mail | intranet |
| `cityhall-dns` | Internal DNS | dns | mgmt |
| `cityhall-backup` | Backup Service | backup | mgmt |

#### `election-commission` — Election Commission (избирком, критическая)

| id | name | kind | exposure |
|---|---|---|---|
| `election-dashboard` | Public Results Dashboard | web | public |
| `election-tally` | Vote Tally Service | api | intranet |
| `election-results-db` | Results Database | db | intranet |
| `election-audit-log` | Audit Log | log | mgmt |

#### `city-courts` — City Courts

| id | name | kind | exposure |
|---|---|---|---|
| `courts-portal` | Court Records Portal | web | public |
| `courts-case-mgmt` | Case Management | api | intranet |
| `courts-db` | Case Database | db | intranet |

#### `civil-registry` — Civil Registry (ЗАГС)

| id | name | kind | exposure |
|---|---|---|---|
| `registry-portal` | Registry Portal | web | public |
| `registry-db` | Records Database | db | intranet |

#### `tax-authority` — Tax Authority

| id | name | kind | exposure |
|---|---|---|---|
| `tax-portal` | Tax Portal | web | public |
| `tax-filing-api` | Filing API | api | intranet |
| `tax-db` | Taxpayer Database | db | intranet |

#### `police-dept` — Police Department

| id | name | kind | exposure |
|---|---|---|---|
| `police-records` | Records System | api | intranet |
| `police-cad` | Dispatch (CAD) | ot | ot |
| `police-db` | Crime Database | db | intranet |

#### `fire-dept` — Fire Department

| id | name | kind | exposure |
|---|---|---|---|
| `fire-cad` | Dispatch (CAD) | ot | ot |
| `fire-db` | Incident Database | db | intranet |

#### `emergency-dispatch` — 911 Dispatch Center

| id | name | kind | exposure |
|---|---|---|---|
| `dispatch-psap` | PSAP Call-Taking | ot | ot |
| `dispatch-radio-gw` | Radio Gateway | ot | ot |

---

### Healthcare

#### `[*] phoenix-pharmacy` — Phoenix Pharmacy (малая, с MSP)

| id | name | kind | exposure |
|---|---|---|---|
| `pharmacy-pos` | POS Terminal | pos | intranet |
| `pharmacy-his` | Pharmacy HIS | api | intranet |
| `pharmacy-vpn-endpoint` | WireGuard Endpoint (MSP) | vpn | mgmt |
| `pharmacy-camera-nvr` | CCTV NVR | cctv | ot |

#### `metro-general-hospital` — Metro General Hospital (крупная)

| id | name | kind | exposure |
|---|---|---|---|
| `hospital-ehr` | Electronic Health Records | api | intranet |
| `hospital-pacs` | Imaging (PACS) | api | intranet |
| `hospital-portal` | Patient Portal | web | public |
| `hospital-iot-monitors` | Patient Monitors | ot | ot |
| `hospital-backup` | Backup Service | backup | mgmt |

#### `st-marys-clinic` — St. Mary's Clinic

| id | name | kind | exposure |
|---|---|---|---|
| `clinic-ehr` | Clinic EHR | api | intranet |
| `clinic-portal` | Patient Portal | web | public |

#### `redcross-blood-bank` — Red Cross Blood Bank

| id | name | kind | exposure |
|---|---|---|---|
| `bloodbank-inventory` | Inventory API | api | intranet |
| `bloodbank-db` | Donors Database | db | intranet |

---

### Infrastructure & Utilities

#### `metro-water` — Metro Water (водоканал)

| id | name | kind | exposure |
|---|---|---|---|
| `water-scada` | Pump SCADA | ot | ot |
| `water-billing-api` | Billing API | api | intranet |
| `water-customer-portal` | Customer Portal | web | public |

#### `power-grid-co` — Power Grid Co. (электросеть)

| id | name | kind | exposure |
|---|---|---|---|
| `grid-scada` | Substation SCADA | ot | ot |
| `grid-outage-api` | Outage Map API | api | public |
| `grid-billing` | Billing API | api | intranet |

#### `metro-transit` — Metro Transit (общественный транспорт)

| id | name | kind | exposure |
|---|---|---|---|
| `transit-fare-api` | Fare API | api | public |
| `transit-ticket-pos` | Ticket POS | pos | intranet |
| `transit-avl` | Fleet GPS (AVL) | ot | ot |
| `traffic-controller` | Traffic Light Controller | ot | ot |

#### `telecom-inc` — Telecom Inc. (оператор связи)

| id | name | kind | exposure |
|---|---|---|---|
| `telecom-billing` | Billing System | api | intranet |
| `telecom-dns-resolver` | DNS Resolver | dns | public |
| `telecom-smsc` | SMSC | api | intranet |

#### `[*] cititech-it` — CitiTech MSP (mgmt-сегмент)

| id | name | kind | exposure |
|---|---|---|---|
| `msp-rmm` | RMM Console | rmm | mgmt |
| `msp-siem` | SIEM Collector | log | mgmt |
| `msp-vpn-hub` | WireGuard Hub | vpn | mgmt |
| `msp-backup-vault` | Backup Vault | backup | mgmt |

---

### Finance

#### `first-bank-cybercity` — First Bank

| id | name | kind | exposure |
|---|---|---|---|
| `bank-core` | Core Banking | api | intranet |
| `bank-ibanking` | Internet Banking | web | public |
| `bank-atm-switch` | ATM Switch | api | intranet |
| `bank-db` | Customer Database | db | intranet |

#### `paynet-processor` — PayNet (платёжный процессинг)

| id | name | kind | exposure |
|---|---|---|---|
| `paynet-hsm` | HSM | api | intranet |
| `paynet-api` | Payment API | api | public |
| `paynet-db` | Transactions Database | db | intranet |

#### `credit-union` — Credit Union

| id | name | kind | exposure |
|---|---|---|---|
| `cu-ibanking` | Internet Banking | web | public |
| `cu-core` | Core Banking | api | intranet |

#### `insurance-hub` — Insurance Hub

| id | name | kind | exposure |
|---|---|---|---|
| `ins-portal` | Customer Portal | web | public |
| `ins-claims` | Claims API | api | intranet |

---

### Retail

#### `megamall-chain` — MegaMall Chain (сеть ТРЦ)

| id | name | kind | exposure |
|---|---|---|---|
| `mall-pos` | POS Terminals | pos | intranet |
| `mall-loyalty` | Loyalty API | api | public |
| `mall-camera-nvr` | CCTV NVR | cctv | ot |
| `mall-wifi` | Guest Wi-Fi | api | public |

#### `grocer-everyday` — Grocer Everyday (продуктовая сеть)

| id | name | kind | exposure |
|---|---|---|---|
| `grocer-pos` | POS Terminals | pos | intranet |
| `grocer-inventory` | Inventory API | api | intranet |
| `grocer-loyalty` | Loyalty API | api | public |

#### `gas-stations-net` — Gas Stations Net (АЗС)

| id | name | kind | exposure |
|---|---|---|---|
| `gas-pos` | Fuel POS | pos | intranet |
| `gas-tank-gauge` | Tank Gauge | ot | ot |
| `gas-loyalty` | Loyalty API | api | public |

---

### Media & Telecom

#### `media-center` — Media Center

| id | name | kind | exposure |
|---|---|---|---|
| `media-portal` | News Portal | web | public |
| `media-cms` | CMS | api | intranet |
| `media-stream` | Live Stream | api | public |

#### `metro-radio` — Metro Radio (местное радио)

| id | name | kind | exposure |
|---|---|---|---|
| `radio-stream` | Audio Stream | api | public |
| `radio-studio-automation` | Studio Automation | api | intranet |

---

### Education

#### `public-schools-district` — Public Schools District

| id | name | kind | exposure |
|---|---|---|---|
| `schools-sis` | Student Information System | api | intranet |
| `schools-portal` | Parent Portal | web | public |
| `schools-wifi` | School Wi-Fi | api | public |

#### `cybercity-university` — CyberCity University

| id | name | kind | exposure |
|---|---|---|---|
| `uni-portal` | University Portal | web | public |
| `uni-lms` | Learning Management System | api | intranet |
| `uni-research-cluster` | Research Cluster | api | intranet |
| `uni-wifi` | Campus Wi-Fi | api | public |

---

### MSP / Узкие провайдеры

#### `cloudnet-hosting` — CloudNet Hosting

| id | name | kind | exposure |
|---|---|---|---|
| `cloudnet-customer-portal` | Customer Portal | web | public |
| `cloudnet-virt` | Virtualization API | api | mgmt |
| `cloudnet-object-storage` | Object Storage | api | mgmt |

#### `logistics-fleet` — Logistics Fleet (курьерская)

| id | name | kind | exposure |
|---|---|---|---|
| `fleet-dispatch` | Dispatch API | api | intranet |
| `fleet-tracking` | Tracking API | api | public |
| `fleet-driver-app` | Driver App | api | intranet |

---

## Decoys

### Шаблон записи

| id | subnet | ip | os | role | ports |
|---|---|---|---|---|---|
| `decoy-corp-ws-0142` | corp | 10.20.4.142 | windows-10 | workstation | 135, 139, 445, 3389 |

Полная YAML-форма (ADR-0003):

```yaml
- id: decoy-corp-ws-0142
  subnet: corp
  ip: 10.20.4.142
  os: windows-10
  hostname: WS-0142
  ports:
    - { port: 135,  service: msrpc,  banner: "Windows 10 Pro 19045" }
    - { port: 139,  service: netbios }
    - { port: 445,  service: microsoft-ds, banner: "Windows 10 Pro 6.3" }
    - { port: 3389, service: ms-wbt-server }
  role: workstation
  notes: "типичная рабочая станция бухгалтерии"
```

### Записи

_(пока нет — заполняется итеративно по одной записи за шаг, с прогоном
`go run ./cmd/validate-network` после каждого изменения)_

---

## Сводка

### Organizations

| Блок | Организаций | Сервисов |
|---|---|---|
| Government | 8 | 26 |
| Healthcare | 4 | 13 |
| Infrastructure & Utilities | 5 | 17 |
| Finance | 4 | 11 |
| Retail | 3 | 10 |
| Media & Telecom | 2 | 5 |
| Education | 2 | 7 |
| MSP / провайдеры | 2 | 6 |
| **Итого** | **30** | **95** |

### Decoys

| Сегмент | Хостов | Сгенерировано |
|---|---|---|
| corp | 0 | — |
| ot | 0 | — |
| mgmt | 0 | — |
| public | 0 | — |
| **Итого** | **0** | — |

## Эталонные организации

Три организации отмечены `[*]` и описываются руками как «золотой стандарт»:

- `city-hall` — крупная гос. организация с публичным веб-порталом и AD;
- `phoenix-pharmacy` — малая организация на аутсорсе у MSP;
- `cititech-it` — сам MSP в mgmt-сегменте.

Все остальные 27 организаций — кандидаты на генерацию LLM-агентом
по шаблону, с обязательной финальной валидацией (`go run ./cmd/validate-network`
+ рендер K8s-манифестов + `kubectl --dry-run`).

## Правила расширения

### Organizations

1. Одна организация за итерацию.
2. Перед генерацией — сжатый индекс уже существующих + запрет на дублирование id.
3. После добавления — прогнать `go run ./cmd/validate-network`.
4. Если три раза подряд валидатор падает на одной и той же записи — остановиться,
   разбирает человек.

### Decoys

1. Один decoy за итерацию.
2. IP внутри сегмента не должен совпадать с IP реального сервиса
   (проверяется в кросс-фазе валидатора, ADR-0007 п. 2).
3. `os` и `role` — из закрытого словаря, пополняемого по мере надобности.
4. После добавления — прогнать `go run ./cmd/validate-network`.
