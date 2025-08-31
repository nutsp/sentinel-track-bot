package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds the logger configuration
type Config struct {
	Level       string   `mapstructure:"level"`        // debug, info, warn, error
	Environment string   `mapstructure:"environment"`  // development, production
	OutputPaths []string `mapstructure:"output_paths"` // stdout, stderr, or file paths
}

// NewLogger creates a new structured logger with the given configuration
func NewLogger(config Config) (*zap.Logger, error) {
	var zapConfig zap.Config

	switch config.Environment {
	case "production":
		zapConfig = zap.NewProductionConfig()
	default:
		zapConfig = zap.NewDevelopmentConfig()
	}

	// Set log level
	level, err := zap.ParseAtomicLevel(config.Level)
	if err != nil {
		level = zap.NewAtomicLevelAt(zap.InfoLevel) // Default to info if invalid level
	}
	zapConfig.Level = level

	// Set output paths
	if len(config.OutputPaths) > 0 {
		zapConfig.OutputPaths = config.OutputPaths
	}

	// Add caller information
	zapConfig.Development = config.Environment != "production"

	// Configure encoder for production
	if config.Environment == "production" {
		zapConfig.EncoderConfig = zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}

// NewDefaultLogger creates a logger with sensible defaults
func NewDefaultLogger() (*zap.Logger, error) {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	config := Config{
		Level:       level,
		Environment: env,
		OutputPaths: []string{"stdout"},
	}

	return NewLogger(config)
}

// NewTestLogger creates a logger suitable for testing
func NewTestLogger() (*zap.Logger, error) {
	config := Config{
		Level:       "debug",
		Environment: "development",
		OutputPaths: []string{"stdout"},
	}

	return NewLogger(config)
}
