package service

import (
	"context"
	"fmt"

	"fix-track-bot/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// channelService implements the ChannelService interface with new schema
type channelService struct {
	channelRepo  domain.ChannelRepository
	customerRepo domain.CustomerRepository
	projectRepo  domain.ProjectRepository
	userRepo     domain.UserRepository
	logger       *zap.Logger
}

// NewChannelService creates a new instance of channel service with new schema support
func NewChannelService(
	channelRepo domain.ChannelRepository,
	customerRepo domain.CustomerRepository,
	projectRepo domain.ProjectRepository,
	userRepo domain.UserRepository,
	logger *zap.Logger,
) domain.ChannelService {
	return &channelService{
		channelRepo:  channelRepo,
		customerRepo: customerRepo,
		projectRepo:  projectRepo,
		userRepo:     userRepo,
		logger:       logger,
	}
}

// getOrCreateCustomer gets existing customer or creates a new one
func (s *channelService) getOrCreateCustomer(ctx context.Context, name, email string) (*domain.Customer, error) {
	customer, err := s.customerRepo.GetByName(ctx, name)
	if err != nil && err != domain.ErrCustomerNotFound {
		return nil, fmt.Errorf("failed to check existing customer: %w", err)
	}

	if customer != nil {
		return customer, nil
	}

	// Create new customer
	customer = &domain.Customer{
		ID:           uuid.New(),
		Name:         name,
		ContactEmail: email,
	}

	if err := s.customerRepo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	s.logger.Info("Created new customer",
		zap.String("customer_id", customer.ID.String()),
		zap.String("name", name),
	)

	return customer, nil
}

// getOrCreateProject gets existing project or creates a new one
func (s *channelService) getOrCreateProject(ctx context.Context, customerID uuid.UUID, name, description string) (*domain.Project, error) {
	project, err := s.projectRepo.GetByName(ctx, customerID, name)
	if err != nil && err != domain.ErrProjectNotFound {
		return nil, fmt.Errorf("failed to check existing project: %w", err)
	}

	if project != nil {
		return project, nil
	}

	// Create new project
	project = &domain.Project{
		ID:          uuid.New(),
		CustomerID:  customerID,
		Name:        name,
		Description: description,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	s.logger.Info("Created new project",
		zap.String("project_id", project.ID.String()),
		zap.String("customer_id", customerID.String()),
		zap.String("name", name),
	)

	return project, nil
}

// getOrCreateUser gets existing user or creates a new one
func (s *channelService) getOrCreateUser(ctx context.Context, discordID, name string) (*domain.User, error) {
	user, err := s.userRepo.GetByDiscordID(ctx, discordID)
	if err != nil && err != domain.ErrUserNotFound {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if user != nil {
		return user, nil
	}

	// Create new user
	user = &domain.User{
		ID:        uuid.New(),
		Name:      name,
		DiscordID: discordID,
		Role:      domain.UserRoleCustomer,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("Created new user",
		zap.String("user_id", user.ID.String()),
		zap.String("discord_id", discordID),
	)

	return user, nil
}

// RegisterChannel registers a new channel with customer and project information
func (s *channelService) RegisterChannel(ctx context.Context, channelID, customerName, customerEmail, projectName, projectDescription, registeredBy, userName, guildID string) (*domain.Channel, error) {
	s.logger.Debug("Registering channel",
		zap.String("channel_id", channelID),
		zap.String("customer_name", customerName),
		zap.String("customer_email", customerEmail),
		zap.String("project_name", projectName),
		zap.String("project_description", projectDescription),
		zap.String("registered_by", registeredBy),
		zap.String("user_name", userName),
		zap.String("guild_id", guildID),
	)

	// Basic validation
	if channelID == "" || customerName == "" || projectName == "" || registeredBy == "" || guildID == "" {
		s.logger.Debug("Invalid channel registration data")
		return nil, domain.ErrInvalidChannelRegistration
	}

	// Get or create customer
	customer, err := s.getOrCreateCustomer(ctx, customerName, customerEmail)
	if err != nil {
		s.logger.Error("Failed to get or create customer",
			zap.Error(err),
			zap.String("customer_name", customerName),
		)
		return nil, fmt.Errorf("failed to get or create customer: %w", err)
	}

	// Get or create project
	project, err := s.getOrCreateProject(ctx, customer.ID, projectName, projectDescription)
	if err != nil {
		s.logger.Error("Failed to get or create project",
			zap.Error(err),
			zap.String("project_name", projectName),
		)
		return nil, fmt.Errorf("failed to get or create project: %w", err)
	}

	// Get or create user
	user, err := s.getOrCreateUser(ctx, registeredBy, userName)
	if err != nil {
		s.logger.Error("Failed to get or create user",
			zap.Error(err),
			zap.String("registered_by", registeredBy),
		)
		return nil, fmt.Errorf("failed to get or create user: %w", err)
	}

	// Check if channel is already registered
	existingChannel, err := s.channelRepo.GetByChannelID(ctx, channelID)
	if err != nil && err != domain.ErrChannelNotFound {
		s.logger.Error("Failed to check existing channel registration",
			zap.Error(err),
			zap.String("channel_id", channelID),
		)
		return nil, fmt.Errorf("failed to check existing channel registration: %w", err)
	}

	if existingChannel != nil {
		s.logger.Debug("Channel already registered",
			zap.String("channel_id", channelID),
			zap.String("project_id", existingChannel.ProjectID.String()),
		)
		return nil, domain.ErrChannelAlreadyRegistered
	}

	// Create new channel registration
	channel := &domain.Channel{
		ID:               uuid.New(),
		ProjectID:        project.ID,
		DiscordChannelID: channelID,
		GuildID:          guildID,
		RegisteredBy:     user.ID,
		IsActive:         true,
	}

	if err := s.channelRepo.Create(ctx, channel); err != nil {
		s.logger.Error("Failed to create channel registration",
			zap.Error(err),
			zap.String("channel_id", channelID),
		)
		return nil, fmt.Errorf("failed to create channel registration: %w", err)
	}

	s.logger.Info("Channel registered successfully",
		zap.String("channel_id", channelID),
		zap.String("customer_name", customerName),
		zap.String("project_name", projectName),
	)

	return channel, nil
}

// GetChannelRegistration retrieves a channel registration by channel ID
func (s *channelService) GetChannelRegistration(ctx context.Context, channelID string) (*domain.Channel, error) {
	s.logger.Debug("Getting channel registration", zap.String("channel_id", channelID))

	channel, err := s.channelRepo.GetByChannelID(ctx, channelID)
	if err != nil {
		s.logger.Error("Failed to get channel registration",
			zap.Error(err),
			zap.String("channel_id", channelID),
		)
		return nil, fmt.Errorf("failed to get channel registration: %w", err)
	}

	return channel, nil
}

// UpdateChannelRegistration updates the customer and project information for a channel
func (s *channelService) UpdateChannelRegistration(ctx context.Context, channelID, customerName, projectName string) error {
	s.logger.Debug("Updating channel registration",
		zap.String("channel_id", channelID),
		zap.String("customer_name", customerName),
		zap.String("project_name", projectName),
	)

	// Get existing channel
	channel, err := s.channelRepo.GetByChannelID(ctx, channelID)
	if err != nil {
		s.logger.Error("Failed to get channel for update",
			zap.Error(err),
			zap.String("channel_id", channelID),
		)
		return fmt.Errorf("failed to get channel for update: %w", err)
	}

	// Get or create customer (no email for update operation)
	customer, err := s.getOrCreateCustomer(ctx, customerName, "")
	if err != nil {
		s.logger.Error("Failed to get or create customer for update",
			zap.Error(err),
			zap.String("customer_name", customerName),
		)
		return fmt.Errorf("failed to get or create customer for update: %w", err)
	}

	// Get or create project (no description for update operation)
	project, err := s.getOrCreateProject(ctx, customer.ID, projectName, "")
	if err != nil {
		s.logger.Error("Failed to get or create project for update",
			zap.Error(err),
			zap.String("project_name", projectName),
		)
		return fmt.Errorf("failed to get or create project for update: %w", err)
	}

	// Update channel
	channel.ProjectID = project.ID

	if err := s.channelRepo.Update(ctx, channel); err != nil {
		s.logger.Error("Failed to update channel registration",
			zap.Error(err),
			zap.String("channel_id", channelID),
		)
		return fmt.Errorf("failed to update channel registration: %w", err)
	}

	s.logger.Info("Channel registration updated successfully",
		zap.String("channel_id", channelID),
		zap.String("customer_name", customerName),
		zap.String("project_name", projectName),
	)

	return nil
}

// DeactivateChannel deactivates a channel registration
func (s *channelService) DeactivateChannel(ctx context.Context, channelID string) error {
	s.logger.Debug("Deactivating channel", zap.String("channel_id", channelID))

	channel, err := s.channelRepo.GetByChannelID(ctx, channelID)
	if err != nil {
		s.logger.Error("Failed to get channel for deactivation",
			zap.Error(err),
			zap.String("channel_id", channelID),
		)
		return fmt.Errorf("failed to get channel for deactivation: %w", err)
	}

	channel.Deactivate()

	if err := s.channelRepo.Update(ctx, channel); err != nil {
		s.logger.Error("Failed to deactivate channel",
			zap.Error(err),
			zap.String("channel_id", channelID),
		)
		return fmt.Errorf("failed to deactivate channel: %w", err)
	}

	s.logger.Info("Channel deactivated successfully", zap.String("channel_id", channelID))
	return nil
}

// ActivateChannel activates a channel registration
func (s *channelService) ActivateChannel(ctx context.Context, channelID string) error {
	s.logger.Debug("Activating channel", zap.String("channel_id", channelID))

	channel, err := s.channelRepo.GetByChannelID(ctx, channelID)
	if err != nil {
		s.logger.Error("Failed to get channel for activation",
			zap.Error(err),
			zap.String("channel_id", channelID),
		)
		return fmt.Errorf("failed to get channel for activation: %w", err)
	}

	channel.Activate()

	if err := s.channelRepo.Update(ctx, channel); err != nil {
		s.logger.Error("Failed to activate channel",
			zap.Error(err),
			zap.String("channel_id", channelID),
		)
		return fmt.Errorf("failed to activate channel: %w", err)
	}

	s.logger.Info("Channel activated successfully", zap.String("channel_id", channelID))
	return nil
}

// ListChannelsForGuild lists all channel registrations for a specific guild
func (s *channelService) ListChannelsForGuild(ctx context.Context, guildID string) ([]*domain.Channel, error) {
	s.logger.Debug("Listing channels for guild", zap.String("guild_id", guildID))

	channels, err := s.channelRepo.GetByGuildID(ctx, guildID)
	if err != nil {
		s.logger.Error("Failed to list channels for guild",
			zap.Error(err),
			zap.String("guild_id", guildID),
		)
		return nil, fmt.Errorf("failed to list channels for guild: %w", err)
	}

	s.logger.Debug("Channels listed successfully",
		zap.String("guild_id", guildID),
		zap.Int("count", len(channels)),
	)

	return channels, nil
}

// IsChannelRegistered checks if a channel is already registered
func (s *channelService) IsChannelRegistered(ctx context.Context, channelID string) (bool, error) {
	s.logger.Debug("Checking if channel is registered", zap.String("channel_id", channelID))

	_, err := s.channelRepo.GetByChannelID(ctx, channelID)
	if err != nil {
		if err == domain.ErrChannelNotFound {
			return false, nil
		}
		s.logger.Error("Failed to check channel registration",
			zap.Error(err),
			zap.String("channel_id", channelID),
		)
		return false, fmt.Errorf("failed to check channel registration: %w", err)
	}

	return true, nil
}
