package ports

import (
	"context"

	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
)

// ServiceAgent — внешний драйвер runtime-цели (vm или container): движок
// посылает ему события, он возвращает наблюдаемые исходы (heartbeat,
// compromise, scan response) как подписанные события от коллектора.
//
// Движок — регистратор, не симулятор: он не вычисляет исходы сам. lite-стабам
// (лёгким stub-контейнерам) отдельный агент не нужен — они наблюдаются
// коллектором наравне с vm/container. honeypot — флаг назначения-наживки
// (свойство сервиса в cybercity-data), а не runtime-вид, агентом не
// управляется. runtime_kind {vm, container, lite} назначается в
// cybercity-manage (deployment-time), см. umbrella ADR-0004.
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
