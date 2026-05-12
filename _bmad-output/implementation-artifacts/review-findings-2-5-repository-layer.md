# Code Review Findings: Story 2-5 - Repository Layer

**Review Date:** 2026-05-13
**Story:** 2-5-implement-repository-layer-for-data-access
**Status:** Requires Action

---

## Summary Statistics
- **Total Findings:** 18 (after deduplication)
- **Decision Needed:** 2
- **Patch Required:** 14
- **Deferred:** 2
- **Dismissed:** 15

---

## Decision Needed (2)

### D-001: User Repository Domain Package Choice
**Source:** auditor
**Severity:** MEDIUM
**Location:** `user_repository_impl.go:11`

**Issue:** User repository imports `internal/user` package instead of `internal/models` like other repositories. The spec mentions repositories should work with entities, but User model from Epic 1 is in a different package.

**Options:**
1. Keep using `internal/user.User` (domain-driven design approach - User is an aggregate root)
2. Migrate User model to `internal/models` for consistency
3. Document why User is treated differently (domain entity vs data entity)

**Recommendation:** Option 1 - Keep using `internal/user.User` as it represents proper DDD (User is an aggregate root with business logic). Update documentation to clarify this pattern.

---

### D-002: Repository File Organization Pattern
**Source:** auditor
**Severity:** LOW
**Location:** `branch_repository.go`, `product_repository_impl.go`

**Issue:** Inconsistent file organization - `branch_repository.go` contains both interface and implementation, while other repositories split them into separate `*_impl.go` files.

**Options:**
1. Split `branch_repository.go` into interface and implementation files (consistent pattern)
2. Merge all repository interfaces and implementations into single files (simpler pattern)
3. Accept inconsistency (documentation burden)

**Recommendation:** Option 1 - Split for consistency and easier navigation in larger codebases.

---

## Patch Required (14)

### P-001: SQL Injection in ORDER BY Clause [CRITICAL SECURITY]
**Source:** blind+edge
**Severity:** HIGH
**AC Violated:** AC3 (Error handling and wrapping - missing input validation)
**Location:** All repository files - `branch_repository.go:156`, `product_repository_impl.go:147`, `transaction_repository_impl.go:128`, `transaction_item_repository_impl.go:106`, `user_repository_impl.go:145`

**Issue:** Direct string concatenation of user input into SQL ORDER BY clause without validation creates SQL injection vulnerability.

**Attack Vector:**
- `SortBy = "id; DROP TABLE branches; --"` → `ORDER BY id; DROP TABLE branches; -- DESC`

**Patch:**
```go
// Add to each repository's List method
allowedSortFields := map[string]bool{"id": true, "name": true, "created_at": true, "updated_at": true, "status": true, "total": true}
if filter.SortBy != "" {
    if !allowedSortFields[filter.SortBy] {
        sortBy = "created_at" // safe default
    } else {
        sortBy = filter.SortBy
    }
}
allowedSortOrders := map[string]bool{"ASC": true, "DESC": true, "asc": true, "desc": true, "": true}
if filter.SortOrder != "" {
    if !allowedSortOrders[filter.SortOrder] {
        sortOrder = "DESC" // safe default
    } else {
        sortOrder = strings.ToUpper(filter.SortOrder)
    }
}
query = query.Order(sortBy + " " + sortOrder)
```

---

### P-002: Race Condition in UpdateStock Method
**Source:** blind+edge
**Severity:** HIGH
**AC Violated:** AC3 (Error handling)
**Location:** `product_repository_impl.go:80-85`

**Issue:** `UpdateStock` sets absolute value instead of incrementing/decrementing, causing race conditions in concurrent transactions.

**Race Scenario:**
1. Transaction A reads stock: 10
2. Transaction B reads stock: 10
3. Both set to 9 (should be 8)

**Patch:**
```go
func (r *productRepository) UpdateStock(ctx context.Context, id uint, delta int64) error {
    // Use atomic increment/decrement with check for negative stock
    err := r.db.WithContext(ctx).Model(&models.Product{}).
        Where("id = ? AND stock_qty + ? >= 0", id, delta).
        Update("stock_qty", gorm.Expr("stock_qty + ?", delta)).Error
    if err != nil {
        return fmt.Errorf("failed to update product stock: %w", err)
    }
    return nil
}
```

---

### P-003: Unbounded Query Results (DoS Risk)
**Source:** blind+edge
**Severity:** MEDIUM
**AC Violated:** AC4 (Pagination support)
**Location:** All repository List methods

**Issue:** When Page/Limit are 0 or unset, all records are returned without limit, causing memory exhaustion and performance issues.

**Patch:**
```go
// Add at the beginning of each List method
if filter.Page < 1 {
    filter.Page = 1
}
if filter.Limit < 1 {
    filter.Limit = 20 // default
}
if filter.Limit > 1000 {
    filter.Limit = 1000 // maximum to prevent DoS
}
```

