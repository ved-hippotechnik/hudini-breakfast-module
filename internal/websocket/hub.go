package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, implement proper origin checking
		return true
	},
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Clients grouped by room
	rooms map[string]map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe operations
	mutex sync.RWMutex
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("Client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mutex.Unlock()
			log.Printf("Client disconnected. Total clients: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastRoomUpdate sends room update to all connected clients
func (h *Hub) BroadcastRoomUpdate(propertyID, roomNumber string, data interface{}) {
	message := map[string]interface{}{
		"type":        "room_update",
		"property_id": propertyID,
		"room_number": roomNumber,
		"data":        data,
		"timestamp":   json.Number("1626451200"), // Current timestamp
	}

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling room update: %v", err)
		return
	}

	select {
	case h.broadcast <- jsonMessage:
	default:
		log.Println("Broadcast channel is full, dropping message")
	}
}

// BroadcastBreakfastUpdate sends breakfast consumption update to all connected clients
func (h *Hub) BroadcastBreakfastUpdate(propertyID, roomNumber string, consumed bool, consumedBy string) {
	message := map[string]interface{}{
		"type":        "breakfast_update",
		"property_id": propertyID,
		"room_number": roomNumber,
		"consumed":    consumed,
		"consumed_by": consumedBy,
		"timestamp":   json.Number("1626451200"), // Current timestamp
	}

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling breakfast update: %v", err)
		return
	}

	select {
	case h.broadcast <- jsonMessage:
	default:
		log.Println("Broadcast channel is full, dropping message")
	}
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// BroadcastToAll broadcasts a message to all connected clients
func (h *Hub) BroadcastToAll(data map[string]interface{}) {
	jsonMessage, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	select {
	case h.broadcast <- jsonMessage:
	default:
		log.Println("Broadcast channel is full, dropping message")
	}
}

// BroadcastToRoom broadcasts a message to all clients in a specific room
func (h *Hub) BroadcastToRoom(room string, data map[string]interface{}) {
	h.mutex.RLock()
	roomClients, exists := h.rooms[room]
	h.mutex.RUnlock()
	
	if !exists || len(roomClients) == 0 {
		return
	}
	
	jsonMessage, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}
	
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	for client := range roomClients {
		select {
		case client.send <- jsonMessage:
		default:
			// Client's send channel is full, close it
			close(client.send)
			delete(h.clients, client)
			delete(roomClients, client)
		}
	}
}

// AddClientToRoom adds a client to a specific room
func (h *Hub) AddClientToRoom(client *Client, room string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*Client]bool)
	}
	h.rooms[room][client] = true
}

// RemoveClientFromRoom removes a client from a specific room
func (h *Hub) RemoveClientFromRoom(client *Client, room string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	if roomClients, exists := h.rooms[room]; exists {
		delete(roomClients, client)
		if len(roomClients) == 0 {
			delete(h.rooms, room)
		}
	}
}

// ServeWS handles websocket requests from the peer
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines
	go client.writePump()
	go client.readPump()
}
