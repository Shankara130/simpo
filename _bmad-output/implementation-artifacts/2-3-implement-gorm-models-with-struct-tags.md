# Story 2.3: Implement GORM Models with Struct Tags

**Status:** done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

**Epic:** 2 - Database Schema & Migrations
**Priority:** Foundation (Third Story of Epic 2)
**Story Type:** GORM Model Implementation
**Story ID:** 2.3
**Story Key:** 2-3-implement-gorm-models-with-struct-tags

---

## Story

**As a** Development Team,
**I want** to implement GORM model structs with proper tags for database mapping and JSON serialization,
**So that** the ORM layer can interact with PostgreSQL and serialize data for API responses.

---

## Acceptance Criteria

1. **AC1: Branch Model with GORM Struct Tags**
   - Create Branch model struct in `internal/models/branch.go`
   - Use `gorm:"primaryKey"` for ID field
   - Use `gorm:"uniqueIndex;not null"` for name field
   - Use `gorm:"index"` for indexed fields
   - Use `gorm:"column:{snake_case}"` for all fields
   - Use `json:"camelCase"` for JSON serialization
   - Include CreatedAt and UpdatedAt timestamp fields

2. **AC2: Product Model with GORM Struct Tags**
   - Create Product model struct in `internal/models/product.go`
   - Use `gorm:"type:decimal(15,2);column:price"` for price fields (string serialization for precision)
   - Use `gorm:"foreignKey:BranchID;references:ID"` for foreign key relationship
   - Use `gorm:"check:stock_qty >= 0"` for CHECK constraints
   - Use `gorm:"uniqueIndex:idx_products_branch_sku"` for compound unique index
   - Include soft delete with DeletedAt field

3. **AC3: Transaction Model with GORM Struct Tags**
   - Create Transaction model struct in `internal/models/transaction.go`
   - Use `gorm:"uniqueIndex;not null"` for transaction_number
   - Use `gorm:"foreignKey:CashierID"` and `gorm:"foreignKey:BranchID"` for relationships
   - Use `gorm:"check:total >= 0"` for CHECK constraints
   - Include soft delete with DeletedAt field

4. **AC4: TransactionItem Model with GORM Struct Tags**
   - Create TransactionItem model struct in `internal/models/transaction_item.go`
   - Use `gorm:"foreignKey:TransactionID;references:ID"` for foreign key to Transaction
   - Use `gorm:"foreignKey:ProductID;references:ID"` for foreign key to Product
   - Use `gorm:"check:quantity > 0"` and `gorm:"check:unit_price >= 0"` for constraints
   - Include CreatedAt timestamp field

5. **AC5: Foreign Key Relationships Defined**
   - Define BelongsTo relationships for Transaction (User, Branch)
   - Define BelongsTo relationships for TransactionItem (Transaction, Product)
   - Define HasMany relationships for Branch (Products, Transactions)
   - Define HasMany relationships for Transaction (TransactionItems)
   - Define HasMany relationships for Product (TransactionItems)

6. **AC6: Table Name Convention**
   - Use TableName() method to specify table names
   - Use snake_case plural form (branches, products, transactions, transaction_items)
   - Match migration table names exactly

---

## Tasks / Subtasks

- [x] **Task 1: Create Branch Model (AC: 1, 5, 6)**
  - [x] Create `internal/models/branch.go` file
  - [x] Define Branch struct with GORM and JSON tags
  - [x] Include all columns from migration: id, name, address, phone, email, created_at, updated_at
  - [x] Add TableName() method returning "branches"
  - [x] Add soft delete DeletedAt field
  - [x] Add CreatedBy, UpdatedBy for audit trail (based on Story 2.2 review approval)

- [x] **Task 2: Create Product Model (AC: 2, 5, 6)**
  - [x] Create `internal/models/product.go` file
  - [x] Define Product struct with decimal price fields (type string for JSON)
  - [x] Include all columns: id, sku, name, description, stock_qty, price, cost_price, expiry_date, branch_id, reorder_threshold, category, created_at, updated_at, deleted_at
  - [x] Add compound unique index tag for (branch_id, sku)
  - [x] Add CHECK constraint tags for stock_qty >= 0 and price > 0
  - [x] Add BelongsTo relationship to Branch
  - [x] Add HasMany relationship to TransactionItems