---

### P-004: Delete Operations Don't Check RowsAffected
**Source:** blind+edge
**Severity:** MEDIUM
**AC Violated:** AC3 (Error handling - not found distinction)
**Location:** All repository Delete methods

**Issue:** GORM Delete succeeds silently on non-existent records (soft delete). Can't distinguish successful delete from attempting to delete non-existent record.

**Patch:**
```go
func (r *branchRepository) Delete(ctx context.Context, id uint) error {
    result := r.db.WithContext(ctx).Delete(&models.Branch{}, id)
    if result.Error != nil {
        return fmt.Errorf("failed to delete branch: %w", result.Error)
    }
    if result.RowsAffected == 0 {
        return ErrNotFound
    }
    return nil
}
```

---

### P-005: Missing Context Cancellation Checks
**Source:** blind+edge
**Severity:** MEDIUM
**AC Violated:** AC2 (GORM-based implementation - should respect context)
**Location:** All repository methods

**Issue:** No explicit context cancellation checks before expensive database operations. Cancelled requests continue executing.

**Patch:**
```go
// Add at the beginning of expensive methods
func (r *productRepository) List(ctx context.Context, filter *ProductFilter) ([]*models.Product, int64, error) {
    select {
    case <-ctx.Done():
        return nil, 0, fmt.Errorf("context cancelled: %w", ctx.Err())
    default:
    }
    // rest of implementation
}
```

---

### P-006: Integer Overflow in Pagination Offset
**Source:** edge
**Severity:** HIGH
**AC Violated:** AC4 (Pagination)
**Location:** All repository List methods with offset calculation

**Issue:** `(filter.Page - 1) * filter.Limit` can overflow with very large page numbers.

**Patch:**
```go
if filter.Page > 1000000 { // reasonable maximum
    return nil, 0, fmt.Errorf("page number exceeds maximum allowed")
}
if filter.Limit > 1000 {
    return nil, 0, fmt.Errorf("limit exceeds maximum allowed")
}
```

---

### P-007: Missing Nil Filter Handling
**Source:** edge
**Severity:** MEDIUM
**AC Violated:** AC3 (Error handling)
**Location:** All repository List methods

**Issue:** No check if filter parameter is nil before accessing its fields.

**Patch:**
```go
func (r *productRepository) List(ctx context.Context, filter *ProductFilter) ([]*models.Product, int64, error) {
    if filter == nil {
        filter = &ProductFilter{}
    }
    // rest of implementation
}
```

---

### P-008: Special Character Handling in Search Queries
**Source:** blind+edge
**Severity:** MEDIUM
**AC Violated:** AC3 (Error handling - missing input sanitization)
**Location:** `branch_repository.go:131-134`, `product_repository_impl.go:121-124`

**Issue:** Special characters (`%`, `_`) in search queries act as wildcards, causing unintended matches and potential DoS.

**Patch:**
```go
if filter.SearchQuery != "" {
    // Remove wildcard characters and limit length
    search := strings.ReplaceAll(filter.SearchQuery, "%", "")
    search = strings.ReplaceAll(search, "_", "")
    search = strings.ReplaceAll(search, "\\", "")
    if len(search) > 100 {
        search = search[:100]
    }
    query = query.Where("name LIKE ? OR address LIKE ?",
        "%"+search+"%", "%"+search+"%")
}
```

---

### P-009: Time Zone Inconsistency in Date Queries
**Source:** edge
**Severity:** MEDIUM
**AC Violated:** AC4 (Complex query support - date filtering)
**Location:** `transaction_repository_impl.go:141-142`, `transaction_repository_impl.go:177-178`

**Issue:** Using `date.Location()` instead of UTC causes boundary issues when DB timezone differs.

**Patch:**
```go
startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
endOfDay := startOfDay.Add(24 * time.Hour)
```

---

### P-010: Empty Items Validation in CreateWithItems
**Source:** edge
**Severity:** MEDIUM
**AC Violated:** AC3 (Error handling)
**Location:** `transaction_repository_impl.go:210-227`

**Issue:** Allows creating transactions without items, violating business logic.

**Patch:**
```go
func (r *transactionRepository) CreateWithItems(ctx context.Context, transaction *models.Transaction, items []*models.TransactionItem) error {
    if len(items) == 0 {
        return fmt.Errorf("transaction must have at least one item")
    }
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // rest of implementation
    })
}
```

---

### P-011: GetByID Missing Zero ID Validation
**Source:** edge
**Severity:** LOW
**AC Violated:** AC3 (Error handling)
**Location:** All repository GetByID methods

**Issue:** ID=0 (zero value for uint) may cause unexpected database behavior.

**Patch:**
```go
func (r *branchRepository) GetByID(ctx context.Context, id uint) (*models.Branch, error) {
    if id == 0 {
        return nil, ErrInvalidInput
    }
    // rest of implementation
}
```

