// Package state владеет изменяемым runtime-состоянием CyberCity Engine.
package state

import (
	"time"

	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
	"github.com/google/uuid"
)

// StateManager — единственный владелец WorldState.
type StateManager struct {
	Topology *models.TopologyGraph
	World    *models.WorldState
}

// NewStateManager создаёт StateManager и инициализирует состояние сервисов
// по топологическому графу.
func NewStateManager(topology *models.TopologyGraph) *StateManager {
	world := &models.WorldState{
		Tick:      0,
		StartedAt: time.Now().UTC(),
		Services:  make(map[string]*models.ServiceState),
		Players:   make(map[string]*models.PlayerState),
		Variables: make(map[string]any),
	}
	sm := &StateManager{
		Topology: topology,
		World:    world,
	}
	sm.initServices()
	return sm
}

func (sm *StateManager) initServices() {
	for id := range sm.Topology.Services {
		sm.World.Services[id] = &models.ServiceState{
			ServiceID: id,
			Status:    models.ServiceStatusUp,
			Health:    1.0,
			SeenBy:    []string{},
			Flags:     make(map[string]any),
			Variables: make(map[string]any),
		}
	}
}

// SetServiceStatus меняет статус сервиса и возвращает STATE_CHANGE событие.
// Если статус не изменился, возвращает nil.
func (sm *StateManager) SetServiceStatus(serviceID string, status models.ServiceStatus, reasonEvent *models.EventNode) *models.EventNode {
	svc := sm.requireService(serviceID)
	if svc.Status == status {
		return nil
	}

	oldStatus := svc.Status
	svc.Status = status
	now := time.Now().UTC()
	svc.LastHeartbeat = &now

	payload := map[string]any{
		"entity":    "service",
		"field":     "status",
		"old_value": string(oldStatus),
		"new_value": string(status),
	}
	if reasonEvent != nil {
		payload["parent_event_id"] = reasonEvent.EventID
	}

	return &models.EventNode{
		EventID:    uuid.NewString(),
		Tick:       sm.World.Tick,
		Timestamp:  time.Now().UTC(),
		SourceType: models.SourceTypeEngine,
		SourceID:   "state-manager",
		EventType:  models.EventTypeStateChange,
		TargetID:   serviceID,
		Payload:    payload,
		Status:     models.EventStatusPending,
	}
}

// RecordHeartbeat обновляет время последнего heartbeat сервиса.
func (sm *StateManager) RecordHeartbeat(serviceID string, timestamp *time.Time) {
	svc := sm.requireService(serviceID)
	if timestamp != nil {
		svc.LastHeartbeat = timestamp
		return
	}
	now := time.Now().UTC()
	svc.LastHeartbeat = &now
}

// MarkSeenBy добавляет observer в список тех, кто видел сервис.
func (sm *StateManager) MarkSeenBy(serviceID, observerID string) {
	svc := sm.requireService(serviceID)
	for _, existing := range svc.SeenBy {
		if existing == observerID {
			return
		}
	}
	svc.SeenBy = append(svc.SeenBy, observerID)
}

// SetHealth устанавливает здоровье сервиса в диапазоне [0, 1].
func (sm *StateManager) SetHealth(serviceID string, health float64) {
	svc := sm.requireService(serviceID)
	if health < 0 {
		health = 0
	}
	if health > 1 {
		health = 1
	}
	svc.Health = health
}

// SetFlag устанавливает runtime-флаг сервиса.
func (sm *StateManager) SetFlag(serviceID, key string, value any) {
	svc := sm.requireService(serviceID)
	if svc.Flags == nil {
		svc.Flags = make(map[string]any)
	}
	svc.Flags[key] = value
}

// AddPlayer добавляет игрока в мир.
func (sm *StateManager) AddPlayer(player *models.PlayerState) {
	sm.World.Players[player.PlayerID] = player
}

// SetScenario устанавливает или сбрасывает активный сценарий.
func (sm *StateManager) SetScenario(scenario *models.ScenarioState) {
	sm.World.ActiveScenario = scenario
}

// IncrementTick увеличивает номер текущего tick.
func (sm *StateManager) IncrementTick() {
	sm.World.Tick++
}

func (sm *StateManager) requireService(serviceID string) *models.ServiceState {
	svc, ok := sm.World.Services[serviceID]
	if !ok {
		panic("unknown service " + serviceID)
	}
	return svc
}
