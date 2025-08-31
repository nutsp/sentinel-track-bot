package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

// CommandManager manages Discord slash commands
type CommandManager struct {
	session *discordgo.Session
	logger  *zap.Logger
}

// NewCommandManager creates a new command manager
func NewCommandManager(session *discordgo.Session, logger *zap.Logger) *CommandManager {
	return &CommandManager{
		session: session,
		logger:  logger,
	}
}

// getCommandDefinitions returns the command definitions
func (cm *CommandManager) getCommandDefinitions() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		// Issue Management
		{
			Name:        "issue",
			Description: "Report a new issue or bug",
		},
		{
			Name:        "issues",
			Description: "List issues in this channel with filters",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "status",
					Description: "Filter by status",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "🔵 Open", Value: "open"},
						{Name: "🔷 In Progress", Value: "in_progress"},
						{Name: "🟡 In Progress", Value: "in_progress"},
						{Name: "🟢 Resolved", Value: "resolved"},
						{Name: "🔷 Verified", Value: "verified"},
						{Name: "✅ Verified", Value: "verified"},
						{Name: "🟣 Closed", Value: "closed"},
						{Name: "🔴 Rejected", Value: "rejected"},
						{Name: "🟠 Reopened", Value: "reopened"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "role",
					Description: "Filter by assignee role",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "👨‍💻 Developer", Value: "dev"},
						{Name: "🧪 QA Tester", Value: "qa"},
						{Name: "👀 Reviewer", Value: "reviewer"},
						{Name: "👤 Other", Value: "other"},
					},
				},
			},
		},
		{
			Name:        "issue-status",
			Description: "Check the status and history of a specific issue",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "id",
					Description: "Issue ID to check",
					Required:    true,
				},
			},
		},

		// Workflow Commands - Development Phase
		{
			Name:        "assign-dev",
			Description: "Assign issue to developer (Open → In Progress)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "issue_id",
					Description: "Issue ID to assign",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "developer",
					Description: "Developer to assign",
					Required:    true,
				},
			},
		},
		{
			Name:        "start-work",
			Description: "Start working on issue (In Progress → Resolved)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "issue_id",
					Description: "Issue ID to start working on",
					Required:    true,
				},
			},
		},
		{
			Name:        "resolve-issue",
			Description: "Mark issue as resolved (In Progress → Resolved)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "issue_id",
					Description: "Issue ID to resolve",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "resolution",
					Description: "How the issue was resolved",
					Required:    false,
				},
			},
		},

		// Workflow Commands - QA Phase
		{
			Name:        "assign-qa",
			Description: "Assign issue to QA tester (Resolved → Verified)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "issue_id",
					Description: "Issue ID to assign for testing",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "qa_tester",
					Description: "QA tester to assign",
					Required:    true,
				},
			},
		},
		{
			Name:        "verify-issue",
			Description: "Verify issue fix (Verified → Closed)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "issue_id",
					Description: "Issue ID to verify",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "notes",
					Description: "QA verification notes",
					Required:    false,
				},
			},
		},
		{
			Name:        "reject-issue",
			Description: "Reject issue fix (Verified → Rejected)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "issue_id",
					Description: "Issue ID to reject",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "reason",
					Description: "Reason for rejection",
					Required:    true,
				},
			},
		},

		// Final Actions
		{
			Name:        "close-issue",
			Description: "Close issue (Verified → Closed)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "issue_id",
					Description: "Issue ID to close",
					Required:    true,
				},
			},
		},
		{
			Name:        "reopen-issue",
			Description: "Reopen closed issue (Closed → Reopened)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "issue_id",
					Description: "Issue ID to reopen",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "reason",
					Description: "Reason for reopening",
					Required:    false,
				},
			},
		},

		// Utility Commands
		{
			Name:        "my-issues",
			Description: "Show issues assigned to you",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "role",
					Description: "Filter by your role",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "👨‍💻 Developer", Value: "dev"},
						{Name: "🧪 QA Tester", Value: "qa"},
						{Name: "👀 Reviewer", Value: "reviewer"},
					},
				},
			},
		},
		{
			Name:        "workflow",
			Description: "Show the issue workflow and your current tasks",
		},

		// Setup Commands
		{
			Name:        "register",
			Description: "Register this channel for issue tracking with customer and project information",
		},
		{
			Name:        "help",
			Description: "Show help information for the bot",
		},
	}
}

// RegisterCommands registers all slash commands with Discord
func (cm *CommandManager) RegisterCommands() error {
	commands := cm.getCommandDefinitions()

	cm.logger.Info("Registering Discord commands")

	// Register commands globally
	for _, command := range commands {
		_, err := cm.session.ApplicationCommandCreate(cm.session.State.User.ID, "", command)
		if err != nil {
			cm.logger.Error("Failed to register command",
				zap.Error(err),
				zap.String("command", command.Name),
			)
			return fmt.Errorf("failed to register command %s: %w", command.Name, err)
		}

		cm.logger.Debug("Registered command", zap.String("command", command.Name))
	}

	cm.logger.Info("All Discord commands registered successfully")
	return nil
}

// RegisterGuildCommands registers commands for a specific guild (server)
func (cm *CommandManager) RegisterGuildCommands(guildID string) error {
	// Use the same commands as global registration
	commands := cm.getCommandDefinitions()

	cm.logger.Info("Registering Discord commands for guild", zap.String("guild_id", guildID))

	for _, command := range commands {
		_, err := cm.session.ApplicationCommandCreate(cm.session.State.User.ID, guildID, command)
		if err != nil {
			cm.logger.Error("Failed to register guild command",
				zap.Error(err),
				zap.String("command", command.Name),
				zap.String("guild_id", guildID),
			)
			return fmt.Errorf("failed to register guild command %s: %w", command.Name, err)
		}

		cm.logger.Debug("Registered guild command",
			zap.String("command", command.Name),
			zap.String("guild_id", guildID),
		)
	}

	cm.logger.Info("All Discord commands registered successfully for guild", zap.String("guild_id", guildID))
	return nil
}

// CleanupCommands removes all registered commands
func (cm *CommandManager) CleanupCommands() error {
	cm.logger.Info("Cleaning up Discord commands")

	// Get all registered commands
	commands, err := cm.session.ApplicationCommands(cm.session.State.User.ID, "")
	if err != nil {
		cm.logger.Error("Failed to get application commands", zap.Error(err))
		return fmt.Errorf("failed to get application commands: %w", err)
	}

	// Delete each command
	for _, command := range commands {
		err := cm.session.ApplicationCommandDelete(cm.session.State.User.ID, "", command.ID)
		if err != nil {
			cm.logger.Error("Failed to delete command",
				zap.Error(err),
				zap.String("command", command.Name),
			)
			continue
		}

		cm.logger.Debug("Deleted command", zap.String("command", command.Name))
	}

	cm.logger.Info("Discord commands cleanup completed")
	return nil
}
