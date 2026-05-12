-- Story 1.10: Add user deactivation tracking fields
-- These fields track when and by whom a user was deactivated, along with the reason

-- Add deactivation timestamp
ALTER TABLE users ADD COLUMN deactivated_at TIMESTAMP;

-- Add foreign key reference to admin who performed deactivation
ALTER TABLE users ADD COLUMN deactivated_by INTEGER;

-- Add reason for deactivation (e.g., "Staff resignation", "Termination", "Contract ended")
ALTER TABLE users ADD COLUMN deactivation_reason TEXT;

-- Add foreign key constraint for deactivated_by
ALTER TABLE users ADD CONSTRAINT fk_users_deactivated_by
    FOREIGN KEY (deactivated_by) REFERENCES users(id) ON DELETE SET NULL;

-- Create index for deactivated users to optimize queries
CREATE INDEX idx_users_deactivated_at ON users(deactivated_at) WHERE deactivated_at IS NOT NULL;
