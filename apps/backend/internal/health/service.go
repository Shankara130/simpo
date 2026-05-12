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
	return HealthResponse{
		Status:      StatusHealthy,
		Version:     s.version,
		Timestamp:   time.Now(),
		Uptime:      s.formatUptime(),
		Environment: s.environment,
		Checks:      make(map[string]CheckResult),
	}
}

func (s *service) GetLiveness(ctx context.Context) HealthResponse {
	return HealthResponse{
		Status:      StatusHealthy,
		Version:     s.version,
		Timestamp:   time.Now(),
		Uptime:      s.formatUptime(),
		Environment: s.environment,
		Checks:      make(map[string]CheckResult),
	}
}

func (s *service) GetReadiness(ctx context.Context) HealthResponse {
	checks := make(map[string]CheckResult)
	overallStatus := StatusHealthy

	for _, checker := range s.checkers {
		result := checker.Check(ctx)
		checks[checker.Name()] = result

		if result.Status == CheckFail {
			overallStatus = StatusUnhealthy
		} else if result.Status == CheckWarn && overallStatus != StatusUnhealthy {
			overallStatus = StatusDegraded
		}
	}

	return HealthResponse{
		Status:      overallStatus,
		Version:     s.version,
		Timestamp:   time.Now(),
		Uptime:      s.formatUptime(),
		Environment: s.environment,
		Checks:      checks,
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
