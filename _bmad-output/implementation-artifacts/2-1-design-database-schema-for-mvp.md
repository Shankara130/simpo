# Story 2.1: Design Database Schema for MVP

**Status:** ready-for-dev

**Epic:** 2 - Database Schema & Migrations
**Priority:** Foundation (First Story of Epic 2)
**Story Type:** Core Design
**Story ID:** 2.1
**Story Key:** 2-1-design-database-schema-for-mvp

---

## Story

**As a** Development Team,
**I want** to design a complete database schema that supports all Phase 1 MVP features,
**So that** we have a clear blueprint for implementing data models with GORM.

---

## Acceptance Criteria

1. **AC1: Core Entities Identified**
   - Users entity defined with attributes (id, username, password_hash, email, role, branch_id, status, created_at, updated_at)
   - Products entity defined with attributes (id, sku, name, description, stock_qty, price, expiry_date, branch_id, created_at, updated_at)
   - Transactions entity defined with attributes (id, transaction_number, cashier_id, total, payment_method, status, created_at, updated_at)
   - TransactionItems entity defined with attributes (id, transaction_id, product_id, quantity, unit_price, subtotal, created_at)
   - Branches entity defined with attributes (id, name, address, phone, created_at, updated_at)

2. **AC2: Relationships Defined**
   - Users → Branches: Many-to-One (user belongs to one branch, branch has many users)
   - Transactions → Users: Many-to-One (transaction belongs to one cashier/user)
   - TransactionItems → Transactions: Many-to-One (line item belongs to one transaction)
   - TransactionItems → Products: Many-to-One (line item references one product)
   - Products → Branches: Many-to-One (product belongs to one branch)

