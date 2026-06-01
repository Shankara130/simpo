-- +migrate Up
-- Create supplier_audit_trail table for Badan POM compliance
-- This table maintains append-only audit trail for all supplier transactions

CREATE TABLE supplier_audit_trail (
    id BIGSERIAL PRIMARY KEY,
    transaction_type VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    user_role VARCHAR(50) NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    action_description TEXT NOT NULL,
    reason TEXT NULL,
    transaction_amount DECIMAL(15,2) NULL,
    affected_items_count INT DEFAULT 0,
    ip_address VARCHAR(45) NULL,
    user_agent VARCHAR(255) NULL,
    branch_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Create indexes for efficient querying
CREATE INDEX idx_supplier_audit_entity ON supplier_audit_trail(entity_type, entity_id);
CREATE INDEX idx_supplier_audit_user ON supplier_audit_trail(user_id, created_at);
CREATE INDEX idx_supplier_audit_date ON supplier_audit_trail(created_at);
CREATE INDEX idx_supplier_audit_branch ON supplier_audit_trail(branch_id, created_at);

-- Add foreign key constraints (will be added after referenced tables exist)
-- ALTER TABLE supplier_audit_trail
--     ADD CONSTRAINT fk_supplier_audit_user
--     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- ALTER TABLE supplier_audit_trail
--     ADD CONSTRAINT fk_supplier_audit_branch
--     FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE;

-- Add table comment
COMMENT ON TABLE supplier_audit_trail IS 'Append-only audit trail for all supplier transactions for Badan POM compliance';

-- Add column comments
COMMENT ON COLUMN supplier_audit_trail.transaction_type IS 'Type of transaction (supplier_operation, purchase_invoice, goods_receipt, payment, return)';
COMMENT ON COLUMN supplier_audit_trail.entity_type IS 'Type of entity affected (supplier, purchase_invoice, supplier_payment)';
COMMENT ON COLUMN supplier_audit_trail.entity_id IS 'ID of affected entity';
COMMENT ON COLUMN supplier_audit_trail.user_id IS 'User who performed the action';
COMMENT ON COLUMN supplier_audit_trail.user_role IS 'Role of user at time of action';
COMMENT ON COLUMN supplier_audit_trail.action_type IS 'Type of action (create, update, delete, receive, pay)';
COMMENT ON COLUMN supplier_audit_trail.action_description IS 'Human-readable description of action';
COMMENT ON COLUMN supplier_audit_trail.reason IS 'Reason for action (if applicable)';
COMMENT ON COLUMN supplier_audit_trail.transaction_amount IS 'Monetary amount if applicable';
COMMENT ON COLUMN supplier_audit_trail.affected_items_count IS 'Number of items affected';
COMMENT ON COLUMN supplier_audit_trail.ip_address IS 'Client IP address';
COMMENT ON COLUMN supplier_audit_trail.user_agent IS 'Client user agent';
COMMENT ON COLUMN supplier_audit_trail.branch_id IS 'Branch where action occurred';
COMMENT ON COLUMN supplier_audit_trail.created_at IS 'When the action occurred';
