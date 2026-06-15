// Package ports определяет интерфейсы, через которые domain-layer движка
// взаимодействует с внешним миром. Все реализации находятся в adapters.
package ports

import (
	"context"

	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
)

// EventStore хранит событийный граф.
type EventStore interface {
	// Add сохраняет событие и создаёт causal-рёбра к родителям.
	Add(ctx context.Context, event models.EventNode) error
	// Link связывает два события неродительской связью.
	Link(ctx context.Context, sourceEventID, targetEventID string, kind models.EventEdgeKind) error
	// Get возвращает событие по id.
	Get(ctx context.Context, eventID string) (models.EventNode, bool)
	// Recent возвращает последние limit событий.
	Recent(ctx context.Context, limit int) ([]models.EventNode, error)
	// Lineage возвращает всех предков события (включая его) от старого к новому.
	Lineage(ctx context.Context, eventID string) ([]models.EventNode, error)
	// Len возвращает число узлов в графе.
	Len(ctx context.Context) int
}