---

### P-012: Deactivate User Doesn't Check Existence
**Source:** edge
**Severity:** LOW
**AC Violated:** AC3 (Error handling)
**Location:** `user_repository_impl.go:88-102`

**Issue:** Deactivating non-existent user returns no error (false success).

**Patch:**
```go
func (r *userRepository) Deactivate(ctx context.Context, id uint, reason string, deactivatedBy uint) error {
    result := r.db.WithContext(ctx).Model(&user.User{}).
        Where("id = ?", id).
        Updates(map[string]interface{}{
            "status":               user.UserStatusInactive,
            "deactivated_at":       &now,
            "deactivated_by":       &deactivatedBy,
            "deactivation_reason":  reason,
        })
    if result.Error != nil {
        return fmt.Errorf("failed to deactivate user: %w", result.Error)
    }
    if result.RowsAffected == 0 {
        return ErrNotFound
    }
    return nil
}
```

---

### P-013: Missing Eager Loading in Product Repository
**Source:** auditor
**Severity:** MEDIUM
**AC Violated:** AC4 (Support eager loading of relationships)
**Location:** `product_repository_impl.go:44` (GetByID method)

**Issue:** Product repository doesn't preload Branch relationship despite foreign key, unlike Transaction repository which preloads TransactionItems.

**Patch:**
```go
func (r *productRepository) GetByID(ctx context.Context, id uint) (*models.Product, error) {
    var product models.Product
    err := r.db.WithContext(ctx).Preload("Branch").First(&product, id).Error
    // rest of implementation
}
```

---

### P-014: Error Wrapping Missing in CreateWithItems
**Source:** auditor
**Severity:** LOW
**AC Violated:** AC3 (Use fmt.Errorf with %w for error wrapping)
**Location:** `transaction_repository_impl.go:211-227`

**Issue:** Transaction function returns raw GORM errors without descriptive wrapping.

**Patch:**
```go
func (r *transactionRepository) CreateWithItems(ctx context.Context, transaction *models.Transaction, items []*models.TransactionItem) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(transaction).Error; err != nil {
            return fmt.Errorf("failed to create transaction: %w", err)
        }
        for _, item := range items {
            item.TransactionID = transaction.ID
            if err := tx.Create(item).Error; err != nil {
                return fmt.Errorf("failed to create transaction item: %w", err)
            }
        }
        return nil
    })
}
```

---

## Deferred (2)

### F-001: Dependency Injection Pattern Not Demonstrated
**Source:** auditor
**Severity:** MEDIUM
**Reason:** Pre-existing architectural pattern. Repository interfaces exist and follow DI pattern correctly. Service layer integration (which would demonstrate the pattern usage) hasn't been implemented yet and is outside the scope of this story.

**Note:** The `NewRepositories()` factory function in `repository.go` demonstrates the DI pattern correctly. Services will use this in future stories.

---

### F-002: No Mock Repositories Provided
**Source:** auditor
**Severity:** LOW
**Reason:** While AC5 mentions "enables testability with mock repositories", the current tests use SQLite in-memory databases which is a valid testing strategy. Creating mock implementations can be deferred until service layer development when they're actually needed.

**Note:** The repository interfaces are properly designed to enable mocking when needed.

---

## Dismissed (15)

1. **Duplicate error handling code** (blind) - Implementation is acceptable; errors are defined once and reused.
2. **Inconsistent naming - user parameter** (blind) - Using `user` as parameter name despite package name `user` is idiomatic Go and works fine with qualified names like `user.User`.
3. **Unsafe type assertion with panic** (blind) - Panic during application startup on DI misconfiguration is appropriate; it fails fast and clearly.
4. **Shared global state in tests** (blind) - Counter implementation is acceptable for simple tests; not a production issue.
5. **N+1 query problem** (blind) - Preload is correctly used; no N+1 issue exists.
6. **Generic error types don't support errors.Is** (auditor) - Custom RepositoryError type is acceptable for domain-specific errors; can use errors.Is() with the exported error variables.
7. **Transaction isolation level** (blind) - GORM default isolation is acceptable for this use case; explicit isolation not required by spec.
8. **No input validation in filter parameters** (blind) - Covered by P-003, P-006, P-007, P-008.
9. **Missing error context in batch operations** (edge) - Error messages are descriptive enough for batch operations.
10. **No input validation on Create operations** (edge) - GORM handles validation via struct tags; database constraints will catch invalid data.
11. **Potential panic in Deactivate with nil parameters** (edge) - Covered by P-007.
12. **Incomplete pagination validation** (auditor) - Covered by P-003.
13. **RepositoryError type doesn't support error chains** (auditor) - Errors are properly wrapped with %w; RepositoryError is for domain errors.
14. **User repository uses different package** (auditor) - Deferred as D-001 for decision.
15. **Branch repository inconsistent file organization** (auditor) - Deferred as D-002 for decision.