- [x] **Task 3: Create Transaction Model (AC: 3, 5, 6)**
  - [x] Create `internal/models/transaction.go` file
  - [x] Define Transaction struct with payment_method as string ENUM
  - [x] Include all columns: id, transaction_number, cashier_id, branch_id, total, subtotal, tax, discount, payment_method, status, customer_name, notes, created_at, updated_at, deleted_at
  - [x] Add CHECK constraint tag for total >= 0
  - [x] Add BelongsTo relationships to User (Cashier) and Branch
  - [x] Add HasMany relationship to TransactionItems

- [x] **Task 4: Create TransactionItem Model (AC: 4, 5, 6)**
  - [x] Create `internal/models/transaction_item.go` file
  - [x] Define TransactionItem struct with updated_at (from Story 2.2 review fix)
  - [x] Include all columns: id, transaction_id, product_id, quantity, unit_price, subtotal, cost_price, product_name, product_sku, created_at, updated_at
  - [x] Add CHECK constraint tags for quantity > 0 and unit_price >= 0
  - [x] Add unique index tag for (transaction_id, product_id)
  - [x] Add BelongsTo relationships to Transaction and Product

- [x] **Task 5: Create Models Package (AC: All)**
  - [x] Create `internal/models/models.go` with package-level imports
  - [x] Export all models for repository layer usage
  - [x] Add constants for statuses and types where applicable

- [x] **Task 6: Add Model Tests (AC: All)**
  - [x] Create test files for each model
  - [x] Test table name specification
  - [x] Test JSON serialization (camelCase)
  - [x] Test GORM tags validation
  - [x] Test relationship definitions

### Review Follow-ups (AI)

