package dev

import (
	"encoding/json"
	"sync"
)

// Broadcaster manages WebSocket client connections and message broadcasting.
// It maintains a set of channels, one per connected client, and provides
// methods to send reload or error messages to all connected clients.
type Broadcaster struct {
	clients map[int]chan string
	mutex   sync.RWMutex
	counter int
}

// NewBroadcaster creates a new Broadcaster with an empty client set.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[int]chan string),
	}
}

// AddClient registers a new client channel and returns a unique client ID.
// The caller should read from the returned channel to receive broadcast messages.
func (b *Broadcaster) AddClient(ch chan string) int {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.counter++
	id := b.counter
	b.clients[id] = ch
	return id
}

// RemoveClient unregisters the client with the given ID.
func (b *Broadcaster) RemoveClient(id int) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	delete(b.clients, id)
}

// ClientCount returns the number of currently connected clients.
func (b *Broadcaster) ClientCount() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return len(b.clients)
}

// SendReload sends the literal string "reload" to all connected clients.
func (b *Broadcaster) SendReload() {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	for _, ch := range b.clients {
		select {
		case ch <- "reload":
		default:
			// Skip slow clients to avoid blocking.
		}
	}
}

// SendError sends a JSON-encoded error message to all connected clients.
// The message format is: {"type":"error","message":"<msg>"}.
func (b *Broadcaster) SendError(msg string) {
	payload, _ := json.Marshal(map[string]string{
		"type":    "error",
		"message": msg,
	})
	message := string(payload)

	b.mutex.RLock()
	defer b.mutex.RUnlock()
	for _, ch := range b.clients {
		select {
		case ch <- message:
		default:
			// Skip slow clients to avoid blocking.
		}
	}
}
