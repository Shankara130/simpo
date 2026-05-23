package health

import (
	"context"
	"fmt"
	"time"
)

type Service interface {
	GetHealth(ctx context.Context) HealthResponse
	GetLiveness(ctx context.Context) HealthResponse
	GetReadiness(ctx context.Context) HealthResponse
	GetDatabaseHealth(ctx context.Context) CheckResult // Story 2.4: Database-specific health check
}

type service struct {
	checkers       []Checker
	dbChecker      Checker // Story 2.4: Cached reference for O(1) lookup
	startTime      time.Time
	version        string
	environment    string
}

func NewService(checkers []Checker, version, environment string) Service {
	// Story 2.4: Cache database checker reference for O(1) lookup
	var dbChecker Checker
	for _, checker := range checkers {
		if checker.Name() == "database" {
			dbChecker = checker
			break
		}
	}

	return &service{
		checkers:    checkers,
		dbChecker:   dbChecker,
		startTime:   time.Now(),
		version:     version,
		environment: environment,
	}
}

func (s *service) GetHealth(ctx context.Context) HealthResponse {
	// Story 9.1, AC1: Enforce 400ms timeout to ensure total response stays under 500ms
	healthCtx, cancel := context.WithTimeout(ctx, 400*time.Millisecond)
	defer cancel()

	checks := make(map[string]CheckResult)

	// Run all checkers with nil check
	for _, checker := range s.checkers {
		if checker == nil {
			continue // Skip nil checkers to prevent panic
		}
		result := checker.Check(healthCtx)
		checks[checker.Name()] = result
	}

	// Calculate overall status and extract database/redis status
	status, database, redis := s.calculateStatus(checks)

	return HealthResponse{
		Status:      status,
		Version:     s.version,
		Timestamp:   time.Now(),
		Uptime:      s.formatUptime(),
		Database:    database,
		Redis:       redis,
		Checks:      checks,
		Environment: s.environment,
	}
}

// calculateStatus determines overall health status and extracts database/redis connectivity
func (s *service) calculateStatus(checks map[string]CheckResult) (HealthStatus, string, string) {
	const (
		DatabaseCheckerName = "database"
		RedisCheckerName    = "redis"
	)

	var dbStatus, redisStatus string
	dbUnhealthy, redisUnhealthy := false, false

	// Extract database status using direct map access
	if dbResult, ok := checks[DatabaseCheckerName]; ok {
		if dbResult.Status == CheckPass {
			dbStatus = "connected"
		} else {
			dbStatus = "disconnected"
			dbUnhealthy = true
		}
	}

	// Extract redis status using direct map access
	if redisResult, ok := checks[RedisCheckerName]; ok {
		if redisResult.Status == CheckPass {
			redisStatus = "connected"
		} else {
			redisStatus = "disconnected"
			redisUnhealthy = true
		}
	}

	// If no database check, assume connected (for environments where DB check is disabled)
	if dbStatus == "" {
		dbStatus = "connected"
	}

	// If no redis check (not configured), show as connected (AC4: Redis is optional)
	if redisStatus == "" {
		redisStatus = "connected"
	}

	// Calculate overall status
	var overallStatus HealthStatus
	if dbUnhealthy {
		// Database is critical - unhealthy if disconnected
		overallStatus = StatusUnhealthy
	} else if redisUnhealthy {
		// Database connected but redis disconnected - degraded (AC5)
		overallStatus = StatusDegraded
	} else {
		// All critical dependencies healthy
		overallStatus = StatusHealthy
	}

	return overallStatus, dbStatus, redisStatus
}

func (s *service) GetLiveness(ctx context.Context) HealthResponse {
	// Liveness check - just verify the service is running
	// Include database/redis status if checkers available
	checks := make(map[string]CheckResult)
	for _, checker := range s.checkers {
		if checker == nil {
			continue
		}
		result := checker.Check(ctx)
		checks[checker.Name()] = result
	}

	// Extract database and redis status for consistency
	_, database, redis := s.calculateStatus(checks)

	return HealthResponse{
		Status:      StatusHealthy,
		Version:     s.version,
		Timestamp:   time.Now(),
		Uptime:      s.formatUptime(),
		Database:    database,
		Redis:       redis,
		Checks:      checks,
		Environment: s.environment,
	}
}

func (s *service) GetReadiness(ctx context.Context) HealthResponse {
	// Story 9.1, AC1: Enforce 400ms timeout to ensure total response stays under 500ms
	readinessCtx, cancel := context.WithTimeout(ctx, 400*time.Millisecond)
	defer cancel()

	checks := make(map[string]CheckResult)

	// Run all checkers with nil check
	for _, checker := range s.checkers {
		if checker == nil {
			continue
		}
		result := checker.Check(readinessCtx)
		checks[checker.Name()] = result
	}

	// Calculate overall status and extract database/redis status
	status, database, redis := s.calculateStatus(checks)

	return HealthResponse{
		Status:      status,
		Version:     s.version,
		Timestamp:   time.Now(),
		Uptime:      s.formatUptime(),
		Database:    database,
		Redis:       redis,
		Checks:      checks,
		Environment: s.environment,
	}
}

func (s *service) formatUptime() string {
	uptime := time.Since(s.startTime)
	days := int(uptime.Hours() / 24)
	hours := int(uptime.Hours()) % 24
	minutes := int(uptime.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// Story 2.4: Database-specific health check for /health/db endpoint
func (s *service) GetDatabaseHealth(ctx context.Context) CheckResult {
	// Use cached checker reference for O(1) lookup
	if s.dbChecker != nil {
		return s.dbChecker.Check(ctx)
	}

	// If no database checker is configured, return unhealthy
	return CheckResult{
		Status:  CheckFail,
		Message: "Database checker not configured (enable with HEALTH_DATABASE_CHECK_ENABLED=true)",
	}
}
