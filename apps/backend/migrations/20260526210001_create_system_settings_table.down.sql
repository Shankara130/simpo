-- Migration: create_system_settings_table
-- Description: Rolls back system_settings table creation

BEGIN;

DROP TABLE IF EXISTS system_settings;

COMMIT;
