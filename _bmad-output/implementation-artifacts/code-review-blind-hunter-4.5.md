# Blind Hunter Review - Story 4.5: Implement Expiry Date Alerts

## Your Role

You are a **Blind Hunter** code reviewer. You receive **ONLY the diff output below** — no spec, no context docs, no project access. Your mission is to find defects using pure code analysis and adversarial thinking.

## Diff to Review

```
diff --git a/_bmad-output/implementation-artifacts/sprint-status.yaml b/_bmad-output/implementation-artifacts/sprint-status.yaml
index a808414..0c69dfb 100644
--- a/_bmad-output/implementation-artifacts/sprint-status.yaml
+++ b/_bmad-output/implementation-artifacts/sprint-status.yaml
@@ -1,5 +1,5 @@
 # generated: 2026-05-09
-# last_updated: 2026-05-14T15:30:00+07:00
+# last_updated: 2026-05-21T12:00:00+07:00
 # project: simpo
 # project_key: NOKEY
 # tracking_system: file-system
@@ -35,7 +35,7 @@
 # - Dev moves story to 'review', then runs code-review (fresh context, different LLM recommended)
 
 generated: 2026-05-09
-last_updated: 2026-05-20T16:30:00+07:00
+last_updated: 2026-05-21T18:30:00+07:00
 project: simpo
 project_key: NOKEY
 tracking_system: file-system
@@ -82,7 +82,7 @@ development_status:
   4-2-implement-real-time-stock-visibility: done # Code review patches applied (2026-05-19)
   4-3-implement-manual-stock-adjustment: done
   4-4-implement-low-stock-notifications: done # All tasks complete, tests passing
-  4-5-implement-expiry-date-alerts: backlog
+  4-5-implement-expiry-date-alerts: review # All implementation complete - backend, web, mobile (2026-05-21)
   4-6-prevent-sale-of-expired-medications: backlog
   epic-4-retrospective: optional

diff --git a/apps/backend/cmd/server/main.go b/apps/backend/cmd/server/main.go
index 45dae7a..f2e8edf 100644
--- a/apps/backend/cmd/server/main.go
+++ b/apps/backend/cmd/server/main.go
@@ -18,6 +18,7 @@ import (
 	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
 	"github.com/vahiiiid/go-rest-api-boilerplate/internal/db"
 	"github.com/vahiiiid/go-rest-api-boilerplate/internal/handlers"
+	"github.com/vahiiiid/go-rest-api-boilerplate/internal/jobs"
 	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
 	"github.com/vahiiiid/go-rest-api-boilerplate/internal/migrate"
 	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
@@ -153,6 +154,13 @@ func run() error {
 	productService := services.NewProductService(productRepo, auditService, stockEventService, stockCacheService, alertService, cfg)
 	productHandler := handlers.NewProductHandler(productService, stockEventService, cfg.JWT.Secret)
 
+	// Story 4.5: Create expiry check service and job
+	expiryCheckService := services.NewExpiryCheckService(productRepo, alertService, redisClient, logger)
+	var expiryCheckJob *jobs.ExpiryCheckJob
+	if expiryCheckService != nil {
+		expiryCheckJob = jobs.NewExpiryCheckJob(expiryCheckService, logger)
+	}
 
 	router := server.SetupRouter(userHandler, newAuthHandler, authServiceForJWT, cfg, database, whitelistHandler, transactionHandler, productHandler)
 
@@ -166,6 +174,13 @@ func run() error {
 		}
 	}
 
+	// Story 4.5, Task 4.5: Start expiry check job as goroutine
+	if expiryCheckJob != nil {
+		ctx := context.Background()
+		go expiryCheckJob.Start(ctx)
+		logger.Info("Expiry check job started")
+	}
 
 	port := cfg.Server.Port
 	if port == "" {
@@ -212,6 +227,13 @@ func run() error {
 		logger.Info("Stock event broadcaster stopped")
 	}
 
+	// Story 4.5, Task 4.3: Stop expiry check job
+	if expiryCheckJob != nil {
+		logger.Info("Stopping expiry check job...")
+		expiryCheckJob.Stop()
+		logger.Info("Expiry check job stopped")
+	}
 
 	sqlDB, err := database.DB()
 	if err == nil {
diff --git a/apps/backend/internal/dto/product_dto.go b/apps/backend/internal/dto/product_dto.go
index c550343..88f85ee 100644
--- a/apps/backend/internal/dto/product_dto.go
+++ b/apps/backend/internal/dto/product_dto.go
@@ -140,3 +140,25 @@ type PaginationMetadata struct {
 	Total      int64 `json:"total"`
 	TotalPages int  `json:"totalPages"`
 }
+
+// ExpiryAlertEvent represents an expiry date alert notification event
+// Story 4.5, AC4, AC6: Event structure for expiry date alerts
+type ExpiryAlertEvent struct {
+	EventID   string            `json:"eventId"`   // UUID for event tracking
+	EventType string            `json:"eventType"` // "product.expiry"
+	Timestamp string            `json:"timestamp"` // ISO 8601 timestamp
+	Data      ProductExpiryData `json:"data"`      // Expiry alert details
+}
+
+// ProductExpiryData contains product information for expiry notifications
+// Story 4.5, AC6: Product details for notification payload
+type ProductExpiryData struct {
+	ProductID     uint   `json:"productId"`
+	SKU           string `json:"sku"`
+	ProductName   string `json:"productName"`
+	ExpiryDate    string `json:"expiryDate"`    // ISO 8601 format
+	DaysRemaining int    `json:"daysRemaining"` // 30, 14, or 7
+	AlertLevel    string `json:"alertLevel"`    // "warning", "critical", "urgent"
+	BranchID      uint   `json:"branchId"`
+	BranchName    string `json:"branchName"`
+}

diff --git a/apps/backend/internal/handlers/product_handler.go b/apps/backend/internal/handlers/product_handler.go
index 783d3d5..79d83d8 100644
--- a/apps/backend/internal/handlers/product_handler.go
+++ b/apps/backend/internal/handlers/product_handler.go
@@ -29,6 +29,7 @@ type ProductHandler interface {
 	SubscribeStockUpdates(c *gin.Context) // Story 4.2, Task 4.1-4.6: WebSocket subscription
 	AdjustStock(c *gin.Context)            // Story 4.3, Task 4.1-4.7: POST /api/v1/products/stock/adjust
 	GetLowStockProducts(c *gin.Context)    // Story 4.4, Task 5.1-5.4: GET /api/v1/products/low-stock
+	GetExpiringProducts(c *gin.Context)    // Story 4.5, Task 5.1-5.5: GET /api/v1/products/expiring
 }
 
 // productHandler implements ProductHandler
@@ -661,3 +662,135 @@ func (h *productHandler) GetLowStockProducts(c *gin.Context) {
 	// Build response
 	c.JSON(http.StatusOK, errors.Success(productItems))
 }
+
+// GetExpiringProducts retrieves products expiring within specified days threshold
+// Story 4.5, Task 5.1-5.5: API endpoint for expiring products
+// GET /api/v1/products/expiring?days={30,14,7}&branch_id={id}
+//
+//	@Summary		Get products expiring soon
+//	@Description	Retrieves products expiring within specified days threshold (7, 14, or 30). Supports optional branch_id filtering for Owners, Cashiers see their assigned branch only.
+//	@Tags			products
+//	@Produce		json
+//	@Param			days		query		int		false	"Days threshold (default: 30, max: 365)"
+//	@Param			branch_id	query		int		false	"Filter by branch ID (Owner only, defaults to user's branch)"
+//	@Success		200			{object}	apiErrors.Response{success=bool,data=[]dto.ProductListItem}	"Success response with expiring products list"
+//	@Failure		400			{object}	apiErrors.Response{success=bool,error=errors.ErrorInfo}	"Validation error - invalid input parameters"
+//	@Failure		401			{object}	apiErrors.Response{success=bool,error=errors.ErrorInfo}	"Unauthorized - authentication required"
+//	@Failure		403			{object}	apiErrors.Response{success=bool,error=errors.ErrorInfo}	"Forbidden - insufficient permissions"
+//	@Failure		500			{object}	apiErrors.Response{success=bool,error=errors.ErrorInfo}	"Server error"
+//	@Router			/api/v1/products/expiring [get]
+func (h *productHandler) GetExpiringProducts(c *gin.Context) {
+	// Story 4.5, Task 5.2: Extract user context for RBAC
+	userRole, exists := c.Get("user_role")
+	if !exists {
+		_ = c.Error(errors.Unauthorized("User role not found"))
+		c.Status(http.StatusUnauthorized)
+		return
+	}
+
+	role, ok := userRole.(string)
+	if !ok {
+		_ = c.Error(errors.InternalServerError(fmt.Errorf("invalid user role type")))
+		c.Status(http.StatusInternalServerError)
+		return
+	}
+
+	// Story 4.5, Task 5.4: RBAC - Apply branch access control
+	// Owners can view all branches or filter by branch_id parameter
+	// Cashiers can only view their assigned branch
+	var branchID uint
+	var userBranchID *uint
+	branchIDValue, branchExists := c.Get("branch_id")
+	if branchExists {
+		if bid, ok := branchIDValue.(uint); ok {
+			userBranchID = &bid
+		}
+	}
+
+	// Parse branch_id query parameter
+	branchIDParam := c.Query("branch_id")
+
+	// Story 4.5, Task 5.4: Cashier branch filtering
+	if role == user.RoleCashier {
+		// Cashier must have a branch assignment
+		if userBranchID == nil {
+			_ = c.Error(errors.Forbidden("Cashier must have a branch assignment"))
+			c.Status(http.StatusForbidden)
+			return
+		}
+		// Cashier: use their assigned branch
+		branchID = *userBranchID
+
+		// If cashier provides branch_id param, verify it matches their assignment
+		if branchIDParam != "" {
+			paramBranchID, err := strconv.ParseUint(branchIDParam, 10, 64)
+			if err == nil && uint(paramBranchID) != branchID {
+				_ = c.Error(errors.Forbidden("Cashiers can only view products from their assigned branch"))
+				c.Status(http.StatusForbidden)
+				return
+			}
+		}
+	} else if role == user.RoleOwner {
+		// Owner: use branch_id from query parameter if provided, otherwise 0 (all branches)
+		if branchIDParam != "" {
+			paramBranchID, err := strconv.ParseUint(branchIDParam, 10, 64)
+			if err != nil {
+				_ = c.Error(errors.BadRequest("Invalid branch_id parameter"))
+				c.Status(http.StatusBadRequest)
+				return
+			}
+			branchID = uint(paramBranchID)
+		}
+		// If branchID remains 0, service will return expiring products from all branches
+	} else {
+		_ = c.Error(errors.Forbidden("Invalid user role"))
+		c.Status(http.StatusForbidden)
+		return
+	}
+
+	// Story 4.5, Task 5.2: Parse days parameter (default: 30)
+	daysParam := c.DefaultQuery("days", "30")
+	daysThreshold, err := strconv.Atoi(daysParam)
+	if err != nil || daysThreshold <= 0 || daysThreshold > 365 {
+		_ = c.Error(errors.BadRequest("Invalid days parameter (must be between 1 and 365)"))
+		c.Status(http.StatusBadRequest)
+		return
+	}
+
+	// Story 4.5, Task 5.3: Call service layer
+	products, err := h.productService.GetExpiringProducts(c.Request.Context(), branchID, daysThreshold)
+	if err != nil {
+		_ = c.Error(errors.InternalServerError(err))
+		c.Status(http.StatusInternalServerError)
+		return
+	}
+
+	// Transform to DTO with expiry information
+	productItems := make([]dto.ProductListItem, 0, len(products))
+	now := time.Now()
+	for _, p := range products {
+		isExpired := p.ExpiryDate != nil && p.ExpiryDate.Before(now)
+		isLowStock := p.StockQty < int64(p.ReorderThreshold)
+
+		productItems = append(productItems, dto.ProductListItem{
+			ID:               p.ID,
+			SKU:              p.SKU,
+			Name:             p.Name,
+			Description:      p.Description,
+			StockQty:         p.StockQty,
+			Price:            p.Price,
+			ExpiryDate:       p.ExpiryDate,
+			BranchID:         p.BranchID,
+			Category:         p.Category,
+			ReorderThreshold: p.ReorderThreshold,
+			IsLowStock:       isLowStock,
+			IsExpired:        isExpired,
+			CreatedAt:        p.CreatedAt,
+			UpdatedAt:        p.UpdatedAt,
+		})
+	}
+
+	// Build response
+	c.JSON(http.StatusOK, errors.Success(productItems))
+}

diff --git a/apps/backend/internal/repositories/product_repository.go b/apps/backend/internal/repositories/product_repository.go
index 1ce0f6e..0e46fad 100644
--- a/apps/backend/internal/repositories/product_repository.go
+++ b/apps/backend/internal/repositories/product_repository.go
@@ -39,6 +39,10 @@ type ProductRepository interface {
 
 	// GetExpiredProducts retrieves products that have expired
 	GetExpiredProducts(ctx context.Context, branchID uint) ([]*models.Product, error)
+
+	// GetExpiringProducts retrieves products expiring within the specified date range
+	// Story 4.5, AC1, AC2, AC3: Find products approaching expiry (30, 14, 7 days)
+	GetExpiringProducts(ctx context.Context, branchID uint, startDate, endDate time.Time) ([]*models.Product, error)
 }

diff --git a/apps/backend/internal/repositories/product_repository_impl.go b/apps/backend/internal/repositories/product_repository_impl.go
index 0ab2883..fe06840 100644
--- a/apps/backend/internal/repositories/product_repository_impl.go
+++ b/apps/backend/internal/repositories/product_repository_impl.go
@@ -247,3 +247,18 @@ func (r *productRepository) GetExpiredProducts(ctx context.Context, branchID uin
 	}
 	return products, nil
 }
+
+// GetExpiringProducts retrieves products expiring within the specified date range
+// Story 4.5, AC1, AC2, AC3: Find products approaching expiry (30, 14, 7 days)
+func (r *productRepository) GetExpiringProducts(ctx context.Context, branchID uint, startDate, endDate time.Time) ([]*models.Product, error) {
+	var products []*models.Product
+	err := r.db.WithContext(ctx).
+		Preload("Branch").
+		Where("branch_id = ? AND expiry_date >= ? AND expiry_date <= ?", branchID, startDate, endDate).
+		Order("expiry_date ASC").
+		Find(&products).Error
+	if err != nil {
+		return nil, fmt.Errorf("failed to get expiring products: %w", err)
+	}
+	return products, nil
+}

diff --git a/apps/backend/internal/server/router.go b/apps/backend/internal/server/router.go
index ad45f3d..66cd1d2 100644
--- a/apps/backend/internal/server/router.go
+++ b/apps/backend/internal/server/router.go
@@ -213,6 +213,8 @@ func SetupRouter(userHandler *user.Handler, authHandler handlers.AuthHandler, au
 				productsGroup.GET("", productHandler.ListProducts) // List with search, filters, and pagination
 				// Story 4.4, Task 5.1-5.4: Low stock products endpoint
 				productsGroup.GET("/low-stock", productHandler.GetLowStockProducts)
+				// Story 4.5, Task 5.1-5.5: Expiring products endpoint
+				productsGroup.GET("/expiring", productHandler.GetExpiringProducts)
 				// Story 4.2, Task 4.2: WebSocket endpoint for real-time stock updates
 				productsGroup.GET("/stock/subscribe", productHandler.SubscribeStockUpdates)
 				// Story 4.3, Task 4.2: Stock adjustment endpoint with admin permissions

diff --git a/apps/backend/internal/services/alert_service.go b/apps/backend/internal/services/alert_service.go
index 4af94f7..1343a6e 100644
--- a/apps/backend/internal/services/alert_service.go
+++ b/apps/backend/internal/services/alert_service.go
@@ -26,6 +26,10 @@ type AlertService interface {
 	// Story 4.4, AC2, AC3, AC6: Publish stock.low event with debounce logic
 	PublishLowStockAlert(ctx context.Context, event *dto.LowStockNotificationEvent) error
 
+	// PublishExpiryAlert publishes expiry notification to Redis pub/sub
+	// Story 4.5, AC4, AC6: Publish product.expiry event with debounce logic
+	PublishExpiryAlert(ctx context.Context, event *dto.ExpiryAlertEvent) error
+
 	// ClearLowStockState clears low stock state when stock returns to normal
 	// Story 4.4, AC7: Debounce logic - remove tracking when stock >= threshold
 	ClearLowStockState(ctx context.Context, productID uint, branchID uint) error

diff --git a/apps/backend/internal/services/alert_service_impl.go b/apps/backend/internal/services/alert_service_impl.go
index d11c897..dd3bee1 100644
--- a/apps/backend/internal/services/alert_service_impl.go
+++ b/apps/backend/internal/services/alert_service_impl.go
@@ -266,3 +266,49 @@ func (s *alertService) PublishExpiryAlert(ctx context.Context, event *dto.ExpiryAlertEvent) error {
+// PublishExpiryAlert publishes expiry notification to Redis pub/sub
+// Story 4.5, AC4: Event published to Redis pub/sub with event type "product.expiry"
+// Story 4.5, Task 3.1-3.5: Implement Redis pub/sub with debounce tracking
+func (s *alertService) PublishExpiryAlert(ctx context.Context, event *dto.ExpiryAlertEvent) error {
+	// Check context cancellation
+	if err := ctx.Err(); err != nil {
+		return fmt.Errorf("operation cancelled: %w", err)
+	}
+
+	// Validate event
+	if event.Data.ProductID == 0 {
+		return &InvalidInputError{Field: "product_id", Message: "product ID is required"}
+	}
+	if event.Data.BranchID == 0 {
+		return &InvalidInputError{Field: "branch_id", Message: "branch ID is required"}
+	}
+
+	// Story 4.5, Task 3.3: Publish to Redis pub/sub channel: product.expiry
+	if s.redisClient != nil {
+		// Marshal event to JSON
+		eventJSON, err := json.Marshal(event)
+		if err != nil {
+			return fmt.Errorf("failed to marshal expiry alert event: %w", err)
+		}
+
+		// Publish to product.expiry channel
+		channel := "product.expiry"
+		if err := s.redisClient.Publish(ctx, channel, eventJSON).Err(); err != nil {
+			// Graceful degradation - notifications are best-effort
+			slog.Error("Failed to publish expiry alert", "error", err, "product_id", event.Data.ProductID)
+			return nil // Don't fail the operation
+		}
+
+		// Story 4.5, Task 3.3: Log publication with structured logging
+		slog.Info("Expiry alert published",
+			"event_id", event.EventID,
+			"product_id", event.Data.ProductID,
+			"sku", event.Data.SKU,
+			"days_remaining", event.Data.DaysRemaining,
+			"alert_level", event.Data.AlertLevel,
+			"branch_id", event.Data.BranchID)
+	}
+
+	return nil
+}

diff --git a/apps/backend/internal/services/product_service.go b/apps/backend/internal/services/product_service.go
index d2f1e58..7080fd8 100644
--- a/apps/backend/internal/services/product_service.go
+++ b/apps/backend/internal/services/product_service.go
@@ -66,6 +66,10 @@ type ProductService interface {
 	// GetLowStockProducts retrieves products with stock below reorder threshold
 	GetLowStockProducts(ctx context.Context, branchID uint) ([]*models.Product, error)
 
+	// GetExpiringProducts retrieves products expiring within specified days threshold
+	// Story 4.5, Task 5: Add expiring products API endpoint
+	GetExpiringProducts(ctx context.Context, branchID uint, daysThreshold int) ([]*models.Product, error)
+
 	// CheckLowStock checks if a product is in low stock state
 	// Story 4.4, Task 1.1-1.5: Low stock detection with debounce logic
 	// Returns true if stock < threshold AND not already in low stock state (for notification triggering)

diff --git a/apps/backend/internal/services/product_service_impl.go b/apps/backend/internal/services/product_service_impl.go
index 785f5c5..e512ea4 100644
--- a/apps/backend/internal/services/product_service_impl.go
+++ b/apps/backend/internal/services/product_service_impl.go
@@ -558,6 +558,32 @@ func (s *productService) GetLowStockProducts(ctx context.Context, branchID uint)
 	return products, nil
 }
+
+// GetExpiringProducts retrieves products expiring within specified days threshold
+// Story 4.5, Task 5.1-5.5: API endpoint for expiring products
+func (s *productService) GetExpiringProducts(ctx context.Context, branchID uint, daysThreshold int) ([]*models.Product, error) {
+	// Check context cancellation
+	if err := ctx.Err(); err != nil {
+		return nil, fmt.Errorf("operation cancelled: %w", err)
+	}
+
+	// Validate days threshold (must be positive, typically 7, 14, or 30)
+	if daysThreshold <= 0 || daysThreshold > 365 {
+		return nil, &InvalidInputError{Field: "days", Message: "days threshold must be between 1 and 365"}
+	}
+
+	// Calculate date range
+	now := time.Now().UTC()
+	endDate := now.AddDate(0, 0, daysThreshold)
+
+	// Get expiring products via repository
+	products, err := s.productRepo.GetExpiringProducts(ctx, branchID, now, endDate)
+	if err != nil {
+		return nil, &ServiceError{Op: "get expiring products", Err: err}
+	}
+
+	return products, nil
+}

diff --git a/apps/backend/internal/services/product_service_impl_test.go b/apps/backend/internal/services/product_service_impl_test.go
index ceb60b0..8748928 100644
--- a/apps/backend/internal/services/product_service_impl_test.go
+++ b/apps/backend/internal/services/product_service_impl_test.go
@@ -78,6 +78,14 @@ func (m *MockProductRepository) GetExpiredProducts(ctx context.Context, branchID
 	return args.Get(0).([]*models.Product), args.Error(1)
 }
 
+func (m *MockProductRepository) GetExpiringProducts(ctx context.Context, branchID uint, startDate, endDate time.Time) ([]*models.Product, error) {
+	args := m.Called(ctx, branchID, startDate, endDate)
+	if args.Get(0) == nil {
+		return nil, args.Error(1)
+	}
+	return args.Get(0).([]*models.Product), args.Error(1)
+}

diff --git a/apps/mobile/src/features/inventory/services/realTimeStockService.ts b/apps/mobile/src/features/inventory/services/realTimeStockService.ts
index 8672f4d..0efc5c3 100644
--- a/apps/mobile/src/features/inventory/services/realTimeStockService.ts
+++ b/apps/mobile/src/features/inventory/services/realTimeStockService.ts
@@ -34,6 +34,19 @@ export interface StockUpdatedEvent {
   updatedAt: string;
 }
 
+// Expiry alert event from backend
+// Story 4.5, Task 11.2: Subscribe to product.expiry events via realTimeStockService
+export interface ExpiryEvent {
+  productId: number;
+  sku: string;
+  productName: string;
+  expiryDate: string;
+  daysRemaining: number;
+  alertLevel: 'warning' | 'critical' | 'urgent';
+  branchId: number;
+  branchName: string;
+}
+
 // Configuration options
 interface RealTimeStockServiceConfig {
   // WebSocket server URL
@@ -53,6 +66,7 @@ interface ServiceEvents {
   'connectionStateChange': (state: ConnectionState) => void;
   'stockUpdate': (event: StockUpdatedEvent) => void;
+  'expiry': (event: ExpiryEvent) => void;
   'error': (error: Error) => void;
 }
 
@@ -217,6 +231,7 @@ class RealTimeStockServiceImpl extends EventEmitter implements RealTimeStockServ
   /**
    * Handle WebSocket message event
    * Story 4.2, Task 11.2: WebSocket connection management
+   * Story 4.5, Task 11.2: Subscribe to product.expiry events via realTimeStockService
    */
   private handleMessage(event: MessageEvent): void {
     try {
@@ -233,6 +248,10 @@ class RealTimeStockServiceImpl extends EventEmitter implements RealTimeStockServ
         if (!this.isOnline) {
           this.queueOfflineEvent(stockEvent);
         }
+      } else if (data.event === 'product.expiry' && data.data) {
+        // Story 4.5, Task 11.2: Handle product.expiry events
+        const expiryEvent: ExpiryEvent = data.data;
+        this.emit('expiry', expiryEvent);
       }
     } catch (error) {
       console.error('[RealTimeStockService] Failed to parse message:', error);

diff --git a/apps/web/app/(auth)/layout.tsx b/apps/web/app/(auth)/layout.tsx
index c953f8d..23f54b8 100644
--- a/apps/web/app/(auth)/layout.tsx
+++ b/apps/web/app/(auth)/layout.tsx
@@ -11,6 +11,7 @@ export default function AuthenticatedLayout({
   const [lowStockCount, setLowStockCount] = useState(0);
+  const [urgentExpiryCount, setUrgentExpiryCount] = useState(0);
 
   // Fetch low stock count periodically
   useEffect(() => {
@@ -39,10 +39,37 @@ export default function AuthenticatedLayout({
     return () => clearInterval(interval);
   }, []);
 
+  // Fetch urgent expiry count periodically (7-day urgent items)
+  useEffect(() => {
+    const fetchUrgentExpiryCount = async () => {
+      try {
+        const response = await fetch('/api/v1/products/expiring?days=7', {
+          credentials: 'include',
+        });
+
+        if (response.ok) {
+          const data = await response.json();
+          setUrgentExpiryCount(data.pagination?.total || 0);
+        }
+      } catch (error) {
+        console.error('Failed to fetch urgent expiry count:', error);
+      }
+    };
+
+    // Fetch immediately
+    fetchUrgentExpiryCount();
+
+    // Fetch every 30 seconds
+    const interval = setInterval(fetchUrgentExpiryCount, 30000);
+
+    return () => clearInterval(interval);
+  }, []);
+
   const navItems = [
     { href: '/dashboard', label: 'Dashboard' },
     { href: '/products', label: 'Products' },
     { href: '/inventory/low-stock', label: 'Low Stock', showBadge: true, badgeCount: lowStockCount },
+    { href: '/inventory/expiring', label: 'Expiring', showBadge: true, badgeCount: urgentExpiryCount, isUrgent: urgentExpiryCount > 0 },
     { href: '/reports', label: 'Reports' },
     { href: '/users', label: 'Users' },
     { href: '/settings', label: 'Settings' },
@@ -77,7 +105,9 @@ export default function AuthenticatedLayout({
               >
                 {item.label}
                 {item.showBadge && item.badgeCount > 0 && (
-                  <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
+                  <span className={`ml-2 inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
+                    (item as any).isUrgent ? 'bg-red-600 text-white animate-pulse' : 'bg-red-100 text-red-800'
+                  }`}>
                     {item.badgeCount}
                   </span>
                 )}

diff --git a/apps/web/hooks/useStockWebSocket.ts b/apps/web/hooks/useStockWebSocket.ts
index 0e417c3..87fec42 100644
--- a/apps/web/hooks/useStockWebSocket.ts
+++ b/apps/web/hooks/useStockWebSocket.ts
@@ -42,8 +42,21 @@ export interface LowStockEvent {
   branchName: string;
 }
 
+// Expiry alert event payload from backend
+// Story 4.5, AC4, AC6: Expiry date alert notification event structure
+export interface ExpiryEvent {
+  productId: number;
+  sku: string;
+  productName: string;
+  expiryDate: string;
+  daysRemaining: number;
+  alertLevel: 'warning' | 'critical' | 'urgent';
+  branchId: number;
+  branchName: string;
+}
+
 // Union type for all stock events
-export type StockEvent = StockUpdatedEvent | LowStockEvent;
+export type StockEvent = StockUpdatedEvent | LowStockEvent | ExpiryEvent;
 
 // Configuration options for the hook
 interface UseStockWebSocketOptions {
@@ -61,6 +74,8 @@ interface UseStockWebSocketOptions {
   onStockUpdate?: (event: StockUpdatedEvent) => void;
   // Story 4.4: Event handler for low stock notifications
   onLowStock?: (event: LowStockEvent) => void;
+  // Story 4.5: Event handler for expiry date alerts
+  onExpiry?: (event: ExpiryEvent) => void;
   // Connection state change handler
   onConnectionStateChange?: (state: ConnectionState) => void;
   // Error handler
@@ -101,6 +116,7 @@ export function useStockWebSocket(options: UseStockWebSocketOptions): UseStockWe
     maxReconnectDelay = 30000,
     onStockUpdate,
     onLowStock,
+    onExpiry,
     onConnectionStateChange,
     onError,
   } = options;
@@ -141,6 +157,7 @@ export function useStockWebSocket(options: UseStockWebSocketOptions): UseStockWe
    * Handle WebSocket message
    * Story 4.2, Task 7.4: Add event handlers for stock updates
    * Story 4.4: Extended to handle stock.low events for low stock notifications
+   * Story 4.5: Extended to handle product.expiry events for expiry date alerts
    */
   const handleMessage = useCallback((event: MessageEvent) => {
     try {
@@ -166,6 +183,14 @@ export function useStockWebSocket(options: UseStockWebSocketOptions): UseStockWe
         if (onLowStock) {
           onLowStock(lowStockEvent);
         }
+      } else if (data.event === 'product.expiry' && data.data) {
+        // Story 4.5: Handle expiry date alerts
+        const expiryEvent: ExpiryEvent = data.data;
+
+        // Call external handler if provided
+        if (onExpiry) {
+          onExpiry(expiryEvent);
+        }
       }
     } catch (error) {
       console.error('[useStockWebSocket] Failed to parse message:', error);

@@ -173,7 +198,7 @@ export function useStockWebSocket(options: UseStockWebSocketOptions): UseStockWe
         onError(error as Error);
       }
     }
-  }, [onStockUpdate, onLowStock, onError]);
+  }, [onStockUpdate, onLowStock, onExpiry, onError]);
```

