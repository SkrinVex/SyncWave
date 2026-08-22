package worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/syncwave/syncwave/internal/domain"
)

type EventType string

const (
	EventTypeProgress EventType = "progress"
	EventTypeLog      EventType = "log"
	EventTypeTrack    EventType = "track_updated"
	EventTypePlaylist EventType = "playlist_updated"
)

type EventMessage struct {
	Type EventType   `json:"type"`
	Data interface{} `json:"data"`
}

type EventHub struct {
	clients map[chan EventMessage]bool
	mu      sync.RWMutex
}

func NewEventHub() *EventHub {
	return &EventHub{
		clients: make(map[chan EventMessage]bool),
	}
}

func (h *EventHub) Subscribe() chan EventMessage {
	h.mu.Lock()
	defer h.mu.Unlock()
	ch := make(chan EventMessage, 64)
	h.clients[ch] = true
	return ch
}

func (h *EventHub) Unsubscribe(ch chan EventMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[ch]; ok {
		delete(h.clients, ch)
		close(ch)
	}
}

func (h *EventHub) Broadcast(msg EventMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.clients {
		select {
		case ch <- msg:
		default:
			// Client channel full or slow, skip to avoid blocking others
		}
	}
}

func (h *EventHub) BroadcastProgress(progress domain.SyncProgress) {
	h.Broadcast(EventMessage{
		Type: EventTypeProgress,
		Data: progress,
	})
}

func (h *EventHub) BroadcastLog(log domain.SyncLog) {
	h.Broadcast(EventMessage{
		Type: EventTypeLog,
		Data: log,
	})
}

func (h *EventHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := h.Subscribe()
	defer h.Unsubscribe(ch)

	// Send initial connected event
	fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
	flusher.Flush()

	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			bytes, err := json.Marshal(msg)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", bytes)
			flusher.Flush()
		}
	}
}
