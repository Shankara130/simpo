-- Story 5.4: Implement Append-Only Audit Trail for Compliance
-- Migration: Create audit_logs table for Badan POM compliance
-- Per NFR-SEC-004: Append-only audit trail with user identification, timestamp, and reason
-- Per NFR-SEC-009: 5-year minimum data retention for Badan POM compliance

-- Create audit_logs table
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    username VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    ip_address VARCHAR(45),
    outcome VARCHAR(50) NOT NULL,
    reason TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for query performance
-- Story 5.4, Task 2.3: Indexes on user_id, action, timestamp for query performance
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);

-- Story 5.4, Task 2.4: Index on timestamp for 5-year retention queries
CREATE INDEX idx_audit_logs_user_timestamp ON audit_logs(user_id, timestamp);

-- Story 5.4, Task 2.6: Set up database role permissions
-- IMPORTANT: Do NOT grant UPDATE or DELETE permissions to enforce append-only behavior
-- This enforces append-only behavior at database level for Badan POM compliance
-- Note: Role-based permissions will be set up during deployment
-- For development, the application user has full access
-- In production, restrict to INSERT only (write-only) and SELECT only (read-only for admins)

-- Comment for production deployment:
-- GRANT INSERT ON audit_logs TO simpo_app; -- Application role (write-only)
-- GRANT SELECT ON audit_logs TO simpo_admin; -- Admin role (read-only for compliance)
-- REVOKE UPDATE, DELETE ON audit_logs FROM simpo_app; -- Enforce append-only at database level
