package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	_ "github.com/vahiiiid/go-rest-api-boilerplate/docs"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/db"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/handlers"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/jobs"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/migrate"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/server"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/whitelist"
)

// System user ID for audit logging (Story 6.4, CRIT-008)
// Use a dedicated system user ID (999) instead of "0" for system operations
const SystemUserID = "999"

//	@title			simpo Pharmacy Management System API
//	@version		1.0
//	@description	API for simpo pharmacy management system supporting POS, inventory, reporting, and multi-branch operations.
//	@termsOfService	https://simpo.com/terms

//	@contact.name	API Support
//	@contact.email	support@simpo.com

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@host		localhost:8080
//	@BasePath	/api/v1

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

// getServerInfo returns server identification information for audit logging
// Story 6.4, Task 2.3: Get server hostname and IP for system startup/shutdown audits
func getServerInfo() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Try to get primary IP address
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return fmt.Sprintf("%s (%s)", hostname, ipnet.IP.String())
				}
			}
		}
	}

	return hostname
}

func run() error {
	logger := slog.Default()
	logger.Info("Starting Go REST API Boilerplate...")

	cfg, err := config.LoadConfig("")
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		return err
	}

	if err := cfg.Validate(); err != nil {
		logger.Error("Configuration validation failed", "error", err)
		return err
	}

	cfg.LogSafeConfig(logger)

	database, err := db.NewPostgresDBFromDatabaseConfig(cfg.Database)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		return err
	}

	if os.Getenv("SKIP_MIGRATION_CHECK") == "" {
		if err := checkMigrationStatus(database, &cfg.Migrations); err != nil {
			logger.Warn("Migration check", "status", "⚠️", "error", err)
		} else {
			logger.Info("Migration check", "status", "✓")
		}
	}

	// Create services
	userRepo := user.NewRepository(database)
	userService := user.NewService(userRepo)

	// Create audit service for login audit logging (Story 1.5, AC7)
	// Story 5.4: Create audit repository for persistent audit trail storage
	auditRepo := repositories.NewAuditRepository(database)
	auditService := services.NewAuditService(auditRepo)

	// Story 10.7: Create supplier audit service for supplier transaction audit trail
	supplierAuditService := services.NewSupplierAuditService(database)

	// Create auth service with audit logging (Story 1.5, username-based login)
	authServiceForJWT := auth.NewServiceWithRepo(&cfg.JWT, database)
	newAuthService := services.NewAuthService(&cfg.JWT, userRepo, auditService)
	newAuthHandler := handlers.NewAuthHandler(newAuthService)

	// Create user handler
	userHandler := user.NewHandler(userService, authServiceForJWT, auditService)

	// Story 1.9: Create whitelist repository and service
	whitelistRepo := whitelist.NewRepository(database)
	whitelistService := whitelist.NewService(whitelistRepo)

	// Create adapter to convert whitelist.Repository to user.WhitelistRepository
	// This allows the user service to use the whitelist repository
	whitelistRepoAdapter := user.NewWhitelistRepoAdapter(whitelistRepo)
	userService.SetWhitelistRepo(whitelistRepoAdapter)

	// Story 1.9: Create verification repository
	verificationRepo := user.NewVerificationRepository(database)
	userService.SetVerificationRepo(verificationRepo)

	// Create whitelist handler
	whitelistHandler := whitelist.NewHandler(whitelistService)
	whitelistHandler.SetAuditService(auditService) // Story 1.9: Wire up audit service for whitelist operations

	// Story 1.8: Create Redis client and session manager
	var redisClient *redis.Client
	if cfg.Redis.Host != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
			Password: cfg.Redis.Password,
			DB:       0,
		})
	}

	// Create session manager for tracking and blocklist (Story 1.8)
	sessionManager := middleware.NewSessionManager(redisClient)
	userHandler.SetSessionManager(sessionManager)

	// Story 4.2: Create stock event service for real-time stock updates
	stockEventService := services.NewStockEventService(redisClient)
	// Story 4.2, Task 15: Create stock cache service for caching stock levels
	var stockCacheService *services.StockCacheService
	if redisClient != nil {
		stockCacheService = services.NewStockCacheService(redisClient)
	}

	// Story 4.4: Create alert service for low stock and expiry notifications
	alertService := services.NewAlertService(nil, auditService, redisClient)

	// Story 6.1: Create system settings repository, service, and handler
	systemSettingRepo := repositories.NewSystemSettingRepository(database)
	systemService := services.NewSystemService(systemSettingRepo, auditService)

	// Story 3.6: Create transaction repositories, service, and handler
	transactionRepo := repositories.NewTransactionRepository(database)
	transactionItemRepo := repositories.NewTransactionItemRepository(database)
	productRepo := repositories.NewProductRepository(database)

	// Story 4.1: Create product service and handler
	// Story 4.4: Add alertService parameter for low stock notifications
	// Story 4.6: Moved before transactionService creation (dependency)
	productService := services.NewProductService(productRepo, auditService, stockEventService, stockCacheService, alertService)
	productHandler := handlers.NewProductHandler(productService, stockEventService, cfg.JWT.Secret)

	// Story 4.6: Create transaction service with productService for expired product validation
	transactionService := services.NewTransactionService(transactionRepo, transactionItemRepo, productRepo, productService, auditService, stockEventService, alertService, systemService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	// Story 4.5: Create expiry check service and job
	expiryCheckService := services.NewExpiryCheckService(productRepo, alertService, redisClient, logger)
	var expiryCheckJob *jobs.ExpiryCheckJob
	if expiryCheckService != nil {
		expiryCheckJob = jobs.NewExpiryCheckJob(expiryCheckService, logger)
	}

	// Story 5.1, 5.2: Create report repository, service, and handler
	reportRepo := repositories.NewReportRepository(database)
	reportService := services.NewReportService(transactionRepo, productRepo, reportRepo, auditService, redisClient)

	// Story 5.3, Task 6: Create file storage service for exports
	exportStoragePath := os.Getenv("EXPORT_STORAGE_PATH")
	if exportStoragePath == "" {
		exportStoragePath = "/tmp/simpo-exports"
	}
	// Code review fix: CRITICAL-003 (Round 4) - Validate export storage path is within expected boundaries
	// Code review fix: CRITICAL-001 (Round 5) - Fix logic flaw: must be relative OR under /tmp
	if strings.Contains(exportStoragePath, "..") || (strings.HasPrefix(exportStoragePath, "/") && !strings.HasPrefix(exportStoragePath, "/tmp/")) {
		slog.Error("Invalid EXPORT_STORAGE_PATH: must be relative path (./) or within /tmp/")
		os.Exit(1)
	}
	fileStorage := services.NewInMemoryFileStorage(exportStoragePath, 100) // 100MB max

	// Story 5.3: Create export service with file storage
	// Story 6.1, AC6: Add systemService for business info in reports
	exportService := services.NewExportService(reportService, fileStorage, systemService)

	reportHandler := handlers.NewReportHandler(reportService, exportService, auditService)

	// Story 5.4: Create audit handler for audit log query and export APIs
	auditHandler := handlers.NewAuditHandler(auditRepo, auditService)

	// Story 6.1: Create system settings handler
	systemSettingsHandler := handlers.NewSystemSettingsHandler(systemService)

	// Story 6.3: Create backup service and handler
	backupService := services.NewBackupService(cfg, auditService) // Story 6.4: Pass audit service for backup audit logging
	backupHandler := handlers.NewBackupHandler(backupService)

	// Story 10.1: Create supplier service and handler
	supplierRepo := repositories.NewSupplierRepository(database)
	supplierService := services.NewSupplierService(supplierRepo, auditService, supplierAuditService)
	supplierHandler := handlers.NewSupplierHandler(supplierService)

	// Story 10.5: Create branch repository and supplier product catalog service early
	// These are needed by purchase invoice service for catalog price integration
	branchRepo := repositories.NewBranchRepository(database)
	supplierProductCatalogRepo := repositories.NewSupplierProductCatalogRepository(database)
	supplierProductCatalogService := services.NewSupplierProductCatalogService(supplierProductCatalogRepo, supplierRepo, productRepo, branchRepo, auditService)

	// Story 10.2: Create purchase invoice service and handler
	// Story 10.5, Task 9: Now receives supplierProductCatalogService for catalog price integration
	// Story 10.7: Added supplierAuditService for audit trail integration
	purchaseInvoiceRepo := repositories.NewPurchaseInvoiceRepository(database)
	purchaseInvoiceService := services.NewPurchaseInvoiceService(purchaseInvoiceRepo, supplierRepo, productRepo, auditService, supplierProductCatalogService, supplierAuditService)
	purchaseInvoiceHandler := handlers.NewPurchaseInvoiceHandler(purchaseInvoiceService)

	// Story 10.3: Create goods receipt service and handler
	// Story 10.7: Added supplierAuditService for audit trail integration
	goodsReceiptRepo := repositories.NewGoodsReceiptRepository(database)
	goodsReceiptService := services.NewGoodsReceiptService(database, goodsReceiptRepo, purchaseInvoiceRepo, productRepo, auditService, alertService, stockEventService, supplierAuditService)
	goodsReceiptHandler := handlers.NewGoodsReceiptHandler(goodsReceiptService)

	// Story 10.4: Create supplier payment service and handler
	// Story 10.7: Added supplierAuditService for audit trail integration
	supplierPaymentRepo := repositories.NewSupplierPaymentRepository(database)
	supplierPaymentService := services.NewSupplierPaymentService(database, supplierPaymentRepo, purchaseInvoiceRepo, supplierRepo, auditService, supplierAuditService)
	supplierPaymentHandler := handlers.NewSupplierPaymentHandler(supplierPaymentService)

	// Story 10.5: Create supplier product catalog handler
	supplierProductCatalogHandler := handlers.NewSupplierProductCatalogHandler(supplierProductCatalogService)

	// Story 10.6: Create supplier aging report handler
	supplierAgingReportService := services.NewSupplierAgingReportService(purchaseInvoiceRepo, supplierPaymentRepo, supplierRepo, auditService)
	supplierAgingReportHandler := handlers.NewSupplierAgingReportHandler(supplierAgingReportService)

	// Story 10.7: Create supplier audit trail handler
	supplierAuditHandler := handlers.NewSupplierAuditHandler(supplierAuditService)

	router := server.SetupRouter(userHandler, newAuthHandler, authServiceForJWT, cfg, database, whitelistHandler, transactionHandler, productHandler, reportHandler, auditHandler, systemSettingsHandler, backupHandler, supplierHandler, purchaseInvoiceHandler, goodsReceiptHandler, supplierPaymentHandler, supplierProductCatalogHandler, supplierAgingReportHandler, supplierAuditHandler, redisClient)

	// Story 4.2, Task 5: Start stock event broadcaster for real-time WebSocket updates
	if stockEventService != nil {
		ctx := context.Background()
		if err := stockEventService.StartBroadcaster(ctx); err != nil {
			logger.Warn("Failed to start stock event broadcaster", "error", err)
		} else {
			logger.Info("Stock event broadcaster started")
		}
	}

	// Story 4.5, Task 4.5: Start expiry check job as goroutine
	if expiryCheckJob != nil {
		ctx := context.Background()
		go expiryCheckJob.Start(ctx)
		logger.Info("Expiry check job started")

		// Story 6.3: Start backup scheduler if enabled
		if cfg.Backup.Enabled {
			ctx := context.Background()
			if err := backupService.StartScheduler(ctx); err != nil {
				logger.Warn("Failed to start backup scheduler", "error", err)
			} else {
				logger.Info("Backup scheduler started", "schedule", cfg.Backup.Schedule)
			}
		}
	}

	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	maxHeaderBytes := cfg.Server.MaxHeaderBytes
	if maxHeaderBytes == 0 {
		maxHeaderBytes = 1 << 20
	}

	srv := &http.Server{
		Addr:           fmt.Sprintf(":%s", port),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(cfg.Server.IdleTimeout) * time.Second,
		MaxHeaderBytes: maxHeaderBytes,
	}

	go func() {
		logger.Info("Server starting", "address", srv.Addr)
		logger.Info("Swagger UI available", "url", fmt.Sprintf("http://localhost:%s/api/docs/index.html", port))
		logger.Info("Health check available", "url", fmt.Sprintf("http://localhost:%s/health", port))
		logger.Info("Liveness probe available", "url", fmt.Sprintf("http://localhost:%s/health/live", port))
		logger.Info("Readiness probe available", "url", fmt.Sprintf("http://localhost:%s/health/ready", port))

		// Story 6.4, Task 2.3: Log system startup for compliance
		serverInfo := getServerInfo()
		if err := auditService.LogSystemStartup(context.Background(), SystemUserID, serverInfo, srv.Addr); err != nil {
			logger.Warn("Failed to log system startup to audit trail", "error", err)
		} else {
			logger.Info("System startup audit log created", "serverInfo", serverInfo, "address", srv.Addr)
		}

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info("Received shutdown signal", "signal", sig)
	logger.Info("Shutting down server gracefully...")

	// Story 4.2, Task 5.6: Stop stock event broadcaster
	if stockEventService != nil {
		logger.Info("Stopping stock event broadcaster...")
		stockEventService.StopBroadcaster()
		logger.Info("Stock event broadcaster stopped")
	}

	// Story 4.5, Task 4.3: Stop expiry check job
	if expiryCheckJob != nil {
		logger.Info("Stopping expiry check job...")
		expiryCheckJob.Stop()
		logger.Info("Expiry check job stopped")

		// Story 6.3: Stop backup scheduler
		if backupService != nil {
			logger.Info("Stopping backup scheduler...")
			ctx := context.Background()
			if err := backupService.StopScheduler(ctx); err != nil {
				logger.Warn("Failed to stop backup scheduler", "error", err)
			} else {
				logger.Info("Backup scheduler stopped")
			}
		}
	}

	// Code review fix: CRITICAL-005 - Shutdown audit service before database close
	if auditService != nil {
		logger.Info("Shutting down audit service...")
		ctx := context.Background()
		if err := auditService.Shutdown(ctx); err != nil {
			logger.Warn("Failed to shutdown audit service gracefully", "error", err)
		} else {
			logger.Info("Audit service shutdown complete")
		}
	}

	sqlDB, err := database.DB()
	if err == nil {
		logger.Info("Closing database connections...")
		if err := sqlDB.Close(); err != nil {
			logger.Error("Error closing database", "error", err)
		} else {
			logger.Info("Database connections closed successfully")
		}
	}

	shutdownTimeout := time.Duration(cfg.Server.ShutdownTimeout) * time.Second
	if shutdownTimeout == 0 {
		shutdownTimeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Story 6.4, Task 2.3: Log system shutdown for compliance
	serverInfo := getServerInfo()
	reason := fmt.Sprintf("Graceful shutdown initiated by signal %s on server %s", sig.String(), serverInfo)
	if err := auditService.LogSystemShutdown(ctx, SystemUserID, reason, srv.Addr); err != nil {
		logger.Warn("Failed to log system shutdown to audit trail", "error", err)
	} else {
		logger.Info("System shutdown audit log created", "reason", reason)
	}

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		return err
	}

	logger.Info("Server exited gracefully")
	return nil
}

func checkMigrationStatus(database *gorm.DB, cfg *config.MigrationsConfig) error {
	sqlDB, err := database.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	migrator, err := migrate.New(sqlDB, migrate.Config{
		MigrationsDir: cfg.Directory,
		Timeout:       time.Duration(cfg.Timeout) * time.Second,
		LockTimeout:   time.Duration(cfg.LockTimeout) * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	version, dirty, err := migrator.Version()
	if err != nil {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database in dirty state at version %d", version)
	}

	slog.Info("Database schema", "version", version)
	return nil
}
