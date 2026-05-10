package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/handlers"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/health"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(userHandler *user.Handler, authHandler handlers.AuthHandler, authService auth.Service, cfg *config.Config, db *gorm.DB) *gin.Engine {
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
	healthService := health.NewService(checkers, cfg.App.Version, cfg.App.Environment)
	healthHandler := health.NewHandler(healthService)

	router.GET("/health", healthHandler.Health)
	router.GET("/health/live", healthHandler.Live)
	router.GET("/health/ready", healthHandler.Ready)

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

	// Story 1.6, AC5: Register Protected Routes
	// Public routes bypass RBAC middleware: login, register, health check
	// Protected routes require JWT auth and RBAC middleware
	// Middleware order: CORS → Rate Limit → Auth → RBAC → Handler

	v1 := router.Group("/api/v1")
	{
		// Public auth routes - no authentication required
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", userHandler.Register)
			authGroup.POST("/login", authHandler.Login) // Story 1.5: Username-based login with JWT
		}

		// Protected auth routes - require authentication but no RBAC (self-access only)
		authProtectedGroup := v1.Group("/auth")
		authProtectedGroup.Use(auth.AuthMiddleware(authService))
		{
			authProtectedGroup.POST("/refresh", userHandler.RefreshToken)
			authProtectedGroup.POST("/logout", userHandler.Logout)
			authProtectedGroup.GET("/me", userHandler.GetMe)
		}

		// Story 1.6, AC5: Protected routes with RBAC middleware
		// These routes enforce role-based access control
		// Middleware chain: Auth (JWT validation) → RBAC (role-based permissions) → Handler

		// User endpoints - OWNER and SYSTEM_ADMIN can access
		usersGroup := v1.Group("/users")
		usersGroup.Use(auth.AuthMiddleware(authService), middleware.RBACMiddleware())
		{
			usersGroup.GET("", userHandler.ListUsers)
			usersGroup.GET("/:id", userHandler.GetUser)
			usersGroup.PUT("/:id", userHandler.UpdateUser)
			usersGroup.DELETE("/:id", userHandler.DeleteUser)
		}

		// Admin endpoints - SYSTEM_ADMIN only
		adminGroup := v1.Group("/admin")
		adminGroup.Use(auth.AuthMiddleware(authService), middleware.RBACMiddleware())
		{
			// Admin settings endpoint (placeholder for future admin functionality)
			adminGroup.GET("/settings", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "admin settings"})
			})
		}
	}

	return router
}
