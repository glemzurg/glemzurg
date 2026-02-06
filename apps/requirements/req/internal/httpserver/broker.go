// Package httpserver provides an HTTP server for serving model documentation.
package httpserver

import "sync"

// Broker manages Server-Sent Events (SSE) for notifying clients of changes.
type Broker struct {
	notifier       chan []byte          // Channel to send notifications to all clients.
	newClients     chan chan []byte     // Channel to register new client channels.
	closingClients chan chan []byte     // Channel to unregister closing client channels.
	clients        map[chan []byte]bool // Map of active client channels for broadcasting.
}

// NewBroker creates and initializes a new broker.
func NewBroker() *Broker {
	b := &Broker{
		notifier:       make(chan []byte, 1),
		newClients:     make(chan chan []byte),
		closingClients: make(chan chan []byte),
		clients:        make(map[chan []byte]bool),
	}
	go b.run()
	return b
}

// run is the main loop for the broker to handle client registration,
// unregistration, and broadcasting.
func (b *Broker) run() {
	for {
		select {
		case client := <-b.newClients:
			b.clients[client] = true
		case client := <-b.closingClients:
			delete(b.clients, client)
			close(client)
		case msg := <-b.notifier:
			for client := range b.clients {
				select {
				case client <- msg:
				default: // Skip if the client channel is full.
				}
			}
		}
	}
}

// Notify sends a notification message to all connected clients.
func (b *Broker) Notify(msg []byte) {
	select {
	case b.notifier <- msg:
	default:
	}
}

// Register adds a new client channel to the broker.
func (b *Broker) Register(client chan []byte) {
	b.newClients <- client
}

// Unregister removes a client channel from the broker.
func (b *Broker) Unregister(client chan []byte) {
	b.closingClients <- client
}

// BrokerRegistry manages brokers keyed by model/file path.
type BrokerRegistry struct {
	brokers sync.Map
}

// NewBrokerRegistry creates a new broker registry.
func NewBrokerRegistry() *BrokerRegistry {
	return &BrokerRegistry{}
}

// GetBroker retrieves or creates a broker for a given key (model/file.md).
func (r *BrokerRegistry) GetBroker(key string) *Broker {
	val, _ := r.brokers.LoadOrStore(key, NewBroker())
	return val.(*Broker)
}

// NotifyAll sends a refresh notification to all brokers.
func (r *BrokerRegistry) NotifyAll() {
	r.brokers.Range(func(key, value interface{}) bool {
		value.(*Broker).Notify([]byte("refresh"))
		return true
	})
}

// NotifyModel sends a refresh notification to all brokers for a specific model.
func (r *BrokerRegistry) NotifyModel(model string) {
	r.brokers.Range(func(key, value interface{}) bool {
		keyStr := key.(string)
		if len(keyStr) > len(model) && keyStr[:len(model)+1] == model+"/" {
			value.(*Broker).Notify([]byte("refresh"))
		}
		return true
	})
}
