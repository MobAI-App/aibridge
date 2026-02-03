package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/MobAI-App/aibridge/internal/bridge"
)

const Version = "1.0.0"

type Handlers struct {
	bridge *bridge.Bridge
}

func NewHandlers(b *bridge.Bridge) *Handlers {
	return &Handlers{bridge: b}
}

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{
		Status:  "ok",
		Version: Version,
	})
}

type StatusResponse struct {
	Idle          bool    `json:"idle"`
	QueueLength   int     `json:"queue_length"`
	ChildRunning  bool    `json:"child_running"`
	ChildTool     string  `json:"child_tool"`
	UptimeSeconds float64 `json:"uptime_seconds"`
}

func (h *Handlers) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, StatusResponse{
		Idle:          h.bridge.IsIdle(),
		QueueLength:   h.bridge.Queue().Len(),
		ChildRunning:  h.bridge.IsChildRunning(),
		ChildTool:     h.bridge.ToolName(),
		UptimeSeconds: h.bridge.UptimeSeconds(),
	})
}

type InjectRequest struct {
	Text     string `json:"text"`
	Priority bool   `json:"priority"`
}

type InjectResponse struct {
	ID       string `json:"id"`
	Queued   bool   `json:"queued"`
	Position int    `json:"position"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *Handlers) Inject(w http.ResponseWriter, r *http.Request) {
	if !h.bridge.IsChildRunning() {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "child process not running"})
		return
	}

	var req InjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON"})
		return
	}

	if req.Text == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "text is required"})
		return
	}

	syncMode := r.URL.Query().Get("sync") == "true"

	inj, err := h.bridge.Queue().EnqueueWithChan(req.Text, req.Priority, syncMode)
	if err == bridge.ErrQueueFull {
		writeJSON(w, http.StatusTooManyRequests, ErrorResponse{Error: "queue is full"})
		return
	}

	h.bridge.NotifyEnqueue()

	if syncMode {
		timeout := 300 * time.Second
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		select {
		case <-inj.SyncChan:
			writeJSON(w, http.StatusOK, InjectResponse{
				ID:       inj.ID,
				Queued:   true,
				Position: 0,
			})
		case <-ctx.Done():
			writeJSON(w, http.StatusRequestTimeout, ErrorResponse{Error: "injection timeout"})
		}
		return
	}

	writeJSON(w, http.StatusOK, InjectResponse{
		ID:       inj.ID,
		Queued:   true,
		Position: h.bridge.Queue().Len(),
	})
}

type QueueClearResponse struct {
	Cleared int `json:"cleared"`
}

func (h *Handlers) QueueClear(w http.ResponseWriter, r *http.Request) {
	count := h.bridge.Queue().Clear()
	writeJSON(w, http.StatusOK, QueueClearResponse{Cleared: count})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
