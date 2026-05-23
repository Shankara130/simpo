package health

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisChecker_Name(t *testing.T) {
	// Arrange
	checker := NewRedisChecker(nil)

	// Act
	name := checker.Name()

	// Assert
	assert.Equal(t, "redis", name)
}

func TestRedisChecker_Check_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	s := miniredis.RunT(t)
	defer s.Close()

	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	defer client.Close()

	checker := NewRedisChecker(client)

	// Act
	result := checker.Check(ctx)

	// Assert
	assert.Equal(t, CheckPass, result.Status)
	assert.Contains(t, result.Message, "connected")
	assert.NotEmpty(t, result.ResponseTime)
}

func TestRedisChecker_Check_Failure(t *testing.T) {
	// Arrange
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:9999", // Invalid port
	})

	checker := NewRedisChecker(client)

	// Act
	result := checker.Check(ctx)

	// Assert
	assert.Equal(t, CheckFail, result.Status)
	assert.Contains(t, result.Message, "disconnected")
	assert.NotEmpty(t, result.ResponseTime)
}

func TestRedisChecker_Check_ContextTimeout(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	s := miniredis.RunT(t)
	defer s.Close()

	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	defer client.Close()

	checker := NewRedisChecker(client)

	// Act
	result := checker.Check(ctx)

	// Assert - should either pass quickly or fail with timeout
	assert.NotEmpty(t, result.Status)
	assert.NotEmpty(t, result.ResponseTime)
}

func TestRedisChecker_Check_NilClient(t *testing.T) {
	// Arrange
	ctx := context.Background()
	checker := NewRedisChecker(nil)

	// Act
	result := checker.Check(ctx)

	// Assert - nil client should return "not configured" as pass (optional dependency)
	assert.Equal(t, CheckPass, result.Status)
	assert.Contains(t, result.Message, "not configured")
}

func TestNewRedisChecker(t *testing.T) {
	// Test with nil client
	checker1 := NewRedisChecker(nil)
	assert.NotNil(t, checker1)
	assert.Equal(t, "redis", checker1.Name())

	// Test with valid client
	client := &redis.Client{}
	checker2 := NewRedisChecker(client)
	assert.NotNil(t, checker2)
}
