package health

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_GetHealth_WithDatabaseAndRedis(t *testing.T) {
	// Arrange
	checkers := []Checker{
		&mockChecker{name: "database", result: CheckResult{Status: CheckPass, Message: "connected"}},
		&mockChecker{name: "redis", result: CheckResult{Status: CheckPass, Message: "connected"}},
	}
	service := NewService(checkers, "1.0.0", "test")

	// Act
	response := service.GetHealth(context.Background())

	// Assert
	assert.Equal(t, StatusHealthy, response.Status)
	assert.Equal(t, "connected", response.Database)
	assert.Equal(t, "connected", response.Redis)
	assert.Equal(t, "1.0.0", response.Version)
	assert.NotEmpty(t, response.Uptime)
	assert.NotEmpty(t, response.Timestamp)
	assert.Equal(t, "test", response.Environment)
	assert.Len(t, response.Checks, 2)
}

func TestService_GetHealth_DatabaseDisconnected(t *testing.T) {
	// Arrange
	checkers := []Checker{
		&mockChecker{name: "database", result: CheckResult{Status: CheckFail, Message: "disconnected"}},
		&mockChecker{name: "redis", result: CheckResult{Status: CheckPass, Message: "connected"}},
	}
	service := NewService(checkers, "1.0.0", "test")

	// Act
	response := service.GetHealth(context.Background())

	// Assert
	assert.Equal(t, StatusUnhealthy, response.Status)
	assert.Equal(t, "disconnected", response.Database)
	assert.Equal(t, "connected", response.Redis)
}

func TestService_GetHealth_RedisDisconnected_Degraded(t *testing.T) {
	// Arrange
	checkers := []Checker{
		&mockChecker{name: "database", result: CheckResult{Status: CheckPass, Message: "connected"}},
		&mockChecker{name: "redis", result: CheckResult{Status: CheckFail, Message: "disconnected"}},
	}
	service := NewService(checkers, "1.0.0", "test")

	// Act
	response := service.GetHealth(context.Background())

	// Assert
	assert.Equal(t, StatusDegraded, response.Status)
	assert.Equal(t, "connected", response.Database)
	assert.Equal(t, "disconnected", response.Redis)
}

func TestService_GetHealth_NoCheckers_DefaultsToConnected(t *testing.T) {
	// Arrange - no checkers (DB check disabled)
	checkers := []Checker{}
	service := NewService(checkers, "1.0.0", "test")

	// Act
	response := service.GetHealth(context.Background())

	// Assert - should default to connected when no checkers
	assert.Equal(t, StatusHealthy, response.Status)
	assert.Equal(t, "connected", response.Database)
	assert.Equal(t, "connected", response.Redis)
}

func TestService_GetHealth_OnlyDatabaseChecker(t *testing.T) {
	// Arrange - only database checker, no redis
	checkers := []Checker{
		&mockChecker{name: "database", result: CheckResult{Status: CheckPass, Message: "connected"}},
	}
	service := NewService(checkers, "1.0.0", "test")

	// Act
	response := service.GetHealth(context.Background())

	// Assert
	assert.Equal(t, StatusHealthy, response.Status)
	assert.Equal(t, "connected", response.Database)
	assert.Equal(t, "connected", response.Redis) // defaults to connected when not configured
}

func TestService_GetReadiness_ResponseFormat(t *testing.T) {
	// Arrange
	checkers := []Checker{
		&mockChecker{name: "database", result: CheckResult{Status: CheckPass, Message: "connected"}},
		&mockChecker{name: "redis", result: CheckResult{Status: CheckPass, Message: "connected"}},
	}
	service := NewService(checkers, "1.0.0", "test")

	// Act
	response := service.GetReadiness(context.Background())

	// Assert
	assert.Equal(t, StatusHealthy, response.Status)
	assert.Equal(t, "connected", response.Database)
	assert.Equal(t, "connected", response.Redis)
	assert.NotEmpty(t, response.Checks)
}

func TestService_GetLiveness_ResponseFormat(t *testing.T) {
	// Arrange
	checkers := []Checker{}
	service := NewService(checkers, "1.0.0", "test")

	// Act
	response := service.GetLiveness(context.Background())

	// Assert
	assert.Equal(t, StatusHealthy, response.Status)
	// Liveness doesn't include checks, so defaults to connected
	assert.Equal(t, "connected", response.Database)
	assert.Equal(t, "connected", response.Redis)
}