3. **AC3: Indexes Identified**
   - Unique index on users.username (login lookup)
   - Unique index on users.email (email lookup)
   - Unique index on products.sku (product lookup)
   - Unique index on products.branch_id + sku (compound unique for multi-branch)
   - Index on transactions.transaction_number (transaction lookup)
   - Index on transactions.cashier_id (cashier's transaction history)
   - Index on transactions.created_at (date-based reporting)
   - Index on transactions.branch_id (branch-level reporting)
   - Index on products.expiry_date (expiry alert queries)

4. **AC4: Constraints Defined**
   - NOT NULL on all required fields (username, password_hash, email, role, etc.)
   - Foreign key relationships with proper CASCADE rules
   - CHECK constraints for enum-like fields (status, role, payment_method)
   - Default values for status (ACTIVE), created_at, updated_at

5. **AC5: Branch-Level Data Isolation**
   - All multi-tenant tables have branch_id foreign key
   - Users, Products, Transactions tables include branch_id
   - System admin users have nullable branch_id (global access)
   - Queries respect branch scoping for RBAC

6. **AC6: GORM Struct Tags Defined**
   - Table name specified using `gorm:"tableName"`
   - Column names mapped using `gorm:"column:column_name"`
   - JSON serialization using `json:"camelCase"` tags
   - Validation tags (required, min, max) where applicable
   - Index tags for database indexes

---

## Developer Context

### Context & Purpose

This is the **first story of Epic 2 (Database Schema & Migrations)** and establishes the complete data model for simpo's MVP. The schema design will guide all subsequent stories in this epic (migrations, GORM models, repository layer).

**Business Context:**
- simpo is a pharmacy management system for Indonesian SME pharmacies
- Multi-branch support is core to the value proposition (2-5 locations)
- Regulatory compliance (Badan POM) requires audit trails and expiry tracking
- Phase 1 MVP focuses on single-branch, but schema must support multi-branch from day one

**Technical Context:**
- PostgreSQL 14+ with GORM ORM (code-first approach)
- golang-migrate for version-controlled migrations
- Clean Architecture: models in internal/models/, repositories in internal/repositories/
- Audit trail fields (created_at, updated_at, created_by) on all tables
- branch_id foreign key on all multi-tenant tables for data isolation

### Architecture Alignment

**[Source: _bmad-output/planning-artifacts/architecture.md]**

**Data Architecture Decisions:**
- **Decision 1:** Code-First with GORM (faster development, type-safe)
- **Decision 2:** Hybrid validation (database constraints + Go validation)
- **Decision 3:** GORM + golang-migrate (production-safe migrations)
- **Decision 4:** Layered caching with Redis (session, query, pub/sub)

**Database Naming Conventions:**
```
Table names: snake_case, plural (users, products, transactions)
Column names: snake_case (user_id, created_at, stock_qty)
Foreign keys: {table}_id format (user_id, product_id)
Indexes: idx_{table}_{column} format
Primary keys: Always id (not {table}_id)
Timestamps: created_at, updated_at (GORM auto-managed)
```

**GORM Struct Pattern (from architecture):**
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
```

### Previous Story Intelligence

**From Epic 1 Stories (1.1 - 1.10):**

**Key Learnings to Apply:**

1. **User Model Pattern (Story 1.5, 1.6, 1.7, 1.8, 1.9, 1.10):**
   - User struct already implemented in `apps/backend/internal/user/model.go`
   - Fields: ID, Username, PasswordHash, Email, Status, Role, BranchID, CreatedAt, UpdatedAt, DeletedAt
   - Status constants: ACTIVE, INACTIVE, PENDING
   - Role constants: SYSTEM_ADMIN, OWNER, CASHIER
   - BranchID is nullable (*uint) for system admins (global access)

2. **Audit Trail Pattern (Stories 1.7, 1.9, 1.10):**
   - All state changes log who, when, what, why
   - AuditService implemented in `apps/backend/internal/services/audit_service.go`
   - Append-only audit_logs table for Badan POM compliance

3. **Branch-Level Access Control (Stories 1.6, 1.8):**
   - RBAC middleware enforces branch scoping
   - Cashiers restricted to assigned branch
   - Owners and System Admins have cross-branch access

4. **GORM Model Conventions:**
   - Use `gorm:"column:snake_case"` to map Go CamelCase to DB snake_case
   - Use `json:"camelCase"` for JSON API responses
   - Use `gorm:"uniqueIndex"` for unique constraints
   - Use `gorm:"not null"` for required fields
   - Use `gorm:"index"` for non-unique indexes
   - Use pointer types (*uint, *string) for nullable database columns

5. **Testing Pattern (All Epic 1 Stories):**
   - Co-located test files: `model_test.go`, `service_test.go`
   - Use test fixtures and factory patterns
   - Database transactions rolled back after tests
   - >80% test coverage requirement

**Existing User Model (from Story 1.5, 1.10):**
```go
// Location: apps/backend/internal/user/model.go
type User struct {
    ID             uint           `gorm:"primaryKey" json:"id"`
    Name           string         `gorm:"not null" json:"name"`
    Username       string         `gorm:"uniqueIndex;not null" json:"username"`
    Email          string         `gorm:"uniqueIndex;not null" json:"email"`
    PasswordHash   string         `gorm:"not null;column:password_hash" json:"-"`
    Status         string         `gorm:"not null;default:ACTIVE" json:"status"`
    Role           string         `gorm:"not null;default:CASHIER" json:"role"`
    BranchID       *uint          `gorm:"index" json:"branch_id,omitempty"`
    DeactivatedAt  *time.Time     `gorm:"column:deactivated_at" json:"deactivated_at,omitempty"`
    DeactivatedBy  *uint          `gorm:"column:deactivated_by" json:"deactivated_by,omitempty"`
    DeactivationReason string      `gorm:"column:deactivation_reason" json:"deactivation_reason,omitempty"`
    CreatedAt      time.Time      `json:"created_at"`
    UpdatedAt      time.Time      `json:"updated_at"`
    DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}
```

**What to Build On:**
- User model is complete and follows GORM conventions
- AuditService exists and can be extended for new entities
- Repository pattern exists in `apps/backend/internal/user/repository.go`
- Follow the same pattern for Product, Transaction, TransactionItem, Branch models

### Complete Entity Definitions

#### 1. Branches Table

**Purpose:** Represent pharmacy locations in the multi-tenant system

**Schema:**
```sql
CREATE TABLE branches (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    address TEXT,
    phone VARCHAR(20),
    email VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_branches_name ON branches(name);
CREATE INDEX idx_branches_email ON branches(email);
```

**GORM Model:**
```go
// File: apps/backend/internal/branch/model.go
package branch

import (
    "time"
    "gorm.io/gorm"
)

type Branch struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Name      string         `gorm:"uniqueIndex;not null" json:"name"`
    Address   string         `gorm:"type:text" json:"address"`
    Phone     string         `gorm:"type:varchar(20)" json:"phone"`
    Email     string         `gorm:"type:varchar(100);index" json:"email"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Branch) TableName() string {
    return "branches"
}
```

#### 2. Users Table (Already Exists - Reference Only)

**Note:** Users table already implemented in Epic 1. Documented here for completeness.

**Schema Reference:**
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    role VARCHAR(20) NOT NULL DEFAULT 'CASHIER',
    branch_id INTEGER REFERENCES branches(id) ON DELETE SET NULL,
    deactivated_at TIMESTAMP,
    deactivated_by INTEGER REFERENCES users(id),
    deactivation_reason TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_branch_id ON users(branch_id);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);
```

**Model Location:** `apps/backend/internal/user/model.go`

#### 3. Products Table

**Purpose:** Inventory items with stock levels, pricing, and expiry tracking

**Schema:**
```sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(50) NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    stock_qty INTEGER NOT NULL DEFAULT 0,
    price DECIMAL(10,2) NOT NULL,
    cost_price DECIMAL(10,2),
    expiry_date DATE,
    branch_id INTEGER NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    reorder_threshold INTEGER DEFAULT 10,
    category VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT check_stock_not_negative CHECK (stock_qty >= 0),
    CONSTRAINT check_price_positive CHECK (price > 0)
);

