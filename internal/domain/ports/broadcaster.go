package ports

import (
	"context"

	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
)

// StateBroadcaster рассылает runtime-состояние подписчикам (WebSocket, UI,
// scenario manager).
type StateBroadcaster interface {
	// Broadcast отправляет снапшот мира подписчикам.
	Broadcast(ctx context.Context, state *models.WorldState) error
}
