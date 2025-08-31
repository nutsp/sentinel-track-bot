package repository

import (
	"context"
	"fmt"

	"fix-track-bot/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// issueRepository implements the IssueRepository interface
type issueRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewIssueRepository creates a new instance of issue repository
func NewIssueRepository(db *gorm.DB, logger *zap.Logger) domain.IssueRepository {
	return &issueRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new issue in the database
func (r *issueRepository) Create(ctx context.Context, issue *domain.Issue) error {
	r.logger.Debug("Creating new issue",
		zap.String("title", issue.Title),
		zap.String("project_id", issue.ProjectID.String()),
		zap.String("reporter_id", issue.ReporterID.String()),
		zap.String("source", issue.Source),
	)

	if err := r.db.WithContext(ctx).Create(issue).Error; err != nil {
		r.logger.Error("Failed to create issue",
			zap.Error(err),
			zap.String("title", issue.Title),
		)
		return fmt.Errorf("failed to create issue: %w", err)
	}

	r.logger.Info("Issue created successfully",
		zap.String("issue_id", issue.ID.String()),
		zap.String("title", issue.Title),
	)

	return nil
}

// GetByID retrieves an issue by its ID
func (r *issueRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Issue, error) {
	r.logger.Debug("Retrieving issue by ID", zap.String("issue_id", id.String()))

	var issue domain.Issue
	if err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Project.Customer").
		Preload("Channel").
		Preload("Channel.Project").
		Preload("Channel.Project.Customer").
		Preload("Reporter").
		Preload("Assignee").
		Where("id = ?", id).
		First(&issue).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("Issue not found", zap.String("issue_id", id.String()))
			return nil, domain.ErrIssueNotFound
		}
		r.logger.Error("Failed to retrieve issue",
			zap.Error(err),
			zap.String("issue_id", id.String()),
		)
		return nil, fmt.Errorf("failed to retrieve issue: %w", err)
	}

	r.logger.Debug("Issue retrieved successfully", zap.String("issue_id", id.String()))
	return &issue, nil
}

// GetByChannelID retrieves all issues for a specific channel by UUID
func (r *issueRepository) GetByChannelID(ctx context.Context, channelID uuid.UUID) ([]*domain.Issue, error) {
	r.logger.Debug("Retrieving issues by channel ID", zap.String("channel_id", channelID.String()))

	var issues []*domain.Issue
	if err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Project.Customer").
		Preload("Channel").
		Preload("Channel.Project").
		Preload("Channel.Project.Customer").
		Preload("Reporter").
		Preload("Assignee").
		Where("channel_id = ?", channelID).
		Order("created_at DESC").
		Find(&issues).Error; err != nil {
		r.logger.Error("Failed to retrieve issues by channel ID",
			zap.Error(err),
			zap.String("channel_id", channelID.String()),
		)
		return nil, fmt.Errorf("failed to retrieve issues by channel ID: %w", err)
	}

	r.logger.Debug("Issues retrieved successfully",
		zap.String("channel_id", channelID.String()),
		zap.Int("count", len(issues)),
	)

	return issues, nil
}

// GetByDiscordChannelID retrieves all issues for a specific Discord channel by string ID
func (r *issueRepository) GetByDiscordChannelID(ctx context.Context, discordChannelID string) ([]*domain.Issue, error) {
	r.logger.Debug("Retrieving issues by Discord channel ID", zap.String("discord_channel_id", discordChannelID))

	var issues []*domain.Issue
	if err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Project.Customer").
		Preload("Channel").
		Preload("Channel.Project").
		Preload("Channel.Project.Customer").
		Preload("Reporter").
		Preload("Assignee").
		Joins("JOIN channels ON issues.channel_id = channels.id").
		Where("channels.channel_id = ?", discordChannelID).
		Order("issues.created_at DESC").
		Find(&issues).Error; err != nil {
		r.logger.Error("Failed to retrieve issues by Discord channel ID",
			zap.Error(err),
			zap.String("discord_channel_id", discordChannelID),
		)
		return nil, fmt.Errorf("failed to retrieve issues by Discord channel ID: %w", err)
	}

	r.logger.Debug("Issues retrieved successfully",
		zap.String("discord_channel_id", discordChannelID),
		zap.Int("count", len(issues)),
	)

	return issues, nil
}

// GetByStatus retrieves all issues with a specific status
func (r *issueRepository) GetByStatus(ctx context.Context, status domain.Status) ([]*domain.Issue, error) {
	r.logger.Debug("Retrieving issues by status", zap.String("status", string(status)))

	var issues []*domain.Issue
	if err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Project.Customer").
		Preload("Channel").
		Preload("Channel.Project").
		Preload("Channel.Project.Customer").
		Preload("Reporter").
		Preload("Assignee").
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&issues).Error; err != nil {
		r.logger.Error("Failed to retrieve issues by status",
			zap.Error(err),
			zap.String("status", string(status)),
		)
		return nil, fmt.Errorf("failed to retrieve issues by status: %w", err)
	}

	r.logger.Debug("Issues retrieved successfully",
		zap.String("status", string(status)),
		zap.Int("count", len(issues)),
	)

	return issues, nil
}

// Update updates an existing issue
func (r *issueRepository) Update(ctx context.Context, issue *domain.Issue) error {
	r.logger.Debug("Updating issue", zap.String("issue_id", issue.ID.String()))

	result := r.db.WithContext(ctx).Save(issue)
	if result.Error != nil {
		r.logger.Error("Failed to update issue",
			zap.Error(result.Error),
			zap.String("issue_id", issue.ID.String()),
		)
		return fmt.Errorf("failed to update issue: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("Issue not found for update", zap.String("issue_id", issue.ID.String()))
		return domain.ErrIssueNotFound
	}

	r.logger.Info("Issue updated successfully",
		zap.String("issue_id", issue.ID.String()),
		zap.String("title", issue.Title),
	)

	return nil
}

// Delete removes an issue from the database
func (r *issueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug("Deleting issue", zap.String("issue_id", id.String()))

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Issue{})
	if result.Error != nil {
		r.logger.Error("Failed to delete issue",
			zap.Error(result.Error),
			zap.String("issue_id", id.String()),
		)
		return fmt.Errorf("failed to delete issue: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("Issue not found for deletion", zap.String("issue_id", id.String()))
		return domain.ErrIssueNotFound
	}

	r.logger.Info("Issue deleted successfully", zap.String("issue_id", id.String()))
	return nil
}

// GetByDiscordChannelID implementation is already added above

// List retrieves all issues with pagination
func (r *issueRepository) List(ctx context.Context, offset, limit int) ([]*domain.Issue, error) {
	r.logger.Debug("Listing issues",
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	var issues []*domain.Issue
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("created_at DESC").Find(&issues).Error; err != nil {
		r.logger.Error("Failed to list issues",
			zap.Error(err),
			zap.Int("offset", offset),
			zap.Int("limit", limit),
		)
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	r.logger.Debug("Issues listed successfully",
		zap.Int("count", len(issues)),
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	return issues, nil
}
