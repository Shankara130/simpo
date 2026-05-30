package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// BackupConfig represents backup configuration
// Story 6.3, AC8: Configurable backup schedule and retention
type BackupConfig struct {
	Schedule      string `mapstructure:"schedule" yaml:"schedule"`             // Cron expression (default: "0 2 * * *")
	RetentionDays int    `mapstructure:"retention_days" yaml:"retention_days"` // Retention period in days (default: 30)
	StoragePath   string `mapstructure:"storage_path" yaml:"storage_path"`     // Backup storage path (default: "/backups")
	Enabled       bool   `mapstructure:"enabled" yaml:"enabled"`               // Enable/disable automated backups (default: true)
}

// CorsConfig represents CORS configuration
// Story 9.4: Environment-based CORS configuration for cross-origin requests
type CorsConfig struct {
	Enabled         bool     `mapstructure:"enabled" yaml:"enabled"`                     // Enable/disable CORS middleware
	AllowedOrigins   []string `mapstructure:"allowed_origins" yaml:"allowed_origins"`   // List of allowed origins (not wildcard for security)
	AllowCredentials bool     `mapstructure:"allow_credentials" yaml:"allow_credentials"` // Allow credentials (cookies, auth headers)
	AllowedMethods   []string `mapstructure:"allowed_methods" yaml:"allowed_methods"`     // Allowed HTTP methods
	AllowedHeaders   []string `mapstructure:"allowed_headers" yaml:"allowed_headers"`     // Allowed request headers
	MaxAge           int      `mapstructure:"max_age" yaml:"max_age"`                     // Pre-flight cache duration in seconds
}

type Config struct {
	App        AppConfig        `mapstructure:"app" yaml:"app"`
	Database   DatabaseConfig   `mapstructure:"database" yaml:"database"`
	JWT        JWTConfig        `mapstructure:"jwt" yaml:"jwt"`
	Server     ServerConfig     `mapstructure:"server" yaml:"server"`
	Logging    LoggingConfig    `mapstructure:"logging" yaml:"logging"`
	Ratelimit  RateLimitConfig  `mapstructure:"ratelimit" yaml:"ratelimit"`
	Migrations MigrationsConfig `mapstructure:"migrations" yaml:"migrations"`
	Health     HealthConfig     `mapstructure:"health" yaml:"health"`
	Redis      RedisConfig      `mapstructure:"redis" yaml:"redis"`
	Backup     BackupConfig     `mapstructure:"backup" yaml:"backup"`
	Cors       CorsConfig       `mapstructure:"cors" yaml:"cors"`
}

type AppConfig struct {
	Name        string `mapstructure:"name" yaml:"name"`
	Version     string `mapstructure:"version" yaml:"version"`
	Environment string `mapstructure:"environment" yaml:"environment"`
	Debug       bool   `mapstructure:"debug" yaml:"debug"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host" yaml:"host"`
	Port     int    `mapstructure:"port" yaml:"port"`
	User     string `mapstructure:"user" yaml:"user"`
	Password string `mapstructure:"password" yaml:"password"`
	Name     string `mapstructure:"name" yaml:"name"`
	SSLMode  string `mapstructure:"sslmode" yaml:"sslmode"`

	// Connection pooling configuration (Story 2.4)
	// MaxOpenConns: Maximum open connections to database (default: 25)
	// MaxIdleConns: Maximum idle connections in pool (default: 5)
	// ConnMaxLifetime: Maximum time a connection can be reused (default: 5m)
	MaxOpenConns    int           `mapstructure:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime"`
}

type JWTConfig struct {
	Secret          string        `mapstructure:"secret" yaml:"secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl" yaml:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl" yaml:"refresh_token_ttl"`
	TTLHours        int           `mapstructure:"ttlhours" yaml:"ttlhours"` // Deprecated: kept for backward compatibility
}

type ServerConfig struct {
	Port            string `mapstructure:"port" yaml:"port"`
	ReadTimeout     int    `mapstructure:"readtimeout" yaml:"readtimeout"`
	WriteTimeout    int    `mapstructure:"writetimeout" yaml:"writetimeout"`
	IdleTimeout     int    `mapstructure:"idletimeout" yaml:"idletimeout"`
	ShutdownTimeout int    `mapstructure:"shutdowntimeout" yaml:"shutdowntimeout"`
	MaxHeaderBytes  int    `mapstructure:"maxheaderbytes" yaml:"maxheaderbytes"`
}

type LoggingConfig struct {
	Level         string   `mapstructure:"level" yaml:"level"`
	RedactEnabled bool     `mapstructure:"redact_enabled" yaml:"redact_enabled"`
	IncludeCaller bool     `mapstructure:"include_caller" yaml:"include_caller"`
	BusinessEvents bool     `mapstructure:"business_events" yaml:"business_events"`
	RedactPatterns []string `mapstructure:"redact_patterns" yaml:"redact_patterns"`
}

