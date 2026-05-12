# Story 9.6: Implement Core Business Services

**Status:** ready-for-dev

**Epic:** 9 - API Foundation & Core Services
**Priority:** CRITICAL (Blocks Epic 3 Point of Sale)
**Story Type:** Infrastructure Implementation
**Story ID:** 9.6
**Story Key:** 9-6-implement-core-business-services

---

## Story

**As a** Development Team,
**I want** to implement the foundational business logic services that power the application,
**So that** handlers can orchestrate business operations without knowing data access details, enabling Clean Architecture separation.

---

## Acceptance Criteria

1. **AC1: Service Interfaces for All Core Domains**
   - Create service interfaces for: UserService, ProductService, TransactionService, ReportService, AlertService, SyncService
   - Each interface defines business methods with clear signatures
   - Interfaces follow Go best practices with dependency injection
   - Interfaces located in `internal/services/` package

2. **AC2: Service Implementations Use Repository Layer**
   - Services use repository interfaces (not concrete implementations) for all data access
   - Services never call GORM directly - always delegate to repositories
   - Services accept repository interfaces via constructor injection
   - Enables testability with mock repositories

3. **AC3: Business Logic Encapsulation**
   - **UserService**: CreateUser, UpdateUser, DeactivateUser, ListUsers, GetUserByID
   - **ProductService**: CreateProduct, UpdateProduct, UpdateStock, CheckAvailability, ListProducts, GetProductByID, GetLowStockProducts
   - **TransactionService**: CreateTransaction, ProcessSale, CalculateTotal, GenerateReceiptData, GetTransactionByID, ListTransactions
   - **ReportService**: GenerateDailySales, GenerateProfitLoss, ExportReport (stub for future)
   - **AlertService**: CheckLowStockAlerts, CheckExpiryAlerts, SendNotification (stub for future)
   - **SyncService**: QueueTransactionSync, ProcessSyncQueue, ResolveConflict (stub for future)

4. **AC4: Error Handling and Domain Errors**
   - Services return domain errors (not repository errors)
   - Business validation errors are descriptive (e.g., "insufficient stock", "product expired")
   - Errors wrap underlying repository errors with context
   - Use structured errors that handlers can convert to RFC 7807

5. **AC5: Transaction Support for Multi-Step Operations**
   - TransactionService.ProcessSale uses transactional operations
   - Stock updates and transaction creation are atomic
   - Rollback on any failure during sale processing
   - All-or-nothing semantics for business operations

6. **AC6: Context Support for Cancellation**
   - All service methods accept context.Context as first parameter
   - Services check context cancellation before expensive operations
   - Long-running queries respect context timeout
   - Graceful handling of context.Done()

---

## Tasks / Subtasks

- [ ] **Task 1: Create Service Interfaces (AC: 1)**
  - [ ] Create `internal/services/user_service.go` with UserService interface
  - [ ] Create `internal/services/product_service.go` with ProductService interface
  - [ ] Create `internal/services/transaction_service.go` with TransactionService interface
  - [ ] Create `internal/services/report_service.go` with ReportService interface
  - [ ] Create `internal/services/alert_service.go` with AlertService interface
  - [ ] Create `internal/services/sync_service.go` with SyncService interface

- [ ] **Task 2: Implement ProductService (AC: 2, 3, 4, 6)**
  - [ ] Create productService struct with ProductRepository and AuditService dependencies
  - [ ] Implement NewProductService factory function with dependency injection
  - [ ] Implement CreateProduct with validation (SKU uniqueness, required fields)
  - [ ] Implement UpdateProduct with business rules (cannot update SKU, preserve created_at)
  - [ ] Implement UpdateStock with atomic increment (learn from Epic 2 retro)
  - [ ] Implement CheckAvailability returning available quantity
  - [ ] Implement ListProducts with filtering and pagination (delegate to repository)
  - [ ] Implement GetProductByID with Preload("Branch")
  - [ ] Implement GetLowStockProducts (stock_qty < reorder_threshold)
  - [ ] Add context cancellation checks before expensive queries
  - [ ] Create unit tests with mock repositories

