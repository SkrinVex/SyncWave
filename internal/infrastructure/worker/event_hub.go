package worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/syncwave/syncwave/internal/domain"
)

type EventType string

const (
	EventTypeProgress     EventType = "progress"
	EventTypeLog          EventType = "log"
	EventTypeTrack        EventType = "track_updated"
	EventTypePlaylist     EventType = "playlist_updated"
	EventTypeCookieStatus EventType = "cookie_status"
)

type EventMessage struct {
	Type EventType   `json:"type"`
	Data interface{} `json:"data"`
}

type EventHub struct {
	clients map[chan EventMessage]string
	mu      sync.RWMutex
}

func NewEventHub() *EventHub {
	return &EventHub{
		clients: make(map[chan EventMessage]string),
	}
}

func (h *EventHub) Subscribe(userID string) chan EventMessage {
	h.mu.Lock()
	defer h.mu.Unlock()
	ch := make(chan EventMessage, 128)
	h.clients[ch] = userID
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
		}
	}
}

func (h *EventHub) BroadcastUser(userID string, msg EventMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch, clientUserID := range h.clients {
		if userID == "" || clientUserID == "" || clientUserID == userID {
			select {
			case ch <- msg:
			default:
			}
		}
	}
}

func (h *EventHub) BroadcastProgress(userID string, progress domain.SyncProgress) {
	h.BroadcastUser(userID, EventMessage{
		Type: EventTypeProgress,
		Data: progress,
	})
}

func (h *EventHub) BroadcastLog(userID string, log domain.SyncLog) {
	h.BroadcastUser(userID, EventMessage{
		Type: EventTypeLog,
		Data: log,
	})
}

func (h *EventHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		if u, ok2 := w.(interface{ Unwrap() http.ResponseWriter }); ok2 {
			flusher, ok = u.Unwrap().(http.Flusher)
		}
	}
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	userID, _ := r.Context().Value("user_id").(string)

	ch := h.Subscribe(userID)
	defer h.Unsubscribe(ch)

	// Send initial connected payload
	fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")
	flusher.Flush()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		case msg, ok := <-ch:
			if !ok {
				return
			}
			bytes, err := json.Marshal(msg)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", bytes)
			flusher.Flush()
		}
	}
}
