package ports

import (
	"context"

	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
)

// SnapshotRepository хранит periodic snapshots WorldState.
type SnapshotRepository interface {
	// Save сохраняет снапшот мира.
	Save(ctx context.Context, tick int, state *models.WorldState) error
	// LoadLatest возвращает последний сохранённый снапшот или nil.
	LoadLatest(ctx context.Context) (*models.WorldState, error)
}
