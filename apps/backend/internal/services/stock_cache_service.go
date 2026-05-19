package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// StockCacheService handles caching of stock levels in Redis
// Story 4.2, Task 15: Implement Caching Strategy (AC: 6)
type StockCacheService struct {
	redisClient *redis.Client
	ttl         time.Duration
}

// StockCacheEntry represents a cached stock level
type StockCacheEntry struct {
	ProductID   uint      `json:"product_id"`
	BranchID    uint      `json:"branch_id"`
	SKU         string    `json:"sku"`
	Name        string    `json:"name"`
	StockQty    int64     `json:"stock_qty"`
	IsLowStock  bool      `json:"is_low_stock"`
	Price       string    `json:"price"`
	CachedAt    time.Time `json:"cached_at"`
}

// NewStockCacheService creates a new stock cache service
// Story 4.2, Task 15.1: Add Redis caching for stock levels (5-minute TTL)
func NewStockCacheService(redisClient *redis.Client) *StockCacheService {
	if redisClient == nil {
		panic("StockCacheService: redisClient cannot be nil")
	}

	return &StockCacheService{
		redisClient: redisClient,
		ttl:         5 * time.Minute, // Story 4.2, Task 15.1: 5-minute TTL
	}
}

// CacheKey generates a cache key for a product stock level
// Format: stock:{product_id}:{branch_id}
func (s *StockCacheService) CacheKey(productID, branchID uint) string {
	return fmt.Sprintf("stock:%d:%d", productID, branchID)
}

// Get retrieves a cached stock level
// Returns (entry, found, error)
func (s *StockCacheService) Get(ctx context.Context, productID, branchID uint) (*StockCacheEntry, bool, error) {
	key := s.CacheKey(productID, branchID)

	data, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil // Cache miss
		}
		return nil, false, fmt.Errorf("failed to get stock cache: %w", err)
	}

	var entry StockCacheEntry
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal stock cache: %w", err)
	}

	return &entry, true, nil
}

// Set stores a stock level in cache
// Story 4.2, Task 15.1: Cache with 5-minute TTL
func (s *StockCacheService) Set(ctx context.Context, entry *StockCacheEntry) error {
	key := s.CacheKey(entry.ProductID, entry.BranchID)

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal stock cache: %w", err)
	}

	entry.CachedAt = time.Now()

	if err := s.redisClient.Set(ctx, key, data, s.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set stock cache: %w", err)
	}

	return nil
}

// Delete removes a stock level from cache
// Story 4.2, Task 15.2: Invalidate cache on stock updates
func (s *StockCacheService) Delete(ctx context.Context, productID, branchID uint) error {
	key := s.CacheKey(productID, branchID)

	if err := s.redisClient.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete stock cache: %w", err)
	}

	return nil
}

// DeleteByPattern removes all stock cache entries matching a pattern
// Useful for bulk invalidation (e.g., all products in a branch)
func (s *StockCacheService) DeleteByPattern(ctx context.Context, pattern string) error {
	iter := s.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := s.redisClient.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete cache entry: %w", err)
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan cache entries: %w", err)
	}

	return nil
}

// WarmCache loads stock levels into cache for a list of products
// Story 4.2, Task 15.4: Add cache warming on application startup
func (s *StockCacheService) WarmCache(ctx context.Context, products []ProductStockData) error {
	pipe := s.redisClient.Pipeline()

	for _, product := range products {
		entry := &StockCacheEntry{
			ProductID:  product.ProductID,
			BranchID:   product.BranchID,
			SKU:        product.SKU,
			Name:       product.Name,
			StockQty:   product.StockQty,
			IsLowStock: product.StockQty < product.ReorderThreshold,
			Price:      product.Price,
		}

		key := s.CacheKey(entry.ProductID, entry.BranchID)
		data, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("failed to marshal stock cache: %w", err)
		}

		entry.CachedAt = time.Now()
		pipe.Set(ctx, key, data, s.ttl)
	}

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to warm stock cache: %w", err)
	}

	return nil
}

// ProductStockData represents minimal data needed for cache warming
type ProductStockData struct {
	ProductID       uint
	BranchID        uint
	SKU             string
	Name            string
	StockQty        int64
	ReorderThreshold int64
	Price           string
}

// GetMultiple retrieves multiple cached stock levels
// Story 4.2, Task 15.3: Use cache as fallback for WebSocket connections
func (s *StockCacheService) GetMultiple(ctx context.Context, keys []string) (map[string]*StockCacheEntry, error) {
	if len(keys) == 0 {
		return make(map[string]*StockCacheEntry), nil
	}

	// Pipeline get for efficiency
	pipe := s.redisClient.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, key := range keys {
		cmds[key] = pipe.Get(ctx, key)
	}

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get multiple stock cache entries: %w", err)
	}

	// Parse results
	results := make(map[string]*StockCacheEntry)
	for key, cmd := range cmds {
		data, err := cmd.Result()
		if err == nil {
			var entry StockCacheEntry
			if err := json.Unmarshal([]byte(data), &entry); err == nil {
				results[key] = &entry
			}
		}
	}

	return results, nil
}

// InvalidateBranch removes all stock cache entries for a branch
func (s *StockCacheService) InvalidateBranch(ctx context.Context, branchID uint) error {
	pattern := fmt.Sprintf("stock:*:%d", branchID)
	return s.DeleteByPattern(ctx, pattern)
}

// InvalidateProduct removes all stock cache entries for a product (all branches)
func (s *StockCacheService) InvalidateProduct(ctx context.Context, productID uint) error {
	pattern := fmt.Sprintf("stock:%d:*", productID)
	return s.DeleteByPattern(ctx, pattern)
}

// ClearAll removes all stock cache entries
// Useful for testing or full cache reset
func (s *StockCacheService) ClearAll(ctx context.Context) error {
	pattern := "stock:*"
	return s.DeleteByPattern(ctx, pattern)
}

// GetCacheStats returns statistics about the cache
func (s *StockCacheService) GetCacheStats(ctx context.Context) (*CacheStats, error) {
	pattern := "stock:*"
	var keys []string

	iter := s.redisClient.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan cache: %w", err)
	}

	stats := &CacheStats{
		TotalKeys: len(keys),
		TTL:       s.ttl,
	}

	// Get memory usage if available
	if info, err := s.redisClient.Info(ctx, "memory").Result(); err == nil {
		stats.MemoryInfo = info
	}

	return stats, nil
}

// WarmFromRepository warms the cache by loading stock levels from repository
// Story 4.2, Task 15.4: Add cache warming on application startup
func (s *StockCacheService) WarmFromRepository(ctx context.Context, repo ProductStockRepository) error {
	// This method would typically load all products and cache their stock levels
	// For now, it's a placeholder - in production, you'd want to:
	// 1. Load all products from repository
	// 2. Batch cache them using WarmCache
	// 3. Handle pagination for large datasets

	// TODO: Implement repository-based cache warming
	// This would require access to the product repository
	return nil
}

// ProductStockRepository defines the interface for loading products to warm cache
type ProductStockRepository interface {
	ListAllStockLevels(ctx context.Context) ([]ProductStockData, error)
}


// CacheStats represents cache statistics
type CacheStats struct {
	TotalKeys  int
	TTL        time.Duration
	MemoryInfo string
}
