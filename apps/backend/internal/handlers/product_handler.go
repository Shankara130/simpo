package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/golang-jwt/jwt/v5"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// ProductHandler defines product handler interface
// Story 4.1, Task 1: Handler interface for product operations
// Story 4.2, Task 4: WebSocket handler for real-time stock updates
type ProductHandler interface {
	ListProducts(c *gin.Context)
	SubscribeStockUpdates(c *gin.Context) // Story 4.2, Task 4.1-4.6: WebSocket subscription
}

// productHandler implements ProductHandler
type productHandler struct {
	productService      services.ProductService
	stockEventService    services.StockEventService // Story 4.2, Task 4.1: Stock event service dependency
	jwtSecret           string                      // Story 4.2, Task 4.5: JWT secret for token validation
	upgrader             *websocket.Upgrader
}

// Story 4.2, Task 4.3: Client represents a WebSocket client connection
type wsClient struct {
	id         string
	branches   []uint
	conn       *websocket.Conn
	messageChan chan services.StockUpdatedEvent
}

// Story 4.2, Task 4.3: Active WebSocket clients registry with mutex protection
var (
	wsClients        = make(map[string]*wsClient)
	wsClientsMutex   sync.RWMutex
	wsRegister       = make(chan *wsClient)
	wsUnregister     = make(chan string)
)

// NewProductHandler creates a new product handler
// Story 4.1, Task 1: Constructor with service dependency injection
// Story 4.2, Task 4.1: Add stockEventService dependency for WebSocket support
// Story 4.2, Task 4.5: Add jwtSecret for JWT validation
func NewProductHandler(productService services.ProductService, stockEventService services.StockEventService, jwtSecret string) ProductHandler {
	if productService == nil {
		panic("productService cannot be nil")
	}
	if jwtSecret == "" {
		panic("jwtSecret cannot be empty")
	}

	// Story 4.2, Task 4.1: Create WebSocket upgrader with JWT authentication
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // TODO: Configure CORS properly for production
		},
	}

	return &productHandler{
		productService:   productService,
		stockEventService: stockEventService,
		jwtSecret:       jwtSecret,
		upgrader:          upgrader,
	}
}