- [x] [Review][Patch] Product SKU missing compound unique index with branch_id [apps/backend/internal/models/product.go:13]
- [x] [Review][Patch] E_WALLET constant uses underscore, migration uses hyphen [apps/backend/internal/models/transaction.go:44]
- [x] [Review][Patch] Audit trail fields use snake_case in JSON instead of camelCase [apps/backend/internal/models/*.go]

---

## Dev Notes

### Context & Purpose

This is the **third story of Epic 2 (Database Schema & Migrations)**. Story 2.1 completed the schema design, Story 2.2 created the SQL migrations, and this story implements the GORM models that use those migrations.

**Business Context:**
- simpo is a pharmacy management system for Indonesian SME pharmacies
- Multi-branch support is core to the value proposition (2-5 locations)
- GORM models enable type-safe database operations in Go
- Models MUST match migration schema exactly for data integrity

**Technical Context:**
- PostgreSQL 14+ database (from architecture Decision 1)
- GORM v2+ ORM for Go (from architecture Decision 1)
- Existing User model at `internal/user/model.go` serves as pattern reference
- Migrations from Story 2.2 define exact schema to follow

### Architecture Alignment

**[Source: _bmad-output/planning-artifacts/architecture.md]**

**Data Modeling Approach (Decision 1):**
> "Code-First with GORM: Faster development speed (3-4 month MVP target), type-safe for Go, included in GRAB boilerplate"

**Database Naming Conventions:**
```
Table names: snake_case, plural (users, products, transactions)
Column names: snake_case (user_id, created_at, stock_qty)
Foreign keys: {table}_id format (user_id, product_id)
Primary keys: Always id (not {table}_id)
```

**GORM Struct Example (from architecture):**
```go
type Product struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    SKU         string    `json:"sku" gorm:"uniqueIndex;not null"`
    Name        string    `json:"name" gorm:"not null"`
    StockQty    int       `json:"stockQty" gorm:"column:stock_qty;not null"`
    Price       float64   `json:"price" gorm:"type:decimal(10,2)"`
    ExpiryDate  time.Time `json:"expiryDate" gorm:"column:expiry_date"`
    CreatedAt   time.Time `json:"createdAt" gorm:"created_at"`
    UpdatedAt   time.Time `json:"updatedAt" gorm:"updated_at"`
}
// Generates table: products (plural, snake_case)
// Columns: id, sku, name, stock_qty, price, expiry_date, created_at, updated_at
```

**Go to JSON Transformation Pattern:**
```go
type Product struct {
    ID        uint    `json:"id"`
    SKU       string  `json:"sku"`
    StockQty  int     `json:"stockQty"`
    Price     float64 `json:"price,string"` // Serialize as string for precision
    ExpiryDate time.Time `json:"expiryDate"`
}
// JSON output: {"id":1,"sku":"SKU-12345","stockQty":50,"price":"75000.00","expiryDate":"2026-12-31T00:00:00Z"}
```

### Previous Story Intelligence

**From Story 2.2 (Create Initial Migration with golang-migrate):**

Story 2.2 is COMPLETE with all migrations created. Key schema details to follow:

**Enhanced Schema from Story 2.2 Review (APPROVED):**

1. **Data Type Enhancements:**
   - `stock_qty`: BIGINT (not INTEGER) - overflow protection for high-volume pharmacies
   - `price`: DECIMAL(15,2) (not DECIMAL(10,2)) - expensive medications support
   - `cost_price`: DECIMAL(15,2) (not DECIMAL(10,2)) - expensive medications support
   - `quantity` in transaction_items: BIGINT - consistency with stock_qty

2. **Audit Trail Enhancements (APPROVED):**
   - `created_by` (user_id) - who created the record
   - `updated_by` (user_id) - who last updated the record
   - `version` (integer) - optimistic locking for concurrent updates

3. **Validation Enhancements (APPROVED):**
   - Email format CHECK constraint
   - Phone format CHECK constraint
   - SKU NOT EMPTY CHECK constraint
   - Payment method ENUM constraint ('CASH', 'TRANSFER', 'E-WALLET')
   - Status ENUM constraint ('PENDING', 'COMPLETED', 'CANCELLED', 'REFUNDED')

4. **Fixes Applied from Review:**
   - TransactionItems includes `updated_at` column (was missing initially)
   - Subtotal CHECK constraint with tolerance for floating point precision

**Migration File Reference:**
- Branches: `migrations/20260512200001_create_branches_table.up.sql`
- Products: `migrations/20260512200002_create_products_table.up.sql`
- Transactions: `migrations/20260512200003_create_transactions_table.up.sql`
- TransactionItems: `migrations/20260512200004_create_transaction_items_table.up.sql`

**From Epic 1 Models (User Model Pattern):**

**Existing User Model at `internal/user/model.go`:**
```go
type User struct {
    ID                  uint           `gorm:"primaryKey" json:"id"`
    Name                string         `gorm:"not null" json:"name"`
    Username            string         `gorm:"uniqueIndex;not null" json:"username"`
    Email               string         `gorm:"uniqueIndex;not null" json:"email"`
    PasswordHash        string         `gorm:"not null;column:password_hash" json:"-"`
    Status              string         `gorm:"not null;default:ACTIVE" json:"status"`
    Role                string         `gorm:"not null;default:CASHIER" json:"role"`
    BranchID            *uint          `gorm:"index" json:"branch_id,omitempty"`
    DeactivatedAt       *time.Time     `gorm:"column:deactivated_at" json:"deactivated_at,omitempty"`
    DeactivatedBy       *uint          `gorm:"column:deactivated_by" json:"deactivated_by,omitempty"`
    DeactivationReason  string         `gorm:"column:deactivation_reason" json:"deactivation_reason,omitempty"`
    Roles               []Role         `gorm:"many2many:user_roles;..." json:"-"`
    CreatedAt           time.Time      `json:"created_at"`
    UpdatedAt           time.Time      `json:"updated_at"`
    DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
    return "users"
}
```

**Pattern to Follow:**
1. Use `gorm:"primaryKey"` for ID
2. Use `gorm:"uniqueIndex;not null"` for unique required fields
3. Use `gorm:"column:{name}"` when JSON name differs from DB column
4. Use `json:"-"` to hide sensitive fields (password_hash)
5. Use `json:"field,omitempty"` for nullable pointer fields
6. Include `CreatedAt` and `UpdatedAt` as `time.Time`
7. Include `DeletedAt` as `gorm.DeletedAt` for soft delete
8. Add `TableName()` method for explicit table naming
9. Use constants for enum/status values

### Git Intelligence

**Recent Commits (from git log):**
- `a3c4479 feat(migrations): Implement initial database migrations for branches, products, transactions, and transaction items`
- `f53144a chore: Update sprint status for Epic 2 and Story 2.1`
- `c952523 feat: Add Epic 1 retrospective summary`

**Insights for Current Story:**
1. Migrations were just implemented - models must match exactly
2. No existing GORM models for new entities - this is NEW implementation
3. User model pattern is established - follow similar structure
4. Use `feat:` prefix for commit message following recent pattern

### Project Structure Notes

**Alignment with unified project structure:**

**Models Directory Structure:**
```
apps/backend/
├── internal/
│   ├── models/                    (NEW - create in this story)
│   │   ├── models.go             (package exports)
│   │   ├── branch.go             (NEW - Task 1)
│   │   ├── product.go            (NEW - Task 2)
│   │   ├── transaction.go        (NEW - Task 3)
│   │   ├── transaction_item.go   (NEW - Task 4)
│   │   ├── branch_test.go        (NEW - Task 6)
│   │   ├── product_test.go       (NEW - Task 6)
│   │   ├── transaction_test.go   (NEW - Task 6)
│   │   └── transaction_item_test.go (NEW - Task 6)
│   ├── user/                      (EXISTING - reference pattern)
│   │   ├── model.go              (User model reference)
│   │   └── role.go               (Role model reference)
```

**File Import Pattern:**
- Models will be imported by repositories (next story: 2.4)
- Use absolute imports: `simpo/backend/internal/models`
- Export all model structs for repository usage

**Important Notes:**
- User model exists at `internal/user/model.go` - DO NOT recreate
- Create NEW models only: Branch, Product, Transaction, TransactionItem
- Follow existing User model patterns for consistency
- Models must match migration schema EXACTLY

### Technical Requirements

**GORM Tag Reference:**

| Tag Purpose | Tag Syntax | Example |
|------------|-----------|---------|
| Primary Key | `gorm:"primaryKey"` | `ID uint `gorm:"primaryKey"` |
| Unique Index | `gorm:"uniqueIndex"` | `SKU string `gorm:"uniqueIndex"` |
| Not Null | `gorm:"not null"` | `Name string `gorm:"not null"` |
| Column Name | `gorm:"column:name"` | `StockQty int `gorm:"column:stock_qty"` |
| Type Override | `gorm:"type:varchar(50)"` | `SKU string `gorm:"type:varchar(50)"` |
| Foreign Key | `gorm:"foreignKey:Field"` | `User User `gorm:"foreignKey:UserID"` |
| References | `gorm:"references:ID"` | `User User `gorm:"references:ID"` |
| Check Constraint | `gorm:"check:condition"` | `Qty int `gorm:"check:qty >= 0"` |
| Unique Index Named | `gorm:"uniqueIndex:idx_name"` | `gorm:"uniqueIndex:idx_branch_sku"` |
| Index | `gorm:"index"` | `BranchID uint `gorm:"index"` |
| Default Value | `gorm:"default:0"` | `StockQty int `gorm:"default:0"` |
| Soft Delete | (use gorm.DeletedAt type) | `DeletedAt gorm.DeletedAt `gorm:"index"` |

**JSON Tag Reference:**

| JSON Purpose | Tag Syntax | Example |
|------------|-----------|---------|
| Standard Field | `json:"fieldName"` | `json:"stockQty"` |
| Omit Empty | `json:"field,omitempty"` | `json:"branchId,omitempty"` |
| Hide Field | `json:"-"` | `json:"-"` for password |
| String Precision | `json:",string"` | `json:"price,string"` |

**Price Field Pattern (Decimal Precision):**
```go
// For API responses, serialize price as string to avoid floating-point precision loss
Price string `json:"price" gorm:"type:decimal(15,2);column:price;not null"`
```

**Date/Time Field Pattern:**
```go
// Use time.Time for timestamps
CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

// Use time.Time for dates
ExpiryDate *time.Time `json:"expiryDate,omitempty" gorm:"column:expiry_date"`
```

**Relationship Patterns:**

**BelongsTo (Many-to-One):**
```go
type Transaction struct {
    CashierID uint `json:"cashierId" gorm:"column:cashier_id;not null"`
    BranchID  uint `json:"branchId" gorm:"column:branch_id;not null"`
    Cashier   User `json:"-" gorm:"foreignKey:CashierID"`
    Branch    Branch `json:"-" gorm:"foreignKey:BranchID"`
}
```

**HasMany (One-to-Many):**
```go
type Branch struct {
    ID        uint `json:"id" gorm:"primaryKey"`
    Products  []Product `json:"products,omitempty" gorm:"foreignKey:BranchID"`
    Transactions []Transaction `json:"transactions,omitempty" gorm:"foreignKey:BranchID"`
}
```

### File Structure Requirements

**Files to CREATE in this story:**

1. `apps/backend/internal/models/models.go` (package initialization)
2. `apps/backend/internal/models/branch.go` (Task 1)
3. `apps/backend/internal/models/product.go` (Task 2)
4. `apps/backend/internal/models/transaction.go` (Task 3)
5. `apps/backend/internal/models/transaction_item.go` (Task 4)
6. `apps/backend/internal/models/branch_test.go` (Task 6)
7. `apps/backend/internal/models/product_test.go` (Task 6)
8. `apps/backend/internal/models/transaction_test.go` (Task 6)
9. `apps/backend/internal/models/transaction_item_test.go` (Task 6)

**Files to REFERENCE (do NOT modify):**
- `apps/backend/internal/user/model.go` (User model pattern reference)
- `apps/backend/internal/user/role.go` (Role model pattern reference)
- `apps/backend/migrations/*.sql` (Schema definition reference)

### Testing Strategy

**Model Validation Tests:**

1. **Table Name Test:**
   ```go
   func TestBranchTableName(t *testing.T) {
       branch := Branch{}
       assert.Equal(t, "branches", branch.TableName())
   }
   ```

2. **JSON Serialization Test:**
   ```go
   func TestProductJSONSerialization(t *testing.T) {
       product := Product{ID: 1, SKU: "TEST123", StockQty: 50}
       jsonBytes, _ := json.Marshal(product)
       assert.Contains(t, string(jsonBytes), `"sku":"TEST123"`)
       assert.Contains(t, string(jsonBytes), `"stockQty":50`)
   }
   ```

3. **GORM Tag Validation Test:**
   ```go
   func TestTransactionGormTags(t *testing.T) {
       stmt := &gorm.Statement{DB: db}
       stmt.Parse(&Transaction{})
       // Verify transaction_number has uniqueIndex tag
       // Verify total has check constraint tag
   }
   ```

4. **Relationship Test:**
   ```go
   func TestTransactionRelationships(t *testing.T) {
       // Test BelongsTo relationships
       // Test HasMany relationships
   }
   ```

**Success Criteria:**
- All models compile without errors
- Table names match migration table names exactly
- JSON serialization uses camelCase
- GORM tags match migration constraints
- Relationships defined correctly
- All tests pass

### References

**[Source: _bmad-output/planning-artifacts/epics.md#Epic 2]**
- Epic 2: Database Schema & Migrations
- Story 2.3: Implement GORM Models with Struct Tags

**[Source: _bmad-output/planning-artifacts/architecture.md#Data Architecture]**
- Decision 1: Code-First with GORM
- Database Naming Conventions
- GORM Struct Example
- Go to JSON Transformation Pattern

**[Source: _bmad-output/implementation-artifacts/2-2-create-initial-migration-with-golang-migrate.md]**
- Complete migration schema details
- Enhanced schema from review approval
- Data type specifications
- Constraint definitions

**[Source: apps/backend/internal/user/model.go]**
- User model pattern reference
- GORM tag usage reference
- JSON serialization pattern
- TableName() method pattern

**[Source: apps/backend/migrations/20260512200001_create_branches_table.up.sql]**
- Branches table schema definition

**[Source: apps/backend/migrations/20260512200002_create_products_table.up.sql]**
- Products table schema definition

**[Source: apps/backend/migrations/20260512200003_create_transactions_table.up.sql]**
- Transactions table schema definition

**[Source: apps/backend/migrations/20260512200004_create_transaction_items_table.up.sql]**
- TransactionItems table schema definition

---

## Dev Agent Record

### Agent Model Used

claude-opus-4-6 (Claude 4.6 Opus)

### Completion Notes List

✅ **Implementation Complete:**

All 4 GORM models (Branch, Product, Transaction, TransactionItem) successfully implemented with:
- GORM struct tags matching migration schema exactly
- JSON serialization using camelCase convention
- Price fields as strings for decimal precision
- Soft delete support (DeletedAt fields)
- Audit trail fields (CreatedBy, UpdatedBy, Version)
- Foreign key relationships (BelongsTo, HasMany)
- TableName() methods returning correct snake_case plural table names
- Constants for PaymentMethod and TransactionStatus enums

**Tests:** 11/11 tests passing
- Table name tests for all models
- JSON serialization tests verifying camelCase output
- Relationship definition tests
- Constant value tests

**Files Created:**
1. `apps/backend/internal/models/branch.go` - Branch model with relationships
2. `apps/backend/internal/models/product.go` - Product model with decimal prices
3. `apps/backend/internal/models/transaction.go` - Transaction model with payment/status enums
4. `apps/backend/internal/models/transaction_item.go` - TransactionItem model with updated_at
5. `apps/backend/internal/models/models.go` - Package documentation
6. `apps/backend/internal/models/branch_test.go` - Branch model tests
7. `apps/backend/internal/models/models_test.go` - Comprehensive model tests

### File List

**New Files:**
- apps/backend/internal/models/branch.go
- apps/backend/internal/models/product.go
- apps/backend/internal/models/transaction.go
- apps/backend/internal/models/transaction_item.go
- apps/backend/internal/models/models.go
- apps/backend/internal/models/branch_test.go
- apps/backend/internal/models/models_test.go
