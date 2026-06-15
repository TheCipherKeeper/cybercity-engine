"""Tests for loading topology from cybercity-data artifacts."""

from __future__ import annotations

import json
import zipfile
from io import BytesIO
from pathlib import Path

from cybercity_engine.bootstrap import load_topology


def _make_engine_zip() -> bytes:
    runtime = {
        "schema_version": 1,
        "tick_ms": 1000,
        "services": [
            {
                "id": "bank-web",
                "org_id": "bank",
                "name": "Bank portal",
                "host": "portal.bank.corp",
                "ip": "10.1.0.10",
                "kind": "web",
                "exposure": "public",
                "network_id": "bank-dmz",
                "auth": "sso",
                "data_classification": "public",
                "criticality": "high",
                "ports": ["tcp/443"],
            },
            {
                "id": "bank-db",
                "org_id": "bank",
                "name": "Bank DB",
                "host": "db.bank.corp",
                "ip": "10.1.1.10",
                "kind": "db",
                "exposure": "intranet",
                "network_id": "bank-lan",
                "auth": "certificate",
                "data_classification": "pci",
                "criticality": "critical",
                "ports": ["tcp/1521"],
            },
        ],
        "links": [
            {
                "from": "bank-web",
                "to": "bank-db",
                "kind": "db-read",
                "protocol": "tcp/1521",
                "encryption": "tls",
            }
        ],
    }
    buf = BytesIO()
    with zipfile.ZipFile(buf, "w") as zf:
        zf.writestr("runtime/engine.json", json.dumps(runtime))
    return buf.getvalue()


def test_load_topology_from_zip(tmp_path: Path) -> None:
    zip_path = tmp_path / "engine.zip"
    zip_path.write_bytes(_make_engine_zip())
    topology = load_topology(str(zip_path))
    assert topology.schema_version == "1"
    assert "bank-web" in topology.services
    assert topology.services["bank-web"].kind == "web"
    assert topology.services["bank-web"].bind_ip == "10.1.0.10"
    assert len(topology.edges) == 1
    assert topology.edges[0].kind == "db-read"


def test_load_topology_empty() -> None:
    topology = load_topology(None)
    assert topology.services == {}
