# Story 10.5: Implement Supplier Product Catalog

Status: done
Completed: 2025-05-31

## Story

As a System Administrator,
I want to maintain supplier product catalogs with purchase prices,
so that cost calculations are accurate and purchase orders can be created efficiently.

## Acceptance Criteria

1. **Given** suppliers are registered in the system
   **When** managing supplier product catalogs
   **Then** the admin can associate products with suppliers and specify purchase prices
   **And** the system maintains current purchase price for each product-supplier combination
   **And** the system uses supplier purchase prices when recording purchase invoices
   **And** the system can display price history to track cost changes over time
   **And** the admin can mark preferred suppliers for each product

## Tasks / Subtasks

- [ ] **Task 1: Create Supplier Product Catalog Database Migration** (AC: 1)
  - [ ] Subtask 1.1: Create migration file `20260531310001_create_supplier_product_catalogs_table.up.sql` with columns:
    - `id` (primary key, auto-increment)
    - `supplier_id` (foreign key to suppliers.id, NOT NULL)
    - `product_id` (foreign key to products.id, NOT NULL)
    - `purchase_price` (decimal(15,2), NOT NULL)
    - `is_preferred` (boolean, NOT NULL, default: false)
    - `sku_code` (varchar(50), nullable) -- supplier's SKU code for this product
    - `minimum_order_quantity` (integer, NOT NULL, default: 1)
    - `lead_time_days` (integer, nullable) -- average lead time in days
    - `branch_id` (foreign key to branches.id, NOT NULL)
    - `created_by` (foreign key to users.id, NOT NULL)
    - `updated_by` (foreign key to users.id, nullable)
    - `created_at` (timestamp, NOT NULL)
    - `updated_at` (timestamp, NOT NULL)
    - `price_effective_from` (date, NOT NULL) -- when this price becomes effective
    - `price_effective_to` (date, nullable) -- when this price ends (null means current)
  - [ ] Subtask 1.2: Create corresponding down migration file
  - [ ] Subtask 1.3: Add unique composite index on `(supplier_id, product_id, price_effective_from)` where `price_effective_to IS NULL`
  - [ ] Subtask 1.4: Add index on `product_id` for product lookup
  - [ ] Subtask 1.5: Add index on `branch_id` for filtering

- [ ] **Task 2: Create Supplier Product Catalog Price History Migration** (AC: 1)
  - [ ] Subtask 2.1: The same table `supplier_product_catalogs` will store price history via `price_effective_from` and `price_effective_to` dates
  - [ ] Subtask 2.2: Current price entries have `price_effective_to IS NULL`
  - [ ] Subtask 2.3: Historical price entries have `price_effective_to` set to the date when price was updated
  - [ ] Subtask 2.4: Add index on `price_effective_from` and `price_effective_to` for date range queries

- [ ] **Task 3: Create Supplier Product Catalog GORM Model** (AC: 1)
  - [ ] Subtask 3.1: Create `apps/backend/internal/models/supplier_product_catalog.go` with SupplierProductCatalog struct
  - [ ] Subtask 3.2: Add GORM tags following project conventions (snake_case DB, camelCase JSON)
  - [ ] Subtask 3.3: Add TableName() method returning "supplier_product_catalogs"
  - [ ] Subtask 3.4: Add validation tags (required for supplier_id, product_id, purchase_price, branch_id)
  - [ ] Subtask 3.5: Add Swagger documentation annotations
  - [ ] Subtask 3.6: Add relationship methods: Supplier() *Supplier, Product() *Product
  - [ ] Subtask 3.7: Add method `IsCurrentPrice()` to check if this is the current active price
  - [ ] Subtask 3.8: Export model in `models.go` package exports

