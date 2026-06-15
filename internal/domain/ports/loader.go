package ports

import (
	"context"

	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
)

// TopologyLoader загружает топологический граф из источника (файл, URL, S3).
// Порт возвращает уже готовый domain TopologyGraph; адаптер отвечает за
// парсинг конкретного формата артефакта.
type TopologyLoader interface {
	// Load возвращает TopologyGraph или ошибку.
	// source может быть путём, URL или иным идентификатором.
	Load(ctx context.Context, source string) (*models.TopologyGraph, error)
}
