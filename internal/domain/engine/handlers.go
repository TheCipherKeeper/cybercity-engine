package engine

import (
	"time"

	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
)

func defaultHandlers() map[models.EventType]Handler {
	return map[models.EventType]Handler{
		models.EventTypeHeartbeat:  handleHeartbeat,
		models.EventTypeScan:       handleScan,
		models.EventTypeAttack:     handleAttack,
		models.EventTypeCompromise: handleCompromise,
		models.EventTypeCommand:    handleCommand,
	}
}

func handleHeartbeat(e *Engine, event models.EventNode) HandlerResult {
	if event.TargetID == "" {
		return HandlerResult{}
	}
	t := event.Timestamp
	if t.IsZero() {
		now := time.Now().UTC()
		t = now
	}
	e.State.RecordHeartbeat(event.TargetID, &t)
	return HandlerResult{}
}

func handleScan(e *Engine, event models.EventNode) HandlerResult {
	if event.TargetID == "" {
		return HandlerResult{}
	}
	e.State.MarkSeenBy(event.TargetID, event.SourceID)
	sourceNode := e.Topology.Services[event.TargetID]
	if sourceNode == nil {
		return HandlerResult{}
	}
	children := e.Router.Propagate(event, sourceNode, e.Topology)
	return HandlerResult{Children: children}
}

func handleAttack(e *Engine, event models.EventNode) HandlerResult {
	if event.TargetID == "" {
		return HandlerResult{}
	}
	success, ok := event.Payload["success"].(bool)
	if !ok || !success {
		return HandlerResult{}
	}
	change := e.State.SetServiceStatus(event.TargetID, models.ServiceStatusCompromised, &event)
	if change == nil {
		return HandlerResult{}
	}
	return HandlerResult{StateChanges: []models.EventNode{*change}}
}

func handleCompromise(e *Engine, event models.EventNode) HandlerResult {
	if event.TargetID == "" {
		return HandlerResult{}
	}
	change := e.State.SetServiceStatus(event.TargetID, models.ServiceStatusCompromised, &event)
	if change == nil {
		return HandlerResult{}
	}
	return HandlerResult{StateChanges: []models.EventNode{*change}}
}

func handleCommand(e *Engine, event models.EventNode) HandlerResult {
	action, _ := event.Payload["action"].(string)
	if event.TargetID == "" {
		return HandlerResult{}
	}
	switch action {
	case "ENABLE_BACKUP_POWER":
		e.State.SetFlag(event.TargetID, "using_backup_power", true)
		change := e.State.SetServiceStatus(event.TargetID, models.ServiceStatusUp, &event)
		if change == nil {
			return HandlerResult{}
		}
		return HandlerResult{StateChanges: []models.EventNode{*change}}
	case "ISOLATE_SERVICE":
		change := e.State.SetServiceStatus(event.TargetID, models.ServiceStatusMaintenance, &event)
		if change == nil {
			return HandlerResult{}
		}
		return HandlerResult{StateChanges: []models.EventNode{*change}}
	}
	return HandlerResult{}
}
