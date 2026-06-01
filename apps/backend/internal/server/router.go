package server

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/handlers"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/health"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/whitelist"
)

// SetupRouter creates and configures the Gin router
// Story 6.1: Added systemSettingsHandler parameter
// Story 6.3: Added backupHandler parameter
// Story 10.1: Added supplierHandler parameter
// Story 10.2: Added purchaseInvoiceHandler parameter
// Story 10.3: Added goodsReceiptHandler parameter
// Story 10.4: Added supplierPaymentHandler parameter
// Story 10.5: Added supplierProductCatalogHandler parameter
// Story 10.6: Added supplierAgingReportHandler parameter
// Story 10.7: Added supplierAuditHandler parameter
func SetupRouter(userHandler *user.Handler, authHandler handlers.AuthHandler, authService auth.Service, cfg *config.Config, db *gorm.DB, whitelistHandler *whitelist.Handler, transactionHandler *handlers.TransactionHandler, productHandler handlers.ProductHandler, reportHandler *handlers.ReportHandler, auditHandler *handlers.AuditHandler, systemSettingsHandler handlers.SystemSettingsHandler, backupHandler *handlers.BackupHandler, supplierHandler *handlers.SupplierHandler, purchaseInvoiceHandler *handlers.PurchaseInvoiceHandler, goodsReceiptHandler *handlers.GoodsReceiptHandler, supplierPaymentHandler *handlers.SupplierPaymentHandler, supplierProductCatalogHandler *handlers.SupplierProductCatalogHandler, supplierAgingReportHandler *handlers.SupplierAgingReportHandler, supplierAuditHandler *handlers.SupplierAuditHandler, redisClient *redis.Client) *gin.Engine {
	router := gin.New()

	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	skipPaths := config.GetSkipPaths(cfg.App.Environment)
	loggerConfig := middleware.NewLoggerConfig(
		cfg.Logging.GetLogLevel(),
		skipPaths,
		cfg.Logging.IncludeCaller,
		cfg.Logging.RedactEnabled,
		cfg.Logging.RedactPatterns,
	)
	router.Use(middleware.Logger(loggerConfig))
	router.Use(errors.ErrorHandler())
	router.Use(gin.Recovery())

	// Story 9.4: Secure CORS configuration with environment-based origin validation
	// SECURITY FIX: Replaced insecure AllowAllOrigins=true with specific allowed origins
	if cfg.Cors.Enabled {
		// Validate CORS configuration to prevent runtime panics
		if len(cfg.Cors.AllowedOrigins) == 0 {
			// Use safe defaults if no origins configured
			cfg.Cors.AllowedOrigins = []string{"http://localhost:3000"}
		}
		if len(cfg.Cors.AllowedMethods) == 0 {
			// Use safe defaults if no methods configured
			cfg.Cors.AllowedMethods = []string{"GET", "POST", "OPTIONS"}
		}
		if len(cfg.Cors.AllowedHeaders) == 0 {
			// Use safe defaults if no headers configured
			cfg.Cors.AllowedHeaders = []string{"Authorization", "Content-Type"}
		}

		// Validate MaxAge to prevent integer overflow
		// MaxAge should be between 0 and 9223372036 seconds (~292 years)
		maxAge := cfg.Cors.MaxAge
		if maxAge < 0 || maxAge > 9223372036 {
			maxAge = 86400 // Default to 24 hours if invalid
		}

		corsConfig := cors.Config{
			AllowOrigins:     cfg.Cors.AllowedOrigins,   // Specific origins from config (not wildcard)
			AllowMethods:     cfg.Cors.AllowedMethods,   // GET, POST, PUT, DELETE, OPTIONS
			AllowHeaders:     cfg.Cors.AllowedHeaders,   // Authorization, Content-Type, X-Requested-With
			AllowCredentials: cfg.Cors.AllowCredentials, // Support cookies and auth headers
			MaxAge:           time.Duration(maxAge) * time.Second, // Pre-flight cache (24h default)
			ExposeHeaders:    []string{"Content-Length"},
		}
		router.Use(cors.New(corsConfig))
	}

	var checkers []health.Checker
	if cfg.Health.DatabaseCheckEnabled {
		dbChecker := health.NewDatabaseChecker(db)
		checkers = append(checkers, dbChecker)
	}
	// Story 9.1: Add Redis checker if Redis is configured
	if redisClient != nil {
		redisChecker := health.NewRedisChecker(redisClient)
		checkers = append(checkers, redisChecker)
	}
	healthService := health.NewService(checkers, cfg.App.Version, cfg.App.Environment)
	healthHandler := health.NewHandler(healthService)

	// Legacy health endpoints (keep for backward compatibility)
	router.GET("/health", healthHandler.Health)
	router.GET("/health/live", healthHandler.Live)
	router.GET("/health/ready", healthHandler.Ready)
	router.GET("/health/db", healthHandler.Database) // Story 2.4: Database-specific health check

	// Story 9.1: API versioned health endpoint
	router.GET("/api/v1/health", healthHandler.Health)

	// Story 6.2: Create admin health monitoring components
	metricsCollector := health.NewMetricsCollector(time.Now(), cfg.App.Version, cfg.App.Environment)

	// Use config values with fallback defaults (ensures alerts work even if config is missing)
	errorRateMax := cfg.Health.ErrorRateMax
	if errorRateMax == 0 {
		errorRateMax = 0.1 // Default: 0.1% error rate threshold
	}
	diskFreeMin := cfg.Health.DiskFreeMin
	if diskFreeMin == 0 {
		diskFreeMin = 20.0 // Default: 20% free disk space threshold
	}

	alertService := health.NewAlertService(dto.AlertThresholdsConfig{
		ErrorRateMax: errorRateMax,
		DiskFreeMin:  diskFreeMin,
	})
	adminHealthHandler := handlers.NewAdminHealthHandler(healthService, metricsCollector, alertService, checkers)

	// Swagger UI for API documentation
	router.GET("/api/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// Legacy swagger endpoint for backward compatibility
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	rlCfg := cfg.Ratelimit
	if rlCfg.Enabled {
		router.Use(
			middleware.NewRateLimitMiddleware(
				rlCfg.Window,
				rlCfg.Requests,
				func(c *gin.Context) string {
					// Story 9.3: Try to get user ID from JWT context first
					// Auth middleware sets "user" context key with auth.Claims
					if userValue, exists := c.Get("user"); exists {
						// Type assertion with safety checks - handle nil and invalid types
						if claims, ok := userValue.(*auth.Claims); ok && claims != nil && claims.UserID > 0 {
							// Track by user ID for authenticated requests
							return fmt.Sprintf("user:%d", claims.UserID)
						}
					}
					// Fallback to IP for unauthenticated requests
					ip := c.ClientIP()
					if ip == "" {
						ip = c.GetHeader("X-Forwarded-For")
						if ip == "" {
							ip = c.GetHeader("X-Real-IP")
						}
						if ip == "" {
							ip = "unknown"
						}
					}
					return fmt.Sprintf("ip:%s", ip)
				},
				nil,
			),
		)
	}

	// Story 1.8: Create Redis client for session tracking and token blocklist
	// Story 1.8, Task 1: Session tracking mechanism (Redis)
	// Use passed redisClient if available, otherwise create local one (backward compatibility)
	if redisClient == nil && cfg.Redis.Host != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
			Password: cfg.Redis.Password,
			DB:       0,
		})
	}

	// Create session manager for tracking and blocklist (Story 1.8)
	sessionManager := middleware.NewSessionManager(redisClient)

	// Story 1.6, AC5: Register Protected Routes
	// Public routes bypass RBAC middleware: login, register, health check
	// Protected routes require JWT auth and RBAC middleware
	// Story 1.8, Task 6: Middleware order: CORS → Rate Limit → Auth (with session tracking) → RBAC → Handler

	v1 := router.Group("/api/v1")
	{
		// Public auth routes - no authentication required
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", userHandler.Register)
			authGroup.POST("/login", authHandler.Login) // Story 1.5: Username-based login with JWT
		}

		// Story 1.9: Staff self-registration endpoints with stricter rate limiting
		// These are public endpoints that are vulnerable to abuse, so we apply stricter limits
		authStrictGroup := v1.Group("/auth")
		if rlCfg.Enabled {
			authStrictGroup.Use(
				middleware.NewRateLimitMiddleware(
					15*time.Minute, // 15 minute window
					5,              // Max 5 requests per window (stricter than global limit)
					func(c *gin.Context) string {
						ip := c.ClientIP()
						if ip == "" {
							ip = c.GetHeader("X-Forwarded-For")
							if ip == "" {
								ip = c.GetHeader("X-Real-IP")
							}
							if ip == "" {
								ip = "unknown"
							}
						}
						return "auth-strict:" + ip
					},
					nil,
				),
			)
		}
		{
			authStrictGroup.POST("/register-staff", userHandler.RegisterStaff)
			authStrictGroup.POST("/verify-email", userHandler.VerifyEmail)
		}

		// Protected auth routes - require authentication with session tracking (Story 1.8)
		authProtectedGroup := v1.Group("/auth")
		// Story 1.8, Task 6: Use SessionAuthMiddleware for session tracking and blocklist checking
		authProtectedGroup.Use(auth.SessionAuthMiddleware(authService, sessionManager))
		{
			authProtectedGroup.POST("/refresh", userHandler.RefreshToken)
			authProtectedGroup.POST("/logout", userHandler.Logout)
			authProtectedGroup.GET("/me", userHandler.GetMe)
		}

		// Story 1.9: Whitelist management routes - SYSTEM_ADMIN only
		whitelistGroup := v1.Group("/whitelist")
		whitelistGroup.Use(auth.SessionAuthMiddleware(authService, sessionManager), middleware.RBACMiddleware())
		{
			whitelistGroup.POST("", whitelistHandler.AddDomain)
			whitelistGroup.GET("", whitelistHandler.ListDomains)
			whitelistGroup.GET("/:id", whitelistHandler.GetDomain)
			whitelistGroup.PUT("/:id", whitelistHandler.UpdateDomain)
			whitelistGroup.DELETE("/:id", whitelistHandler.DeleteDomain)
		}

		// Story 1.6, AC5: Protected routes with RBAC middleware
		// These routes enforce role-based access control
		// Story 1.8, Task 6: Middleware chain: Auth (JWT validation + session tracking) → RBAC → Handler

		// User endpoints - OWNER and SYSTEM_ADMIN can access
		// POST /api/v1/users requires SYSTEM_ADMIN only (via permissions)
		usersGroup := v1.Group("/users")
		// Story 1.8, Task 6: Use SessionAuthMiddleware for session tracking
		usersGroup.Use(auth.SessionAuthMiddleware(authService, sessionManager), middleware.RBACMiddleware())
		{
			usersGroup.POST("", userHandler.CreateUser) // Story 1.7: SYSTEM_ADMIN only
			usersGroup.GET("", userHandler.ListUsers)
			usersGroup.GET("/:id", userHandler.GetUser)
			usersGroup.PUT("/:id", userHandler.UpdateUser)
			usersGroup.DELETE("/:id", userHandler.DeleteUser)
			usersGroup.PUT("/:id/deactivate", userHandler.DeactivateUser) // Story 1.10: SYSTEM_ADMIN only
		}

		// Admin endpoints - SYSTEM_ADMIN only
		adminGroup := v1.Group("/admin")
		adminGroup.Use(auth.SessionAuthMiddleware(authService, sessionManager), middleware.RBACMiddleware())
		{
			// Story 6.1: System settings endpoints - SYSTEM_ADMIN only
			if systemSettingsHandler != nil {
				adminGroup.GET("/settings", systemSettingsHandler.GetSettings)
				adminGroup.PUT("/settings", systemSettingsHandler.UpdateSettings)
			}
			// Story 6.2: Admin health monitoring endpoints - ADMIN and SYSTEM_ADMIN only
			adminGroup.GET("/health/dashboard", adminHealthHandler.GetDashboard)
			adminGroup.GET("/health/alerts", adminHealthHandler.GetAlerts)
			adminGroup.GET("/health/metrics", adminHealthHandler.GetMetrics)
		}

		// Story 6.1: Public settings endpoint (no authentication required - for receipts/reports)
		if systemSettingsHandler != nil {
			v1.GET("/settings/public", systemSettingsHandler.GetPublicSettings)
		}

		// Story 3.6: Transaction endpoints - require authentication
		// Cashiers can create transactions for sales
		transactionsGroup := v1.Group("/transactions")
		transactionsGroup.Use(auth.SessionAuthMiddleware(authService, sessionManager))

		// MEDIUM FIX: Rate limiting should be applied to prevent DoS attacks
		// TODO: Add rate limiting middleware: transactionsGroup.Use(middleware.RateLimit(100, time.Minute))
		{
			transactionsGroup.POST("", transactionHandler.CreateTransaction)
			// Story 3.7: Transaction history and detail endpoints
			transactionsGroup.GET("", transactionHandler.ListTransactions)       // List with filters and pagination
			transactionsGroup.GET("/:id", transactionHandler.GetTransactionByID) // Get transaction details
		}

		// Story 4.1: Product endpoints - require authentication
		// Owners and Cashiers can view products with RBAC for branch access
		if productHandler != nil {
			productsGroup := v1.Group("/products")
			productsGroup.Use(auth.SessionAuthMiddleware(authService, sessionManager), middleware.RBACMiddleware())
			{
				productsGroup.GET("", productHandler.ListProducts) // List with search, filters, and pagination
				// Story 4.4, Task 5.1-5.4: Low stock products endpoint
				productsGroup.GET("/low-stock", productHandler.GetLowStockProducts)
				// Story 4.5, Task 5.1-5.5: Expiring products endpoint
				productsGroup.GET("/expiring", productHandler.GetExpiringProducts)
				// Story 4.2, Task 4.2: WebSocket endpoint for real-time stock updates
				productsGroup.GET("/stock/subscribe", productHandler.SubscribeStockUpdates)
				// Story 4.3, Task 4.2: Stock adjustment endpoint with admin permissions
				productsGroup.POST("/stock/adjust", productHandler.AdjustStock)
				// Story 4.6, Task 6: Barcode scan endpoint with expired blocking
				productsGroup.GET("/sku/:sku", productHandler.GetProductBySKU)
			}
		}

		// Story 5.1, 5.2, 5.3: Financial report endpoints - require authentication and RBAC
		// Only Owner and Admin can access financial reports
		if reportHandler != nil {
			reportsGroup := v1.Group("/reports")
			reportsGroup.Use(auth.SessionAuthMiddleware(authService, sessionManager), middleware.RBACMiddleware())
			{
				// Story 5.1: Daily sales summary report
				reportsGroup.GET("/daily", reportHandler.GetDailySalesReport)
				// Story 5.2: Profit/Loss report
				reportsGroup.GET("/profit-loss", reportHandler.GetProfitLossReport)
				// Story 5.3, Task 4.1-4.8: Report export endpoints (PDF and Excel)
				reportsGroup.GET("/daily/export", reportHandler.ExportDailySalesReport)
				reportsGroup.GET("/profit-loss/export", reportHandler.ExportProfitLossReport)
			}

			// Story 5.4: Audit log endpoints - require authentication and RBAC
			// Only Owner, Admin, and SystemAdmin can access audit logs
			if auditHandler != nil {
				auditGroup := v1.Group("/audit")
				auditGroup.Use(auth.SessionAuthMiddleware(authService, sessionManager), middleware.RBACMiddleware())
				{
					// Story 5.4, Task 5: Query audit logs with filters and pagination
					auditGroup.GET("/logs", auditHandler.GetAuditLogs)
					// Story 5.4, Task 6: Export audit logs in CSV or JSON format
					auditGroup.GET("/logs/export", auditHandler.GetAuditLogsExport)
					// Story 5.4, Task 7: Manual retention cleanup (SystemAdmin only)
					auditGroup.POST("/cleanup", auditHandler.CleanupAuditLogs)
				}
			}

			// Story 10.7: Supplier audit trail endpoints - require authentication and RBAC
			// Only Admin and Owner can access supplier audit trail
			if supplierAuditHandler != nil {
				supplierAuditGroup := v1.Group("/audit/supplier")
				supplierAuditGroup.Use(auth.SessionAuthMiddleware(authService, sessionManager), middleware.RBACMiddleware())
				{
					// Story 10.7, AC2: Query supplier audit trail with filters
					supplierAuditGroup.GET("", supplierAuditHandler.QueryAuditTrail)
					// Story 10.7, AC2: Get audit trail for specific entity
					supplierAuditGroup.GET("/entity/:type/:id", supplierAuditHandler.GetAuditByEntity)
					// Story 10.7, AC2: Get audit trail for specific user
					supplierAuditGroup.GET("/user/:id", supplierAuditHandler.GetAuditByUser)
					// Story 10.7, AC3: Export audit trail in CSV format
					supplierAuditGroup.GET("/export/csv", supplierAuditHandler.ExportAuditTrailCSV)
					// Story 10.7, AC3: Export audit trail in PDF format
					supplierAuditGroup.GET("/export/pdf", supplierAuditHandler.ExportAuditTrailPDF)
				}
			}
		}

		// Story 10.1: Supplier management endpoints - require authentication and RBAC
		// Only Admin can manage suppliers (CRUD), Owner and Admin can view
		if supplierHandler != nil {
			suppliersGroup := v1.Group("/suppliers")
			suppliersGroup.Use(auth.SessionAuthMiddleware(authService, sessionManager), middleware.RBACMiddleware())
			{
				// Story 10.1, AC1: Create supplier - Admin only
				suppliersGroup.POST("", supplierHandler.CreateSupplier)
				// Story 10.1, AC1: Get supplier by ID - Admin and Owner
				suppliersGroup.GET("/:id", supplierHandler.GetSupplier)
				// Story 10.1, AC2: List suppliers - Admin and Owner
				suppliersGroup.GET("", supplierHandler.ListSuppliers)
				// Story 10.1, AC2: Update supplier - Admin only
				suppliersGroup.PUT("/:id", supplierHandler.UpdateSupplier)
				// Story 10.1, AC3: Deactivate supplier - Admin only
				suppliersGroup.DELETE("/:id", supplierHandler.DeactivateSupplier)
			}
		}

		// Story 10.2: Purchase invoice management endpoints - require authentication and RBAC
		// Only Admin can manage purchase invoices (CRUD), Owner and Admin can view
		if purchaseInvoiceHandler != nil {
			purchaseInvoicesGroup := v1.Group("/purchase-invoices")
			purchaseInvoicesGroup.Use(auth.SessionAuthMiddleware(authService, sessionManager), middleware.RBACMiddleware())
			{
				// Story 10.2, AC1: Create purchase invoice - Admin only
				purchaseInvoicesGroup.POST("", purchaseInvoiceHandler.CreatePurchaseInvoice)
				// Story 10.2, AC3: Get purchase invoice by ID - Admin and Owner
				purchaseInvoicesGroup.GET("/:id", purchaseInvoiceHandler.GetPurchaseInvoice)
				// Story 10.2, AC2: List purchase invoices - Admin and Owner
				purchaseInvoicesGroup.GET("", purchaseInvoiceHandler.ListPurchaseInvoices)
				// Story 10.2: Update purchase invoice - Admin only
				purchaseInvoicesGroup.PUT("/:id", purchaseInvoiceHandler.UpdatePurchaseInvoice)
				// Story 10.2: Delete purchase invoice - Admin only
				purchaseInvoicesGroup.DELETE("/:id", purchaseInvoiceHandler.DeletePurchaseInvoice)
			}
		}

		// Story 10.3: Goods receipt management endpoints - require authentication and RBAC
		// Only Admin and Owner can process goods receipts
		if goodsReceiptHandler != nil {
			goodsReceiptsGroup := v1.Group("/goods-receipts")
			goodsReceiptsGroup.Use(auth.SessionAuthMiddleware(authService, sessionManager), middleware.RBACMiddleware())
			{
				// Story 10.3, AC1: Process goods receipt - Admin and Owner
				goodsReceiptsGroup.POST("/process", goodsReceiptHandler.ProcessGoodsReceipt)
				// Story 10.3: Get goods receipt by ID - Admin and Owner
				goodsReceiptsGroup.GET("/:id", goodsReceiptHandler.GetGoodsReceipt)
				// Story 10.3: List goods receipts - Admin and Owner
				goodsReceiptsGroup.GET("", goodsReceiptHandler.ListGoodsReceipts)
			}
		}

		// Story 10.4: Supplier payment management endpoints - require authentication and RBAC
		// Only Admin and Owner can record and view payments
		if supplierPaymentHandler != nil {
			supplierPaymentsGroup := v1.Group("/supplier-payments")
			supplierPaymentsGroup.Use(auth.SessionAuthMiddleware(authService, sessionManager), middleware.RBACMiddleware())
			{
				// Story 10.4, AC1: Record payment - Admin and Owner
				supplierPaymentsGroup.POST("", supplierPaymentHandler.RecordPayment)
				// Story 10.4: Get payment by ID - Admin and Owner
				supplierPaymentsGroup.GET("/:id", supplierPaymentHandler.GetSupplierPayment)
				// Story 10.4: List payments - Admin and Owner
				supplierPaymentsGroup.GET("", supplierPaymentHandler.ListSupplierPayments)
			}
		}

		// Story 10.4, AC2: Supplier payment history endpoints - require authentication and RBAC
		// Only Admin and Owner can view payment history by supplier
		if supplierPaymentHandler != nil {
			supplierHistoryGroup := v1.Group("/suppliers/:id")
			supplierHistoryGroup.Use(auth.SessionAuthMiddleware(authService, sessionManager), middleware.RBACMiddleware())
			{
				// Story 10.4, AC2: Get payment history by supplier - Admin and Owner
				supplierHistoryGroup.GET("/payment-history", supplierPaymentHandler.GetPaymentHistoryBySupplier)
			}
			}
			// Story 10.5: Supplier product catalog management endpoints - require authentication and RBAC
			// Only Admin and Owner can manage product catalogs and view pricing
			if supplierProductCatalogHandler != nil {
				catalogGroup := v1.Group("/supplier-product-catalogs")
				catalogGroup.Use(auth.SessionAuthMiddleware(authService, sessionManager), middleware.RBACMiddleware())
				{
					// Story 10.5, AC1: Associate product with supplier - Admin only
					catalogGroup.POST("", supplierProductCatalogHandler.AssociateProduct)
					// Story 10.5: Get catalog entry by ID - Admin and Owner
					catalogGroup.GET("/:id", supplierProductCatalogHandler.GetProductCatalog)
					// Story 10.5: List catalog entries - Admin and Owner
					catalogGroup.GET("", supplierProductCatalogHandler.ListProductCatalogs)
					// Story 10.5, AC1: Update purchase price - Admin only
					catalogGroup.PUT("/:id/price", supplierProductCatalogHandler.UpdatePurchasePrice)
					// Story 10.5, AC1: Set preferred supplier - Admin only
					catalogGroup.PUT("/:id/preferred", supplierProductCatalogHandler.SetPreferredSupplier)
				}

				// Story 10.5, AC1: Product price history endpoint - require authentication and RBAC
				productPriceGroup := v1.Group("/products/:id")
				productPriceGroup.Use(auth.SessionAuthMiddleware(authService, sessionManager), middleware.RBACMiddleware())
				{
					// Story 10.5, AC1: Get price history for a product - Admin and Owner
					productPriceGroup.GET("/price-history", supplierProductCatalogHandler.GetPriceHistory)
					// Story 10.5: Get preferred supplier for a product - Admin and Owner
					productPriceGroup.GET("/preferred-supplier", supplierProductCatalogHandler.GetPreferredSupplier)
				}

				// Story 10.5: Supplier catalog endpoint - require authentication and RBAC
				supplierCatalogGroup := v1.Group("/suppliers/:id")
				supplierCatalogGroup.Use(auth.SessionAuthMiddleware(authService, sessionManager), middleware.RBACMiddleware())
				{
					// Story 10.5: Get supplier's product catalog - Admin and Owner
					supplierCatalogGroup.GET("/product-catalog", supplierProductCatalogHandler.GetSupplierCatalog)
				}
			}
		}



		// Story 10.6: Supplier aging report endpoints - require authentication and RBAC
		// Only Owner can generate and export aging reports (critical financial data)
		if supplierAgingReportHandler != nil {
			reportsGroup := v1.Group("/reports")
			reportsGroup.Use(auth.SessionAuthMiddleware(authService, sessionManager), middleware.RBACMiddleware())
			{
				// Story 10.6, AC1: Generate aging report - Owner only
				reportsGroup.POST("/supplier-aging", supplierAgingReportHandler.GenerateAgingReport)
				// Story 10.6, AC1: Export aging report as PDF - Owner only
				reportsGroup.POST("/supplier-aging/export/pdf", supplierAgingReportHandler.ExportAgingReportPDF)
				// Story 10.6, AC1: Export aging report as Excel - Owner only
				reportsGroup.POST("/supplier-aging/export/excel", supplierAgingReportHandler.ExportAgingReportExcel)
			}
		}

	return router
}