- [ ] **Task 4: Implement Supplier Product Catalog Repository** (AC: 1)
  - [ ] Subtask 4.1: Create `apps/backend/internal/repositories/supplier_product_catalog_repository.go` interface with methods:
    - `Create(catalog *models.SupplierProductCatalog) error`
    - `GetByID(id uint) (*models.SupplierProductCatalog, error)`
    - `GetBySupplierAndProduct(supplierID, productID uint) ([]models.SupplierProductCatalog, error)`
    - `GetCurrentPrice(supplierID, productID uint) (*models.SupplierProductCatalog, error)`
    - `List(filters SupplierProductCatalogFilter) ([]models.SupplierProductCatalog, error)`
    - `GetPriceHistory(productID uint, startDate, endDate time.Time) ([]models.SupplierProductCatalog, error)`
    - `UpdatePrice(catalogID uint, newPrice float64, updatedBy uint) error`
    - `SetPreferredSupplier(supplierID, productID uint, isPreferred bool) error`
    - `GetPreferredSupplier(productID uint) (*models.SupplierProductCatalog, error)`
  - [ ] Subtask 4.2: Create `apps/backend/internal/repositories/supplier_product_catalog_repository_impl.go` implementation
  - [ ] Subtask 4.3: Use GORM for database operations with error handling
  - [ ] Subtask 4.4: Add eager loading for Supplier and Product relationships
  - [ ] Subtask 4.5: Implement price history logic: when updating price, set `price_effective_to` on old entry and create new entry
  - [ ] Subtask 4.6: Add repository to `Repository` container in `repository.go`
  - [ ] Subtask 4.7: Create unit tests following existing test patterns

- [ ] **Task 5: Create Supplier Product Catalog Service** (AC: 1)
  - [ ] Subtask 5.1: Create `apps/backend/internal/services/supplier_product_catalog_service.go` interface with business logic methods
  - [ ] Subtask 5.2: Create `apps/backend/internal/services/supplier_product_catalog_service_impl.go` implementation
  - [ ] Subtask 5.3: Add `AssociateProduct(catalog *AssociateProductRequest, createdBy uint) (*SupplierProductCatalog, error)` method
  - [ ] Subtask 5.4: Implement price history tracking: when updating price, archive old price and create new entry
  - [ ] Subtask 5.5: Add `UpdatePurchasePrice(catalogID uint, newPrice float64, updatedBy uint) error` method with transaction wrapping
  - [ ] Subtask 5.6: Add `SetPreferredSupplier(supplierID, productID uint, isPreferred bool) error` method
  - [ ] Subtask 5.7: Add `GetPreferredSupplier(productID uint) (*SupplierProductCatalog, error)` method
  - [ ] Subtask 5.8: Add `GetPriceHistory(productID uint, filters PriceHistoryFilter) ([]*PriceHistoryEntry, error)` method
  - [ ] Subtask 5.9: Add `GetCatalogBySupplier(supplierID uint, filters SupplierProductCatalogFilter) ([]*SupplierProductCatalogResponse, error)` method
  - [ ] Subtask 5.10: Integrate with AuditService for logging all catalog operations
  - [ ] Subtask 5.11: Add service to service container
  - [ ] Subtask 5.12: Create unit tests following existing test patterns

- [ ] **Task 6: Create Supplier Product Catalog DTOs** (AC: 1)
  - [ ] Subtask 6.1: Create `apps/backend/internal/dto/supplier_product_catalog_dto.go` with:
    - `AssociateProductRequest` struct with validation tags (supplierID, productID, purchasePrice, isPreferred, skuCode, minimumOrderQuantity, leadTimeDays)
    - `UpdatePriceRequest` struct (catalogID, newPrice)
    - `SupplierProductCatalogResponse` struct for API responses
    - `SupplierProductCatalogListResponse` struct with pagination support
    - `SupplierProductCatalogFilter` struct for filtering
    - `PriceHistoryEntry` struct for price history display
    - `PriceHistoryFilter` struct for date range filtering
  - [ ] Subtask 6.2: Add Swagger annotations for all fields

