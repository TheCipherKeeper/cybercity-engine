package memory

import (
	"context"
	"sync"

	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/ports"
)

// MessageBus — in-memory шина для unit-тестов и локальной разработки.
type MessageBus struct {
	mu       sync.RWMutex
	handlers map[string][]ports.EventHandler
	closed   bool
}

// NewMessageBus создаёт in-memory шину.
func NewMessageBus() *MessageBus {
	return &MessageBus{
		handlers: make(map[string][]ports.EventHandler),
	}
}

// Publish доставляет событие всем подписчикам топика.
func (b *MessageBus) Publish(ctx context.Context, topic string, event models.EventNode) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.closed {
		return nil
	}
	for _, h := range b.handlers[topic] {
		_ = h(ctx, event)
	}
	return nil
}

// Subscribe регистрирует handler на топики.
func (b *MessageBus) Subscribe(ctx context.Context, topics []string, handler ports.EventHandler) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, topic := range topics {
		b.handlers[topic] = append(b.handlers[topic], handler)
	}
	return nil
}

// Close помечает шину закрытой.
func (b *MessageBus) Close(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
	return nil
}
