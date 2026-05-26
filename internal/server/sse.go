package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/burak/linux-dashboard/internal/collector"
	"github.com/burak/linux-dashboard/internal/event"
)

// sseClient represents one connected SSE subscriber.
type sseClient struct {
	id     uint64
	send   chan []byte
	closed chan struct{}
}

// sseHub manages all SSE clients and broadcasts events to them.
type sseHub struct {
	mu      sync.RWMutex
	clients map[uint64]*sseClient
	nextID  uint64
}

func newSSEHub(emitter *event.Emitter) *sseHub {
	hub := &sseHub{clients: make(map[uint64]*sseClient)}
	if emitter != nil {
		emitter.Subscribe(func(eventType string, data any) {
			hub.broadcast(eventType, data)
		})
	}
	return hub
}

func (h *sseHub) broadcast(eventType string, data any) {
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients {
		select {
		case c.send <- payload:
		default:
			// Client is slow; drop the event for this client.
		}
	}
}

// Handler returns the SSE HTTP handler.
func (h *sseHub) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		// CORS headers for embedded web UI.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		client := &sseClient{
			send:   make(chan []byte, 64),
			closed: make(chan struct{}),
		}
		h.mu.Lock()
		h.nextID++
		client.id = h.nextID
		h.clients[client.id] = client
		h.mu.Unlock()

		defer func() {
			h.mu.Lock()
			delete(h.clients, client.id)
			h.mu.Unlock()
			close(client.closed)
		}()

		// Initial hello message.
		fmt.Fprintf(w, "event: hello\ndata: {\"client\":%d}\n\n", client.id)
		flusher.Flush()

		heartbeat := time.NewTicker(25 * time.Second)
		defer heartbeat.Stop()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case <-heartbeat.C:
				fmt.Fprintf(w, ": ping %d\n\n", time.Now().Unix())
				flusher.Flush()
			case payload := <-client.send:
				fmt.Fprintf(w, "event: metrics.snapshot\ndata: %s\n\n", payload)
				flusher.Flush()
			}
		}
	}
}

// ClientCount returns the number of active SSE clients.
func (h *sseHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// handleSSE is the HTTP handler for the SSE stream endpoint.
func handleSSE(emitter *event.Emitter, col *collector.Manager) http.HandlerFunc {
	hub := newSSEHub(emitter)
	return hub.Handler()
}