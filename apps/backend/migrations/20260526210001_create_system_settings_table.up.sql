-- Migration: create_system_settings_table
-- Description: Creates system_settings table for storing pharmacy configuration
-- Related to Story: 6.1 - Implement System Settings Configuration

BEGIN;

CREATE TABLE IF NOT EXISTS system_settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    description VARCHAR(255),
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    updated_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_system_settings_key ON system_settings(key);
CREATE INDEX IF NOT EXISTS idx_system_settings_updated_by ON system_settings(updated_by);

COMMENT ON TABLE system_settings IS 'System configuration settings for pharmacy';
COMMENT ON COLUMN system_settings.id IS 'Primary key';
COMMENT ON COLUMN system_settings.key IS 'Setting key (e.g., pharmacy.name, pharmacy.address)';
COMMENT ON COLUMN system_settings.value IS 'Setting value';
COMMENT ON COLUMN system_settings.description IS 'Setting description for UI display';
COMMENT ON COLUMN system_settings.created_by IS 'User who created this setting';
COMMENT ON COLUMN system_settings.updated_by IS 'User who last updated this setting';
COMMENT ON COLUMN system_settings.created_at IS 'Timestamp when setting was created';
COMMENT ON COLUMN system_settings.updated_at IS 'Timestamp when setting was last updated';

-- Insert default settings
INSERT INTO system_settings (key, value, description) VALUES
    ('pharmacy.name', 'Simpo Pharmacy', 'Pharmacy business name'),
    ('pharmacy.address', '', 'Pharmacy street address'),
    ('pharmacy.phone', '', 'Pharmacy phone number'),
    ('pharmacy.email', '', 'Pharmacy email address'),
    ('pharmacy.logo_url', '', 'Pharmacy logo URL (future enhancement)');

COMMIT;