type RateLimitConfig struct {
	Enabled  bool          `mapstructure:"enabled" yaml:"enabled"`
	Requests int           `mapstructure:"requests" yaml:"requests"`
	Window   time.Duration `mapstructure:"window" yaml:"window"`
}

type MigrationsConfig struct {
	Directory   string `mapstructure:"directory" yaml:"directory"`
	Timeout     int    `mapstructure:"timeout" yaml:"timeout"`
	LockTimeout int    `mapstructure:"locktimeout" yaml:"locktimeout"`
}

type HealthConfig struct {
	Timeout              int  `mapstructure:"timeout" yaml:"timeout"`
	DatabaseCheckEnabled bool `mapstructure:"database_check_enabled" yaml:"database_check_enabled"`
	// Story 6.2: Alert thresholds for health monitoring
	ErrorRateMax float64 `mapstructure:"error_rate_max" yaml:"error_rate_max"` // 0.1% = 0.001
	DiskFreeMin  float64 `mapstructure:"disk_free_min" yaml:"disk_free_min"`   // 20% = 0.20
}

// RedisConfig represents Redis configuration for session tracking and token blocklist
// Story 1.8, Task 1: Redis for session storage and token blocklist
type RedisConfig struct {
	Host     string `mapstructure:"host" yaml:"host"`
	Port     string `mapstructure:"port" yaml:"port"`
	Password string `mapstructure:"password" yaml:"password"`
}

