package server

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/handlers"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/health"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/whitelist"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(userHandler *user.Handler, authHandler handlers.AuthHandler, authService auth.Service, cfg *config.Config, db *gorm.DB, whitelistHandler *whitelist.Handler, transactionHandler *handlers.TransactionHandler, productHandler handlers.ProductHandler, reportHandler *handlers.ReportHandler, auditHandler *handlers.AuditHandler, redisClient *redis.Client) *gin.Engine {
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
	)
	router.Use(middleware.Logger(loggerConfig))
	router.Use(errors.ErrorHandler())
	router.Use(gin.Recovery())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	router.Use(cors.New(corsConfig))

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

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	rlCfg := cfg.Ratelimit
	if rlCfg.Enabled {
		router.Use(
			middleware.NewRateLimitMiddleware(
				rlCfg.Window,
				rlCfg.Requests,
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
					return ip
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
					5,               // Max 5 requests per window (stricter than global limit)
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
			usersGroup.POST("", userHandler.CreateUser)    // Story 1.7: SYSTEM_ADMIN only
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
			// Admin settings endpoint (placeholder for future admin functionality)
			adminGroup.GET("/settings", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "admin settings"})
			})
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
			transactionsGroup.GET("", transactionHandler.ListTransactions)  // List with filters and pagination
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
		}
	}

	return router
}
