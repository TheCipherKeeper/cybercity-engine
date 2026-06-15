// Package application связывает adapters и domain engine в рабочее
// приложение. Это composition root: здесь создаются все конкретные
// зависимости и передаются в domain.
package application

import (
	"context"
	"fmt"

	"github.com/TheCipherKeeper/cybercity-engine/internal/adapters/api"
	"github.com/TheCipherKeeper/cybercity-engine/internal/adapters/loader"
	"github.com/TheCipherKeeper/cybercity-engine/internal/adapters/memory"
	"github.com/TheCipherKeeper/cybercity-engine/internal/config"
	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/engine"
	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
)

// Runtime содержит собранное приложение.
type Runtime struct {
	Engine *engine.Engine
	Server *api.Server
	Bus    *memory.MessageBus
}

// NewRuntime собирает runtime из конфигурации.
func NewRuntime(cfg config.EngineConfig) (*Runtime, error) {
	topologyLoader := loader.NewLoader()
	ctx := context.Background()

	source := cfg.EngineZipURL
	if cfg.EngineZipPath != "" {
		source = cfg.EngineZipPath
	}
	if source == "" {
		source = cfg.EngineZipURL
	}

	topology, err := topologyLoader.Load(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("load topology: %w", err)
	}

	store := memory.NewEventStore(10000)
	snapshots := memory.NewSnapshotStore()
	bus := memory.NewMessageBus()
	broadcaster := memory.NewBroadcaster()

	domainCfg := engine.Config{
		TickMs:                cfg.TickMs,
		SnapshotIntervalTicks: cfg.SnapshotIntervalTicks,
	}

	eng := engine.NewEngine(
		topology,
		domainCfg,
		nil,
		store,
		snapshots,
		bus,
		broadcaster,
	)

	server := api.NewServer(eng, cfg.Host, cfg.Port)

	// Подписываем WebSocket broadcast на снапшоты.
	broadcaster.Subscribe(func(state *models.WorldState) {
		server.Broadcast(state)
	})

	return &Runtime{
		Engine: eng,
		Server: server,
		Bus:    bus,
	}, nil
}
