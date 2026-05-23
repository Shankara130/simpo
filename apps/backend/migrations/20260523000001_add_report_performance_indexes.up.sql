-- Migration: add_report_performance_indexes
-- Description: Add composite indexes for report performance optimization
-- Story 5.1, Task 2.6, AC3: Performance optimization with indexes for <10s report generation

BEGIN;

-- Composite index on (created_at, branch_id) for combined date and branch filtering
-- This optimizes queries that filter by both date range and branch
-- Story 5.1, AC1, AC2: Daily sales summary with date and optional branch filter
CREATE INDEX IF NOT EXISTS idx_transactions_created_branch
    ON transactions(created_at, branch_id)
    WHERE deleted_at IS NULL;

-- Note: Individual indexes already exist from previous migrations:
-- - idx_transactions_created_at (created_at) exists
-- - idx_transactions_branch (branch_id) exists
-- - idx_transaction_items_product (product_id) exists
--
-- This composite index complements the existing indexes for optimal query performance
-- when both date and branch filters are used together.

-- Documentation
COMMENT ON INDEX idx_transactions_created_branch IS
    'Composite index for date+branch filtering in reports (Story 5.1)';

COMMIT;
