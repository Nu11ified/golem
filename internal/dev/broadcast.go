package dev

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

// HMRMessage represents a message sent to connected HMR clients.
type HMRMessage struct {
	Type   string `json:"type"`
	Module string `json:"module,omitempty"`
	Error  string `json:"error,omitempty"`
}

// Broadcaster manages WebSocket connections and broadcasts HMR messages
// to all connected clients.
type Broadcaster struct {
	clients map[*websocket.Conn]struct{}
	mu      sync.RWMutex
}

// NewBroadcaster creates a new Broadcaster.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[*websocket.Conn]struct{}),
	}
}

// HandleWebSocket is an HTTP handler that upgrades connections to WebSocket
// and registers them for HMR broadcasts.
func (b *Broadcaster) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // Allow connections from any origin in dev
	})
	if err != nil {
		log.Printf("WebSocket accept error: %v", err)
		return
	}

	b.mu.Lock()
	b.clients[c] = struct{}{}
	b.mu.Unlock()

	log.Println("HMR client connected")

	// Keep connection alive; remove on disconnect
	defer func() {
		b.mu.Lock()
		delete(b.clients, c)
		b.mu.Unlock()
		c.Close(websocket.StatusNormalClosure, "connection closed")
		log.Println("HMR client disconnected")
	}()

	// Read loop to detect disconnection
	for {
		_, _, err := c.Read(r.Context())
		if err != nil {
			return
		}
	}
}

// SendReload broadcasts a full page reload message to all connected clients.
func (b *Broadcaster) SendReload() {
	b.broadcast(HMRMessage{Type: "reload"})
}

// SendModuleUpdate broadcasts a module update message to all connected clients.
func (b *Broadcaster) SendModuleUpdate(module string) {
	b.broadcast(HMRMessage{Type: "module-update", Module: module})
}

// SendError broadcasts an error message to all connected clients.
func (b *Broadcaster) SendError(msg string) {
	b.broadcast(HMRMessage{Type: "error", Error: msg})
}

func (b *Broadcaster) broadcast(msg HMRMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal HMR message: %v", err)
		return
	}

	b.mu.RLock()
	clients := make([]*websocket.Conn, 0, len(b.clients))
	for c := range b.clients {
		clients = append(clients, c)
	}
	b.mu.RUnlock()

	for _, c := range clients {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := c.Write(ctx, websocket.MessageText, data); err != nil {
			log.Printf("Failed to send HMR message to client: %v", err)
			b.mu.Lock()
			delete(b.clients, c)
			b.mu.Unlock()
		}
		cancel()
	}
}

// ClientCount returns the number of connected clients.
func (b *Broadcaster) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}
