# 0007 — network model and K8s projection

Status: accepted
Date: 2026-06-02

## Context

The city must be described in a way that both humans and the
runtime can agree on. K8s is the chosen runtime. Manually keeping
K8s manifests and a city description in sync is a known failure mode.

## Decision

`network.yml` is the **single canonical description** of the city:
organizations, services, decoy hosts, allowed channels. K8s manifests
(Namespace, NetworkPolicy, Service, Ingress) are a **projection** of
this YAML, rendered by `cmd/render-manifests`.

## Consequences

- One validator (`cmd/validate-network`) covers all three phases:
  structural, intra-organisation, cross-organisation vs decoys.
- No separate `validate-decoys` — single contract.
- Editing K8s manifests by hand is forbidden; they are generated.
