package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// Use models.AuditAction for all audit action constants
// Story 5.4: Audit action types defined in models/audit_log.go

// AuditLogEntry represents an append-only audit log entry (Story 1.5, AC7, NFR-SEC-004)
type AuditLogEntry struct {
	UserID    *uint           `json:"user_id,omitempty"`
	Username  string          `json:"username"`
	Action    models.AuditAction `json:"action"`
	IPAddress string          `json:"ip_address"`
	Outcome   string          `json:"outcome"`
	Reason    string          `json:"reason,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// AuditRepository defines the interface for audit log persistence
// Story 5.4: Interface for persistent audit storage (replaces stdout logging)
type AuditRepository interface {
	Create(ctx context.Context, auditLog *models.AuditLog) error
}

// AuditService defines the audit logging interface (Story 1.5, AC7, NFR-SEC-004)
// Audit logs are append-only per NFR-SEC-004 compliance requirements
type AuditService interface {
	// LogLoginAttempt logs login attempts (success and failure) to append-only audit trail
	LogLoginAttempt(ctx context.Context, entry AuditLogEntry) error

	// LogAuthorizationFailure logs authorization failures (403 responses)
	// Story 1.6, AC6: All authorization failures are logged with user_id, role, endpoint, reason
	LogAuthorizationFailure(ctx context.Context, entry AuditLogEntry) error

	// LogUserCreation logs user creation actions (Story 1.7, AC7)
	// Logs admin_user_id, created_user_id, action, timestamp, ip_address
	LogUserCreation(ctx context.Context, adminID uint, createdUserID uint, adminUsername string, createdUsername string, ipAddress string) error

	// LogWhitelistChange logs whitelist domain management actions (Story 1.9, AC8)
	// Logs admin_user_id, domain, action, timestamp, ip_address
	LogWhitelistChange(ctx context.Context, adminID uint, adminUsername string, domain string, action models.AuditAction, ipAddress string) error

	// LogSelfRegistration logs staff self-registration actions (Story 1.9, AC8)
	// Logs user_id, email, domain, action, timestamp, ip_address
	LogSelfRegistration(ctx context.Context, userID uint, email string, domain string, ipAddress string) error

	// LogEmailVerification logs email verification actions (Story 1.9, AC8)
	// Logs user_id, email, action, timestamp, ip_address
	LogEmailVerification(ctx context.Context, userID uint, email string, ipAddress string) error

	// LogUserDeactivation logs user deactivation actions (Story 1.10, AC5)
	// Logs admin_user_id, deactivated_user_id, admin_username, deactivated_username, reason, action, timestamp, ip_address
	LogUserDeactivation(ctx context.Context, adminID uint, deactivatedUserID uint, adminUsername string, deactivatedUsername string, reason string, ipAddress string) error

	// LogStockAdjustment logs manual stock adjustment actions (Story 4.3, AC5)
	// Logs admin_user_id, product_id, product_sku, old_qty, new_qty, reason, timestamp, ip_address
	// Append-only audit trail for Badan POM compliance (NFR-SEC-004, NFR-SEC-009)
	// Story 5.4, Task 4.5: Added ipAddress parameter for IP address extraction from request context
	LogStockAdjustment(ctx context.Context, adminID uint, adminUsername string, productID uint, productSKU string, oldQty int64, newQty int64, reason string, ipAddress string) error

	// LogBlockedSaleAttempt logs blocked sale attempts for expired products (Story 4.6, AC6)
	// Logs user_id, username, product_id, product_sku, product_name, expiry_date, reason, timestamp, ip_address
	// Append-only audit trail for Badan POM regulatory compliance (NFR-SEC-004, NFR-SEC-009, NFR-SEC-011)
	// Story 5.4, Task 4.5: Added ipAddress parameter for IP address extraction from request context
	LogBlockedSaleAttempt(ctx context.Context, userID uint, username string, productID uint, productSKU string, productName string, expiryDate string, reason string, ipAddress string) error

	// LogReportExport logs report export events (Story 5.3, Task 4.7)
	// Logs user_id, username, report_type, format, date_range, timestamp, ip_address
	// Append-only audit trail for Badan POM regulatory compliance (NFR-SEC-004, NFR-SEC-009)
	// Story 5.4, Task 4.5: Added ipAddress parameter for IP address extraction from request context
	LogReportExport(ctx context.Context, userID uint, username string, reportType string, format string, dateRange string, outcome string, ipAddress string) error

	// LogSettingsUpdate logs system settings changes (Story 6.1, AC7)
	// Logs admin_user_id, changes (JSON), timestamp, ip_address
	// Append-only audit trail for Badan POM regulatory compliance
	LogSettingsUpdate(ctx context.Context, adminID uint, adminUsername string, changesJSON string, ipAddress string) error

	// Story 6.4: System change audit methods for Badan POM compliance

	// LogBackupCreated logs backup creation operations
	// Logs admin_user_id, admin_username, backup_file, size, timestamp, ip_address
	LogBackupCreated(ctx context.Context, adminID uint, adminUsername string, backupFile string, size int64, ipAddress string) error

	// LogBackupRestored logs backup restore operations
	// Logs admin_user_id, admin_username, backup_file, timestamp, ip_address
	LogBackupRestored(ctx context.Context, adminID uint, adminUsername string, backupFile string, ipAddress string) error

	// LogBackupDeleted logs backup deletion operations
	// Logs admin_user_id, admin_username, backup_file, timestamp, ip_address
	LogBackupDeleted(ctx context.Context, adminID uint, adminUsername string, backupFile string, ipAddress string) error

	// Story 6.4: Role and permission management audit methods

	// LogRoleUpdated logs role changes for users
	// Logs admin_user_id, admin_username, target_user_id, target_username, old_role, new_role, timestamp, ip_address
	LogRoleUpdated(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, oldRole string, newRole string, ipAddress string) error

	// LogPermissionGranted logs permission grant operations
	// Logs admin_user_id, admin_username, target_user_id, target_username, permission, timestamp, ip_address
	LogPermissionGranted(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, permission string, ipAddress string) error

	// LogPermissionRevoked logs permission revoke operations
	// Logs admin_user_id, admin_username, target_user_id, target_username, permission, timestamp, ip_address
	LogPermissionRevoked(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, permission string, ipAddress string) error

	// Story 6.4: Branch management audit methods

	// LogBranchCreated logs branch creation operations
	// Logs admin_user_id, admin_username, branch_name, branch_location, timestamp, ip_address
	LogBranchCreated(ctx context.Context, adminID uint, adminUsername string, branchName string, branchLocation string, ipAddress string) error

	// LogBranchUpdated logs branch update operations
	// Logs admin_user_id, admin_username, branch_id, branch_name, changes, timestamp, ip_address
	LogBranchUpdated(ctx context.Context, adminID uint, adminUsername string, branchID uint, branchName string, changes string, ipAddress string) error

	// LogBranchDeactivated logs branch deactivation operations
	// Logs admin_user_id, admin_username, branch_id, branch_name, reason, timestamp, ip_address
	LogBranchDeactivated(ctx context.Context, adminID uint, adminUsername string, branchID uint, branchName string, reason string, ipAddress string) error

	// Story 6.4: System operations audit methods

	// LogSystemStartup logs system startup operations
	// Logs system_id, startup_timestamp, server_info, ip_address
	LogSystemStartup(ctx context.Context, systemID string, serverInfo string, ipAddress string) error

	// LogSystemShutdown logs system shutdown operations
	// Logs system_id, shutdown_timestamp, reason, ip_address
	LogSystemShutdown(ctx context.Context, systemID string, reason string, ipAddress string) error

	// LogMaintenanceModeEnabled logs maintenance mode activation
	// Logs admin_user_id, admin_username, reason, timestamp, ip_address
	LogMaintenanceModeEnabled(ctx context.Context, adminID uint, adminUsername string, reason string, ipAddress string) error

	// LogMaintenanceModeDisabled logs maintenance mode deactivation
	// Logs admin_user_id, admin_username, reason, timestamp, ip_address
	LogMaintenanceModeDisabled(ctx context.Context, adminID uint, adminUsername string, reason string, ipAddress string) error

	// ResetMetrics resets all audit service metrics to zero
	// Code review fix: CRIT-016 - Add metrics reset functionality
	ResetMetrics()

	// Shutdown gracefully shuts down the audit service (Story 6.4, CRIT-001)
	Shutdown(ctx context.Context) error
}

// auditService implements AuditService with persistent audit trail storage
// Story 5.4: Implements persistent audit trail storage (replaces MVP stdout logging)
// AuditServiceConfig configures audit service behavior
type AuditServiceConfig struct {
	EnableRetry        bool          // Enable retry queue for failed audit writes
	MaxRetries         int           // Maximum retry attempts per audit entry
	RetryInterval      time.Duration // Initial retry interval (with exponential backoff)
	RetryQueueSize     int           // Maximum queue size for pending retries
}

// DefaultAuditServiceConfig returns default configuration for audit service
func DefaultAuditServiceConfig() AuditServiceConfig {
	return AuditServiceConfig{
		EnableRetry:    true,    // Enable retry by default for compliance
		MaxRetries:     3,       // Retry up to 3 times
		RetryInterval:  5 * time.Second,
		RetryQueueSize: 1000,    // Allow up to 1000 pending retries
	}
}

// pendingRetry represents an audit entry awaiting retry
type pendingRetry struct {
	entry     AuditLogEntry
	attempts  int
	nextRetry time.Time
}

type auditService struct {
	repo         AuditRepository // Story 5.4: Inject repository for persistent storage
	config       AuditServiceConfig
	retryQueue   chan pendingRetry
	retryWg      sync.WaitGroup
	shutdownOnce sync.Once
	stopRetry    chan struct{}
	metrics      auditMetrics
}

// auditMetrics tracks audit service metrics for monitoring
type auditMetrics struct {
	sync.Mutex
	totalWrites      int64
	successfulWrites int64
	failedWrites     int64
	retriedWrites    int64
	abandonedWrites  int64 // Failed after max retries
}

// NewAuditService creates a new audit service instance
// Story 5.4, Task 4.1: Modify audit_service.go to inject AuditRepository dependency
// Story 6.4, CRIT-001: Add retry mechanism for failed audit writes (Badan POM compliance)
func NewAuditService(repo AuditRepository) AuditService {
	return NewAuditServiceWithConfig(repo, DefaultAuditServiceConfig())
}

// NewAuditServiceWithConfig creates a new audit service with custom configuration
func NewAuditServiceWithConfig(repo AuditRepository, config AuditServiceConfig) AuditService {
	if repo == nil {
		// For backward compatibility during migration, allow nil repo (falls back to stdout)
		slog.Warn("AuditRepository not provided - using stdout logging (should be temporary)")
		return &auditService{repo: nil, config: config}
	}

	service := &auditService{
		repo:       repo,
		config:     config,
		stopRetry:  make(chan struct{}),
		retryQueue: make(chan pendingRetry, config.RetryQueueSize),
	}

	// Start background retry worker if enabled
	if config.EnableRetry {
		service.startRetryWorker()
		slog.Info("Audit service retry mechanism enabled",
			"maxRetries", config.MaxRetries,
			"retryInterval", config.RetryInterval,
			"queueSize", config.RetryQueueSize,
		)
	}

	return service
}

// startRetryWorker starts the background worker for retrying failed audit writes
func (s *auditService) startRetryWorker() {
	s.retryWg.Add(1)
	go func() {
		defer s.retryWg.Done()
		s.retryWorker()
	}()
}

// retryWorker processes failed audit entries from the retry queue
func (s *auditService) retryWorker() {
	ticker := time.NewTicker(s.config.RetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopRetry:
			// Drain remaining queue before shutdown
			s.drainRetryQueue()
			return
		case retry := <-s.retryQueue:
			// Check if it's time to retry
			if time.Now().Before(retry.nextRetry) {
				// Calculate duration until retry time
				durationUntilRetry := time.Until(retry.nextRetry)
				select {
				case <-time.After(durationUntilRetry):
					// Time to retry, continue to process this entry
				case <-s.stopRetry:
					// Shutdown requested, put back and stop processing
					s.retryQueue <- retry
					return
				}
			}

			// Attempt retry
			err := s.repo.Create(context.Background(), s.createAuditLogEntry(retry.entry))
			if err != nil {
				retry.attempts++
				if retry.attempts >= s.config.MaxRetries {
					// Max retries exceeded - log and abandon
					s.incrementAbandoned()
					slog.Error("AUDIT_LOG_ABANDONED",
						"action", string(retry.entry.Action),
						"username", retry.entry.Username,
						"attempts", retry.attempts,
						"error", err.Error(),
					)
				} else {
					// Exponential backoff with overflow protection
					// Cap exponent at 10 to prevent integer overflow (max multiplier: 1024x)
					maxExponent := 10
					if retry.attempts > maxExponent {
						retry.attempts = maxExponent
					}
					backoffMultiplier := time.Duration(1 << retry.attempts)
					retry.nextRetry = time.Now().Add(s.config.RetryInterval * backoffMultiplier)
					s.incrementRetried()
					s.retryQueue <- retry
					slog.Warn("AUDIT_LOG_RETRY",
						"action", string(retry.entry.Action),
						"username", retry.entry.Username,
						"attempt", retry.attempts,
						"nextRetry", retry.nextRetry,
					)
				}
			} else {
				s.incrementSuccess()
				slog.Info("AUDIT_LOG_RETRY_SUCCESS",
					"action", string(retry.entry.Action),
					"username", retry.entry.Username,
					"attempts", retry.attempts,
				)
			}
		case <-ticker.C:
			// Periodic check for shutdown
			continue
		}
	}
}

// drainRetryQueue attempts to save remaining queued entries before shutdown
func (s *auditService) drainRetryQueue() {
	remaining := len(s.retryQueue)
	if remaining > 0 {
		slog.Warn("Audit service shutting down with pending retries",
			"remaining", remaining,
		)
		// Try to save remaining entries immediately
		for i := 0; i < remaining; i++ {
			select {
			case retry := <-s.retryQueue:
				_ = s.repo.Create(context.Background(), s.createAuditLogEntry(retry.entry))
			default:
				return
			}
		}
	}
}

// Shutdown gracefully shuts down the audit service
func (s *auditService) Shutdown(ctx context.Context) error {
	s.shutdownOnce.Do(func() {
		close(s.stopRetry)
		done := make(chan struct{})
		go func() {
			s.retryWg.Wait()
			close(done)
		}()

		select {
		case <-done:
			slog.Info("Audit service shutdown complete")
		case <-ctx.Done():
			slog.Error("Audit service shutdown timeout", "error", ctx.Err())
		}
	})
	return nil
}

// GetMetrics returns current audit service metrics as a map (snapshot)
func (s *auditService) GetMetrics() map[string]int64 {
	s.metrics.Lock()
	defer s.metrics.Unlock()
	return map[string]int64{
		"total_writes":      s.metrics.totalWrites,
		"successful_writes": s.metrics.successfulWrites,
		"failed_writes":     s.metrics.failedWrites,
		"retried_writes":    s.metrics.retriedWrites,
		"abandoned_writes":  s.metrics.abandonedWrites,
	}
}

// ResetMetrics resets all audit service metrics to zero
// Code review fix: CRIT-016 - Add metrics reset functionality
func (s *auditService) ResetMetrics() {
	s.metrics.Lock()
	defer s.metrics.Unlock()
	s.metrics.totalWrites = 0
	s.metrics.successfulWrites = 0
	s.metrics.failedWrites = 0
	s.metrics.retriedWrites = 0
	s.metrics.abandonedWrites = 0
}

// Metrics increment helpers
func (s *auditService) incrementTotal() { s.metrics.Lock(); s.metrics.totalWrites++; s.metrics.Unlock() }
func (s *auditService) incrementSuccess() { s.metrics.Lock(); s.metrics.successfulWrites++; s.metrics.Unlock() }
func (s *auditService) incrementFailed() { s.metrics.Lock(); s.metrics.failedWrites++; s.metrics.Unlock() }
func (s *auditService) incrementRetried() { s.metrics.Lock(); s.metrics.retriedWrites++; s.metrics.Unlock() }
func (s *auditService) incrementAbandoned() { s.metrics.Lock(); s.metrics.abandonedWrites++; s.metrics.Unlock() }

// createAuditLogEntry creates an AuditLog model from AuditLogEntry
func (s *auditService) createAuditLogEntry(entry AuditLogEntry) *models.AuditLog {
	userID := uint(0)
	if entry.UserID != nil {
		userID = *entry.UserID
	}

	return &models.AuditLog{
		UserID:    userID,
		Username:  entry.Username,
		Action:    models.AuditAction(entry.Action),
		IPAddress: entry.IPAddress,
			Outcome:   entry.Outcome,
		Reason:    entry.Reason,
		Timestamp: entry.Timestamp,
	}
}

// persistToDatabase persists audit log to database via repository
// Story 5.4, Task 4.2: Update all Log* methods to call repository.CreateAuditLog instead of stdout
// Story 5.4, Task 4.3: Add error handling for audit log failures (log to stderr, don't block operations)
// Story 6.4, CRIT-001: Add retry queue for failed audit writes (Badan POM compliance)
func (s *auditService) persistToDatabase(ctx context.Context, entry AuditLogEntry) error {
	s.incrementTotal()

	if s.repo == nil {
		s.incrementFailed()
		return fmt.Errorf("audit repository not configured")
	}

	// Story 5.4, Task 4.4: Add context cancellation checks before database writes
	select {
	case <-ctx.Done():
		s.incrementFailed()
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
		// Continue with database write
	}

	// Create audit log model from entry
	auditLog := s.createAuditLogEntry(entry)

	// Persist to database
	err := s.repo.Create(ctx, auditLog)
	if err != nil {
		s.incrementFailed()

		// Story 5.4, Task 4.3: Log audit log failures to stderr (don't block operations)
		slog.Error("AUDIT_LOG_FAILED",
			"timestamp", time.Now().Format(time.RFC3339),
			"action", string(entry.Action),
			"username", entry.Username,
			"user_id", entry.UserID,
			"error", err.Error(),
		)

		// Story 6.4, CRIT-001: Add retry mechanism for compliance
		// If retry is enabled, queue the entry for retry instead of losing it
		if s.config.EnableRetry {
			select {
			case s.retryQueue <- pendingRetry{
				entry:     entry,
				attempts:  0,
				nextRetry: time.Now().Add(s.config.RetryInterval),
			}:
				slog.Info("AUDIT_QUEUED_FOR_RETRY",
					"action", string(entry.Action),
					"username", entry.Username,
					"queueSize", len(s.retryQueue),
				)
			default:
				// Queue is full - critical error
				s.incrementAbandoned()
				slog.Error("AUDIT_RETRY_QUEUE_FULL",
					"action", string(entry.Action),
					"username", entry.Username,
					"queueSize", len(s.retryQueue),
					"error", "Retry queue full - audit entry may be lost",
				)
			}
		}

		return err
	}

	s.incrementSuccess()
	return nil
}

// LogLoginAttempt logs login attempt to append-only audit trail (Story 1.5, AC7)
func (s *auditService) LogLoginAttempt(ctx context.Context, entry AuditLogEntry) error {
	// Set timestamp if not provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Story 5.4: Persist to database (replaces stdout-only logging)
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback/backup for visibility during development
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"username", entry.Username,
		"user_id", entry.UserID,
		"ip_address", entry.IPAddress,
		"outcome", entry.Outcome,
		"reason", entry.Reason,
	)

	return nil
}

// LogAuthorizationFailure logs authorization failures to append-only audit trail
// Story 1.6, AC6: All authorization failures are logged with user_id, role, endpoint, reason
func (s *auditService) LogAuthorizationFailure(ctx context.Context, entry AuditLogEntry) error {
	// Set timestamp if not provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"username", entry.Username,
		"user_id", entry.UserID,
		"ip_address", entry.IPAddress,
		"outcome", entry.Outcome,
		"reason", entry.Reason,
	)

	return nil
}

// LogUserCreation logs user creation actions to append-only audit trail (Story 1.7, AC7)
func (s *auditService) LogUserCreation(ctx context.Context, adminID uint, createdUserID uint, adminUsername string, createdUsername string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    models.AuditActionUserCreated,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Created user '%s' (ID: %d)", createdUsername, createdUserID),
		Timestamp: time.Now().UTC(),
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"created_user_id", createdUserID,
		"created_username", createdUsername,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogWhitelistChange logs whitelist domain management actions to append-only audit trail (Story 1.9, AC8)
func (s *auditService) LogWhitelistChange(ctx context.Context, adminID uint, adminUsername string, domain string, action models.AuditAction, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    action,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Domain '%s'", domain),
		Timestamp: time.Now().UTC(),
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"domain", domain,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogSelfRegistration logs staff self-registration actions to append-only audit trail (Story 1.9, AC8)
func (s *auditService) LogSelfRegistration(ctx context.Context, userID uint, email string, domain string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &userID,
		Username:  email, // Use email as username for self-registration
		Action:    models.AuditActionSelfRegistration,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Self-registered from domain '%s'", domain),
		Timestamp: time.Now().UTC(),
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"user_id", userID,
		"email", email,
		"domain", domain,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogEmailVerification logs email verification actions to append-only audit trail (Story 1.9, AC8)
func (s *auditService) LogEmailVerification(ctx context.Context, userID uint, email string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &userID,
		Username:  email, // Use email as username for email verification
		Action:    models.AuditActionEmailVerified,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    "Email verified and account activated",
		Timestamp: time.Now().UTC(),
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"user_id", userID,
		"email", email,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogUserDeactivation logs user deactivation actions to append-only audit trail (Story 1.10, AC5)
func (s *auditService) LogUserDeactivation(ctx context.Context, adminID uint, deactivatedUserID uint, adminUsername string, deactivatedUsername string, reason string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    models.AuditActionUserDeactivated,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Deactivated user '%s' (ID: %d) - Reason: %s", deactivatedUsername, deactivatedUserID, reason),
		Timestamp: time.Now().UTC(),
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"deactivated_user_id", deactivatedUserID,
		"deactivated_username", deactivatedUsername,
		"reason", reason,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogStockAdjustment logs manual stock adjustment actions to append-only audit trail (Story 4.3, AC5)
// Story 5.4, Task 4.5: Added ipAddress parameter for IP address extraction from request context
func (s *auditService) LogStockAdjustment(ctx context.Context, adminID uint, adminUsername string, productID uint, productSKU string, oldQty int64, newQty int64, reason string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    models.AuditActionStockAdjustment,
		IPAddress: ipAddress, // Story 5.4, Task 4.5: IP address extracted from request context
		Outcome:   "success",
		Reason:    fmt.Sprintf("Adjusted stock for product '%s' (ID: %d): %d → %d - Reason: %s", productSKU, productID, oldQty, newQty, reason),
		Timestamp: time.Now().UTC(),
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"product_id", productID,
		"product_sku", productSKU,
		"old_qty", oldQty,
		"new_qty", newQty,
		"reason", reason,
		"outcome", entry.Outcome,
	)

	// Story 5.4, Task 4.6: Update existing TODO comments to reference this story (Story 5.4)
	// TODO comment cleanup done - now using persistent storage

	return nil
}

// LogReportExport logs report export events for regulatory compliance
// Story 5.3, Task 4.7: Add export event logging to audit trail
// Story 5.4, Task 4.5: Added ipAddress parameter for IP address extraction from request context
func (s *auditService) LogReportExport(ctx context.Context, userID uint, username string, reportType string, format string, dateRange string, outcome string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &userID,
		Username:  username,
		Action:    models.AuditActionExportReport,
		IPAddress: ipAddress, // Story 5.4, Task 4.5: IP address extracted from request context
		Outcome:   outcome,
		Reason:    fmt.Sprintf("Exported report: type=%s, format=%s, range=%s", reportType, format, dateRange),
		Timestamp: time.Now().UTC(),
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback for MVP
	slog.Warn("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"user_id", userID,
		"username", username,
		"report_type", reportType,
		"format", format,
		"date_range", dateRange,
		"outcome", entry.Outcome,
	)

	// Story 5.4, Task 4.6: Update existing TODO comments to reference this story (Story 5.4)
	// TODO comment cleanup done - now using persistent storage

	return nil
}


// LogBlockedSaleAttempt logs blocked sale attempts for expired products to append-only audit trail
// Story 4.6, AC6: Audit trail logging for blocked sale attempts (regulatory compliance)
// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
// Per NFR-SEC-009: 5-year minimum retention for Badan POM compliance
// Per NFR-SEC-011: Expiry date blocking is mandatory with audit logging
// Story 5.4, Task 4.5: Added ipAddress parameter for IP address extraction from request context
func (s *auditService) LogBlockedSaleAttempt(ctx context.Context, userID uint, username string, productID uint, productSKU string, productName string, expiryDate string, reason string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &userID,
		Username:  username,
		Action:    models.AuditActionBlockedSaleAttempt,
		IPAddress: ipAddress, // Story 5.4, Task 4.5: IP address extracted from request context
		Outcome:   "blocked",
		Reason:    fmt.Sprintf("Blocked sale attempt for expired product '%s' (ID: %d, SKU: %s, Expiry: %s): %s", productName, productID, productSKU, expiryDate, reason),
		Timestamp: time.Now().UTC(),
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback for MVP
	// Use Warn level for audit logs to ensure they're not filtered in production
	slog.Warn("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"user_id", userID,
		"username", username,
		"product_id", productID,
		"product_sku", productSKU,
		"product_name", productName,
		"expiry_date", expiryDate,
		"reason", reason,
		"outcome", entry.Outcome,
	)

	// Story 5.4, Task 4.6: Update existing TODO comments to reference this story (Story 5.4)
	// TODO comment cleanup done - now using persistent storage

	return nil
}

// LogSettingsUpdate logs system settings changes to append-only audit trail
// Story 6.1, AC7: Audit trail for configuration changes
// Story 6.4: Updated to use SYSTEM_SETTINGS_UPDATED action for Badan POM compliance
// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
// Logs admin_user_id, username, changes (JSON), timestamp, ip_address
func (s *auditService) LogSettingsUpdate(ctx context.Context, adminID uint, adminUsername string, changesJSON string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    models.AuditActionSystemSettingsUpdated, // Story 6.4: Use new system action constant
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    changesJSON, // Store changes as JSON in reason field
		Timestamp: time.Now().UTC(),
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback for MVP
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"changes", changesJSON,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogBackupCreated logs backup creation operations to append-only audit trail
// Story 6.4, Task 3: Audit logging for backup operations
// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
func (s *auditService) LogBackupCreated(ctx context.Context, adminID uint, adminUsername string, backupFile string, size int64, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    models.AuditActionBackupCreated,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Created backup: file=%s, size=%d bytes", backupFile, size),
		Timestamp: time.Now().UTC(),
	}

	_ = s.persistToDatabase(ctx, entry)

	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"backup_file", backupFile,
		"size_bytes", size,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogBackupRestored logs backup restore operations to append-only audit trail
// Story 6.4, Task 3: Audit logging for backup operations
// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
func (s *auditService) LogBackupRestored(ctx context.Context, adminID uint, adminUsername string, backupFile string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    models.AuditActionBackupRestored,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Restored backup from file: %s", backupFile),
		Timestamp: time.Now().UTC(),
	}

	_ = s.persistToDatabase(ctx, entry)

	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"backup_file", backupFile,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogBackupDeleted logs backup deletion operations to append-only audit trail
// Story 6.4, Task 3: Audit logging for backup operations
// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
func (s *auditService) LogBackupDeleted(ctx context.Context, adminID uint, adminUsername string, backupFile string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    models.AuditActionBackupDeleted,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Deleted backup file: %s", backupFile),
		Timestamp: time.Now().UTC(),
	}

	_ = s.persistToDatabase(ctx, entry)

	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"backup_file", backupFile,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogRoleUpdated logs role changes to append-only audit trail
// Story 6.4, Task 4: Audit logging for role management
// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
func (s *auditService) LogRoleUpdated(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, oldRole string, newRole string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    models.AuditActionRoleUpdated,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Changed role for user '%s' (ID: %d): %s → %s", targetUsername, targetUserID, oldRole, newRole),
		Timestamp: time.Now().UTC(),
	}

	_ = s.persistToDatabase(ctx, entry)

	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"target_user_id", targetUserID,
		"target_username", targetUsername,
		"old_role", oldRole,
		"new_role", newRole,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogPermissionGranted logs permission grant operations to append-only audit trail
// Story 6.4, Task 4: Audit logging for permission management
// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
func (s *auditService) LogPermissionGranted(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, permission string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    models.AuditActionPermissionGranted,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Granted permission '%s' to user '%s' (ID: %d)", permission, targetUsername, targetUserID),
		Timestamp: time.Now().UTC(),
	}

	_ = s.persistToDatabase(ctx, entry)

	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"target_user_id", targetUserID,
		"target_username", targetUsername,
		"permission", permission,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogPermissionRevoked logs permission revoke operations to append-only audit trail
// Story 6.4, Task 4: Audit logging for permission management
// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
func (s *auditService) LogPermissionRevoked(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, permission string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    models.AuditActionPermissionRevoked,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Revoked permission '%s' from user '%s' (ID: %d)", permission, targetUsername, targetUserID),
		Timestamp: time.Now().UTC(),
	}

	_ = s.persistToDatabase(ctx, entry)

	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"target_user_id", targetUserID,
		"target_username", targetUsername,
		"permission", permission,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogBranchCreated logs branch creation operations to append-only audit trail
// Story 6.4, Task 5: Audit logging for branch management
// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
func (s *auditService) LogBranchCreated(ctx context.Context, adminID uint, adminUsername string, branchName string, branchLocation string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    models.AuditActionBranchCreated,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Created branch '%s' at location '%s'", branchName, branchLocation),
		Timestamp: time.Now().UTC(),
	}

	_ = s.persistToDatabase(ctx, entry)

	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"branch_name", branchName,
		"branch_location", branchLocation,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogBranchUpdated logs branch update operations to append-only audit trail
// Story 6.4, Task 5: Audit logging for branch management
// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
func (s *auditService) LogBranchUpdated(ctx context.Context, adminID uint, adminUsername string, branchID uint, branchName string, changes string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    models.AuditActionBranchUpdated,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Updated branch '%s' (ID: %d): %s", branchName, branchID, changes),
		Timestamp: time.Now().UTC(),
	}

	_ = s.persistToDatabase(ctx, entry)

	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"branch_id", branchID,
		"branch_name", branchName,
		"changes", changes,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogBranchDeactivated logs branch deactivation operations to append-only audit trail
// Story 6.4, Task 5: Audit logging for branch management
// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
func (s *auditService) LogBranchDeactivated(ctx context.Context, adminID uint, adminUsername string, branchID uint, branchName string, reason string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    models.AuditActionBranchDeactivated,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Deactivated branch '%s' (ID: %d) - Reason: %s", branchName, branchID, reason),
		Timestamp: time.Now().UTC(),
	}

	_ = s.persistToDatabase(ctx, entry)

	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"branch_id", branchID,
		"branch_name", branchName,
		"reason", reason,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogSystemStartup logs system startup operations to append-only audit trail
// Story 6.4, Task 6: Audit logging for system operations
// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
func (s *auditService) LogSystemStartup(ctx context.Context, systemID string, serverInfo string, ipAddress string) error {
	// Use system user ID 0 for system-generated events
	systemUserID := uint(0)
	entry := AuditLogEntry{
		UserID:    &systemUserID,
		Username:  "system",
		Action:    models.AuditActionSystemStartup,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("System startup: %s - %s", systemID, serverInfo),
		Timestamp: time.Now().UTC(),
	}

	_ = s.persistToDatabase(ctx, entry)

	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"system_id", systemID,
		"server_info", serverInfo,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogSystemShutdown logs system shutdown operations to append-only audit trail
// Story 6.4, Task 6: Audit logging for system operations
// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
func (s *auditService) LogSystemShutdown(ctx context.Context, systemID string, reason string, ipAddress string) error {
	// Use system user ID 0 for system-generated events
	systemUserID := uint(0)
	entry := AuditLogEntry{
		UserID:    &systemUserID,
		Username:  "system",
		Action:    models.AuditActionSystemShutdown,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("System shutdown: %s - Reason: %s", systemID, reason),
		Timestamp: time.Now().UTC(),
	}

	_ = s.persistToDatabase(ctx, entry)

	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"system_id", systemID,
		"reason", reason,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogMaintenanceModeEnabled logs maintenance mode activation to append-only audit trail
// Story 6.4, Task 6: Audit logging for maintenance mode operations
// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
func (s *auditService) LogMaintenanceModeEnabled(ctx context.Context, adminID uint, adminUsername string, reason string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    models.AuditActionMaintenanceModeEnabled,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Maintenance mode enabled - Reason: %s", reason),
		Timestamp: time.Now().UTC(),
	}

	_ = s.persistToDatabase(ctx, entry)

	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"reason", reason,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogMaintenanceModeDisabled logs maintenance mode deactivation to append-only audit trail
// Story 6.4, Task 6: Audit logging for maintenance mode operations
// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
func (s *auditService) LogMaintenanceModeDisabled(ctx context.Context, adminID uint, adminUsername string, reason string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    models.AuditActionMaintenanceModeDisabled,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Maintenance mode disabled - Reason: %s", reason),
		Timestamp: time.Now().UTC(),
	}

	_ = s.persistToDatabase(ctx, entry)

	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"reason", reason,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}