- [ ] **Task 3: Implement TransactionService (AC: 2, 3, 4, 5, 6)**
  - [ ] Create transactionService struct with repositories (Transaction, TransactionItem, Product)
  - [ ] Implement NewTransactionService factory function
  - [ ] Implement CalculateTotal with item quantity * price summation
  - [ ] Implement CreateTransaction with transaction_number generation
  - [ ] Implement ProcessSale with transactional operations:
    - [ ] Validate all products exist and have sufficient stock
    - [ ] Begin database transaction
    - [ ] Deduct stock for all items (atomic increments)
    - [ ] Create transaction record
    - [ ] Create transaction items
    - [ ] Commit transaction
    - [ ] Rollback on any error
  - [ ] Implement GenerateReceiptData returning receipt structure
  - [ ] Implement GetTransactionByID with Preload("Items") and Preload("Cashier")
  - [ ] Implement ListTransactions with filtering (date range, cashier, payment method)
  - [ ] Add context cancellation checks
  - [ ] Create unit tests with mock repositories

- [ ] **Task 4: Implement UserService (AC: 2, 3, 4, 6)**
  - [ ] Create userService struct with UserRepository and AuditService
  - [ ] Implement NewUserService factory function
  - [ ] Implement CreateUser with validation (username uniqueness, required fields)
  - [ ] Implement UpdateUser with business rules (role changes require admin, cannot change own role)
  - [ ] Implement DeactivateUser with audit logging
  - [ ] Implement ListUsers with filtering (role, branch, status)
  - [ ] Implement GetUserByID
  - [ ] Add context cancellation checks
  - [ ] Create unit tests with mock repositories

- [ ] **Task 5: Implement ReportService (AC: 2, 3, 4, 6)**
  - [ ] Create reportService struct with TransactionRepository and ProductRepository
  - [ ] Implement NewReportService factory function
  - [ ] Implement GenerateDailySales with date range and branch filtering
  - [ ] Implement GenerateProfitLoss (revenue - COGS calculation)
  - [ ] Implement ExportReport as stub (returns "not implemented" - for future story)
  - [ ] Add context cancellation checks
  - [ ] Create unit tests with mock repositories

- [ ] **Task 6: Implement AlertService (AC: 2, 3, 4, 6)**
  - [ ] Create alertService struct with ProductRepository
  - [ ] Implement NewAlertService factory function
  - [ ] Implement CheckLowStockAlerts (stock_qty <= reorder_threshold)
  - [ ] Implement CheckExpiryAlerts (expiry within 30/14/7 days)
  - [ ] Implement SendNotification as stub (for future Redis pub/sub story)
  - [ ] Add context cancellation checks
  - [ ] Create unit tests with mock repositories

- [ ] **Task 7: Implement SyncService (AC: 2, 3, 4, 6)**
  - [ ] Create syncService struct with TransactionRepository
  - [ ] Implement NewSyncService factory function
  - [ ] Implement QueueTransactionSync as stub (for future offline sync story)
  - [ ] Implement ProcessSyncQueue as stub (for future offline sync story)
  - [ ] Implement ResolveConflict as stub (for future offline sync story)
  - [ ] Add context cancellation checks
  - [ ] Create unit tests

- [ ] **Task 8: Create Service Error Types (AC: 4)**
  - [ ] Create `internal/services/errors.go` with domain error types
  - [ ] Define InsufficientStockError, ProductExpiredError, InvalidInputError, etc.
  - [ ] Ensure errors implement error interface with descriptive messages
  - [ ] Add error types that handlers can convert to RFC 7807 format

---

## Dev Notes

### Architecture Context

**Clean Architecture Pattern (Handler → Service → Repository):**

```
┌─────────────────────────────────────────────────────────────┐
│                        HTTP Layer                            │
│  handlers/ (Gin handlers) - Parse request, call service     │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                      Service Layer                           │
│  services/ (Business Logic) - Orchestrate, validate, calc   │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                   Repository Layer                           │
│  repositories/ (Data Access) - GORM queries, pagination     │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                    PostgreSQL Database                       │
└─────────────────────────────────────────────────────────────┘
```

