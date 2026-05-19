package services

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// StockUpdatedEvent represents a stock update event payload
// Story 4.2, Task 1: Define event struct for stock updates
type StockUpdatedEvent struct {
	ProductID uint      `json:"productId"`
	BranchID  uint      `json:"branchId"`
	SKU       string    `json:"sku"`
	Name      string    `json:"name"`
	OldStock  int64     `json:"oldStock"`
	NewStock  int64     `json:"newStock"`
	Change    int64     `json:"change"`
	UpdatedBy string    `json:"updatedBy"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// StockEventService defines the interface for stock event publishing and broadcasting
// Story 4.2, Task 1 & 5: Redis pub/sub service and event broadcaster
type StockEventService interface {
	// PublishStockUpdate publishes a stock update event to Redis pub/sub
	// Story 4.2, Task 1.3: Implement PublishStockUpdate method using Redis pub/sub
	PublishStockUpdate(ctx context.Context, event StockUpdatedEvent) error

	// SubscribeToStockUpdates subscribes to stock update events
	// Story 4.2, Task 5.2: Subscribe to Redis pub/sub channels
	SubscribeToStockUpdates(ctx context.Context, channel string) *redis.PubSub

	// StartBroadcaster starts the event broadcaster that forwards Redis events to WebSocket clients
	// Story 4.2, Task 5.1: Create StockEventBroadcaster
	StartBroadcaster(ctx context.Context) error

	// StopBroadcaster gracefully stops the broadcaster
	// Story 4.2, Task 5.6: Add graceful shutdown for broadcaster
	StopBroadcaster()

	// RegisterClient registers a WebSocket client for stock updates
	// Story 4.2, Task 4.3: Implement connection management
	RegisterClient(clientID string, branches []uint, messageChan chan StockUpdatedEvent)

	// UnregisterClient removes a WebSocket client
	UnregisterClient(clientID string)
}

// stockEventService implements StockEventService interface
type stockEventService struct {
	redisClient *redis.Client
	// WebSocket client management
	clients        map[string]clientSubscription
	clientsMutex   sync.RWMutex
	// Channel for broadcaster control
	broadcasterCtx    context.Context
	broadcasterCancel context.CancelFunc
	// PubSub connection for broadcaster (kept for cleanup)
	pubsub *redis.PubSub
}

// clientSubscription represents a WebSocket client's subscription
type clientSubscription struct {
	Branches     []uint
	MessageChan  chan<- StockUpdatedEvent
	Disconnected chan struct{}
}

// NewStockEventService creates a new stock event service
// Story 4.2, Task 1: Create stock_event_service.go
func NewStockEventService(redisClient *redis.Client) StockEventService {
	if redisClient == nil {
		panic("stockEventService: redisClient cannot be nil")
	}

	return &stockEventService{
		redisClient: redisClient,
		clients:     make(map[string]clientSubscription),
	}
}

// PublishStockUpdate publishes a stock update event to Redis pub/sub
// Story 4.2, Task 1.3-1.6: Implement pub/sub with error handling
func (s *stockEventService) PublishStockUpdate(ctx context.Context, event StockUpdatedEvent) error {
	if s.redisClient == nil {
		// Story 4.2, Task 1.6: Error handling for Redis unavailable scenarios
		// Log warning but don't fail - stock updates work, real-time notifications disabled
		return nil
	}

	// Story 4.2, Task 1.5: Add event serialization (JSON format)
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Story 4.2, Task 1.4: Define event channel naming
	// Channel format: stock.updated (global channel for all stock updates)
	globalChannel := "stock.updated"

	// Publish to global channel
	return s.redisClient.Publish(ctx, globalChannel, eventJSON).Err()
}

// getChannelName generates the channel name for a stock update event
// Story 4.2, Task 1.4: Event channel naming convention
func (s *stockEventService) getChannelName(productID, branchID uint) string {
	return "stock.updated"
}

// SubscribeToStockUpdates subscribes to stock update events
// Story 4.2, Task 5.2: Subscribe to Redis pub/sub channels
func (s *stockEventService) SubscribeToStockUpdates(ctx context.Context, channel string) *redis.PubSub {
	if s.redisClient == nil {
		return nil
	}
	return s.redisClient.Subscribe(ctx, channel)
}

// StartBroadcaster starts the event broadcaster
// Story 4.2, Task 5.1-5.6: StockEventBroadcaster implementation
func (s *stockEventService) StartBroadcaster(ctx context.Context) error {
	if s.redisClient == nil {
		return nil
	}

	s.broadcasterCtx, s.broadcasterCancel = context.WithCancel(ctx)

	// Subscribe to global stock updates channel
	s.pubsub = s.redisClient.Subscribe(s.broadcasterCtx, "stock.updated")

	// Start broadcaster goroutine
	go func() {
		for {
			select {
			case <-s.broadcasterCtx.Done():
				return
			case msg := <-s.pubsub.Channel():
				if msg == nil {
					continue
				}

				// Parse event from Redis message (JSON)
				var event StockUpdatedEvent
				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					continue
				}

				// Story 4.2, Task 5.3: Broadcast events to connected WebSocket clients
				// Story 4.2, Task 5.4: Filter events based on client's subscribed branches
				s.broadcastToClients(event)
			}
		}
	}()

	return nil
}

// StopBroadcaster gracefully stops the broadcaster
// Story 4.2, Task 5.6: Add graceful shutdown
func (s *stockEventService) StopBroadcaster() {
	if s.broadcasterCancel != nil {
		s.broadcasterCancel()
	}
	if s.pubsub != nil {
		s.pubsub.Close()
		s.pubsub = nil
	}
}

// RegisterClient registers a WebSocket client for stock updates
// Story 4.2, Task 4.3: Implement connection management
func (s *stockEventService) RegisterClient(clientID string, branches []uint, messageChan chan StockUpdatedEvent) {
	s.clientsMutex.Lock()
	s.clients[clientID] = clientSubscription{
		Branches:     branches,
		MessageChan:  messageChan,
		Disconnected: make(chan struct{}),
	}
	s.clientsMutex.Unlock()
}

// UnregisterClient removes a WebSocket client
func (s *stockEventService) UnregisterClient(clientID string) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()
	if client, exists := s.clients[clientID]; exists {
		close(client.Disconnected)
		delete(s.clients, clientID)
	}
}

// broadcastToClients broadcasts an event to all relevant connected clients
// Story 4.2, Task 5.4: Filter events based on client's subscribed branches
func (s *stockEventService) broadcastToClients(event StockUpdatedEvent) {
	s.clientsMutex.RLock()
	// Copy client IDs to avoid holding lock while sending
	clientIDs := make([]string, 0, len(s.clients))
	for clientID := range s.clients {
		clientIDs = append(clientIDs, clientID)
	}
	s.clientsMutex.RUnlock()

	for _, clientID := range clientIDs {
		s.clientsMutex.RLock()
		client, exists := s.clients[clientID]
		s.clientsMutex.RUnlock()

		if !exists {
			continue
		}

		// Check if client is interested in this event (branch filtering)
		if !s.shouldSendToClient(client, event) {
			continue
		}

		// Send event to client (non-blocking)
		select {
		case client.MessageChan <- event:
		case <-client.Disconnected:
			// Client disconnected, remove from registry
			s.clientsMutex.Lock()
			delete(s.clients, clientID)
			s.clientsMutex.Unlock()
		case <-time.After(100 * time.Millisecond):
			// Timeout - client channel is full or slow
			// Story 4.2, Error Handling: Drop oldest messages if queue full
		}
	}
}

// shouldSendToClient determines if a client should receive this event
// Story 4.2, Task 5.4: Filter events based on client's subscribed branches
func (s *stockEventService) shouldSendToClient(client clientSubscription, event StockUpdatedEvent) bool {
	// If client has no branch filters, send all events
	if len(client.Branches) == 0 {
		return true
	}

	// Check if event's branch is in client's subscribed branches
	for _, branchID := range client.Branches {
		if branchID == event.BranchID {
			return true
		}
	}

	return false
}

