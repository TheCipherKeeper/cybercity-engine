package ports

import (
	"context"

	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
)

// ServiceAgent представляет внешнего агента, управляющего real или decoy
// сервисом. Движок посылает агенту события, агент возвращает результаты
// (heartbeat, compromise, scan response).
type ServiceAgent interface {
	// ServiceID возвращает id сервиса, за который отвечает агент.
	ServiceID() string
	// ApplyEvent применяет к агенту событие, сгенерированное движком.
	ApplyEvent(ctx context.Context, event models.EventNode) error
	// Heartbeat возвращает heartbeat-событие или nil, если не пришло.
	Heartbeat(ctx context.Context) (*models.EventNode, error)
	// Close завершает работу агента.
	Close(ctx context.Context) error
}
