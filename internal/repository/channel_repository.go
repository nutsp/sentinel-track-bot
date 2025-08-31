package repository

import (
	"context"
	"fmt"

	"fix-track-bot/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// channelRepository implements the ChannelRepository interface
type channelRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewChannelRepository creates a new instance of channel repository
func NewChannelRepository(db *gorm.DB, logger *zap.Logger) domain.ChannelRepository {
	return &channelRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new channel registration in the database
func (r *channelRepository) Create(ctx context.Context, channel *domain.Channel) error {
	r.logger.Debug("Creating new channel registration",
		zap.String("channel_id", channel.DiscordChannelID),
		zap.String("project_id", channel.ProjectID.String()),
		zap.String("registered_by", channel.RegisteredBy.String()),
	)

	if err := r.db.WithContext(ctx).Create(channel).Error; err != nil {
		r.logger.Error("Failed to create channel registration",
			zap.Error(err),
			zap.String("channel_id", channel.DiscordChannelID),
		)
		return fmt.Errorf("failed to create channel registration: %w", err)
	}

	r.logger.Info("Channel registration created successfully",
		zap.String("registration_id", channel.ID.String()),
		zap.String("channel_id", channel.DiscordChannelID),
		zap.String("project_id", channel.ProjectID.String()),
	)

	return nil
}

// GetByChannelID retrieves a channel registration by its Discord channel ID
func (r *channelRepository) GetByChannelID(ctx context.Context, channelID string) (*domain.Channel, error) {
	r.logger.Debug("Retrieving channel registration by channel ID", zap.String("channel_id", channelID))

	var channel domain.Channel
	if err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Project.Customer").
		Preload("RegisteredByUser").
		Where("discord_channel_id = ?", channelID).
		First(&channel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("Channel registration not found", zap.String("channel_id", channelID))
			return nil, domain.ErrChannelNotFound
		}
		r.logger.Error("Failed to retrieve channel registration",
			zap.Error(err),
			zap.String("channel_id", channelID),
		)
		return nil, fmt.Errorf("failed to retrieve channel registration: %w", err)
	}

	r.logger.Debug("Channel registration retrieved successfully", zap.String("channel_id", channelID))
	return &channel, nil
}

// GetByGuildID retrieves all channel registrations for a specific guild
func (r *channelRepository) GetByGuildID(ctx context.Context, guildID string) ([]*domain.Channel, error) {
	r.logger.Debug("Retrieving channel registrations by guild ID", zap.String("guild_id", guildID))

	var channels []*domain.Channel
	if err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Project.Customer").
		Preload("RegisteredByUser").
		Where("guild_id = ?", guildID).
		Order("created_at DESC").
		Find(&channels).Error; err != nil {
		r.logger.Error("Failed to retrieve channel registrations by guild ID",
			zap.Error(err),
			zap.String("guild_id", guildID),
		)
		return nil, fmt.Errorf("failed to retrieve channel registrations by guild ID: %w", err)
	}

	r.logger.Debug("Channel registrations retrieved successfully",
		zap.String("guild_id", guildID),
		zap.Int("count", len(channels)),
	)

	return channels, nil
}

// Update updates an existing channel registration
func (r *channelRepository) Update(ctx context.Context, channel *domain.Channel) error {
	r.logger.Debug("Updating channel registration", zap.String("registration_id", channel.ID.String()))

	result := r.db.WithContext(ctx).Save(channel)
	if result.Error != nil {
		r.logger.Error("Failed to update channel registration",
			zap.Error(result.Error),
			zap.String("registration_id", channel.ID.String()),
		)
		return fmt.Errorf("failed to update channel registration: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("Channel registration not found for update", zap.String("registration_id", channel.ID.String()))
		return domain.ErrChannelNotFound
	}

	r.logger.Info("Channel registration updated successfully",
		zap.String("registration_id", channel.ID.String()),
		zap.String("channel_id", channel.DiscordChannelID),
	)

	return nil
}

// Delete removes a channel registration from the database
func (r *channelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug("Deleting channel registration", zap.String("registration_id", id.String()))

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Channel{})
	if result.Error != nil {
		r.logger.Error("Failed to delete channel registration",
			zap.Error(result.Error),
			zap.String("registration_id", id.String()),
		)
		return fmt.Errorf("failed to delete channel registration: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("Channel registration not found for deletion", zap.String("registration_id", id.String()))
		return domain.ErrChannelNotFound
	}

	r.logger.Info("Channel registration deleted successfully", zap.String("registration_id", id.String()))
	return nil
}

// List retrieves all channel registrations with pagination
func (r *channelRepository) List(ctx context.Context, offset, limit int) ([]*domain.Channel, error) {
	r.logger.Debug("Listing channel registrations",
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	var channels []*domain.Channel
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("created_at DESC").Find(&channels).Error; err != nil {
		r.logger.Error("Failed to list channel registrations",
			zap.Error(err),
			zap.Int("offset", offset),
			zap.Int("limit", limit),
		)
		return nil, fmt.Errorf("failed to list channel registrations: %w", err)
	}

	r.logger.Debug("Channel registrations listed successfully",
		zap.Int("count", len(channels)),
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	return channels, nil
}

// GetActiveChannels retrieves all active channel registrations
func (r *channelRepository) GetActiveChannels(ctx context.Context) ([]*domain.Channel, error) {
	r.logger.Debug("Retrieving active channel registrations")

	var channels []*domain.Channel
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("created_at DESC").Find(&channels).Error; err != nil {
		r.logger.Error("Failed to retrieve active channel registrations", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve active channel registrations: %w", err)
	}

	r.logger.Debug("Active channel registrations retrieved successfully", zap.Int("count", len(channels)))
	return channels, nil
}