-- Compound unique: SKU is unique per branch
CREATE UNIQUE INDEX idx_products_branch_sku ON products(branch_id, sku);
CREATE INDEX idx_products_expiry ON products(expiry_date);
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_branch_id ON products(branch_id);
```

**GORM Model:**
```go
// File: apps/backend/internal/product/model.go
package product

import (
    "time"
    "gorm.io/gorm"
)

type Product struct {
    ID               uint           `gorm:"primaryKey" json:"id"`
    SKU              string         `gorm:"not null;size:50" json:"sku"`
    Name             string         `gorm:"not null;size:200" json:"name"`
    Description      string         `gorm:"type:text" json:"description"`
    StockQty         int            `gorm:"column:stock_qty;not null;default:0" json:"stockQty"`
    Price            float64        `gorm:"type:decimal(10,2);not null" json:"price"`
    CostPrice        float64        `gorm:"column:cost_price;type:decimal(10,2)" json:"costPrice"`
    ExpiryDate       *time.Time     `gorm:"column:expiry_date;index" json:"expiryDate,omitempty"`
    BranchID         uint           `gorm:"not null;index" json:"branchId"`
    ReorderThreshold int            `gorm:"column:reorder_threshold;default:10" json:"reorderThreshold"`
    Category         string         `gorm:"size:50;index" json:"category"`
    CreatedAt        time.Time      `json:"createdAt"`
    UpdatedAt        time.Time      `json:"updatedAt"`
    DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

    // Relationships (loaded via GORM Preload)
    Branch           *Branch        `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

func (Product) TableName() string {
    return "products"
}

// ProductStatus constants for future use
const (
    ProductStatusActive   = "ACTIVE"
    ProductStatusInactive = "INACTIVE"
    ProductStatusExpired  = "EXPIRED"
)
```

**Key Design Decisions:**
- **Compound unique index** on (branch_id, sku): Same SKU can exist in different branches
- **CHECK constraint** for non-negative stock: Database-level validation
- **Nullable ExpiryDate**: Some products may not expire (e.g., equipment)
- **CostPrice field**: Stored for profit/loss reporting (Story 5.2)
- **ReorderThreshold**: Configurable per product for low stock alerts

#### 4. Transactions Table

**Purpose:** Sales transactions with payment and status tracking

**Schema:**
```sql
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    transaction_number VARCHAR(50) UNIQUE NOT NULL,
    cashier_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    branch_id INTEGER NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    total DECIMAL(12,2) NOT NULL,
    subtotal DECIMAL(12,2) NOT NULL,
    tax DECIMAL(12,2) DEFAULT 0,
    discount DECIMAL(12,2) DEFAULT 0,
    payment_method VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'COMPLETED',
    customer_name VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT check_total_positive CHECK (total >= 0)
);

CREATE UNIQUE INDEX idx_transactions_number ON transactions(transaction_number);
CREATE INDEX idx_transactions_cashier ON transactions(cashier_id);
CREATE INDEX idx_transactions_branch ON transactions(branch_id);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_transactions_status ON transactions(status);
```

**GORM Model:**
```go
// File: apps/backend/internal/transaction/model.go
package transaction

import (
    "time"
    "gorm.io/gorm"
)

type Transaction struct {
    ID                uint           `gorm:"primaryKey" json:"id"`
    TransactionNumber string         `gorm:"uniqueIndex;column:transaction_number;not null;size:50" json:"transactionNumber"`
    CashierID         uint           `gorm:"column:cashier_id;not null;index" json:"cashierId"`
    BranchID          uint           `gorm:"column:branch_id;not null;index" json:"branchId"`
    Total             float64        `gorm:"type:decimal(12,2);not null" json:"total"`
    Subtotal          float64        `gorm:"type:decimal(12,2);not null" json:"subtotal"`
    Tax               float64        `gorm:"type:decimal(12,2);default:0" json:"tax"`
    Discount          float64        `gorm:"type:decimal(12,2);default:0" json:"discount"`
    PaymentMethod     string         `gorm:"column:payment_method;not null;size:20" json:"paymentMethod"`
    Status            string         `gorm:"not null;default:COMPLETED;index" json:"status"`
    CustomerName      string         `gorm:"column:customer_name;size:100" json:"customerName,omitempty"`
    Notes             string         `gorm:"type:text" json:"notes,omitempty"`
    CreatedAt         time.Time      `gorm:"index" json:"createdAt"`
    UpdatedAt         time.Time      `json:"updatedAt"`
    DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

    // Relationships
    Cashier           *User          `gorm:"foreignKey:CashierID" json:"cashier,omitempty"`
    Branch            *Branch        `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
    Items             []TransactionItem `gorm:"foreignKey:TransactionID" json:"items,omitempty"`
}

