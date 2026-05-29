package health

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Health godoc
//
//	@Summary		API health check endpoint
//	@Description	Check if the application is running and its dependencies (Story 9.1 - API versioned endpoint)
//	@Tags			Health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	HealthResponse		"System is healthy - all critical dependencies connected"
//	@Failure		503	{object}	HealthResponse		"System is unhealthy - critical dependency disconnected"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Router			/api/v1/health [get]
func (h *Handler) Health(c *gin.Context) {
	ctx := c.Request.Context()
	response := h.service.GetHealth(ctx)

	// Story 9.1, AC8: Log health check requests for audit purposes with explicit timestamp
	slog.Info("Health check",
		"timestamp", time.Now().Format(time.RFC3339),
		"path", "/api/v1/health",
		"status", response.Status,
		"database", response.Database,
		"redis", response.Redis,
	)

	statusCode := http.StatusOK
	if response.Status == StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// Live godoc
//
//	@Summary		Liveness probe
//	@Description	Check if the application is alive (not deadlocked)
//	@Tags			Health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	HealthResponse
//	@Router			/health/live [get]
func (h *Handler) Live(c *gin.Context) {
	ctx := c.Request.Context()
	response := h.service.GetLiveness(ctx)
	c.JSON(http.StatusOK, response)
}

// Ready godoc
//
//	@Summary		Readiness probe
//	@Description	Check if the application and its dependencies are ready to serve traffic
//	@Tags			Health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	HealthResponse	"Service is ready"
//	@Success		503	{object}	HealthResponse	"Service is not ready"
//	@Router			/health/ready [get]
func (h *Handler) Ready(c *gin.Context) {
	ctx := c.Request.Context()
	response := h.service.GetReadiness(ctx)

	statusCode := http.StatusOK
	if response.Status == StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// Database godoc
//
//	@Summary		Database health check
//	@Description	Check if the database connection is healthy (Story 2.4)
//	@Tags			Health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	CheckResult	"Database is healthy"
//	@Success		503	{object}	CheckResult	"Database is unhealthy"
//	@Router			/health/db [get]
func (h *Handler) Database(c *gin.Context) {
	ctx := c.Request.Context()
	response := h.service.GetDatabaseHealth(ctx)

	statusCode := http.StatusOK
	if response.Status == CheckFail {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}