- [ ] **Task 7: Create Supplier Product Catalog Handler** (AC: 1)
  - [ ] Subtask 7.1: Create `apps/backend/internal/handlers/supplier_product_catalog_handler.go`
  - [ ] Subtask 7.2: Implement handler methods:
    - `AssociateProduct` - POST /api/v1/supplier-product-catalogs
    - `GetProductCatalog` - GET /api/v1/supplier-product-catalogs/:id
    - `ListProductCatalogs` - GET /api/v1/supplier-product-catalogs with pagination
    - `UpdatePurchasePrice` - PUT /api/v1/supplier-product-catalogs/:id/price
    - `SetPreferredSupplier` - PUT /api/v1/supplier-product-catalogs/:id/preferred
    - `GetPriceHistory` - GET /api/v1/products/:id/price-history
    - `GetPreferredSupplier` - GET /api/v1/products/:id/preferred-supplier
    - `GetSupplierCatalog` - GET /api/v1/suppliers/:id/product-catalog
  - [ ] Subtask 7.3: Add RBAC middleware (Admin, Owner roles)
  - [ ] Subtask 7.4: Add error handling with RFC 7807 format
  - [ ] Subtask 7.5: Add input validation with meaningful error messages
  - [ ] Subtask 7.6: Create handler tests following existing patterns

- [ ] **Task 8: Register Supplier Product Catalog Routes** (AC: 1)
  - [ ] Subtask 8.1: Update `apps/backend/internal/server/router.go`
  - [ ] Subtask 8.2: Add supplierProductCatalogHandler parameter to SetupRouter
  - [ ] Subtask 8.3: Register supplier product catalog routes with proper middleware (auth, RBAC)
  - [ ] Subtask 8.4: Add route groups: `/api/v1/supplier-product-catalogs`, `/api/v1/products/:id/price-history`, `/api/v1/suppliers/:id/product-catalog`

- [ ] **Task 9: Integrate with Purchase Invoice Recording** (AC: 1)
  - [ ] Subtask 9.1: Update `PurchaseInvoiceService` from Story 10-2 to use supplier catalog prices
  - [ ] Subtask 9.2: Add method `GetSuggestedPrice(supplierID, productID uint) (float64, error)` to invoice service
  - [ ] Subtask 9.3: When recording purchase invoice, pre-populate unit cost from supplier catalog if available
  - [ ] Subtask 9.4: Allow manual override of catalog price during invoice recording
  - [ ] Subtask 9.5: Log when catalog price is used vs manual price override

- [ ] **Task 10: Add Integration Tests** (AC: 1)
  - [ ] Subtask 10.1: Create `apps/backend/internal/handlers/supplier_product_catalog_handler_test.go`
  - [ ] Subtask 10.2: Test product association with supplier
  - [ ] Subtask 10.3: Test price update and history tracking
  - [ ] Subtask 10.4: Test preferred supplier marking
  - [ ] Subtask 10.5: Test price history query with date ranges
  - [ ] Subtask 10.6: Test integration with purchase invoice recording
  - [ ] Subtask 10.7: Test audit trail logging
  - [ ] Subtask 10.8: Test authentication and authorization
  - [ ] Subtask 10.9: Test error cases (duplicate association, invalid supplier/product, etc.)

## Dev Notes

### Project Structure Notes

Following the established project structure in `apps/backend/`:

```
apps/backend/
├── internal/
│   ├── models/
│   │   ├── supplier_product_catalog.go          [NEW] - GORM model
│   │   └── models.go                             [UPDATE] - Export new model
│   ├── repositories/
│   │   ├── supplier_product_catalog_repository.go          [NEW] - Interface
│   │   ├── supplier_product_catalog_repository_impl.go     [NEW] - Implementation
│   │   └── repository.go                          [UPDATE] - Add to container
│   ├── services/
│   │   ├── supplier_product_catalog_service.go             [NEW] - Interface
│   │   ├── supplier_product_catalog_service_impl.go        [NEW] - Implementation
│   │   ├── purchase_invoice_service.go          [UPDATE] - Integrate catalog prices
│   │   └── services.go                           [UPDATE] - Add to container
│   ├── handlers/
│   │   └── supplier_product_catalog_handler.go  [NEW] - HTTP handlers
│   ├── dto/
│   │   └── supplier_product_catalog_dto.go       [NEW] - Request/Response DTOs
│   └── server/
│       └── router.go                               [UPDATE] - Register routes
└── migrations/
    ├── 20260531310001_create_supplier_product_catalogs_table.up.sql   [NEW]
    └── 20260531310001_create_supplier_product_catalogs_table.down.sql  [NEW]
```

