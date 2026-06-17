package models

import (
	"testing"
)

func TestTopologyNodeDefaults(t *testing.T) {
	node := NewTopologyNode("bank-web", "bank", "Bank portal", "web", ExposurePublic, "portal.bank.corp")
	if node.Auth != "local" {
		t.Errorf("expected auth local, got %q", node.Auth)
	}
	if node.Criticality != CriticalityMedium {
		t.Errorf("expected criticality medium, got %q", node.Criticality)
	}
	if node.IsHoneypot {
		t.Error("expected is_honeypot false")
	}
}

func TestTopologyGraphNeighbors(t *testing.T) {
	graph := &TopologyGraph{
		SchemaVersion: "3.0.0",
		SourceVersion: "0.1.0",
		Services: map[string]*TopologyNode{
			"a": {ID: "a", OrgID: "x", Name: "A", Kind: "web", Exposure: ExposurePublic, Host: "a.corp"},
			"b": {ID: "b", OrgID: "x", Name: "B", Kind: "db", Exposure: ExposureIntranet, Host: "b.corp"},
		},
		Edges: []TopologyEdge{{Source: "a", Target: "b", Kind: "api-call"}},
	}
	neighbors := graph.Neighbors("a")
	if len(neighbors) != 1 {
		t.Fatalf("expected 1 neighbor, got %d", len(neighbors))
	}
	if neighbors[0].Target != "b" {
		t.Errorf("expected target b, got %q", neighbors[0].Target)
	}
}

func TestEventSpawnChild(t *testing.T) {
	parent := NewEvent(SourceTypePlayer, "p1", EventTypeScan, "bank-web", map[string]any{"ports": []int{443}})
	parent.Tick = 1

	child := parent.SpawnChild(EventTypePropagation, SourceTypeEngine, "bank-web", "bank-log", map[string]any{"noise_level": 0.8}, 2)

	if len(child.ParentEventIDs) != 1 || child.ParentEventIDs[0] != parent.EventID {
		t.Errorf("expected parent %s, got %v", parent.EventID, child.ParentEventIDs)
	}
	if child.CorrelationID != parent.CorrelationID {
		t.Error("expected same correlation id")
	}
	if child.Tick != 2 {
		t.Errorf("expected tick 2, got %d", child.Tick)
	}
	if child.EventType != EventTypePropagation {
		t.Errorf("expected event type propagation, got %q", child.EventType)
	}
}

func TestServiceStateHealthNoHardBounds(t *testing.T) {
	svc := &ServiceState{ServiceID: "x", Health: 2.0}
	if svc.Health != 2.0 {
		t.Errorf("expected health 2.0, got %f", svc.Health)
	}
}
