-- Story 5.4: Implement Append-Only Audit Trail for Compliance
-- DOWN Migration: Drop audit_logs table
-- WARNING: Dropping audit_logs table will lose all compliance data
-- Only run this migration if you have proper backup and understand the implications

DROP INDEX IF EXISTS idx_audit_logs_user_timestamp;
DROP INDEX IF EXISTS idx_audit_logs_timestamp;
DROP INDEX IF EXISTS idx_audit_logs_action;
DROP INDEX IF EXISTS idx_audit_logs_user_id;
DROP TABLE IF EXISTS audit_logs;
