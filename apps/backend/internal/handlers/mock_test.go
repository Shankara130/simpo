package handlers

import (
	"github.com/gin-gonic/gin"
)

// MockAuthHandler is a mock implementation of AuthHandler for testing
type MockAuthHandler struct{}

// Login is a mock login handler that returns a simple response
func (m *MockAuthHandler) Login(c *gin.Context) {
	c.JSON(200, gin.H{"message": "mock login"})
}
