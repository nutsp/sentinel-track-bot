package config

import (
	"fmt"
	"strings"

	"fix-track-bot/pkg/logger"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Discord  DiscordConfig  `mapstructure:"discord"`
	Database DatabaseConfig `mapstructure:"database"`
	Logger   logger.Config  `mapstructure:"logger"`
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
	Debug       bool   `mapstructure:"debug"`
}

// DiscordConfig holds Discord bot configuration
type DiscordConfig struct {
	Token  string `mapstructure:"token"`
	Prefix string `mapstructure:"prefix"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"ssl_mode"`
	FilePath string `mapstructure:"file_path"` // For SQLite
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("$HOME/.fix-track-bot")

	// Set default values
	setDefaults()

	// Enable reading from environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, continue with environment variables and defaults
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// App defaults
	viper.SetDefault("app.name", "fix-track-bot")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.debug", false)

	// Discord defaults
	viper.SetDefault("discord.prefix", "!")

	// Database defaults
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.file_path", "./data/fix-track.db")

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.environment", "development")
	viper.SetDefault("logger.output_paths", []string{"stdout"})
}

// validate validates the configuration
func validate(config *Config) error {
	// Validate Discord token
	if strings.TrimSpace(config.Discord.Token) == "" {
		return fmt.Errorf("discord token is required")
	}

	// Validate database configuration
	if config.Database.Driver == "" {
		return fmt.Errorf("database driver is required")
	}

	if config.Database.Driver != "sqlite" && config.Database.Driver != "postgres" {
		return fmt.Errorf("unsupported database driver: %s", config.Database.Driver)
	}

	// For SQLite, ensure file path is provided
	if config.Database.Driver == "sqlite" && strings.TrimSpace(config.Database.FilePath) == "" {
		return fmt.Errorf("database file path is required for SQLite")
	}

	// For PostgreSQL, ensure host and database are provided
	if config.Database.Driver == "postgres" {
		if strings.TrimSpace(config.Database.Host) == "" {
			return fmt.Errorf("database host is required for PostgreSQL")
		}
		if strings.TrimSpace(config.Database.Database) == "" {
			return fmt.Errorf("database name is required for PostgreSQL")
		}
	}

	return nil
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	switch c.Driver {
	case "sqlite":
		return c.FilePath
	case "postgres":
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode,
		)
	default:
		return ""
	}
}
