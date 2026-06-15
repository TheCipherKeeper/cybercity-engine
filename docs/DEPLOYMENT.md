# CyberCity Engine — Deployment

This document describes how to deploy the CyberCity engine and its
dependencies. It covers local development, home lab, and sketches production
considerations.

## Local development

### Requirements

- Python 3.12+
- Docker / Docker Compose
- `uv`

### 1. Start dependencies

```bash
docker compose up -d postgres redpanda minio
```

This starts:

- PostgreSQL on `localhost:5432`
- Redpanda on `localhost:9092`
- MinIO on `localhost:9000`

### 2. Sync Python environment

```bash
uv sync
```

### 3. Build or download a city artifact

From `cybercity-data`:

```bash
cd ../cybercity-data
uv run cybercity-data build . --clean --seed 42
```

Copy `build/engine.zip` to the engine directory or make it available via MinIO.

### 4. Run the engine

```bash
uv run cybercity-engine --engine-zip engine.zip
```

Or with hot reload for development:

```bash
uv run uvicorn cybercity_engine.api:create_app --reload
```

Note: `create_app` requires an engine instance, so for reload use a small
wrapper module. This will be documented later.

### 5. Verify

```bash
curl http://localhost:8000/health
curl http://localhost:8000/topology
```

## Home lab

### Target hardware

| Profile | CPU | RAM | Storage | Services |
|---------|-----|-----|---------|----------|
| Entry | 8c/16t | 32 GB | 1 TB NVMe | Mostly simulated, 0–2 real VMs |
| Comfort | 16c/32t | 64 GB | 2 TB NVMe | Mixed, ~10 real VMs |
| Sweet spot | 24c/48t | 128 GB | 4 TB NVMe | Full hybrid, ~30 real VMs |

### Layout

```text
Proxmox VE
├── k8s-cp-01,02,03          (K8s control plane)
├── k8s-worker-01,02,03      (engine, UI, simulated services)
├── db-01,02                   (PostgreSQL HA)
├── redpanda-01,02,03          (messaging)
├── router-01                  (VyOS/pfSense)
├── bank-web-01                (real VM)
├── hospital-pacs-01           (real VM)
├── windows-ad-01              (real VM)
├── kali-workstation-01        (player VM)
└── mon-01                     (monitoring)
```

### K8s

Use Talos Linux or kubeadm on Ubuntu. Deploy engine, UI, simulated services,
Redpanda, PostgreSQL, and monitoring via Helm/Kustomize.

### Networking

- Management network: `192.168.100.0/24`
- City networks: `10.0.0.0/8` allocated per org by `cybercity-data`
- Player VPN/VDI: separate segment
- Use Cilium + Multus for pod network in city segments.

### Public access

Use Cloudflare Tunnel or similar to expose UI read-only view without opening
home router ports.

## Production sketch

| Component | Suggested |
|-----------|-----------|
| Orchestration | Managed Kubernetes (EKS/GKE/AKS) |
| Database | Managed PostgreSQL or CloudNativePG |
| Messaging | Redpanda Cloud or self-hosted cluster |
| Real VMs | Proxmox / VMware / cloud bare metal |
| Networking | Cilium cluster mesh, dedicated firewalls |
| Observability | Prometheus/Grafana/Loki + custom dashboards |
| Secrets | Vault or cloud KMS |
| CI/CD | ArgoCD, GitHub Actions |
| Artifacts | S3/MinIO + signed releases |

## Environments

| Env | Purpose | Trigger |
|-----|---------|---------|
| `dev` | Experiments | every push to main |
| `staging` | Pre-release validation | tag `v*-rc*` |
| `prod` | Live exercises | manual promotion |

## Allocation seed

For `staging` and `prod`, the `cybercity-data` allocation seed must be fixed so
that IP addresses do not change between releases. Dev may use random seeds.

## Backup and recovery

- PostgreSQL: continuous WAL archiving + daily snapshots.
- Event audit log: write-once, replicated object storage.
- City artifact: versioned in artifact store.
- VM real services: Proxmox Backup Server or snapshot tooling.

## Security hardening

- Segment networks by organization and exposure.
- Use mutual TLS for agent-to-bus communication.
- Store secrets in Vault/KMS, never in repositories.
- Restrict player VPN to city networks only.
- Keep OT networks air-gapped from public segments.
- Regular patching of real VM templates.

## Related

- `docs/ARCHITECTURE.md` — system context.
- `compose.yaml` — local dependency setup.
- `cybercity-data` docs for artifact generation.
