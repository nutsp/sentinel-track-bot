// Package main provides a simple entry point that delegates to the main application.
// The actual application logic is in cmd/bot/main.go following Go project conventions.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"fix-track-bot/internal/config"
	"fix-track-bot/internal/repository"
	"fix-track-bot/internal/service"
	"fix-track-bot/internal/transport/discord"
	"fix-track-bot/pkg/logger"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

// App represents the main application
type App struct {
	config    *config.Config
	logger    *zap.Logger
	dbManager *repository.DatabaseManager
	session   *discordgo.Session
	handler   *discord.Handler
	cmdMgr    *discord.CommandManager
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.NewLogger(cfg.Logger)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting Fix Track Bot",
		zap.String("version", cfg.App.Version),
		zap.String("environment", cfg.App.Environment),
	)

	// Create and run application
	app, err := NewApp(cfg, log)
	if err != nil {
		log.Fatal("Failed to create application", zap.Error(err))
	}

	if err := app.Run(); err != nil {
		log.Fatal("Application failed to run", zap.Error(err))
	}
}

// NewApp creates a new application instance
func NewApp(cfg *config.Config, logger *zap.Logger) (*App, error) {
	// Initialize database
	dbManager, err := repository.NewDatabaseManager(&cfg.Database, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run migrations
	if err := dbManager.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	// Initialize Discord session
	session, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	// Set Discord intents
	// session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages
	session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent

	// Initialize repository layer
	customerRepo := repository.NewCustomerRepository(dbManager.GetDB(), logger)
	projectRepo := repository.NewProjectRepository(dbManager.GetDB(), logger)
	userRepo := repository.NewUserRepository(dbManager.GetDB(), logger)
	issueRepo := repository.NewIssueRepository(dbManager.GetDB(), logger)
	channelRepo := repository.NewChannelRepository(dbManager.GetDB(), logger)
	issueAssigneeRepo := repository.NewIssueAssigneeRepository(dbManager.GetDB(), logger)

	// Initialize service layer
	issueService := service.NewIssueService(issueRepo, channelRepo, userRepo, logger)
	channelService := service.NewChannelService(channelRepo, customerRepo, projectRepo, userRepo, logger)
	issueAssigneeService := service.NewIssueAssigneeService(issueAssigneeRepo, userRepo, logger)

	// Initialize transport layer
	handler := discord.NewHandler(session, issueService, channelService, issueAssigneeService, logger)
	cmdMgr := discord.NewCommandManager(session, logger)

	return &App{
		config:    cfg,
		logger:    logger,
		dbManager: dbManager,
		session:   session,
		handler:   handler,
		cmdMgr:    cmdMgr,
	}, nil
}

// Run starts the application
func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register Discord handlers
	a.handler.RegisterHandlers()

	a.logger.Info("Opening Discord connection")

	// Open Discord connection
	if err := a.session.Open(); err != nil {
		return fmt.Errorf("failed to open Discord connection: %w", err)
	}
	defer a.session.Close()

	a.logger.Info("Discord connection established")

	// Register slash commands
	if err := a.cmdMgr.RegisterCommands(); err != nil {
		a.logger.Error("Failed to register commands", zap.Error(err))
		// Continue without commands for now
	}

	// Register commands in all guilds
	for _, guild := range a.session.State.Guilds {
		a.logger.Info("Registering commands in guild",
			zap.String("guild_name", guild.Name),
			zap.String("guild_id", guild.ID),
		)

		if err := a.cmdMgr.RegisterGuildCommands(guild.ID); err != nil {
			a.logger.Error("Failed to register commands in guild",
				zap.Error(err),
				zap.String("guild_id", guild.ID),
			)
		}
	}

	a.logger.Info("Bot is now running. Press CTRL-C to exit.")

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	select {
	case <-stop:
		a.logger.Info("Received shutdown signal")
	case <-ctx.Done():
		a.logger.Info("Context cancelled")
	}

	return a.Shutdown(ctx)
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down application")

	// Cleanup Discord commands
	if err := a.cmdMgr.CleanupCommands(); err != nil {
		a.logger.Error("Failed to cleanup Discord commands", zap.Error(err))
	}

	// Close Discord session
	if err := a.session.Close(); err != nil {
		a.logger.Error("Failed to close Discord session", zap.Error(err))
	}

	// Close database connection
	if err := a.dbManager.Close(); err != nil {
		a.logger.Error("Failed to close database connection", zap.Error(err))
	}

	a.logger.Info("Application shutdown complete")
	return nil
}
