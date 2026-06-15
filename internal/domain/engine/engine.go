// Package engine содержит основной tick-цикл и обработчик событий.
// Engine — единственный мутатор runtime-состояния.
package engine

import (
	"context"
	"sync"
	"time"

	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/ports"
	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/router"
	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/state"
)

// Config — параметры, необходимые domain engine.
type Config struct {
	TickMs                int
	SnapshotIntervalTicks int
}

// HandlerResult — результат обработки события.
type HandlerResult struct {
	Children     []models.EventNode
	StateChanges []models.EventNode
}

// Handler — функция обработки события.
type Handler func(e *Engine, event models.EventNode) HandlerResult

// Engine — CyberCity simulation engine.
type Engine struct {
	Topology    *models.TopologyGraph
	Config      Config
	State       *state.StateManager
	Router      *router.EventRouter
	Store       ports.EventStore
	Snapshots   ports.SnapshotRepository
	Bus         ports.MessageBus
	Broadcaster ports.StateBroadcaster

	handlers  map[models.EventType]Handler
	commandCh chan models.EventNode
	pending   []models.EventNode
	running   bool
	mu        sync.Mutex
	stopCh    chan struct{}
}

// NewEngine создаёт новый экземпляр движка. Все зависимости передаются
// извне (dependency injection из application layer).
func NewEngine(
	topology *models.TopologyGraph,
	cfg Config,
	routerInst *router.EventRouter,
	store ports.EventStore,
	snapshots ports.SnapshotRepository,
	bus ports.MessageBus,
	broadcaster ports.StateBroadcaster,
) *Engine {
	if routerInst == nil {
		routerInst = router.NewEventRouter()
	}
	return &Engine{
		Topology:    topology,
		Config:      cfg,
		State:       state.NewStateManager(topology),
		Router:      routerInst,
		Store:       store,
		Snapshots:   snapshots,
		Bus:         bus,
		Broadcaster: broadcaster,
		handlers:    defaultHandlers(),
		commandCh:   make(chan models.EventNode, 256),
		pending:     make([]models.EventNode, 0),
		stopCh:      make(chan struct{}),
	}
}

// Start запускает tick-цикл движка. Блокирует до вызова Stop.
func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	e.running = true
	e.mu.Unlock()

	ticker := time.NewTicker(time.Duration(e.Config.TickMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			e.Stop()
			return ctx.Err()
		case <-e.stopCh:
			return nil
		case cmd := <-e.commandCh:
			e.mu.Lock()
			e.pending = append(e.pending, cmd)
			e.mu.Unlock()
		case <-ticker.C:
			if err := e.tick(ctx); err != nil {
				return err
			}
		}
	}
}

// Stop останавливает tick-цикл.
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.running {
		return
	}
	e.running = false
	close(e.stopCh)
}

// SubmitCommand отправляет команду в очередь движка.
func (e *Engine) SubmitCommand(command models.EventNode) error {
	select {
	case e.commandCh <- command:
	default:
	}
	return nil
}

// IsRunning возвращает true, если движок запущен.
func (e *Engine) IsRunning() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.running
}

func (e *Engine) tick(ctx context.Context) error {
	e.State.IncrementTick()
	e.drainCommands()

	for len(e.pending) > 0 {
		event := e.pending[0]
		e.pending = e.pending[1:]
		e.ProcessEvent(ctx, &event)
	}

	if e.Broadcaster != nil && e.State.World.Tick%e.Config.SnapshotIntervalTicks == 0 {
		_ = e.Broadcaster.Broadcast(ctx, e.State.World)
	}

	if e.Config.SnapshotIntervalTicks > 0 && e.State.World.Tick%e.Config.SnapshotIntervalTicks == 0 {
		_ = e.Snapshots.Save(ctx, e.State.World.Tick, e.State.World)
	}

	return nil
}

func (e *Engine) drainCommands() {
	for {
		select {
		case cmd := <-e.commandCh:
			e.pending = append(e.pending, cmd)
		default:
			return
		}
	}
}

// ProcessEvent обрабатывает одно событие и распространяет его эффекты.
// Экспортируется для тестов.
func (e *Engine) ProcessEvent(ctx context.Context, event *models.EventNode) {
	_ = e.Store.Add(ctx, *event)

	handler := e.handlers[event.EventType]
	if handler == nil {
		event.Status = models.EventStatusSuppressed
		_ = e.Store.Add(ctx, *event)
		return
	}

	result := handler(e, *event)

	for _, child := range result.Children {
		_ = e.Store.Add(ctx, child)
		_ = e.Store.Link(ctx, event.EventID, child.EventID, models.EventEdgeKindPropagatedTo)
		e.pending = append(e.pending, child)
		if e.Bus != nil {
			_ = e.Bus.Publish(ctx, "city.events", child)
		}
	}

	for _, change := range result.StateChanges {
		_ = e.Store.Add(ctx, change)
		_ = e.Store.Link(ctx, event.EventID, change.EventID, models.EventEdgeKindTriggeredRule)
		if e.Bus != nil {
			_ = e.Bus.Publish(ctx, "city.events", change)
		}
		if change.TargetID != "" {
			sourceNode := e.Topology.Services[change.TargetID]
			if sourceNode != nil {
				propagated := e.Router.Propagate(change, sourceNode, e.Topology)
				e.pending = append(e.pending, propagated...)
			}
		}
	}

	event.Status = models.EventStatusProcessed
	_ = e.Store.Add(ctx, *event)
	if e.Bus != nil {
		_ = e.Bus.Publish(ctx, "city.events", *event)
	}
}
