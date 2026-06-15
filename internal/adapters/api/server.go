// Package api предоставляет HTTP и WebSocket endpoints для взаимодействия
// с движком. Это adapter: он не содержит бизнес-логики, только преобразует
// вход в события и отдаёт состояние.
package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/engine"
	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
)

// Server — HTTP/WebSocket сервер движка.
type Server struct {
	engine   *engine.Engine
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
	mu       sync.RWMutex
	host     string
	port     int
}

// NewServer создаёт новый сервер, привязанный к экземпляру движка.
func NewServer(e *engine.Engine, host string, port int) *Server {
	return &Server{
		engine: e,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients: make(map[*websocket.Conn]bool),
		host:    host,
		port:    port,
	}
}

// Handler возвращает http.Handler для регистрации в http.Server.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/state", s.handleState)
	mux.HandleFunc("/topology", s.handleTopology)
	mux.HandleFunc("/events/recent", s.handleRecentEvents)
	mux.HandleFunc("/commands", s.handleCommand)
	mux.HandleFunc("/ws", s.handleWebSocket)
	return mux
}

// Run запускает HTTP сервер и ждёт завершения контекста.
func (s *Server) Run(ctx context.Context) error {
	addr := s.host + ":" + strconv.Itoa(s.port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: s.Handler(),
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	slog.Info("starting http server", "addr", addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"tick":     s.engine.State.World.Tick,
		"services": len(s.engine.State.World.Services),
	})
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.engine.State.World)
}

func (s *Server) handleTopology(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.engine.Topology)
}

func (s *Server) handleRecentEvents(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
			limit = n
		}
	}
	recent, err := s.engine.Store.Recent(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, recent)
}

func (s *Server) handleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var cmd map[string]any
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	event := commandEvent(cmd)
	if err := s.engine.SubmitCommand(event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{
		"status":   "queued",
		"event_id": event.EventID,
	})
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
	}()

	if err := s.sendSnapshot(conn); err != nil {
		slog.Error("websocket snapshot failed", "error", err)
		return
	}

	for {
		var msg map[string]any
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("websocket read error", "error", err)
			}
			return
		}
		event := commandEvent(msg)
		if err := s.engine.SubmitCommand(event); err != nil {
			_ = conn.WriteJSON(map[string]any{
				"type":   "COMMAND_RESULT",
				"status": "REJECTED",
				"error":  err.Error(),
			})
			continue
		}
		_ = conn.WriteJSON(map[string]any{
			"type":     "COMMAND_RESULT",
			"status":   "ACCEPTED",
			"event_id": event.EventID,
		})
	}
}

// Broadcast отправляет снапшот всем WebSocket-клиентам.
func (s *Server) Broadcast(state *models.WorldState) {
	s.mu.RLock()
	clients := make([]*websocket.Conn, 0, len(s.clients))
	for c := range s.clients {
		clients = append(clients, c)
	}
	s.mu.RUnlock()

	msg := map[string]any{
		"type": "SNAPSHOT",
		"data": state,
	}
	for _, c := range clients {
		if err := c.WriteJSON(msg); err != nil {
			slog.Warn("websocket broadcast failed", "error", err)
		}
	}
}

func (s *Server) sendSnapshot(conn *websocket.Conn) error {
	return conn.WriteJSON(map[string]any{
		"type": "SNAPSHOT",
		"data": s.engine.State.World,
	})
}

func commandEvent(cmd map[string]any) models.EventNode {
	playerID := "anonymous"
	if p, ok := cmd["player_id"].(string); ok && p != "" {
		playerID = p
	}
	targetID, _ := cmd["target"].(string)
	action, _ := cmd["action"].(string)
	params := make(map[string]any)
	if p, ok := cmd["params"].(map[string]any); ok {
		params = p
	}
	return models.NewEvent(
		models.SourceTypePlayer,
		playerID,
		models.EventTypeCommand,
		targetID,
		map[string]any{
			"action": action,
			"params": params,
		},
	)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
