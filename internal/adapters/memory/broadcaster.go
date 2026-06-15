package memory

import (
	"context"
	"sync"

	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
)

// Broadcaster — in-memory реализация StateBroadcaster.
type Broadcaster struct {
	mu        sync.RWMutex
	listeners []func(*models.WorldState)
}

// NewBroadcaster создаёт in-memory broadcaster.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{}
}

// Subscribe регистрирует callback для получения снапшотов.
func (b *Broadcaster) Subscribe(fn func(*models.WorldState)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.listeners = append(b.listeners, fn)
}

// Broadcast рассылает снапшот подписчикам.
func (b *Broadcaster) Broadcast(ctx context.Context, state *models.WorldState) error {
	b.mu.RLock()
	listeners := make([]func(*models.WorldState), len(b.listeners))
	copy(listeners, b.listeners)
	b.mu.RUnlock()

	for _, fn := range listeners {
		fn(state)
	}
	return nil
}
