// Package memory содержит in-memory реализации портов для тестов и демо.
package memory

import (
	"container/list"
	"context"
	"sync"

	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
)

// EventStore — in-memory хранилище событийного графа.
type EventStore struct {
	mu          sync.RWMutex
	nodes       map[string]models.EventNode
	edges       []models.EventEdge
	recent      *list.List
	recentIndex map[string]*list.Element
	maxRecent   int
}

// NewEventStore создаёт новый in-memory store.
func NewEventStore(maxRecent int) *EventStore {
	if maxRecent <= 0 {
		maxRecent = 10000
	}
	return &EventStore{
		nodes:       make(map[string]models.EventNode),
		edges:       make([]models.EventEdge, 0),
		recent:      list.New(),
		recentIndex: make(map[string]*list.Element),
		maxRecent:   maxRecent,
	}
}

// Add сохраняет событие и создаёт causal-рёбра.
func (s *EventStore) Add(ctx context.Context, event models.EventNode) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nodes[event.EventID] = event
	s.promoteRecent(event.EventID)

	for _, parentID := range event.ParentEventIDs {
		s.edges = append(s.edges, models.EventEdge{
			SourceEventID: parentID,
			TargetEventID: event.EventID,
			Kind:          models.EventEdgeKindCausedBy,
		})
	}
	return nil
}

// Link связывает два события.
func (s *EventStore) Link(ctx context.Context, sourceEventID, targetEventID string, kind models.EventEdgeKind) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.edges = append(s.edges, models.EventEdge{
		SourceEventID: sourceEventID,
		TargetEventID: targetEventID,
		Kind:          kind,
	})
	return nil
}

// Get возвращает событие по id.
func (s *EventStore) Get(ctx context.Context, eventID string) (models.EventNode, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	event, ok := s.nodes[eventID]
	return event, ok
}

// Recent возвращает последние limit событий.
func (s *EventStore) Recent(ctx context.Context, limit int) ([]models.EventNode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		return nil, nil
	}
	var ids []string
	for e := s.recent.Front(); e != nil; e = e.Next() {
		ids = append(ids, e.Value.(string))
	}
	if len(ids) > limit {
		ids = ids[len(ids)-limit:]
	}

	out := make([]models.EventNode, 0, len(ids))
	for _, id := range ids {
		if event, ok := s.nodes[id]; ok {
			out = append(out, event)
		}
	}
	return out, nil
}

// Lineage возвращает всех предков события от старого к новому.
func (s *EventStore) Lineage(ctx context.Context, eventID string) ([]models.EventNode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	seen := make(map[string]bool)
	var order []string
	stack := []string{eventID}

	for len(stack) > 0 {
		currentID := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[currentID] {
			continue
		}
		seen[currentID] = true
		order = append(order, currentID)
		if current, ok := s.nodes[currentID]; ok {
			for _, parentID := range current.ParentEventIDs {
				if !seen[parentID] {
					stack = append(stack, parentID)
				}
			}
		}
	}

	out := make([]models.EventNode, 0, len(order))
	for i := len(order) - 1; i >= 0; i-- {
		if event, ok := s.nodes[order[i]]; ok {
			out = append(out, event)
		}
	}
	return out, nil
}

// Len возвращает число узлов.
func (s *EventStore) Len(ctx context.Context) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.nodes)
}

func (s *EventStore) promoteRecent(eventID string) {
	if el, ok := s.recentIndex[eventID]; ok {
		s.recent.MoveToBack(el)
		return
	}
	el := s.recent.PushBack(eventID)
	s.recentIndex[eventID] = el

	for s.recent.Len() > s.maxRecent {
		front := s.recent.Front()
		if front == nil {
			break
		}
		id := front.Value.(string)
		s.recent.Remove(front)
		delete(s.recentIndex, id)
	}
}
