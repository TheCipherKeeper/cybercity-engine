package ports

import (
	"context"

	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
)

// EventHandler обрабатывает события из шины.
type EventHandler func(ctx context.Context, event models.EventNode) error

// MessageBus — абстракция над Redpanda/Kafka/очередью.
type MessageBus interface {
	// Publish отправляет событие в указанный топик.
	Publish(ctx context.Context, topic string, event models.EventNode) error
	// Subscribe подписывается на топики и вызывает handler для каждого события.
	Subscribe(ctx context.Context, topics []string, handler EventHandler) error
	// Close закрывает соединение.
	Close(ctx context.Context) error
}
