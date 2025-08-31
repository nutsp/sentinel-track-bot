package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"fix-track-bot/internal/config"
	"fix-track-bot/internal/domain"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseManager manages database connections and migrations
type DatabaseManager struct {
	db     *gorm.DB
	config *config.DatabaseConfig
	logger *zap.Logger
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager(config *config.DatabaseConfig, zapLogger *zap.Logger) (*DatabaseManager, error) {
	// Configure GORM logger
	gormLogger := logger.Default
	if config.Driver == "sqlite" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	var db *gorm.DB
	var err error

	switch config.Driver {
	case "sqlite":
		// Ensure data directory exists
		if err := os.MkdirAll(filepath.Dir(config.FilePath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}

		db, err = gorm.Open(sqlite.Open(config.FilePath), &gorm.Config{
			Logger: gormLogger,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
		}

		zapLogger.Info("Connected to SQLite database", zap.String("path", config.FilePath))

	case "postgres":
		dsn := config.GetDSN()
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogger,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
		}

		zapLogger.Info("Connected to PostgreSQL database",
			zap.String("host", config.Host),
			zap.Int("port", config.Port),
			zap.String("database", config.Database),
		)

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	return &DatabaseManager{
		db:     db,
		config: config,
		logger: zapLogger,
	}, nil
}

// GetDB returns the database connection
func (dm *DatabaseManager) GetDB() *gorm.DB {
	return dm.db
}

// Migrate runs database migrations
func (dm *DatabaseManager) Migrate() error {
	dm.logger.Info("Running database migrations")

	// For PostgreSQL, we expect the schema to be created manually using the DDL script
	// This prevents GORM from misinterpreting relationships and creating wrong foreign keys
	// if dm.config.Driver == "postgres" {
	// 	dm.logger.Info("PostgreSQL detected - skipping AutoMigrate. Please ensure schema is created using DDL script.")
	// 	return nil
	// }

	// For SQLite, continue using AutoMigrate for development
	models := []interface{}{
		&domain.Customer{},
		&domain.Project{},
		&domain.User{},
		&domain.Channel{},
		&domain.Issue{},
		&domain.IssueAssignee{},
		&domain.IssueStatusLog{},
	}

	for _, model := range models {
		if err := dm.db.AutoMigrate(model); err != nil {
			dm.logger.Error("Failed to run migrations", zap.Error(err))
			return fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	dm.logger.Info("Database migrations completed successfully")
	return nil
}

// Close closes the database connection
func (dm *DatabaseManager) Close() error {
	dm.logger.Info("Closing database connection")

	sqlDB, err := dm.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	dm.logger.Info("Database connection closed")
	return nil
}

// Health checks if the database connection is healthy
func (dm *DatabaseManager) Health() error {
	sqlDB, err := dm.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}
