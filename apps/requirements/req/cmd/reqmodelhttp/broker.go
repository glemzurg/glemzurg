// broker.go
package main

import "sync"

type broker struct { // Defines a broker for Server-Sent Events (SSE) to notify clients of changes.
	notifier       chan []byte          // Channel to send notifications (e.g., "refresh") to all clients.
	newClients     chan chan []byte     // Channel to register new client channels.
	closingClients chan chan []byte     // Channel to unregister closing client channels.
	clients        map[chan []byte]bool // Map of active client channels for broadcasting.
}

func newBroker() *broker { // Factory function to create and initialize a new broker.
	b := &broker{ // Initializes the broker struct.
		notifier:       make(chan []byte, 1),       // Buffered channel for notifications to prevent blocking.
		newClients:     make(chan chan []byte),     // Unbuffered channel for adding clients.
		closingClients: make(chan chan []byte),     // Unbuffered channel for removing clients.
		clients:        make(map[chan []byte]bool), // Empty map to hold client channels.
	}
	go b.run() // Starts the broker's event loop in a separate goroutine.
	return b
}

func (b *broker) run() { // Main loop for the broker to handle client registration, unregistration, and broadcasting.
	for { // Infinite loop to keep the broker running.
		select { // Multiplexes over channels for non-blocking operations.
		case client := <-b.newClients: // Handles new client registration.
			b.clients[client] = true // Adds the client to the map.
		case client := <-b.closingClients: // Handles client unregistration.
			delete(b.clients, client) // Removes the client from the map.
			close(client)             // Closes the client's channel to signal end.
		case msg := <-b.notifier: // Handles incoming notifications to broadcast.
			for client := range b.clients { // Iterates over all clients.
				select { // Non-blocking send to avoid hanging on slow clients.
				case client <- msg: // Sends the message to the client.
				default: // Skips if the client channel is full.
				}
			}
		}
	}
}

var brokers sync.Map // Thread-safe map to store brokers keyed by model/file.md, allowing concurrent access.

func getBroker(key string) *broker { // Retrieves or creates a broker for a given key (model/file.md).
	val, _ := brokers.LoadOrStore(key, newBroker()) // Loads existing or stores new broker atomically.
	return val.(*broker)                            // Casts and returns the broker.
}
