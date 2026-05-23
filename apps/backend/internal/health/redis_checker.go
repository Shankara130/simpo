package health

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisChecker struct {
	client *redis.Client
}

// NewRedisChecker creates a new Redis health checker.
// If client is nil, Redis is considered not configured (optional dependency).
func NewRedisChecker(client *redis.Client) Checker {
	return &redisChecker{
		client: client,
	}
}

// Name returns the name of this health checker.
func (r *redisChecker) Name() string {
	return "redis"
}

// Check performs a ping to Redis to verify connectivity.
// If client is nil, returns pass with "not configured" message.
func (r *redisChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	// Redis is optional - nil client means not configured
	if r.client == nil {
		return CheckResult{
			Status:  CheckPass,
			Message: "Redis not configured (optional dependency)",
		}
	}

	// Use the context timeout from handler (AC1: 400ms max for total response)
	// Only add timeout if parent context has no deadline
	pingCtx := ctx
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		pingCtx, cancel = context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancel()
	}

	err := r.client.Ping(pingCtx).Err()
	duration := time.Since(start)

	if err != nil {
		return CheckResult{
			Status:       CheckFail,
			Message:      fmt.Sprintf("Redis disconnected: %v", err),
			ResponseTime: duration.String(),
		}
	}

	return CheckResult{
		Status:       CheckPass,
		Message:      "Redis connected",
		ResponseTime: duration.String(),
	}
}