// ListProducts handles product listing with search, filters, and pagination
// Story 4.1, AC1, AC2, AC3, AC4, AC7: Product list with search, filters, and pagination
//
//	@Summary		List products
//	@Description	Get products with search, filters, and pagination. Owners can filter by branch/category. Cashiers see only their branch.
//	@Tags			products
//	@Accept			json
//	@Produce		json
//	@Param			search	query		string	false	"Search by name or SKU"
//	@Param			category	query		string	false	"Filter by category"
//	@Param			branch_id	query		int		false	"Filter by branch (Owner only)"
//	@Param			low_stock	query		bool	false	"Filter for low stock items"
//	@Param			expired	query		bool	false	"Filter for expired items"
//	@Param			page		query		int		false	"Page number (default 1)"
//	@Param			limit		query		int		false	"Items per page (default 20, max 1000)"
//	@Param			sort_by		query		string	false	"Field to sort by"	Enums(id, name, sku, price, stock_qty, category, created_at)
//	@Param			sort_order	query		string	false	"Sort order"	Enums(asc, desc)
//	@Success		200			{object}	apiErrors.Response{success=bool,data=dto.ProductListResponse}	"Success response with product list"
//	@Failure		400			{object}	apiErrors.Response{success=bool,error=errors.ErrorInfo}	"Validation error - invalid input parameters"
//	@Failure		401			{object}	apiErrors.Response{success=bool,error=errors.ErrorInfo}	"Unauthorized - authentication required"
//	@Failure		403			{object}	apiErrors.Response{success=bool,error=errors.ErrorInfo}	"Forbidden - insufficient permissions"
//	@Failure		500			{object}	apiErrors.Response{success=bool,error=errors.ErrorInfo}	"Server error"
//	@Router			/api/v1/products [get]
func (h *productHandler) ListProducts(c *gin.Context) {
	// Story 4.1, Task 1.3: Bind and validate query parameters
	var req dto.ProductListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		_ = c.Error(errors.FromGinValidation(err))
		c.Status(http.StatusBadRequest)
		return
	}

	// Apply defaults for pagination
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20 // Default page size
	}
	if req.Limit > 1000 {
		req.Limit = 1000 // Maximum to prevent DoS
	}

	// Story 4.1, Task 5.1, 5.2, 5.3: Extract user context for RBAC
	userRole, exists := c.Get("user_role")
	if !exists {
		_ = c.Error(errors.Unauthorized("User role not found"))
		c.Status(http.StatusUnauthorized)
		return
	}

	role, ok := userRole.(string)
	if !ok {
		_ = c.Error(errors.InternalServerError(fmt.Errorf("invalid user role type")))
		c.Status(http.StatusInternalServerError)
		return
	}

	// Story 4.1, AC3: Apply branch access control
	// Owners can filter by branch, Cashiers restricted to their assigned branch
	var userBranchID *uint
	branchIDValue, branchExists := c.Get("branch_id")
	if branchExists {
		if bid, ok := branchIDValue.(uint); ok {
			userBranchID = &bid
		}
	}

	// Story 4.1, Task 5.2, 5.3: RBAC - Branch access control
	// Owners can view all branches or filter by branch_id parameter
	// Cashiers can only view their assigned branch
	if role == user.RoleCashier {
		// Story 4.1 Code Review (2026-05-18): Ensure cashier has branch assignment
		if userBranchID == nil {
			_ = c.Error(errors.Forbidden("Cashier must have a branch assignment"))
			c.Status(http.StatusForbidden)
			return
		}
		// Cashier: Override branch_id filter with their assigned branch
		if req.BranchID != nil && *req.BranchID != *userBranchID {
			_ = c.Error(errors.Forbidden("Cashiers can only view products from their assigned branch"))
			c.Status(http.StatusForbidden)
			return
		}
		req.BranchID = userBranchID
	}
	// For Owner: use the branch_id from query parameter if provided, otherwise nil (all branches)

	// Story 4.1 Code Review (2026-05-18): Validate SortBy field to prevent SQL injection
	allowedSortFields := map[string]bool{
		"id": true, "name": true, "sku": true, "price": true,
		"stock_qty": true, "category": true, "created_at": true,
	}
	if req.SortBy != "" && !allowedSortFields[req.SortBy] {
		req.SortBy = "created_at" // Default fallback
	}

	// Build service filter from request
	filter := &services.ProductFilter{
		BranchID:     req.BranchID,
		Category:     req.Category,
		SearchQuery:  req.Search,
		LowStock:     req.LowStock != nil && *req.LowStock,
		Expired:      req.Expired != nil && *req.Expired,
		Page:         req.Page,
		Limit:        req.Limit,
		SortBy:       req.SortBy,
		SortOrder:    req.SortOrder,
	}

	// Call service layer
	products, total, err := h.productService.ListProducts(c.Request.Context(), filter)
	if err != nil {
		_ = c.Error(errors.InternalServerError(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	// Story 4.1, Task 4: Transform to DTO with indicators
	// Story 4.1, AC5: Calculate low stock indicator
	// Story 4.1, AC6: Calculate expired indicator
	productItems := make([]dto.ProductListItem, 0, len(products))
	now := time.Now()
	for _, p := range products {
		isLowStock := p.StockQty < int64(p.ReorderThreshold)
		isExpired := p.ExpiryDate != nil && p.ExpiryDate.Before(now)

		productItems = append(productItems, dto.ProductListItem{
			ID:               p.ID,
			SKU:              p.SKU,
			Name:             p.Name,
			Description:      p.Description,
			StockQty:         p.StockQty,
			Price:            p.Price,
			ExpiryDate:       p.ExpiryDate,
			BranchID:         p.BranchID,
			Category:         p.Category,
			ReorderThreshold: p.ReorderThreshold,
			IsLowStock:       isLowStock,
			IsExpired:        isExpired,
			CreatedAt:        p.CreatedAt,
			UpdatedAt:        p.UpdatedAt,
		})
	}

	// Calculate total pages
	totalPages := int(total) / req.Limit
	if int(total)%req.Limit > 0 {
		totalPages++
	}

	// Build response
	response := dto.ProductListResponse{
		Data: productItems,
		Pagination: dto.PaginationMetadata{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	c.JSON(http.StatusOK, errors.Success(response))
}

// SubscribeStockUpdates handles WebSocket connections for real-time stock updates
// Story 4.2, Task 4.1-4.6: WebSocket endpoint with JWT auth and branch filtering
//
//	@Summary		Subscribe to real-time stock updates
//	@Description	WebSocket endpoint for receiving real-time stock updates. JWT token required via query parameter.
//	@Tags			products
//	@Param			token	query		string	true	"JWT authentication token"
//	@Param			branches	query		string	false	"Comma-separated branch IDs to filter (Owner only, defaults to user's branch)"
//	@Router			/api/v1/products/stock/subscribe [get]
func (h *productHandler) SubscribeStockUpdates(c *gin.Context) {
	// Story 4.2, Task 4.5: JWT authentication validation
	token := c.Query("token")
	if token == "" {
		_ = c.Error(errors.Unauthorized("JWT token required"))
		c.Status(http.StatusUnauthorized)
		return
	}

	// Validate JWT token and extract user info
	// Story 4.2, Task 4.5: Proper JWT validation for WebSocket
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.jwtSecret), nil
	})

	if err != nil || !parsedToken.Valid {
		_ = c.Error(errors.Unauthorized(fmt.Sprintf("Invalid JWT token: %v", err)))
		c.Status(http.StatusUnauthorized)
		return
	}

	// Extract claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		_ = c.Error(errors.Unauthorized("Invalid token claims"))
		c.Status(http.StatusUnauthorized)
		return
	}

	// Extract user role and branch ID from validated claims
	userRole, ok := claims["role"].(string)
	if !ok {
		_ = c.Error(errors.Unauthorized("Invalid token: missing role"))
		c.Status(http.StatusUnauthorized)
		return
	}

	// Extract branch_id (can be nil for system users)
	var userBranchID *uint
	if branchIDFloat, ok := claims["branch_id"].(float64); ok {
		bid := uint(branchIDFloat)
		userBranchID = &bid
	}

	// Verify user role is valid
	if userRole != user.RoleOwner && userRole != user.RoleCashier {
		_ = c.Error(errors.Unauthorized("Invalid token: invalid user role"))
		c.Status(http.StatusUnauthorized)
		return
	}

	// Story 4.2, Task 4.4: Branch-based subscription filtering
	// Parse branches parameter
	var branches []uint
	branchesParam := c.Query("branches")
	if branchesParam != "" {
		// Only Owners can specify multiple branches
		if userRole == user.RoleOwner {
			// Fix: Split by comma to get individual branch IDs
			branchStrings := strings.Split(branchesParam, ",")
			for _, bs := range branchStrings {
				bs = strings.TrimSpace(bs)
				if bs == "" {
					continue
				}
				bid, err := strconv.ParseUint(bs, 10, 64)
				if err == nil {
					branches = append(branches, uint(bid))
				}
			}
		}
	}

	// Cashiers can only subscribe to their assigned branch
	if len(branches) == 0 {
		if userBranchID != nil {
			branches = []uint{*userBranchID}
		} else {
			_ = c.Error(errors.Unauthorized("Cashier must have a branch assignment"))
			c.Status(http.StatusUnauthorized)
			return
		}
	}

	// Story 4.2, Task 4.1-4.2: Upgrade to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("WebSocket upgrade failed", "error", err)
		return
	}

	// Story 4.2, Task 4.3: Generate client ID and register
	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())
	
	// Create message channel for this client (increased to 1000 for backpressure handling)
	messageChan := make(chan services.StockUpdatedEvent, 1000)
	
	client := &wsClient{
		id:         clientID,
		branches:   branches,
		conn:       conn,
		messageChan: messageChan,
	}

	// Story 4.2, Task 4.3: Register client with stock event service
	if h.stockEventService != nil {
		h.stockEventService.RegisterClient(clientID, branches, messageChan)
	}
	wsClientsMutex.Lock()
		wsClients[clientID] = client
		wsClientsMutex.Unlock()

	// Start goroutine to send messages to this client
	go h.handleClientMessages(client)

	// Story 4.2, Task 4.6: Connection cleanup on disconnect
	// Wait for client to disconnect
	defer func() {
		if h.stockEventService != nil {
			h.stockEventService.UnregisterClient(clientID)
		}
		delete(wsClients, clientID)
		close(messageChan)
		conn.Close()
	}()

	// Keep connection alive and handle incoming messages
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		
		// Handle incoming messages if needed (e.g., ping/pong, subscription changes)
		_ = messageType
		_ = message
	}
}

// handleClientMessages sends stock update events to WebSocket client
// Story 4.2, Task 4.3: Broadcast events to connected WebSocket clients
func (h *productHandler) handleClientMessages(client *wsClient) {
	for {
		select {
		case event, ok := <-client.messageChan:
			if !ok {
				return
			}
			
			// Wrap event in the expected format for frontend
			// Frontend expects: {event: "stock.updated", data: {...}}
			wrappedEvent := map[string]interface{}{
				"event": "stock.updated",
				"data":  event,
			}

			// Send event to client
			data, err := json.Marshal(wrappedEvent)
			if err != nil {
				slog.Error("Failed to marshal stock event", "error", err, "client", client.id)
				continue
			}

			err = client.conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				slog.Error("Failed to send message to client", "error", err, "client", client.id)
				return
			}
		}
	}
}
