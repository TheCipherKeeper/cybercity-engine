// Package models содержит основные доменные модели CyberCity Engine.
//
// Движок работает с двумя связанными структурами:
//  1. Топологический граф — статический blueprint, загружаемый из cybercity-data.
//  2. Событийный граф — динамический causal-граф runtime-событий.
package models

import (
	"time"

	"github.com/google/uuid"
)

// ServiceStatus — runtime-статус сервиса.
type ServiceStatus string

const (
	ServiceStatusUp          ServiceStatus = "up"
	ServiceStatusDown        ServiceStatus = "down"
	ServiceStatusCompromised ServiceStatus = "compromised"
	ServiceStatusMaintenance ServiceStatus = "maintenance"
)

// Exposure — уровень экспозиции сервиса.
type Exposure string

const (
	ExposurePublic   Exposure = "public"
	ExposureIntranet Exposure = "intranet"
	ExposureOT       Exposure = "ot"
	ExposureMgmt     Exposure = "mgmt"
)

// Criticality — критичность сервиса.
type Criticality string

const (
	CriticalityCritical Criticality = "critical"
	CriticalityHigh     Criticality = "high"
	CriticalityMedium   Criticality = "medium"
	CriticalityLow      Criticality = "low"
)

// EventType — тип runtime-события.
type EventType string

const (
	EventTypeHeartbeat        EventType = "heartbeat"
	EventTypeScan             EventType = "scan"
	EventTypeAttack           EventType = "attack"
	EventTypeCompromise       EventType = "compromise"
	EventTypeRestore          EventType = "restore"
	EventTypeStateChange      EventType = "state_change"
	EventTypeCommand          EventType = "command"
	EventTypeScenarioStart    EventType = "scenario_start"
	EventTypeScenarioStop     EventType = "scenario_stop"
	EventTypeFlagCaptured     EventType = "flag_captured"
	EventTypeBackgroundEffect EventType = "background_effect"
	EventTypePropagation      EventType = "propagation"
)

// SourceType — кто породил событие.
type SourceType string

const (
	SourceTypeEngine     SourceType = "engine"
	SourceTypeService    SourceType = "service"
	SourceTypeScenario   SourceType = "scenario"
	SourceTypePlayer     SourceType = "player"
	SourceTypeSystem     SourceType = "system"
	SourceTypeBackground SourceType = "background"
)

// EventStatus — статус обработки события.
type EventStatus string

const (
	EventStatusPending    EventStatus = "pending"
	EventStatusProcessed  EventStatus = "processed"
	EventStatusFailed     EventStatus = "failed"
	EventStatusSuppressed EventStatus = "suppressed"
)

// EventEdgeKind — вид связи между событиями.
type EventEdgeKind string

const (
	EventEdgeKindCausedBy      EventEdgeKind = "caused_by"
	EventEdgeKindPropagatedTo  EventEdgeKind = "propagated_to"
	EventEdgeKindTriggeredRule EventEdgeKind = "triggered_rule"
	EventEdgeKindResponseTo    EventEdgeKind = "response_to"
)

// TopologyNode — статическое описание сервиса из blueprint города.
type TopologyNode struct {
	ID                 string         `json:"id"`
	OrgID              string         `json:"org_id"`
	Name               string         `json:"name"`
	Kind               string         `json:"kind"`
	Exposure           Exposure       `json:"exposure"`
	Host               string         `json:"host"`
	NetworkID          string         `json:"network_id,omitempty"`
	BindIP             string         `json:"bind_ip,omitempty"`
	Auth               string         `json:"auth,omitempty"`
	DataClassification string         `json:"data_classification,omitempty"`
	Criticality        Criticality    `json:"criticality,omitempty"`
	Ports              []string       `json:"ports,omitempty"`
	IsHoneypot         bool           `json:"is_honeypot,omitempty"`
	HoneypotKind       string         `json:"honeypot_kind,omitempty"`
	Software           map[string]any `json:"software,omitempty"`
	OSHint             string         `json:"os_hint,omitempty"`
}

// NewTopologyNode создаёт TopologyNode с разумными дефолтами.
func NewTopologyNode(id, orgID, name, kind string, exposure Exposure, host string) *TopologyNode {
	return &TopologyNode{
		ID:                 id,
		OrgID:              orgID,
		Name:               name,
		Kind:               kind,
		Exposure:           exposure,
		Host:               host,
		Auth:               "local",
		DataClassification: "internal",
		Criticality:        CriticalityMedium,
	}
}

// TopologyEdge — статическая или inferred-связь между сервисами.
type TopologyEdge struct {
	Source     string `json:"source"`
	Target     string `json:"target"`
	Kind       string `json:"kind"`
	Protocol   string `json:"protocol,omitempty"`
	Encryption string `json:"encryption,omitempty"`
	Inferred   bool   `json:"inferred,omitempty"`
}