## Your Mission

Hunt for code defects focusing on:

### 1. Code Quality & Maintainability
- Unused variables, dead code, commented-out code
- Inconsistent naming conventions
- Magic numbers without constants
- Duplicate code that could be extracted
- Poor separation of concerns
- Functions/methods that are too long or complex

### 2. Error Handling & Edge Cases
- Missing nil/null checks
- Unhandled error cases
- Missing validation on user input
- Race conditions or concurrency issues
- Resource leaks (unclosed connections, files, etc.)
- Missing context cancellation checks

### 3. Security Issues
- SQL injection vectors
- XSS vulnerabilities
- Missing authentication/authorization checks
- Sensitive data in logs
- Injection attacks

### 4. Performance & Scalability
- N+1 query problems
- Missing indexes or inefficient queries
- Unnecessary allocations in loops
- Missing connection pooling
- Blocking operations in async contexts

### 5. Logic Bugs
- Off-by-one errors
- Incorrect comparisons
- Missing edge case handling
- Incorrect date/time calculations
- State synchronization issues

## Output Format

Provide your findings as a Markdown list. Each finding must have:

```markdown
### [Severity] [Category]: One-line title

**Evidence:** [Specific code snippet or line reference from the diff]

**Risk:** [What could go wrong?]

**Suggested Fix:** [Concise fix recommendation]
```

Severity levels: `Critical`, `High`, `Medium`, `Low`, `Info`

Categories: `Security`, `Performance`, `Logic`, `Error Handling`, `Code Quality`, `Maintainability`

**Important:** Be adversarial. Assume this code will be deployed to production handling sensitive pharmacy data. Think like a hacker, a stressed user, and a maintainer who didn't write this code.
