// Package router отвечает за распространение событий по топологическому графу.
package router

import (
	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
)

// PropagationRule — чистая функция, которая по событию, исходному узлу и ребру
// решает, нужно ли породить дочернее событие для соседа.
type PropagationRule func(event models.EventNode, source *models.TopologyNode, edge models.TopologyEdge, topology *models.TopologyGraph) *models.EventNode

// EventRouter — graph-aware propagation engine.
type EventRouter struct {
	rules []PropagationRule
}

// NewEventRouter создаёт роутер с дефолтными правилами.
func NewEventRouter(rules ...PropagationRule) *EventRouter {
	if len(rules) == 0 {
		rules = defaultRules()
	}
	return &EventRouter{rules: rules}
}

// Propagate возвращает дочерние события для всех соседей, затронутых
// исходным событием.
func (r *EventRouter) Propagate(event models.EventNode, sourceNode *models.TopologyNode, topology *models.TopologyGraph) []models.EventNode {
	var children []models.EventNode
	for _, edge := range topology.Neighbors(sourceNode.ID) {
		for _, rule := range r.rules {
			child := rule(event, sourceNode, edge, topology)
			if child != nil {
				children = append(children, *child)
			}
		}
	}
	return children
}

func defaultRules() []PropagationRule {
	return []PropagationRule{
		scanAlertRule,
		compromisePropagationRule,
		stateChangePropagationRule,
	}
}

func scanAlertRule(event models.EventNode, source *models.TopologyNode, edge models.TopologyEdge, topology *models.TopologyGraph) *models.EventNode {
	if event.EventType != models.EventTypeScan {
		return nil
	}
	if edge.Kind != "log-sink" && edge.Kind != "trusts" {
		return nil
	}
	noiseRaw, ok := event.Payload["noise_level"]
	if !ok {
		return nil
	}
	noise := toFloat(noiseRaw)
	if noise < 0.5 {
		return nil
	}

	child := event.SpawnChild(
		models.EventTypePropagation,
		models.SourceTypeEngine,
		source.ID,
		edge.Target,
		map[string]any{
			"kind":              "scan_alert",
			"original_event_id": event.EventID,
			"noise_level":       noise,
			"via":               edge.Kind,
		},
		event.Tick,
	)
	return &child
}

func compromisePropagationRule(event models.EventNode, source *models.TopologyNode, edge models.TopologyEdge, topology *models.TopologyGraph) *models.EventNode {
	if event.EventType != models.EventTypeCompromise {
		return nil
	}
	if edge.Kind != "trusts" && edge.Kind != "auth" && edge.Kind != "vendor-vpn" {
		return nil
	}
	severity := toFloat(event.Payload["severity"])
	if severity < 0.3 {
		severity = 0.3
	}
	if severity < 0.3 {
		return nil
	}

	child := event.SpawnChild(
		models.EventTypeAttack,
		models.SourceTypeEngine,
		source.ID,
		edge.Target,
		map[string]any{
			"kind":              "lateral_movement",
			"original_event_id": event.EventID,
			"severity":          severity * 0.8,
			"via":               edge.Kind,
		},
		event.Tick,
	)
	return &child
}

func stateChangePropagationRule(event models.EventNode, source *models.TopologyNode, edge models.TopologyEdge, topology *models.TopologyGraph) *models.EventNode {
	if event.EventType != models.EventTypeStateChange {
		return nil
	}
	if event.Payload["field"] != "status" {
		return nil
	}
	newStatus, ok := event.Payload["new_value"].(string)
	if !ok {
		return nil
	}
	if newStatus != "down" && newStatus != "compromised" {
		return nil
	}
	if edge.Kind != "api-call" && edge.Kind != "db-read" && edge.Kind != "db-write" && edge.Kind != "backup-of" && edge.Kind != "log-sink" {
		return nil
	}

	child := event.SpawnChild(
		models.EventTypeBackgroundEffect,
		models.SourceTypeEngine,
		source.ID,
		edge.Target,
		map[string]any{
			"kind":              "dependency_impact",
			"original_event_id": event.EventID,
			"status":            newStatus,
			"via":               edge.Kind,
		},
		event.Tick,
	)
	return &child
}

func toFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	}
	return 0
}
