package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"hudini-breakfast-module/internal/cache"
	"hudini-breakfast-module/internal/logging"
	
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	// Redis pub/sub channel for WebSocket messages
	WebSocketChannel = "hudini:websocket:broadcast"
	// Channel for room status updates
	RoomUpdateChannel = "hudini:room:updates"
)

// DistributedHub manages WebSocket connections across multiple servers
type DistributedHub struct {
	hub      *Hub
	cache    *cache.RedisCache
	pubsub   *redis.PubSub
	serverID string
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// Message represents a WebSocket message that can be distributed
type DistributedMessage struct {
	ServerID  string                 `json:"server_id"`
	Type      string                 `json:"type"`
	Room      string                 `json:"room,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewDistributedHub creates a new distributed WebSocket hub
func NewDistributedHub(hub *Hub, cache *cache.RedisCache, serverID string) *DistributedHub {
	ctx, cancel := context.WithCancel(context.Background())
	
	dh := &DistributedHub{
		hub:      hub,
		cache:    cache,
		serverID: serverID,
		ctx:      ctx,
		cancel:   cancel,
	}
	
	// Subscribe to Redis channels
	dh.pubsub = cache.Subscribe(ctx, WebSocketChannel, RoomUpdateChannel)
	
	// Start listening for messages
	go dh.listenForMessages()
	
	return dh
}

// listenForMessages listens for messages from Redis and broadcasts to local clients
func (dh *DistributedHub) listenForMessages() {
	ch := dh.pubsub.Channel()
	
	for {
		select {
		case <-dh.ctx.Done():
			return
		case msg := <-ch:
			if msg == nil {
				continue
			}
			
			var distMsg DistributedMessage
			if err := json.Unmarshal([]byte(msg.Payload), &distMsg); err != nil {
				logging.WithError(err).Error("Failed to unmarshal distributed message")
				continue
			}
			
			// Don't process our own messages
			if distMsg.ServerID == dh.serverID {
				continue
			}
			
			// Broadcast to local clients based on message type
			dh.handleDistributedMessage(&distMsg)
		}
	}
}

// handleDistributedMessage handles incoming distributed messages
func (dh *DistributedHub) handleDistributedMessage(msg *DistributedMessage) {
	switch msg.Type {
	case "room_update":
		// Broadcast room update to all local clients
		dh.hub.BroadcastToAll(msg.Data)
		
	case "consumption_update":
		// Broadcast consumption update to specific room
		if msg.Room != "" {
			dh.hub.BroadcastToRoom(msg.Room, msg.Data)
		}
		
	case "analytics_update":
		// Broadcast analytics update to all clients
		dh.hub.BroadcastToAll(msg.Data)
		
	default:
		logging.WithFields(logrus.Fields{
			"type":      msg.Type,
			"server_id": msg.ServerID,
		}).Warn("Unknown distributed message type")
	}
}

// BroadcastUpdate broadcasts an update to all servers
func (dh *DistributedHub) BroadcastUpdate(msgType string, room string, data map[string]interface{}) error {
	msg := DistributedMessage{
		ServerID:  dh.serverID,
		Type:      msgType,
		Room:      room,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
	
	return dh.cache.Publish(dh.ctx, WebSocketChannel, msg)
}

// BroadcastRoomUpdate broadcasts a room update to all servers
func (dh *DistributedHub) BroadcastRoomUpdate(roomNumber string, status map[string]interface{}) error {
	return dh.BroadcastUpdate("room_update", roomNumber, status)
}

// BroadcastConsumptionUpdate broadcasts a consumption update to all servers
func (dh *DistributedHub) BroadcastConsumptionUpdate(roomNumber string, consumption map[string]interface{}) error {
	return dh.BroadcastUpdate("consumption_update", roomNumber, consumption)
}

// GetConnectedClients returns the count of connected clients across all servers
func (dh *DistributedHub) GetConnectedClients(ctx context.Context) (int, error) {
	// Store local count in Redis
	localCount := dh.hub.GetClientCount()
	key := fmt.Sprintf("hudini:server:%s:clients", dh.serverID)
	
	err := dh.cache.Set(ctx, key, []byte(fmt.Sprintf("%d", localCount)), 30*time.Second)
	if err != nil {
		return localCount, err
	}
	
	// Get all server client counts
	// This is a simplified version - in production, you'd want to scan for all server keys
	// For now, just return local count
	return localCount, nil
}

// Close closes the distributed hub
func (dh *DistributedHub) Close() error {
	dh.cancel()
	return dh.pubsub.Close()
}

// Health checks if the distributed hub is healthy
func (dh *DistributedHub) Health(ctx context.Context) error {
	// Check Redis connection
	if err := dh.cache.Health(ctx); err != nil {
		return fmt.Errorf("redis unhealthy: %w", err)
	}
	
	// Check if we're still subscribed
	if dh.pubsub == nil {
		return fmt.Errorf("not subscribed to redis channels")
	}
	
	return nil
}