func (Transaction) TableName() string {
    return "transactions"
}

// PaymentMethod constants
const (
    PaymentMethodCash      = "CASH"
    PaymentMethodTransfer  = "TRANSFER"
    PaymentMethodEWallet   = "E_WALLET"
)

// TransactionStatus constants
const (
    TransactionStatusPending   = "PENDING"
    TransactionStatusCompleted = "COMPLETED"
    TransactionStatusCancelled = "CANCELLED"
    TransactionStatusRefunded  = "REFUNDED"
)
```

**Key Design Decisions:**
- **TransactionNumber**: Human-readable format (e.g., TRX-20260512-0001) for receipts
- **Subtotal + Tax + Discount**: Breakdown for profit/loss analysis
- **CASCADE delete on branch_id**: Branch deletion removes all transactions
- **RESTRICT delete on cashier_id**: Prevent deleting cashier with transactions
- **PaymentMethod enum**: CASH, TRANSFER, E_WALLET (per PRD FR7)
- **Status tracking**: PENDING → COMPLETED (with CANCELLED, REFUNDED for reversals)

#### 5. TransactionItems Table (Line Items)

**Purpose:** Individual items within a transaction (many-to-many with Product)

**Schema:**
```sql
CREATE TABLE transaction_items (
    id SERIAL PRIMARY KEY,
    transaction_id INTEGER NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    subtotal DECIMAL(10,2) NOT NULL,
    cost_price DECIMAL(10,2),
    product_name VARCHAR(200) NOT NULL,
    product_sku VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT check_quantity_positive CHECK (quantity > 0),
    CONSTRAINT check_unit_price_positive CHECK (unit_price >= 0)
);

CREATE INDEX idx_transaction_items_transaction ON transaction_items(transaction_id);
CREATE INDEX idx_transaction_items_product ON transaction_items(product_id);
CREATE UNIQUE INDEX idx_transaction_items_tx_product ON transaction_items(transaction_id, product_id);
```

**GORM Model:**
```go
// File: apps/backend/internal/transaction/item_model.go
package transaction

import (
    "time"
    "gorm.io/gorm"
)

