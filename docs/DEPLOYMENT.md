# CyberCity Engine — Развёртывание

Этот документ описывает, как развёртывать движок CyberCity и его
зависимости. Покрывает локальную разработку, home lab и набросок production.

## Локальная разработка

### Требования

- Go 1.22+
- Docker / Docker Compose

### 1. Запуск зависимостей

```bash
docker compose up -d postgres redpanda minio
```

Запускает:

- PostgreSQL на `localhost:5432`
- Redpanda на `localhost:9092`
- MinIO на `localhost:9000`

### 2. Сборка движка

```bash
go mod tidy
go build ./cmd/cybercity-engine
```

### 3. Сборка или загрузка артефакта города

Из `cybercity-data`:

```bash
cd ../cybercity-data
cybercity-data build . --clean --seed 42
```

Скопируйте `build/engine.zip` в директорию движка или сделайте доступным через
MinIO.

### 4. Запуск движка

```bash
go run ./cmd/cybercity-engine --engine-zip engine.zip
# или, после сборки:
./cybercity-engine --engine-zip engine.zip
```

### 5. Проверка

```bash
curl http://localhost:8000/health
curl http://localhost:8000/topology
```

## Home lab

### Целевое железо

| Профиль | CPU | RAM | Storage | Сервисы |
|---------|-----|-----|---------|---------|
| Entry | 8c/16t | 32 GB | 1 TB NVMe | В основном simulated, 0–2 real VM |
| Comfort | 16c/32t | 64 GB | 2 TB NVMe | Смешанный, ~10 real VM |
| Sweet spot | 24c/48t | 128 GB | 4 TB NVMe | Full hybrid, ~30 real VM |

### Схема

```text
Proxmox VE
├── k8s-cp-01,02,03          (control plane K8s)
├── k8s-worker-01,02,03      (движок, UI, simulated-сервисы)
├── db-01,02                   (PostgreSQL HA)
├── redpanda-01,02,03          (messaging)
├── router-01                  (VyOS/pfSense)
├── bank-web-01                (real VM)
├── hospital-pacs-01           (real VM)
├── windows-ad-01              (real VM)
├── kali-workstation-01        (player VM)
└── mon-01                     (мониторинг)
```

### Kubernetes

Используйте Talos Linux или kubeadm на Ubuntu. Деплой движка, UI,
simulated-сервисов, Redpanda, PostgreSQL и мониторинга через Helm/Kustomize.

### Сеть

- Management network: `192.168.100.0/24`
- City networks: `10.0.0.0/8`, выделенные per org `cybercity-data`
- Player VPN/VDI: отдельный сегмент
- Cilium + Multus для pod-сети в city-сегментах

### Публичный доступ

Используйте Cloudflare Tunnel или аналог, чтобы открыть read-only UI без
проброса портов домашнего роутера.

## Production-набросок

| Компонент | Рекомендация |
|-----------|--------------|
| Оркестрация | Managed Kubernetes (EKS/GKE/AKS) |
| БД | Managed PostgreSQL или CloudNativePG |
| Messaging | Redpanda Cloud или self-hosted cluster |
| Real VMs | Proxmox / VMware / cloud bare metal |
| Сеть | Cilium cluster mesh, выделенные firewall |
| Observability | Prometheus/Grafana/Loki + кастомные дашборды |
| Secrets | Vault или cloud KMS |
| CI/CD | ArgoCD, GitHub Actions |
| Артефакты | S3/MinIO + signed releases |

## Окружения

| Окружение | Назначение | Триггер |
|-----------|-----------|---------|
| `dev` | Эксперименты | каждый push в main |
| `staging` | Pre-release валидация | tag `v*-rc*` |
| `prod` | Живые учения | ручное promotion |

## Allocation seed

Для `staging` и `prod` seed `cybercity-data` должен быть фиксированным, чтобы
IP-адреса не менялись между релизами. Dev может использовать random seeds.

## Backup и восстановление

- PostgreSQL: continuous WAL archiving + daily снапшоты.
- Audit log событий: write-once, реплицированный object storage.
- City artifact: версионированный в artifact store.
- Real VM: Proxmox Backup Server или snapshot-утилиты.

## Усиление безопасности

- Сегментировать сети по организациям и exposure.
- Использовать mutual TLS для agent-to-bus коммуникаций.
- Хранить секреты в Vault/KMS, никогда в репозиториях.
- Ограничить player VPN только city-сетями.
- Держать OT-сети air-gapped от public-сегментов.
- Регулярно патчить шаблоны real VM.

## Связанные документы

- [`docs/ARCHITECTURE.md`](ARCHITECTURE.md) — системный контекст.
- [`compose.yaml`](../compose.yaml) — локальная поднималка зависимостей.
- Документация `cybercity-data` для генерации артефактов.
