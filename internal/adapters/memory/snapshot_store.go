package memory

import (
	"context"
	"sync"

	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
)

// SnapshotStore — in-memory реализация SnapshotRepository.
type SnapshotStore struct {
	mu         sync.RWMutex
	latest     *models.WorldState
	latestTick int
}

// NewSnapshotStore создаёт in-memory snapshot store.
func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{}
}

// Save сохраняет снапшот мира.
func (s *SnapshotStore) Save(ctx context.Context, tick int, state *models.WorldState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// shallow-копия для безопасности; в реальной БД сериализуем JSON.
	copyState := *state
	s.latest = &copyState
	s.latestTick = tick
	return nil
}

// LoadLatest возвращает последний сохранённый снапшот.
func (s *SnapshotStore) LoadLatest(ctx context.Context) (*models.WorldState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latest, nil
}
