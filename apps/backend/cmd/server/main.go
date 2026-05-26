package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	_ "github.com/vahiiiid/go-rest-api-boilerplate/api/docs"
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

// @title Go REST API Boilerplate
// @version 1.0
// @description A production-ready REST API boilerplate in Go with JWT authentication
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
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
		transactionService := services.NewTransactionService(transactionRepo, transactionItemRepo, productRepo, productService, auditService, stockEventService, alertService)
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
		exportService := services.NewExportService(reportService, fileStorage)

		reportHandler := handlers.NewReportHandler(reportService, exportService, auditService)

		// Story 5.4: Create audit handler for audit log query and export APIs
		auditHandler := handlers.NewAuditHandler(auditRepo, auditService)

		router := server.SetupRouter(userHandler, newAuthHandler, authServiceForJWT, cfg, database, whitelistHandler, transactionHandler, productHandler, reportHandler, auditHandler, redisClient)

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
		logger.Info("Swagger UI available", "url", fmt.Sprintf("http://localhost:%s/swagger/index.html", port))
		logger.Info("Health check available", "url", fmt.Sprintf("http://localhost:%s/health", port))
		logger.Info("Liveness probe available", "url", fmt.Sprintf("http://localhost:%s/health/live", port))
		logger.Info("Readiness probe available", "url", fmt.Sprintf("http://localhost:%s/health/ready", port))

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
