"""Load the city blueprint from ``cybercity-data`` artifacts.

The engine consumes ``engine.zip`` or a local ``engine.json`` produced by
``cybercity-data build``. This module converts the canonical data layer into the
engine's internal topology graph.
"""

from __future__ import annotations

import json
import zipfile
from io import BytesIO
from pathlib import Path
from typing import Any, cast
from urllib.parse import urlparse

import httpx

from .models import TopologyEdge, TopologyGraph, TopologyNode

__all__ = ["load_topology"]


def _is_url(value: str) -> bool:
    parsed = urlparse(value)
    return parsed.scheme in {"http", "https"}


def _fetch_bytes(url: str) -> bytes:
    with httpx.Client(timeout=30.0) as client:
        response = client.get(url)
        response.raise_for_status()
        return response.content


def _read_bytes(path: str) -> bytes:
    return Path(path).read_bytes()


def _load_raw(source: str) -> dict[str, Any]:
    data = _fetch_bytes(source) if _is_url(source) else _read_bytes(source)
    return cast(dict[str, Any], json.loads(data))


def _load_from_zip(source: str) -> dict[str, Any]:
    data = _fetch_bytes(source) if _is_url(source) else _read_bytes(source)
    with zipfile.ZipFile(BytesIO(data)) as zf:
        if "runtime/engine.json" in zf.namelist():
            return cast(
                dict[str, Any],
                json.loads(zf.read("runtime/engine.json").decode("utf-8")),
            )
        if "model/network.json" in zf.namelist():
            return cast(
                dict[str, Any],
                json.loads(zf.read("model/network.json").decode("utf-8")),
            )
    raise ValueError(f"engine.zip from {source!r} has no engine.json or network.json")


def load_topology(source: str | None = None) -> TopologyGraph:
    """Load a topology graph from an ``engine.zip`` or JSON source.

    ``source`` may be:
      * a local path to ``engine.zip`` or ``engine.json``;
      * an HTTP(S) URL to the same;
      * ``None`` — returns an empty graph.
    """
    if source is None:
        return TopologyGraph(schema_version="3.0.0", source_version="0.0.0")

    raw: dict[str, Any]
    if source.endswith(".zip"):
        raw = _load_from_zip(source)
    else:
        raw = _load_raw(source)

    # Accept both runtime/engine.json and model/network.json shapes.
    services_raw = raw.get("services", [])
    if not services_raw:
        services_raw = raw.get("model", {}).get("services", [])

    links_raw = raw.get("links", [])
    if not links_raw:
        links_raw = raw.get("model", {}).get("links", [])

    services = {}
    for svc in services_raw:
        decoy = svc.get("decoy") or svc.get("decoy_profile")
        node = TopologyNode(
            id=svc["id"],
            org_id=svc["org_id"],
            name=svc.get("name", svc["id"]),
            kind=svc["kind"],
            exposure=svc["exposure"],
            host=svc["host"],
            network_id=svc.get("network_id"),
            bind_ip=svc.get("ip") or svc.get("bind_ip"),
            auth=svc.get("auth", "local"),
            data_classification=svc.get("data_classification", "internal"),
            criticality=svc.get("criticality", "medium"),
            ports=svc.get("ports", []),
            is_decoy=decoy is not None,
            decoy_kind=decoy.get("kind") if isinstance(decoy, dict) else None,
            software=svc.get("software", {}),
            os_hint=svc.get("os_hint"),
        )
        services[node.id] = node

    edges = []
    for link in links_raw:
        edges.append(
            TopologyEdge(
                source=link["from"] if "from" in link else link["from_service"],
                target=link["to"] if "to" in link else link["to_service"],
                kind=link["kind"],
                protocol=link.get("protocol"),
                encryption=link.get("encryption"),
            )
        )

    return TopologyGraph(
        schema_version=str(raw.get("version", raw.get("schema_version", "3.0.0"))),
        source_version=str(raw.get("source_version", "unknown")),
        services=services,
        edges=edges,
    )