// TopologyGraph — иммутабельный blueprint, загружаемый из cybercity-data.
type TopologyGraph struct {
	SchemaVersion string                   `json:"schema_version"`
	SourceVersion string                   `json:"source_version"`
	Services      map[string]*TopologyNode `json:"services"`
	Edges         []TopologyEdge           `json:"edges"`
}

// Neighbors возвращает все исходящие рёбра из сервиса.
func (g *TopologyGraph) Neighbors(serviceID string) []TopologyEdge {
	out := make([]TopologyEdge, 0)
	for _, e := range g.Edges {
		if e.Source == serviceID {
			out = append(out, e)
		}
	}
	return out
}

// EventNode — одно событие в динамическом событийном графе.
// События иммутабельны на уровне API; поле Status может меняться при обработке.
type EventNode struct {
	EventID        string         `json:"event_id"`
	ParentEventIDs []string       `json:"parent_event_ids,omitempty"`
	CorrelationID  string         `json:"correlation_id"`
	Tick           int            `json:"tick"`
	Timestamp      time.Time      `json:"timestamp"`
	SourceType     SourceType     `json:"source_type"`
	SourceID       string         `json:"source_id"`
	EventType      EventType      `json:"event_type"`
	TargetID       string         `json:"target_id,omitempty"`
	Payload        map[string]any `json:"payload,omitempty"`
	Status         EventStatus    `json:"status"`
}

// NewEvent создаёт событие с новыми UUID.
func NewEvent(sourceType SourceType, sourceID string, eventType EventType, targetID string, payload map[string]any) EventNode {
	if payload == nil {
		payload = make(map[string]any)
	}
	return EventNode{
		EventID:       uuid.NewString(),
		CorrelationID: uuid.NewString(),
		Timestamp:     time.Now().UTC(),
		SourceType:    sourceType,
		SourceID:      sourceID,
		EventType:     eventType,
		TargetID:      targetID,
		Payload:       payload,
		Status:        EventStatusPending,
	}
}

// SpawnChild создаёт дочернее событие, вызванное текущим.
func (e EventNode) SpawnChild(eventType EventType, sourceType SourceType, sourceID, targetID string, payload map[string]any, tick int) EventNode {
	if payload == nil {
		payload = make(map[string]any)
	}
	parents := []string{e.EventID}
	parents = append(parents, e.ParentEventIDs...)
	return EventNode{
		EventID:        uuid.NewString(),
		ParentEventIDs: parents,
		CorrelationID:  e.CorrelationID,
		Tick:           tick,
		Timestamp:      time.Now().UTC(),
		SourceType:     sourceType,
		SourceID:       sourceID,
		EventType:      eventType,
		TargetID:       targetID,
		Payload:        payload,
		Status:         EventStatusPending,
	}
}

// EventEdge — связь между двумя событиями.
type EventEdge struct {
	SourceEventID string        `json:"source_event_id"`
	TargetEventID string        `json:"target_event_id"`
	Kind          EventEdgeKind `json:"kind"`
}

// ServiceState — изменяемое runtime-состояние, привязанное к узлу топологии.
type ServiceState struct {
	ServiceID        string         `json:"service_id"`
	Status           ServiceStatus  `json:"status"`
	Health           float64        `json:"health"`
	CompromiseVector string         `json:"compromise_vector,omitempty"`
	LastHeartbeat    *time.Time     `json:"last_heartbeat,omitempty"`
	SeenBy           []string       `json:"seen_by,omitempty"`
	Flags            map[string]any `json:"flags,omitempty"`
	Variables        map[string]any `json:"variables,omitempty"`
}

// PlayerState — игрок в учении.
type PlayerState struct {
	PlayerID string   `json:"player_id"`
	Name     string   `json:"name"`
	OrgID    string   `json:"org_id,omitempty"`
	Score    int      `json:"score"`
	Flags    []string `json:"flags,omitempty"`
	Status   string   `json:"status"`
}

// ScenarioState — активный учебный сценарий.
type ScenarioState struct {
	ScenarioID string         `json:"scenario_id"`
	Name       string         `json:"name"`
	Status     string         `json:"status"`
	StartedAt  time.Time      `json:"started_at"`
	EndedAt    *time.Time     `json:"ended_at,omitempty"`
	Config     map[string]any `json:"config,omitempty"`
}

// WorldState — полный runtime-снапшот симуляции.
type WorldState struct {
	Tick           int                      `json:"tick"`
	StartedAt      time.Time                `json:"started_at"`
	Services       map[string]*ServiceState `json:"services"`
	Players        map[string]*PlayerState  `json:"players"`
	ActiveScenario *ScenarioState           `json:"active_scenario,omitempty"`
	Variables      map[string]any           `json:"variables,omitempty"`
}

// Service возвращает состояние сервиса по id.
func (w *WorldState) Service(serviceID string) *ServiceState {
	return w.Services[serviceID]
}
