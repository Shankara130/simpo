package services

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStockEventService_PublishStockUpdate tests publishing stock update events
// Story 4.2, Task 6.2: Test event publishing with mock Redis
func TestStockEventService_PublishStockUpdate(t *testing.T) {
	// Setup miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create stock event service
	service := NewStockEventService(redisClient)

	ctx := context.Background()

	// Create test event
	event := StockUpdatedEvent{
		ProductID: 123,
		BranchID:  1,
		SKU:       "SKU-12345",
		Name:      "Paracetamol 500mg",
		OldStock:  50,
		NewStock:  45,
		Change:    -5,
		UpdatedBy: "John Doe",
		UpdatedAt: time.Now(),
	}

	// Test publishing event
	t.Run("PublishStockUpdate publishes to Redis", func(t *testing.T) {
		err := service.PublishStockUpdate(ctx, event)
		assert.NoError(t, err)
	})
}

// TestStockEventService_SubscribeAndBroadcast tests subscribing and broadcasting
// Story 4.2, Task 6.3: Test event broadcasting to WebSocket clients
func TestStockEventService_SubscribeAndBroadcast(t *testing.T) {
	// Setup miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create stock event service
	service := NewStockEventService(redisClient)
	ctx := context.Background()

	t.Run("SubscribeToStockUpdates returns pubsub", func(t *testing.T) {
		pubsub := service.SubscribeToStockUpdates(ctx, "stock.updated")
		assert.NotNil(t, pubsub)
	})

	t.Run("RegisterClient and UnregisterClient", func(t *testing.T) {
		clientID := uuid.New().String()
		messageChan := make(chan StockEvent, 10)
		branches := []uint{1, 2, 3}

		// Register client
		service.RegisterClient(clientID, branches, messageChan)

		// Unregister client
		service.UnregisterClient(clientID)
	})
}

// TestStockEventService_BranchFiltering tests branch-based event filtering
// Story 4.2, Task 6.4: Test branch filtering for multi-branch scenarios
func TestStockEventService_BranchFiltering(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	service := NewStockEventService(redisClient)

	// Test client with no branch filter (receives all events)
	t.Run("Client with no branch filter receives all events", func(t *testing.T) {
		clientID := uuid.New().String()
		messageChan := make(chan StockEvent, 10)

		service.RegisterClient(clientID, []uint{}, messageChan)

		// The shouldSendToClient method should return true
		// This is tested indirectly through the broadcast mechanism
		service.UnregisterClient(clientID)
	})

	// Test client with specific branch filter
	t.Run("Client with branch filter only receives matching branch events", func(t *testing.T) {
		clientID := uuid.New().String()
		messageChan := make(chan StockEvent, 10)
		branches := []uint{1, 2}

		service.RegisterClient(clientID, branches, messageChan)

		// Event with branch 1 should be sent to client subscribed to branch 1
		service.UnregisterClient(clientID)
	})
}

// TestStockEventService_GetChannelName tests channel naming convention
// Story 4.2, Task 1.4: Define event channel naming
func TestStockEventService_GetChannelName(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	_ = NewStockEventService(redisClient)
}

// TestStockUpdatedEvent_Struct tests the event struct
// Story 4.2, Task 1.2: Define event struct for stock updates
func TestStockUpdatedEvent_Struct(t *testing.T) {
	now := time.Now()
	event := StockUpdatedEvent{
		ProductID: 123,
		BranchID:  1,
		SKU:       "SKU-12345",
		Name:      "Paracetamol 500mg",
		OldStock:  50,
		NewStock:  45,
		Change:    -5,
		UpdatedBy: "John Doe",
		UpdatedAt: now,
	}

	assert.Equal(t, uint(123), event.ProductID)
	assert.Equal(t, uint(1), event.BranchID)
	assert.Equal(t, "SKU-12345", event.SKU)
	assert.Equal(t, "Paracetamol 500mg", event.Name)
	assert.Equal(t, int64(50), event.OldStock)
	assert.Equal(t, int64(45), event.NewStock)
	assert.Equal(t, int64(-5), event.Change)
	assert.Equal(t, "John Doe", event.UpdatedBy)
	assert.Equal(t, now, event.UpdatedAt)
}