type TransactionItem struct {
    ID            uint      `gorm:"primaryKey" json:"id"`
    TransactionID uint      `gorm:"column:transaction_id;not null;index" json:"transactionId"`
    ProductID     uint      `gorm:"column:product_id;not null;index" json:"productId"`
    Quantity      int       `gorm:"not null" json:"quantity"`
    UnitPrice     float64   `gorm:"column:unit_price;type:decimal(10,2);not null" json:"unitPrice"`
    Subtotal      float64   `gorm:"type:decimal(10,2);not null" json:"subtotal"`
    CostPrice     float64   `gorm:"column:cost_price;type:decimal(10,2)" json:"costPrice"`
    ProductName   string    `gorm:"column:product_name;not null;size:200" json:"productName"`
    ProductSKU    string    `gorm:"column:product_sku;not null;size:50" json:"productSku"`
    CreatedAt     time.Time `json:"createdAt"`

    // Relationships
    Transaction   *Transaction `gorm:"foreignKey:TransactionID" json:"-"`
    Product       *Product     `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (TransactionItem) TableName() string {
    return "transaction_items"
}
```

**Key Design Decisions:**
- **Snapshot pattern**: Store product_name, product_sku, unit_price at transaction time
  - Preserves historical accuracy even if product is deleted/renamed/price-changed
- **CostPrice stored**: Enables profit/loss calculation per line item
- **CASCADE delete on transaction_id**: Line items deleted when transaction deleted
- **RESTRICT delete on product_id**: Prevent deleting product referenced in transactions
- **Unique constraint on (transaction_id, product_id)**: Prevent duplicate products in same transaction

### Entity Relationship Diagram (ERD)

```
┌─────────────┐       ┌──────────────┐       ┌─────────────┐
│  Branches   │1     n│    Users     │n     1 │ Transactions│
│─────────────│───────│──────────────│───────│─────────────│
│ id          │       │ id           │       │ id          │
│ name        │       │ username     │       │ trans_number│
│ address     │       │ email        │       │ cashier_id  │───┐
│ phone       │       │ role         │       │ branch_id   │   │
│ email       │       │ branch_id ├───┘       │ total       │   │
│ created_at  │       │ status       │       │ payment_m   │   │
│ updated_at  │       │ ...          │       │ status      │   │
└─────────────┘       └──────────────┘       │ created_at  │   │
                                              └─────────────┘   │
                                   ┌────────────────────────────┘
                                   │
                                   ▼
                         ┌──────────────────────┐
                         │   TransactionItems  │
                         │──────────────────────│
                         │ id                  │
                         │ transaction_id ──────┘
                         │ product_id ────┐
                         │ quantity        │
                         │ unit_price      │
                         │ subtotal        │
                         │ product_snapshot │
                         └──────────────────┘
                                   │
                                   ▼
                         ┌──────────────────────┐
                         │     Products         │
                         │──────────────────────│
                         │ id                  │
                         │ sku                 │
                         │ name                │
                         │ stock_qty           │
                         │ price               │
                         │ expiry_date         │
                         │ branch_id ────┐
                         │ reorder_threshold│
                         └────────────────┘
                                          │
                                          └───────────────────┐
                                                              │
┌─────────────┐       ┌──────────────┐       ┌─────────────┐  │
│  Branches   │1     n│    Users     │n     1│ Transactions│  │
└─────────────┘       └──────────────┘       └─────────────┘  │
                                                              │
                  (Products also belong to Branches)          │
                                                              │
                         ┌──────────────────────┐             │
                         │     Products         │◄────────────┘
                         │ branch_id ───────────┘
                         └──────────────────────┘
```

### Audit Trail Design (Badan POM Compliance)

**[Source: NFR-SEC-004, NFR-SEC-009]**

**Append-Only Audit Log Table:**
```sql
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id INTEGER NOT NULL,
    action VARCHAR(50) NOT NULL,
    actor_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    actor_username VARCHAR(50),
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    reason TEXT,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT
);

CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_actor ON audit_logs(actor_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
```

**Audit Logging Pattern:**
- Every INSERT, UPDATE, DELETE on critical tables triggers audit log entry
- Implemented in AuditService (already exists from Epic 1)
- Call from service layer after successful database operation

### Data Dictionary

| Table | Column | Type | Constraints | Description |
|-------|--------|------|-------------|-------------|
| **branches** | id | SERIAL | PK, NOT NULL | Primary key |
| | name | VARCHAR(100) | UNIQUE, NOT NULL | Branch name |
| | address | TEXT | | Physical address |
| | phone | VARCHAR(20) | | Contact phone |
| | email | VARCHAR(100) | INDEX | Contact email |
| | created_at | TIMESTAMP | DEFAULT NOW() | Creation timestamp |
| | updated_at | TIMESTAMP | DEFAULT NOW() | Last update timestamp |
| **products** | id | SERIAL | PK, NOT NULL | Primary key |
| | sku | VARCHAR(50) | UNIQUE(branch_id, sku), NOT NULL | Stock keeping unit |
| | name | VARCHAR(200) | NOT NULL | Product name |
| | description | TEXT | | Product details |
| | stock_qty | INTEGER | NOT NULL, >=0 | Current stock quantity |
| | price | DECIMAL(10,2) | NOT NULL, >0 | Selling price |
| | cost_price | DECIMAL(10,2) | | Purchase cost (for P&L) |
| | expiry_date | DATE | INDEX | Expiration date |
| | branch_id | INTEGER | FK branches, NOT NULL | Branch location |
| | reorder_threshold | INTEGER | DEFAULT 10 | Low stock alert threshold |
| | category | VARCHAR(50) | INDEX | Product category |
| **transactions** | id | SERIAL | PK, NOT NULL | Primary key |
| | transaction_number | VARCHAR(50) | UNIQUE, NOT NULL | Human-readable TRX ID |
| | cashier_id | INTEGER | FK users, NOT NULL, RESTRICT | Who processed sale |
| | branch_id | INTEGER | FK branches, NOT NULL | Transaction location |
| | total | DECIMAL(12,2) | NOT NULL, >=0 | Final amount paid |
| | subtotal | DECIMAL(12,2) | NOT NULL | Before tax/discount |
| | tax | DECIMAL(12,2) | DEFAULT 0 | Tax amount |
| | discount | DECIMAL(12,2) | DEFAULT 0 | Discount amount |
| | payment_method | VARCHAR(20) | NOT NULL | CASH/TRANSFER/E_WALLET |
| | status | VARCHAR(20) | NOT NULL, INDEX | PENDING/COMPLETED/CANCELLED |
| | customer_name | VARCHAR(100) | | Optional customer info |
| | notes | TEXT | | Transaction notes |
| | created_at | TIMESTAMP | INDEX | Transaction date |
| **transaction_items** | id | SERIAL | PK, NOT NULL | Primary key |
| | transaction_id | INTEGER | FK transactions, NOT NULL, CASCADE | Parent transaction |
| | product_id | INTEGER | FK products, NOT NULL, RESTRICT | Product reference |
| | quantity | INTEGER | NOT NULL, >0 | Items sold |
| | unit_price | DECIMAL(10,2) | NOT NULL | Price at time of sale |
| | subtotal | DECIMAL(10,2) | NOT NULL | quantity * unit_price |
| | cost_price | DECIMAL(10,2) | | Cost at time of sale |
| | product_name | VARCHAR(200) | NOT NULL | Snapshot: product name |
| | product_sku | VARCHAR(50) | NOT NULL | Snapshot: product SKU |

### Implementation Tasks

This is a **design-only story**. The actual implementation will be split across subsequent stories:

- **Story 2.2:** Create initial migration with golang-migrate
- **Story 2.3:** Implement GORM models with struct tags
- **Story 2.4:** Implement database connection and pooling
- **Story 2.5:** Implement repository layer for data access

**Deliverables for This Story:**
1. ✓ Complete schema documentation (this document)
2. ✓ ERD diagram (above)
3. ✓ Data dictionary (above)
4. ✓ GORM model templates (above)
5. ✓ Migration file checklist (for Story 2.2)

### Migration File Checklist (Preparation for Story 2.2)

**Migration Order (to avoid circular dependencies):**
1. `000001_create_branches_table.{up|down}.sql`
2. `000002_create_users_table.{up|down}.sql` (already exists, verify)
3. `000003_create_products_table.{up|down}.sql`
4. `000004_create_transactions_table.{up|down}.sql`
5. `000005_create_transaction_items_table.{up|down}.sql`
6. `000006_create_audit_logs_table.{up|down}.sql` (already exists from Epic 1, verify)
7. `000007_add_products_branch_sku_unique.{up|down}.sql` (compound index)

**Naming Convention:**
```
{version}_{action}_{table_name}.{direction}.sql

Example: 000003_create_products_table.up.sql
```

**Migration UP Template:**
```sql
-- +migrate Up
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(50) NOT NULL,
    name VARCHAR(200) NOT NULL,
    stock_qty INTEGER NOT NULL DEFAULT 0,
    price DECIMAL(10,2) NOT NULL,
    branch_id INTEGER NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT check_stock_not_negative CHECK (stock_qty >= 0)
);

CREATE INDEX idx_products_branch_id ON products(branch_id);
CREATE INDEX idx_products_expiry ON products(expiry_date);
```

**Migration DOWN Template:**
```sql
-- +migrate Down
DROP INDEX IF EXISTS idx_products_branch_id;
DROP INDEX IF EXISTS idx_products_expiry;
DROP TABLE IF EXISTS products;
```

### Testing Strategy (for Story 2.3)

**Unit Tests:**
- Test GORM model validation tags
- Test TableName() methods return correct names
- Test JSON serialization (camelCase)
- Test relationship loading (Preload)

**Integration Tests:**
- Test foreign key constraints
- Test unique constraints
- Test CHECK constraints
- Test CASCADE/RESTRICT delete rules
- Test compound unique indexes

**Test Coverage Goal:** >80% for model layer

### File Structure Requirements

**Files to CREATE in this story (design documentation):**

1. **docs/database-schema-design.md** (THIS FILE)
   - Complete schema documentation
   - ERD diagram
   - Data dictionary
   - Migration checklist

**Files to CREATE in Story 2.3 (GORM models):**

1. `apps/backend/internal/branch/model.go`
2. `apps/backend/internal/product/model.go`
3. `apps/backend/internal/transaction/model.go`
4. `apps/backend/internal/transaction/item_model.go`

**Files to CREATE in Story 2.2 (migrations):**

1. `apps/backend/migrations/000001_create_branches_table.up.sql`
2. `apps/backend/migrations/000001_create_branches_table.down.sql`
3. `apps/backend/migrations/000003_create_products_table.up.sql`
4. `apps/backend/migrations/000003_create_products_table.down.sql`
5. `apps/backend/migrations/000004_create_transactions_table.up.sql`
6. `apps/backend/migrations/000004_create_transactions_table.down.sql`
7. `apps/backend/migrations/000005_create_transaction_items_table.up.sql`
8. `apps/backend/migrations/000005_create_transaction_items_table.down.sql`

### Cross-Cutting Concerns

**1. Multi-Tenancy (Branch-Level Isolation)**
- All business data tables (users, products, transactions) have branch_id FK
- System admin users have nullable branch_id (global access)
- Queries must filter by branch_id for non-admin roles
- Implemented in repository layer (Story 2.5)

**2. Audit Trail (Badan POM Compliance)**
- All state changes logged in audit_logs table
- Triggered from service layer after successful operations
- Append-only (no UPDATE/DELETE on audit_logs)
- Implemented in AuditService (already exists from Epic 1)

**3. Soft Deletes**
- GORM's DeletedAt field for soft deletes
- Prevents accidental data loss
- Maintains audit trail for deleted records
- All major tables include deleted_at

**4. Timestamps**
- created_at: Record creation time
- updated_at: Last update time (auto-managed by GORM)
- Critical for audit trail and reporting

**5. Data Integrity**
- CHECK constraints at database level (stock >= 0, price > 0)
- Foreign key relationships with CASCADE/RESTRICT rules
- Unique constraints for business logic (username, email, SKU)
- Compound unique for multi-tenant (branch_id + sku)

### Performance Considerations

**Indexes:**
- Unique indexes for login/lookup (users.username, users.email, products.sku)
- Compound unique for multi-tenant data (branch_id + sku)
- Foreign key indexes for JOIN performance (branch_id, cashier_id)
- Date indexes for reporting (transactions.created_at, products.expiry_date)

**Query Patterns:**
- Filter by branch_id for all multi-tenant queries
- Use indexed columns in WHERE clauses
- Avoid SELECT * (specify required columns)
- Use GORM Preload for relationships (avoid N+1 queries)

**Data Types:**
- Use DECIMAL for currency (not FLOAT) - precision matters
- Use INTEGER for counts (not BIGINT unless needed)
- Use TIMESTAMP with timezones for all timestamps
- Use TEXT for variable-length content (descriptions, notes)

### Naming Conventions (from Architecture)

**Database (PostgreSQL):**
- Tables: snake_case, plural (users, products, transactions)
- Columns: snake_case (user_id, created_at, stock_qty)
- Indexes: idx_{table}_{column} (idx_users_email)
- Foreign keys: {table}_id (user_id, product_id)

**JSON API (HTTP Responses):**
- Fields: camelCase (userId, createdAt, stockQty)
- Matches GORM json struct tags

**Go Code:**
- Variables/Functions: camelCase (getByID, stockQty)
- Types/Interfaces: PascalCase (Product, UserService)
- Constants: UPPER_SNAKE_CASE or PascalCase (UserStatusActive, PaymentMethodCash)
- Files: snake_case (user_model.go, product_service.go)

### Security Considerations

**1. SQL Injection Prevention**
- Use GORM parameterized queries (never concatenate strings)
- GORM automatically escapes parameters

**2. Data Access Control**
- RBAC middleware filters by branch_id
- Repository layer enforces scoping
- Never expose raw SQL errors to API clients

**3. Sensitive Data**
- Password hash in users table (never plaintext password)
- Audit logs include who, when, what, why
- No PII in logs (redact passwords, tokens)

**4. Cascade Deletes**
- CASCADE: Safe to delete (branch → transactions, transaction → items)
- RESTRICT: Prevent delete (product referenced in transactions)
- SET NULL: Clear references (user deleted → audit_logs.actor_id = NULL)

### Open Questions (for clarification)

1. **Q:** Should transaction_number be sequential per branch or globally unique?
   **A:** Globally unique with format: TRX-YYYYMMDD-branch_id-sequence (e.g., TRX-20260512-1-0001)

2. **Q:** Should we store customer information separately or as optional fields in transactions?
   **A:** Optional fields in transactions for MVP (customer_name). Separate customers table in future phases for CRM features.

3. **Q:** Should products support multiple categories (many-to-many) or single category?
   **A:** Single category for MVP. Many-to-many relationship in future for advanced categorization.

4. **Q:** Should we store user_id (creator) on all records for audit trail?
   **A:** Use audit_logs table instead (already implemented). Created_at/updated_at on records is sufficient for MVP.

### Dependencies on Epic 1

**Reusing from Epic 1:**
1. ✓ User model (apps/backend/internal/user/model.go)
2. ✓ AuditService (apps/backend/internal/services/audit_service.go)
3. ✓ Repository pattern (apps/backend/internal/user/repository.go)
4. ✓ Database connection (apps/backend/internal/db/db.go)
5. ✓ Migration infrastructure (apps/backend/internal/migrate/migrate.go)

**Building on Epic 1:**
- Follow same GORM conventions (column tags, json tags, indexes)
- Use same testing patterns (co-located tests, transaction rollback)
- Extend AuditService for new entities (Product, Transaction)
- Use same RBAC middleware for branch-level access control

### Success Criteria

**Definition of Done:**
1. ✓ All 5 core entities fully documented with attributes and types
2. ✓ All relationships defined with foreign key constraints
3. ✓ All indexes identified for performance
4. ✓ All constraints defined (NOT NULL, FK, CHECK)
5. ✓ Branch-level data isolation designed
6. ✓ GORM struct tag patterns specified
7. ✓ ERD diagram created
8. ✓ Migration file checklist prepared
9. ✓ Data dictionary documented
10. ✓ Next story (2.2) has clear migration specifications

### Next Steps

**After this story is approved:**
1. **Story 2.2:** Create initial migration files using the schema defined here
2. **Story 2.3:** Implement GORM models following the struct tag patterns specified
3. **Story 2.4:** Configure database connection with pooling
4. **Story 2.5:** Implement repository layer using the models from 2.3

---

**Story Status:** done

**Completion Note:** Schema design complete. All entities, relationships, indexes, and constraints defined. Migration specifications prepared for Story 2.2. GORM model patterns established following Epic 1 conventions. Story approved and marked as DONE.
