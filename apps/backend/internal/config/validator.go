package config

import (
	"fmt"
)

func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is required - generate with: make generate-jwt-secret")
	}

	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf(
			"JWT_SECRET must be at least 32 characters (current: %d)\nGenerate secure secret: make generate-jwt-secret",
			len(c.JWT.Secret),
		)
	}

	if c.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}

	// Story 2.4: Validate database connection pooling configuration
	if c.Database.MaxOpenConns <= 0 {
		// Set default if not configured
		c.Database.MaxOpenConns = 25
	}
	if c.Database.MaxIdleConns < 0 {
		// Set default if not configured
		c.Database.MaxIdleConns = 5
	}
	if c.Database.ConnMaxLifetime <= 0 {
		// Set default if not configured (5 minutes)
		c.Database.ConnMaxLifetime = 5 * 60 * 1000000000 // 5 minutes in nanoseconds (time.Duration)
	}

	// Validate pooling constraints
	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return fmt.Errorf("database.max_idle_conns (%d) cannot exceed database.max_open_conns (%d)",
			c.Database.MaxIdleConns, c.Database.MaxOpenConns)
	}
	if c.Database.MaxOpenConns > 100 {
		return fmt.Errorf("database.max_open_conns (%d) exceeds maximum recommended value (100)",
			c.Database.MaxOpenConns)
	}
	// Add minimum validation to prevent bottlenecks
	if c.Database.MaxOpenConns > 0 && c.Database.MaxOpenConns < 2 {
		return fmt.Errorf("database.max_open_conns (%d) must be at least 2 for proper operation",
			c.Database.MaxOpenConns)
	}
	if c.Database.MaxIdleConns > 0 && c.Database.MaxIdleConns < 1 {
		return fmt.Errorf("database.max_idle_conns (%d) must be at least 1 to maintain warm connections",
			c.Database.MaxIdleConns)
	}
	// Add bounds validation for ConnMaxLifetime (max 24 hours)
	if c.Database.ConnMaxLifetime > 86400*1000000000 {
		return fmt.Errorf("database.conn_max_lifetime (%d) exceeds maximum recommended value (24 hours)",
			c.Database.ConnMaxLifetime/1000000000)
	}

	// Story 2.4: Validate SSLMode against allowed values
	validSSLModes := map[string]bool{
		"disable":     true,
		"allow":       true,
		"prefer":      true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	if c.Database.SSLMode != "" && !validSSLModes[c.Database.SSLMode] {
		return fmt.Errorf("database.sslmode '%s' is invalid (allowed: disable, allow, prefer, require, verify-ca, verify-full)",
			c.Database.SSLMode)
	}

	if c.Server.ReadTimeout < 0 {
		return fmt.Errorf("server.readtimeout must be non-negative")
	}

	if c.Server.WriteTimeout < 0 {
		return fmt.Errorf("server.writetimeout must be non-negative")
	}

	if c.Server.IdleTimeout < 0 {
		return fmt.Errorf("server.idletimeout must be non-negative")
	}

	if c.Server.ShutdownTimeout < 0 {
		return fmt.Errorf("server.shutdowntimeout must be non-negative")
	}

	if c.Server.MaxHeaderBytes < 0 {
		return fmt.Errorf("server.maxheaderbytes must be non-negative")
	}

	if c.App.Environment == "production" {
		if c.Database.Password == "" {
			return fmt.Errorf("database.password is required in production")
		}

		if c.Database.SSLMode == "disable" {
			return fmt.Errorf("database SSL mode cannot be 'disable' in production")
		}
	}

	return nil
}