### Code Pattern References

**GORM Model Pattern** [Source: `internal/models/purchase_invoice.go` (Story 10-2)]:
- Use snake_case for DB columns, camelCase for JSON
- Include relationship methods: `Supplier() *Supplier`, `Product() *Product`
- Use pointer types for optional fields
- Implement `TableName()` method
- Add Swagger annotations for API documentation

**Price History Pattern** [Source: Architecture.md#Data Architecture]:
- Price history stored in same table with effective date ranges
- Current price: `price_effective_to IS NULL`
- Historical price: `price_effective_to` set to end date
- Query current price: `WHERE supplier_id = ? AND product_id = ? AND price_effective_to IS NULL`

**Transaction Pattern** [Source: PATCH-001 from Story 10-2 Code Review]:
- Wrap multi-table operations in database transactions
- Rollback entire transaction if any step fails
- Use `db.Transaction()` helper for proper transaction handling

**Repository Pattern** [Source: `internal/repositories/purchase_invoice_repository_impl.go` (Story 10-2)]:
- Interface in separate file
- Implementation in `{name}_impl.go`
- Add to Repository container in `repository.go`
- Use GORM for database operations
- Eager load relationships using `Preload`

**Service Layer Pattern** [Source: `internal/services/purchase_invoice_service_impl.go` (Story 10-2)]:
- Business logic validation (supplier exists, product exists, unique association)
- Integration with AuditService for logging
- Transaction wrapping for atomic operations
- Price history tracking logic

### Naming Conventions

**Database** [Source: Architecture.md#Naming Patterns]:
- Tables: `supplier_product_catalogs` (snake_case, plural)
- Columns: `id`, `supplier_id`, `product_id`, `purchase_price`, `is_preferred`, `sku_code`, `minimum_order_quantity`, `lead_time_days`, `branch_id`, `created_by`, `updated_by`, `created_at`, `updated_at`, `price_effective_from`, `price_effective_to`

**Go Code** [Source: Architecture.md#Naming Patterns]:
- Structs: `SupplierProductCatalog` (PascalCase)
- Methods: `AssociateProduct`, `UpdatePurchasePrice`, `SetPreferredSupplier` (PascalCase)
- Variables: `catalogRepo`, `catalogService` (camelCase)
- Files: `supplier_product_catalog.go`, `supplier_product_catalog_repository.go` (snake_case)

**API/JSON** [Source: Architecture.md#Naming Patterns]:
- Request DTOs: `AssociateProductRequest`, `UpdatePriceRequest`
- Response DTOs: `SupplierProductCatalogResponse`, `SupplierProductCatalogListResponse`, `PriceHistoryEntry`
- JSON fields: `id`, `supplierId`, `productId`, `purchasePrice`, `isPreferred`, `skuCode`, `minimumOrderQuantity`, `leadTimeDays`, `branchId`, `createdBy`, `updatedBy`, `createdAt`, `updatedAt`, `priceEffectiveFrom`, `priceEffectiveTo`

### Architecture Compliance

**Clean Architecture Layers** [Source: Architecture.md#Core Architectural Decisions]:
- Handler → Service → Repository → Model (GORM)
- Handlers handle HTTP concerns only
- Services contain business logic and validation
- Repositories handle data access only
- Models are simple GORM structs

**API Security** [Source: Architecture.md#Decision 6]:
- Apply JWT authentication middleware
- Apply RBAC middleware (Admin, Owner roles for catalog management)
- Use RFC 7807 for error responses
- Validate all input with struct tags

**Audit Trail** [Source: PRD#FR42, Architecture.md#Security Requirements]:
- Append-only audit logging for all supplier catalog operations
- Include: user ID, timestamp, action, affected entity (supplier, product, price)
- Log to `audit_logs` table via AuditService
- Log format: "catalog.product_associated", "catalog.price_updated", "catalog.preferred_set"

**Data Integrity** [Source: Architecture.md#Data Architecture]:
- Use database constraints for critical validations (NOT NULL, foreign keys, unique constraints)
- Use application-level validation for user-friendly error messages
- Transaction wrapping for price updates (archive old price, create new price)
- Unique constraint: one current price per supplier-product combination

**Price History Tracking** [Source: Architecture.md#Data Architecture]:
- Current price entries have `price_effective_to IS NULL`
- When updating price: (1) Set `price_effective_to = NOW()` on old entry, (2) Create new entry with `price_effective_from = NOW()`, `price_effective_to IS NULL`
- Price history queryable by date range
- All operations wrapped in transaction

### Testing Requirements

**Unit Tests** [Source: Existing test patterns in Story 10-2]:
- Test model validation
- Test repository operations (CRUD, price history logic)
- Test service business logic (association validation, price update logic, preferred supplier logic)
- Test handler request/response
- Use table-driven tests for multiple scenarios
- Mock external dependencies (AuditService)

**Integration Tests** [Source: Existing test patterns in Story 10-2]:
- Test product association with supplier
- Test price update and history tracking
- Test preferred supplier marking
- Test price history query with date ranges
- Test integration with purchase invoice recording
- Test audit trail logging
- Test authentication and authorization (Admin, Owner roles)
- Test error cases (duplicate association, invalid supplier/product, negative price, etc.)
- Use test database with cleanup

### Database Schema

**supplier_product_catalogs table** (NEW):
```sql
CREATE TABLE supplier_product_catalogs (
    id SERIAL PRIMARY KEY,
    supplier_id INTEGER NOT NULL REFERENCES suppliers(id),
    product_id INTEGER NOT NULL REFERENCES products(id),
    purchase_price DECIMAL(15,2) NOT NULL,
    is_preferred BOOLEAN NOT NULL DEFAULT false,
    sku_code VARCHAR(50),
    minimum_order_quantity INTEGER NOT NULL DEFAULT 1,
    lead_time_days INTEGER,
    branch_id INTEGER NOT NULL REFERENCES branches(id),
    created_by INTEGER NOT NULL REFERENCES users(id),
    updated_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    price_effective_from DATE NOT NULL DEFAULT CURRENT_DATE,
    price_effective_to DATE
);

CREATE UNIQUE INDEX idx_supplier_product_catalog_current ON supplier_product_catalogs(supplier_id, product_id, price_effective_from)
WHERE price_effective_to IS NULL;
CREATE INDEX idx_supplier_product_catalog_product ON supplier_product_catalogs(product_id);
CREATE INDEX idx_supplier_product_catalog_branch ON supplier_product_catalogs(branch_id);
CREATE INDEX idx_supplier_product_catalog_dates ON supplier_product_catalogs(price_effective_from, price_effective_to);
```

**Price History Logic:**
- Current price: `WHERE price_effective_to IS NULL`
- Historical prices: `WHERE price_effective_to IS NOT NULL`
- Price update transaction:
  1. UPDATE current entry: `SET price_effective_to = CURRENT_DATE WHERE id = ?`
  2. INSERT new entry: `price_effective_from = CURRENT_DATE, price_effective_to = NULL`

### API Endpoints

**POST** `/api/v1/supplier-product-catalogs` - Associate product with supplier
- Auth: Required (Admin, Owner roles)
- Request: `AssociateProductRequest` (supplierID, productID, purchasePrice, isPreferred, skuCode, minimumOrderQuantity, leadTimeDays)
- Response: `SupplierProductCatalogResponse` (201)

**PUT** `/api/v1/supplier-product-catalogs/:id/price` - Update purchase price
- Auth: Required (Admin, Owner roles)
- Request: `UpdatePriceRequest` (newPrice)
- Response: `SupplierProductCatalogResponse` (200)

**PUT** `/api/v1/supplier-product-catalogs/:id/preferred` - Set preferred supplier
- Auth: Required (Admin, Owner roles)
- Request: `{ isPreferred: boolean }`
- Response: `SupplierProductCatalogResponse` (200)

**GET** `/api/v1/supplier-product-catalogs/:id` - Get catalog entry by ID
- Auth: Required (Admin, Owner roles)
- Response: `SupplierProductCatalogResponse` (200)

**GET** `/api/v1/supplier-product-catalogs` - List catalog entries
- Auth: Required (Admin, Owner roles)
- Query: `?page=1&limit=20&supplier_id=&product_id=&branch_id=&is_preferred=`
- Response: `SupplierProductCatalogListResponse` with pagination

**GET** `/api/v1/products/:id/price-history` - Get price history for product
- Auth: Required (Admin, Owner roles)
- Query: `?start_date=&end_date=&supplier_id=`
- Response: `PriceHistoryResponse` with historical prices

**GET** `/api/v1/products/:id/preferred-supplier` - Get preferred supplier for product
- Auth: Required (Admin, Owner roles)
- Response: `SupplierProductCatalogResponse` (200) or 404 if no preferred supplier

**GET** `/api/v1/suppliers/:id/product-catalog` - Get supplier's product catalog
- Auth: Required (Admin, Owner roles)
- Query: `?page=1&limit=20`
- Response: `SupplierProductCatalogListResponse` with pagination

### Dependencies

**Existing Components to Integrate**:
- Supplier model and repository (from Story 10-1)
- Product model and repository (from Epic 4)
- Branch model and repository (from Epic 2)
- PurchaseInvoiceService from Story 10-2 (for price integration)
- AuditService (for logging catalog operations)
- RBAC middleware (for Admin, Owner role enforcement)
- Error handling middleware (for RFC 7807 responses)
- JWT authentication middleware

**No New External Dependencies Required**
- Uses existing GORM, Gin, and project libraries

### Cross-Story Context

This is the **fifth story in Epic 10**. Follow patterns established in previous stories.

**Previous Story (10-4) Intelligence:**
- Transaction wrapping for multi-table operations (CRITICAL from code review)
- Audit trail logging using structured logging
- Branch validation for data isolation
- Optimistic locking with version field validation
- Integration with AuditService for all operations
- Payment status calculation logic pattern (applies to price history)

**Key Learnings from Story 10-2 Code Review Patches:**
- Apply all 22 code review patches from Story 10-2 to prevent similar issues
- Use parameterized queries instead of string concatenation (PATCH-001)
- Use transactions for multi-table operations (PATCH-001 concept)
- Add branch access authorization (PATCH-003)
- Add overflow protection for calculations (PATCH-004)
- Validate date ranges and formats (PATCH-005, PATCH-010)
- Validate zero IDs (PATCH-006)
- Add minimum length checks after normalization (PATCH-007)
- Validate enum values (PATCH-008)
- Validate negative values (PATCH-009) - CRITICAL for price validation
- Add URL format validation (PATCH-011)
- Add overflow checks for pagination (PATCH-012)
- Use generic error messages (PATCH-013)
- Add audit failure warnings (PATCH-014)
- Add limits for array sizes (PATCH-015)
- Add nil pointer checks (PATCH-016)
- Normalize unicode consistently (PATCH-017)
- Use UTC for dates (PATCH-018)
- Handle empty pagination results (PATCH-019)
- Sanitize search queries (PATCH-021)
- Add truncation indicators (PATCH-022)

**Key Learnings from Story 10-3 Code Review Patches:**
- All 18 code review patches incorporated
- CRITICAL: Transaction wrapping for atomic operations
- CRITICAL: Audit trail logging for all operations
- HIGH: Optimistic locking for concurrent modification prevention
- HIGH: Pagination defaults to prevent division by zero
- MEDIUM: Nil pointer checks for relationship loading

**Related Stories for Context:**
- Story 10-1: Supplier Master Data Management (completed) - provides supplier data
- Story 10-2: Purchase Invoice Recording (completed) - integrates with catalog prices
- Story 10-3: Goods Receipt Processing (completed) - receives goods from catalog
- Story 10-4: Supplier Payment Tracking (completed) - manages payments for catalog purchases
- Story 10-6: Supplier Aging Reports (future) - uses catalog data for analysis
- Story 10-7: Supplier Transaction Audit Trail (future) - logs catalog operations

### Business Logic Requirements

**Product Association Logic:**
- Validate supplier exists and is active (not deleted)
- Validate product exists and is active (not deleted)
- Validate unique association: only one current price per supplier-product combination
- Validate purchase price > 0
- Validate minimum order quantity >= 1
- Create new catalog entry with `price_effective_from = CURRENT_DATE`, `price_effective_to = NULL`
- Log all operations to audit trail

**Price Update Logic:**
- Validate catalog entry exists
- Validate new price > 0
- Transaction wrapping:
  1. UPDATE current entry: `SET price_effective_to = CURRENT_DATE, updated_by = ?, updated_at = NOW() WHERE id = ?`
  2. INSERT new entry with new price, `price_effective_from = CURRENT_DATE`, `price_effective_to = NULL`, same supplier_id, product_id, other fields
- All historical prices preserved in same table
- Log price change to audit trail with old and new values

**Preferred Supplier Logic:**
- A product can have at most one preferred supplier per branch
- When setting new preferred supplier, unset `is_preferred` on previous preferred supplier for same product
- `is_preferred` is per product-supplier combination (branch-scoped)

**Price History Query Logic:**
- Query all entries (current and historical) for product
- Filter by date range if provided
- Group by supplier to show price evolution per supplier
- Show effective date ranges for each price
- Most recent price first

**Catalog by Supplier Logic:**
- List all products associated with supplier
- Show current purchase price for each product
- Indicate preferred status if applicable
- Support pagination for large catalogs
- Filter by branch for multi-branch support

**Integration with Purchase Invoice Recording:**
- When recording purchase invoice (Story 10-2), pre-populate unit cost from supplier catalog
- Query: `WHERE supplier_id = ? AND product_id = ? AND price_effective_to IS NULL`
- Allow manual override if catalog price is not applicable
- Log whether price came from catalog or manual override

**Validation Rules:**
- Supplier must exist and be active (DeletedAt is null)
- Product must exist and be active (DeletedAt is null)
- Purchase price must be > 0
- Minimum order quantity must be >= 1
- Lead time days must be >= 0 if provided
- SKU code max length 50 if provided
- Only one current association per supplier-product combination
- Branch access control: users can only manage catalogs for their branch (unless Owner/Admin)
- User must have Admin or Owner role

**Transaction Safety:**
- Wrap price update operation in database transaction
- Steps: (1) Update old entry price_effective_to, (2) Create new entry, (3) Log audit entries
- Rollback entire transaction if any step fails
- Ensure atomicity - all updates or none

**Multi-Branch Support:**
- Each catalog entry is associated with a branch via `branch_id`
- Users can only manage catalogs for their assigned branch (unless Owner/Admin)
- Preferred supplier is per branch (different branches can have different preferred suppliers)
- Catalog views respect branch-level access control

### Critical Implementation Notes

**PRICE HISTORY TRACKING [CRITICAL]:**
- MUST use `price_effective_from` and `price_effective_to` for history tracking
- Current price: `price_effective_to IS NULL`
- Historical price: `price_effective_to IS NOT NULL`
- MUST update price in transaction: archive old price, create new price
- Reference: Data Architecture pattern for temporal data

**TRANSACTION WRAPPING [CRITICAL]:**
- MUST wrap price update in single transaction (archive old + create new)
- Reference PATCH-001 from Story 10-2 and CRITICAL patches from Story 10-3
- Use `db.Transaction()` helper for proper transaction handling
- Rollback on any error during the operation

**UNIQUE ASSOCIATION VALIDATION [CRITICAL]:**
- MUST validate only one current price per supplier-product combination
- Database unique index enforces this at DB level
- Application-level validation for user-friendly error message
- Check: `WHERE supplier_id = ? AND product_id = ? AND price_effective_to IS NULL`

**PRICE VALIDATION [CRITICAL]:**
- MUST validate purchase_price > 0 (no zero or negative prices)
- Reference PATCH-009 from Story 10-2 code review
- Return validation error for non-positive prices
- Apply to both association and price update

**AUDIT LOGGING [CRITICAL]:**
- MUST log "catalog.product_associated" with supplier ID, product ID, price
- MUST log "catalog.price_updated" with catalog ID, old price, new price
- MUST log "catalog.preferred_set" with product ID, supplier ID, is_preferred flag
- MUST include user context (created_by/updated_by user ID)
- MUST use structured logging (slog.InfoContext) following Story 10-3 pattern
- Reference: AuditService pattern from Story 10-2

**BRANCH ACCESS VALIDATION [CRITICAL]:**
- MUST validate user's branch_id matches catalog's branch_id
- MUST check at handler level before calling service
- MUST apply to all endpoints (GET, POST, PUT)
- Reference: PATCH-003 from Story 10-2 code review

**INTEGRATION WITH PURCHASE INVOICE [CRITICAL]:**
- MUST update PurchaseInvoiceService to use catalog prices
- Add `GetSuggestedPrice(supplierID, productID uint)` method
- Pre-populate unit_cost in invoice recording from catalog
- Allow manual override with audit logging
- Reference: PurchaseInvoiceService from Story 10-2

**PREFERRED SUPPLIER LOGIC [CRITICAL]:**
- MUST ensure only one preferred supplier per product per branch
- When setting new preferred: unset `is_preferred = false` on previous preferred
- Transaction wrapping for atomic preferred supplier update
- Return 404 if no preferred supplier set

### References

- [Source: `_bmad-output/planning-artifacts/epics.md#Epic 10 Story 10.5`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Naming Patterns`]
- [Source: `_bmad-output/planning-artifacts/architecture.md#Data Architecture`]
- [Source: `_bmad-output/planning-artifacts/prd.md#FR39, FR41`]
- [Source: `apps/backend/internal/models/supplier.go` (Story 10-1)]
- [Source: `apps/backend/internal/models/product.go` (Epic 4)]
- [Source: `apps/backend/internal/models/purchase_invoice.go` (Story 10-2)]
- [Source: `apps/backend/internal/services/purchase_invoice_service.go` (Story 10-2)]
- [Source: `_bmad-output/implementation-artifacts/10-1-implement-supplier-master-data-management.md`]
- [Source: `_bmad-output/implementation-artifacts/10-2-implement-purchase-invoice-recording.md`]
- [Source: `_bmad-output/implementation-artifacts/10-4-implement-supplier-payment-tracking.md`]
- [Source: `_bmad-output/implementation-artifacts/10-2-code-review-triage.md`]
- [Source: `_bmad-output/implementation-artifacts/10-3-code-review-triage.md`]

## Dev Agent Record

### Agent Model Used

claude-opus-4-6

### Debug Log References

### Completion Notes List

_Story created: 2026-05-31_
_Story status: ready-for-dev_

### File List

_Story file created at:_ `/_bmad-output/implementation-artifacts/10-5-implement-supplier-product-catalog.md`

**Database Migrations:**
- `apps/backend/migrations/20260531310001_create_supplier_product_catalogs_table.up.sql` [PENDING]
- `apps/backend/migrations/20260531310001_create_supplier_product_catalogs_table.down.sql` [PENDING]

**Models:**
- `apps/backend/internal/models/supplier_product_catalog.go` [PENDING]
- `apps/backend/internal/models/models.go` [PENDING - to be updated]

**Repositories:**
- `apps/backend/internal/repositories/supplier_product_catalog_repository.go` [PENDING]
- `apps/backend/internal/repositories/supplier_product_catalog_repository_impl.go` [PENDING]
- `apps/backend/internal/repositories/repository.go` [PENDING - to be updated]

**Services:**
- `apps/backend/internal/services/supplier_product_catalog_service.go` [PENDING]
- `apps/backend/internal/services/supplier_product_catalog_service_impl.go` [PENDING]
- `apps/backend/internal/services/purchase_invoice_service.go` [PENDING - to be updated for catalog price integration]
- `apps/backend/internal/services/services.go` [PENDING - to be updated]

**DTOs:**
- `apps/backend/internal/dto/supplier_product_catalog_dto.go` [PENDING]

**Handlers:**
- `apps/backend/internal/handlers/supplier_product_catalog_handler.go` [PENDING]
- `apps/backend/internal/handlers/supplier_product_catalog_handler_test.go` [PENDING]

**Modified Files:**
- `apps/backend/internal/server/router.go` [PENDING - to be updated]
- `apps/backend/cmd/server/main.go` [PENDING - to be updated]