**Key Principle:** Services MUST use repository interfaces, never call GORM directly. This enables testing with mocks and separates business logic from data access.

### Project Structure

**Service Layer Directory:**
```
apps/backend/internal/services/
├── user_service.go           # UserService interface + impl
├── product_service.go        # ProductService interface + impl
├── transaction_service.go    # TransactionService interface + impl
├── report_service.go         # ReportService interface + impl
├── alert_service.go          # AlertService interface + impl
├── sync_service.go           # SyncService interface + impl
├── errors.go                 # Domain error types
├── auth_service.go           # Already exists from GRAB
└── audit_service.go          # Already exists from GRAB
```

**Repository Layer (Already Complete - Epic 2):**
```
apps/backend/internal/repositories/
├── product_repository.go         # ProductRepository interface
├── product_repository_impl.go    # Implementation with security patches
├── transaction_repository.go     # TransactionRepository interface
├── transaction_repository_impl.go
├── user_repository.go            # UserRepository interface
├── user_repository_impl.go
├── branch_repository.go          # BranchRepository interface
├── branch_repository_impl.go
├── transaction_item_repository.go
└── errors.go                     # Common repository errors
```

### Existing Patterns to Follow

**AuthService Pattern (from GRAB boilerplate):**
```go
// Interface for handlers
type AuthInterface interface {
    Login(ctx context.Context, username, password, ipAddress string) (*dto.LoginResponse, error)
}

// Service struct with dependencies
type AuthService struct {
    jwtSecret       string
    accessTokenTTL  time.Duration
    userRepo        UserFinder  // Repository interface
    auditService    AuditService
}

// Factory function with dependency injection
func NewAuthService(cfg *config.JWTConfig, userRepo UserFinder, auditService AuditService) *AuthService {
    // Panic on nil dependencies (fail fast)
    if cfg == nil {
        panic("authService: config cannot be nil")
    }
    // ... validation ...
    return &AuthService{...}
}
```

**Repository Interface Pattern (from Epic 2 Story 2-5):**
```go
// ProductRepository interface (from product_repository.go)
type ProductRepository interface {
    Create(ctx context.Context, product *models.Product) error
    GetByID(ctx context.Context, id uint) (*models.Product, error)
    GetBySKU(ctx context.Context, branchID uint, sku string) (*models.Product, error)
    Update(ctx context.Context, product *models.Product) error
    UpdateStock(ctx context.Context, id uint, quantity int64) error
    Delete(ctx context.Context, id uint) error
    List(ctx context.Context, filter *ProductFilter) ([]*models.Product, int64, error)
    GetLowStockProducts(ctx context.Context, branchID uint) ([]*models.Product, error)
    GetExpiredProducts(ctx context.Context, branchID uint) ([]*models.Product, error)
}
```

### Epic 2 Retrospective Key Learnings

**Security Learnings (14 patches applied to Story 2-5):**

1. **SQL Injection Prevention:** Whitelist validation for user input in ORDER BY clauses
2. **Race Conditions:** Use atomic operations for concurrent stock updates (UpdateStock)
3. **Unbounded Queries:** Always apply pagination limits (default 20, max 1000)
4. **Integer Overflow:** Bounds checking on page/limit values
5. **Context Cancellation:** Check context.Done() before expensive operations
6. **Nil Pointer Safety:** Always check for nil filters with default empty filter
7. **Special Character Sanitization:** Remove wildcard characters (% and _) from search input
8. **Time Zone Consistency:** Use UTC for all date boundaries
9. **Empty Validation:** Validate that slices are not empty before processing
10. **Zero ID Validation:** Reject id == 0 in GetByID methods

