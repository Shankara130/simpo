package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
)

// Story 2.4: Default connection pooling constants
const (
	DefaultMaxOpenConns    = 25
	DefaultMaxIdleConns    = 5
	DefaultConnMaxLifetime = 5 * time.Minute
)

// customLogger wraps the default logger to ignore ErrRecordNotFound
type customLogger struct {
	logger.Interface
}

func (l customLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// Don't log "record not found" errors as they are expected in many cases
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	l.Interface.Trace(ctx, begin, fc, err)
}

func (l customLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	// Don't log "record not found" errors as they are expected in many cases
	if len(data) > 0 {
		if err, ok := data[0].(error); ok && errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}
	}
	l.Interface.Error(ctx, msg, data...)
}

// Config holds database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: customLogger{logger.Default.LogMode(logger.Info)},
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established")
	return db, nil
}

// NewPostgresDBFromDatabaseConfig creates a new PostgreSQL DB connection from typed config
// Story 2.4: Enhanced with configurable pooling, Ping verification, and structured logging
func NewPostgresDBFromDatabaseConfig(cfg config.DatabaseConfig) (*gorm.DB, error) {
	// Story 2.4: Validate required database fields
	if cfg.Host == "" {
		return nil, fmt.Errorf("database host not configured (set DB_HOST or DATABASE_HOST)")
	}
	if cfg.Port <= 0 {
		return nil, fmt.Errorf("database port must be positive (got: %d)", cfg.Port)
	}
	if cfg.User == "" {
		return nil, fmt.Errorf("database user not configured (set DB_USER or DATABASE_USER)")
	}
	if cfg.Password == "" {
		return nil, fmt.Errorf("database password not configured (set DB_PASSWORD or DATABASE_PASSWORD)")
	}
	if cfg.Name == "" {
		return nil, fmt.Errorf("database name not configured (set DB_NAME or DATABASE_NAME)")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: customLogger{logger.Default.LogMode(logger.Info)},
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database at %s:%d: %w (check if PostgreSQL is running and credentials are correct)", cfg.Host, cfg.Port, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm DB: %w", err)
	}

	// Story 2.4: Apply configurable connection pooling settings
	// Use validated defaults if not configured (set in validator.go)
	maxOpenConns := cfg.MaxOpenConns
	if maxOpenConns == 0 {
		maxOpenConns = DefaultMaxOpenConns
	}
	maxIdleConns := cfg.MaxIdleConns
	if maxIdleConns == 0 {
		maxIdleConns = DefaultMaxIdleConns
	}
	connMaxLifetime := cfg.ConnMaxLifetime
	if connMaxLifetime == 0 {
		connMaxLifetime = DefaultConnMaxLifetime
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	// Story 2.4: Ping database to verify connectivity
	// Use context with timeout to prevent indefinite hangs
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		if closeErr := sqlDB.Close(); closeErr != nil {
			slog.Error("Failed to close database connection after ping failure", "error", closeErr)
		}
		return nil, fmt.Errorf("database ping failed at %s:%d (database '%s'): %w (verify database is accessible and network connectivity is working)",
			cfg.Host, cfg.Port, cfg.Name, err)
	}

	// Story 2.4: Log successful connection with pooling details
	slog.Info("Database connection established",
		slog.String("host", cfg.Host),
		slog.Int("port", cfg.Port),
		slog.String("database", cfg.Name),
		slog.Int("max_open_connections", maxOpenConns),
		slog.Int("max_idle_connections", maxIdleConns),
		slog.Duration("conn_max_lifetime", connMaxLifetime),
	)

	return db, nil
}

// LogPoolStats logs current connection pool statistics
// Story 2.4: Pool health monitoring for operations visibility
func LogPoolStats(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("Failed to get sql.DB for pool stats", "error", err)
		return
	}

	stats := sqlDB.Stats()
	slog.Info("Database connection pool stats",
		slog.Int("open_connections", stats.OpenConnections),
		slog.Int("in_use", stats.InUse),
		slog.Int("idle", stats.Idle),
		slog.Int64("wait_count", stats.WaitCount),
		slog.Duration("wait_duration", stats.WaitDuration),
		slog.Int64("max_idle_closed", stats.MaxIdleClosed),
		slog.Int64("max_idle_time_closed", stats.MaxIdleTimeClosed),
		slog.Int64("max_lifetime_closed", stats.MaxLifetimeClosed),
	)
}

// NewSQLiteDB creates a new SQLite database connection (for testing)
func NewSQLiteDB(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlite database: %w", err)
	}

	return db, nil
}

// LoadConfigFromEnv loads database configuration using Viper (env overrides + defaults)
func LoadConfigFromEnv() Config {
	return Config{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		User:     viper.GetString("database.user"),
		Password: viper.GetString("database.password"),
		Name:     viper.GetString("database.name"),
		SSLMode:  viper.GetString("database.sslmode"),
	}
}
