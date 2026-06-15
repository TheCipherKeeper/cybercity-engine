package engine

import (
	"context"
	"testing"
	"time"

	"github.com/TheCipherKeeper/cybercity-engine/internal/adapters/memory"
	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
)

func minimalTopology() *models.TopologyGraph {
	return &models.TopologyGraph{
		SchemaVersion: "3.0.0",
		SourceVersion: "0.1.0",
		Services: map[string]*models.TopologyNode{
			"bank-web": {ID: "bank-web", OrgID: "bank", Name: "Bank portal", Kind: "web", Exposure: models.ExposurePublic, Host: "portal.bank.corp"},
			"bank-db":  {ID: "bank-db", OrgID: "bank", Name: "Bank DB", Kind: "db", Exposure: models.ExposureIntranet, Host: "db.bank.corp"},
			"bank-log": {ID: "bank-log", OrgID: "bank", Name: "Bank log sink", Kind: "log", Exposure: models.ExposureIntranet, Host: "log.bank.corp"},
		},
		Edges: []models.TopologyEdge{
			{Source: "bank-web", Target: "bank-db", Kind: "db-read"},
			{Source: "bank-web", Target: "bank-log", Kind: "log-sink"},
		},
	}
}

func makeEngine() *Engine {
	return NewEngine(
		minimalTopology(),
		Config{TickMs: 0, SnapshotIntervalTicks: 10},
		nil,
		memory.NewEventStore(10000),
		memory.NewSnapshotStore(),
		nil,
		nil,
	)
}

func TestHeartbeatUpdatesService(t *testing.T) {
	eng := makeEngine()
	event := models.NewEvent(models.SourceTypeService, "bank-web", models.EventTypeHeartbeat, "bank-web", nil)
	eng.ProcessEvent(context.Background(), &event)

	svc := eng.State.World.Service("bank-web")
	if svc == nil {
		t.Fatal("service bank-web not found")
	}
	if svc.LastHeartbeat == nil {
		t.Fatal("expected last_heartbeat to be set")
	}
	if event.Status != models.EventStatusProcessed {
		t.Errorf("expected event status processed, got %q", event.Status)
	}
}

func TestScanPropagatesToLogSink(t *testing.T) {
	eng := makeEngine()
	event := models.NewEvent(models.SourceTypePlayer, "p1", models.EventTypeScan, "bank-web", map[string]any{"noise_level": 0.9})
	eng.ProcessEvent(context.Background(), &event)

	recent, err := eng.Store.Recent(context.Background(), 10)
	if err != nil {
		t.Fatalf("recent: %v", err)
	}
	hasScan := false
	for _, e := range recent {
		if e.EventType == models.EventTypeScan {
			hasScan = true
		}
	}
	if !hasScan {
		t.Error("expected scan event in recent events")
	}
}

func TestAttackCompromisesService(t *testing.T) {
	eng := makeEngine()
	event := models.NewEvent(models.SourceTypePlayer, "p1", models.EventTypeAttack, "bank-web", map[string]any{"success": true, "vector": "sqli"})
	eng.ProcessEvent(context.Background(), &event)

	svc := eng.State.World.Service("bank-web")
	if svc == nil {
		t.Fatal("service bank-web not found")
	}
	if svc.Status != models.ServiceStatusCompromised {
		t.Errorf("expected status compromised, got %q", svc.Status)
	}
}

func TestCommandIsolatesService(t *testing.T) {
	eng := makeEngine()
	event := models.NewEvent(models.SourceTypePlayer, "p1", models.EventTypeCommand, "bank-web", map[string]any{"action": "ISOLATE_SERVICE"})
	eng.ProcessEvent(context.Background(), &event)

	svc := eng.State.World.Service("bank-web")
	if svc == nil {
		t.Fatal("service bank-web not found")
	}
	if svc.Status != models.ServiceStatusMaintenance {
		t.Errorf("expected status maintenance, got %q", svc.Status)
	}
}

func TestTickIncrements(t *testing.T) {
	eng := makeEngine()
	if eng.State.World.Tick != 0 {
		t.Fatalf("expected initial tick 0, got %d", eng.State.World.Tick)
	}
	eng.tick(context.Background())
	if eng.State.World.Tick != 1 {
		t.Errorf("expected tick 1, got %d", eng.State.World.Tick)
	}
}

func TestSubmitCommand(t *testing.T) {
	eng := makeEngine()
	event := models.NewEvent(models.SourceTypePlayer, "p1", models.EventTypeCommand, "bank-web", map[string]any{"action": "ISOLATE_SERVICE"})
	if err := eng.SubmitCommand(event); err != nil {
		t.Fatalf("submit command: %v", err)
	}

	eng.drainCommands()
	if len(eng.pending) != 1 {
		t.Fatalf("expected 1 pending event, got %d", len(eng.pending))
	}
}

func TestEngineStartStop(t *testing.T) {
	eng := makeEngine()
	eng.Config.TickMs = 50

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- eng.Start(ctx)
	}()

	time.Sleep(120 * time.Millisecond)
	eng.Stop()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("engine start/stop error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("engine did not stop in time")
	}

	if eng.State.World.Tick < 1 {
		t.Errorf("expected at least one tick, got %d", eng.State.World.Tick)
	}
}