**Apply These Patterns in Services:**
- Always check context cancellation before repository calls
- Validate input parameters before delegating to repositories
- Use atomic operations for stock updates (don't read-then-write)
- Return descriptive domain errors (not raw repository errors)

### Testing Standards

**Unit Test Pattern (from Epic 2):**
```go
func TestProductService_CreateProduct(t *testing.T) {
    // Arrange: Create mock repository
    mockRepo := &MockProductRepository{
        products: make(map[uint]*models.Product),
    }
    service := NewProductService(mockRepo, &mockAuditService)

    // Act: Call service method
    err := service.CreateProduct(context.Background(), &models.Product{
        SKU:    "TEST-001",
        Name:   "Test Product",
        StockQty: 100,
        Price:  50000,
    })

    // Assert: Verify behavior
    assert.NoError(t, err)
    assert.Equal(t, 1, mockRepo.createCallCount)
}
```

**Test Coverage:** Target >80% for all service methods. Focus on:
- Business logic validation (stock checks, expiry validation)
- Error handling (insufficient stock, product not found)
- Context cancellation (return early when context done)
- Edge cases (empty inputs, zero IDs, nil pointers)

### Business Rules Summary

**ProductService Business Rules:**
- SKU must be unique within a branch
- Stock cannot go negative (UpdateStock validates this)
- CheckAvailability returns min(stock_qty, requested_qty)
- Low stock threshold: stock_qty <= reorder_threshold
- Expired products: expiry_date < NOW

**TransactionService Business Rules:**
- Transaction must have at least 1 item (validated by repository)
- All products must exist and have sufficient stock
- Sale process is atomic (all-or-nothing)
- Transaction number format: TRX-{YYYYMMDD}-{sequential}
- Total calculation: SUM(quantity * unit_price)

**UserService Business Rules:**
- Username must be unique globally
- Email must be unique globally
- Role changes require Admin role
- Cannot deactivate yourself
- Branch ID required for Cashier and Owner roles

**ReportService Business Rules:**
- Daily sales: 00:00:00 to 23:59:59 in local timezone
- Profit/Loss: Revenue - Cost of Goods Sold
- Reports are branch-scoped (except Admin can see all)

**AlertService Business Rules:**
- Low stock alert: stock_qty <= reorder_threshold
- Expiry alerts: 30, 14, 7 days before expiry_date
- Alerts are branch-specific

### Dependencies Between Services

**TransactionService Dependencies:**
- ProductRepository (for stock validation and updates)
- TransactionRepository (for creating transactions)
- TransactionItemRepository (for creating line items)
- AuditService (for logging sales)

**ProductService Dependencies:**
- ProductRepository (all product operations)
- AuditService (for logging stock adjustments)

**UserService Dependencies:**
- UserRepository (all user operations)
- AuditService (for logging user management)

**ReportService Dependencies:**
- TransactionRepository (sales data)
- ProductRepository (inventory value)
- TransactionItemRepository (line item details)

**AlertService Dependencies:**
- ProductRepository (stock and expiry checks)

### References

- [Source: epics.md#Epic-9] - Epic 9 requirements
- [Source: epics.md#Story-9.6] - Story 9.6 acceptance criteria
- [Source: architecture.md#Decision-7] - Clean Architecture with Handler → Service → Repository
- [Source: architecture.md#Implementation-Patterns] - Service layer patterns
- [Source: 2-5-implement-repository-layer-for-data-access.md] - Repository layer implementation
- [Source: epic-2-retro-summary.md] - Epic 2 retrospective with security learnings
- [Source: apps/backend/internal/services/auth_service.go] - Existing AuthService pattern
- [Source: apps/backend/internal/services/audit_service.go] - Existing AuditService pattern
- [Source: apps/backend/internal/repositories/product_repository.go] - ProductRepository interface
- [Source: apps/backend/internal/repositories/transaction_repository.go] - TransactionRepository interface

### Integration Points

**Before Epic 3 (Point of Sale) can start:**
- ✅ Repository layer complete (Epic 2)
- ⏳ **Service layer (Story 9-6)** ← CURRENT STORY
- ⏳ API handlers (will be implemented in Epic 3)

**After This Story:**
- Handlers can be implemented for POS endpoints
- Mobile app can consume business logic via API
- Epic 3 (Point of Sale) can proceed with implementation

---

## Dev Agent Record

### Agent Model Used

claude-opus-4-6

### Debug Log References

### Completion Notes List

### File List