// LoadConfig loads configuration using Viper. If configPath is non-empty it
// will be used as the exact config file path, otherwise Viper searches common locations.
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Story 6.2: Set default values for health alert thresholds
	v.SetDefault("health.error_rate_max", 0.1) // 0.1% threshold
	v.SetDefault("health.disk_free_min", 20.0) // 20% free threshold
	// Story 9.4: Set default values for CORS configuration
	v.SetDefault("cors.enabled", true)
	v.SetDefault("cors.allow_credentials", true)
	v.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowed_headers", []string{"Authorization", "Content-Type", "X-Requested-With"})
	v.SetDefault("cors.max_age", 86400) // 24 hours
	// SECURITY: No default origins - require explicit configuration to prevent production exposure
	v.SetDefault("cors.allowed_origins", []string{}) // Empty by default - must be explicitly configured

	// Logging defaults - Story 9.5: Enhanced structured logging
	v.SetDefault("logging.level", "info")               // Default to info level
	v.SetDefault("logging.redact_enabled", true)       // Enable sensitive data redaction by default
	v.SetDefault("logging.include_caller", false)     // Disable caller info by default (performance)
	v.SetDefault("logging.business_events", true)    // Enable business event logging by default
	v.SetDefault("logging.redact_patterns", []string{"password", "token", "secret", "authorization", "cookie"}) // Default redaction patterns

	bindEnvVariables(v)

	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
		}
	} else {
		env := v.GetString("APP_ENVIRONMENT")
		if env == "" {
			env = "development"
		}

		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("configs")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")

		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("failed to read base config file: %w", err)
			}
		}

		v.SetConfigName(fmt.Sprintf("config.%s", env))
		if err := v.MergeInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("failed to merge environment config: %w", err)
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Story 9.4: Validate CORS configuration to prevent security issues
	// Reject wildcard origins when credentials are enabled (forbidden by CORS spec)
	if cfg.Cors.Enabled && cfg.Cors.AllowCredentials {
		for _, origin := range cfg.Cors.AllowedOrigins {
			if origin == "*" {
				return nil, fmt.Errorf("CORS security violation: wildcard origin '*' is not allowed when AllowCredentials is true")
			}
			// Validate origin format (must include protocol)
			if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
				return nil, fmt.Errorf("CORS security violation: origin '%s' must include http:// or https:// protocol", origin)
			}
		}
	}

	if cfg.App.Environment == "" {
		if e := v.GetString("app.environment"); e != "" {
			cfg.App.Environment = e
		} else if e := v.GetString("ENV"); e != "" {
			cfg.App.Environment = e
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func bindEnvVariables(v *viper.Viper) {
	envBindings := map[string]string{
		"app.name":                      "APP_NAME",
		"app.version":                   "APP_VERSION",
		"app.environment":               "APP_ENVIRONMENT",
		"app.debug":                     "APP_DEBUG",
		"database.host":                 "DATABASE_HOST",
		"database.port":                 "DATABASE_PORT",
		"database.user":                 "DATABASE_USER",
		"database.password":             "DATABASE_PASSWORD",
		"database.name":                 "DATABASE_NAME",
		"database.sslmode":              "DATABASE_SSLMODE",
		"database.max_open_conns":       "DB_MAX_OPEN_CONNECTIONS",
		"database.max_idle_conns":       "DB_MAX_IDLE_CONNECTIONS",
		"database.conn_max_lifetime":    "DB_CONNECTION_MAX_LIFETIME",
		"jwt.secret":                    "JWT_SECRET",
		"jwt.access_token_ttl":          "JWT_ACCESS_TOKEN_TTL",
		"jwt.refresh_token_ttl":         "JWT_REFRESH_TOKEN_TTL",
		"jwt.ttlhours":                  "JWT_TTLHOURS",
		"server.port":                   "SERVER_PORT",
		"server.readtimeout":            "SERVER_READTIMEOUT",
		"server.writetimeout":           "SERVER_WRITETIMEOUT",
		"server.idletimeout":            "SERVER_IDLETIMEOUT",
		"server.shutdowntimeout":        "SERVER_SHUTDOWNTIMEOUT",
		"server.maxheaderbytes":         "SERVER_MAXHEADERBYTES",
		"logging.level":                 "LOGGING_LEVEL",
		"logging.redact_enabled":       "LOGGING_REDACT_ENABLED",
		"logging.include_caller":       "LOGGING_INCLUDE_CALLER",
		"logging.business_events":      "LOGGING_BUSINESS_EVENTS",
		"logging.redact_patterns":      "LOGGING_REDACT_PATTERNS",
		"ratelimit.enabled":             "RATELIMIT_ENABLED",
		"ratelimit.requests":            "RATELIMIT_REQUESTS",
		"ratelimit.window":              "RATELIMIT_WINDOW",
		"cors.enabled":                  "CORS_ENABLED",
		"cors.allowed_origins":         "CORS_ALLOWED_ORIGINS",
		"cors.allow_credentials":        "CORS_ALLOW_CREDENTIALS",
		"cors.allowed_methods":          "CORS_ALLOWED_METHODS",
		"cors.allowed_headers":          "CORS_ALLOWED_HEADERS",
		"cors.max_age":                  "CORS_MAX_AGE",
		"migrations.directory":          "MIGRATIONS_DIRECTORY",
		"migrations.timeout":            "MIGRATIONS_TIMEOUT",
		"migrations.locktimeout":        "MIGRATIONS_LOCKTIMEOUT",
		"health.timeout":                "HEALTH_TIMEOUT",
		"health.database_check_enabled": "HEALTH_DATABASE_CHECK_ENABLED",
	}
	for key, env := range envBindings {
		_ = v.BindEnv(key, env)
	}
}

func (l *LoggingConfig) GetLogLevel() slog.Level {
	switch strings.ToLower(l.Level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // Default to info level
	}
}

func GetSkipPaths(env string) []string {
	switch env {
	case "production":
		return []string{"/health", "/health/live", "/health/ready", "/metrics", "/debug", "/pprof"}
	case "development":
		return []string{"/health", "/health/live", "/health/ready"}
	case "test":
		return []string{"/health", "/health/live", "/health/ready"}
	default:
		return []string{"/health", "/health/live", "/health/ready"}
	}
}

func GetConfigPath() string {
	paths := []string{
		"configs/config.yaml",
		"./configs/config.yaml",
		"../configs/config.yaml",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	return "configs/config.yaml"
}

func (c *Config) LogSafeConfig(logger *slog.Logger) {
	logger.Info("Loaded Configuration:")
	logger.Info("App", "Name", c.App.Name, "Environment", c.App.Environment, "Debug", c.App.Debug)
	logger.Info("Database", "Host", c.Database.Host, "Port", c.Database.Port, "User", c.Database.User, "Password", "<redacted>", "Name", c.Database.Name, "SSLMode", c.Database.SSLMode)
	logger.Info("JWT", "Secret", "<redacted>", "AccessTokenTTL", c.JWT.AccessTokenTTL, "RefreshTokenTTL", c.JWT.RefreshTokenTTL)
	logger.Info("Server", "Port", c.Server.Port, "ReadTimeout", c.Server.ReadTimeout, "WriteTimeout", c.Server.WriteTimeout, "IdleTimeout", c.Server.IdleTimeout, "ShutdownTimeout", c.Server.ShutdownTimeout, "MaxHeaderBytes", c.Server.MaxHeaderBytes)
	logger.Info("Logging", "Level", c.Logging.Level)
	logger.Info("RateLimit", "Enabled", c.Ratelimit.Enabled, "Requests", c.Ratelimit.Requests, "Window", c.Ratelimit.Window)
	logger.Info("CORS", "Enabled", c.Cors.Enabled, "AllowedOrigins", "<redacted>", "AllowCredentials", c.Cors.AllowCredentials, "MaxAge", c.Cors.MaxAge)
	logger.Info("Migrations", "Directory", c.Migrations.Directory, "Timeout", c.Migrations.Timeout, "LockTimeout", c.Migrations.LockTimeout)
}
