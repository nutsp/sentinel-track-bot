package service

import (
	"context"
	"fmt"
	"strings"

	"fix-track-bot/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// issueService implements the IssueService interface with new schema
type issueService struct {
	issueRepo   domain.IssueRepository
	channelRepo domain.ChannelRepository
	userRepo    domain.UserRepository
	logger      *zap.Logger
}

// NewIssueService creates a new instance of issue service with new schema support
func NewIssueService(
	issueRepo domain.IssueRepository,
	channelRepo domain.ChannelRepository,
	userRepo domain.UserRepository,
	logger *zap.Logger,
) domain.IssueService {
	return &issueService{
		issueRepo:   issueRepo,
		channelRepo: channelRepo,
		userRepo:    userRepo,
		logger:      logger,
	}
}

// getOrCreateUser gets existing user or creates a new one
func (s *issueService) getOrCreateUser(ctx context.Context, discordID string) (*domain.User, error) {
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
		DiscordID: discordID,
		Role:      domain.UserRoleCustomer,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("Created new user for issue",
		zap.String("user_id", user.ID.String()),
		zap.String("discord_id", discordID),
	)

	return user, nil
}

// CreateIssue creates a new issue
func (s *issueService) CreateIssue(ctx context.Context, title, description, imageURL, reporterID, channelID string) (*domain.Issue, error) {
	s.logger.Debug("Creating issue",
		zap.String("title", title),
		zap.String("reporter_id", reporterID),
		zap.String("channel_id", channelID),
	)

	// Validate input
	if title == "" || description == "" || reporterID == "" || channelID == "" {
		s.logger.Debug("Invalid issue input")
		return nil, fmt.Errorf("invalid issue input: title, description, reporter_id, and channel_id are required")
	}

	// Get channel registration by Discord channel ID
	channel, err := s.channelRepo.GetByChannelID(ctx, channelID)
	if err != nil {
		s.logger.Error("Failed to get channel registration",
			zap.Error(err),
			zap.String("discord_channel_id", channelID),
		)
		return nil, fmt.Errorf("failed to get channel registration: %w", err)
	}

	// Get or create user
	user, err := s.getOrCreateUser(ctx, reporterID)
	if err != nil {
		s.logger.Error("Failed to get or create user",
			zap.Error(err),
			zap.String("reporter_id", reporterID),
		)
		return nil, fmt.Errorf("failed to get or create user: %w", err)
	}

	// Create new issue
	issue := &domain.Issue{
		ID:          uuid.New(),
		ProjectID:   channel.ProjectID, // Project from channel
		ChannelID:   &channel.ID,       // Optional channel reference
		ReporterID:  user.ID,
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(description),
		ImageURL:    strings.TrimSpace(imageURL),
		Priority:    domain.PriorityMedium,        // Default priority
		Status:      domain.StatusOpen,            // Default status
		Source:      string(domain.SourceDiscord), // Mark as Discord issue
		PublicHash:  uuid.New().String(),
	}

	// Save to repository
	if err := s.issueRepo.Create(ctx, issue); err != nil {
		s.logger.Error("Failed to create issue",
			zap.Error(err),
			zap.String("title", title),
		)
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	s.logger.Info("Issue created successfully",
		zap.String("issue_id", issue.ID.String()),
		zap.String("title", title),
		zap.String("project_id", issue.ProjectID.String()),
		zap.String("source", issue.Source),
	)

	return issue, nil
}

// GetIssue retrieves an issue by its ID
func (s *issueService) GetIssue(ctx context.Context, id uuid.UUID) (*domain.Issue, error) {
	s.logger.Debug("Getting issue", zap.String("issue_id", id.String()))

	issue, err := s.issueRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get issue",
			zap.Error(err),
			zap.String("issue_id", id.String()),
		)
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	return issue, nil
}

// GetIssuesByChannel retrieves all issues for a specific Discord channel
func (s *issueService) GetIssuesByChannel(ctx context.Context, discordChannelID string) ([]*domain.Issue, error) {
	s.logger.Debug("Getting issues by Discord channel", zap.String("discord_channel_id", discordChannelID))

	issues, err := s.issueRepo.GetByDiscordChannelID(ctx, discordChannelID)
	if err != nil {
		s.logger.Error("Failed to get issues by Discord channel",
			zap.Error(err),
			zap.String("discord_channel_id", discordChannelID),
		)
		return nil, fmt.Errorf("failed to get issues by Discord channel: %w", err)
	}

	s.logger.Debug("Issues retrieved successfully",
		zap.String("discord_channel_id", discordChannelID),
		zap.Int("count", len(issues)),
	)

	return issues, nil
}

// UpdateIssuePriority updates the priority of an issue
func (s *issueService) UpdateIssuePriority(ctx context.Context, id uuid.UUID, priority domain.Priority) error {
	s.logger.Debug("Updating issue priority",
		zap.String("issue_id", id.String()),
		zap.String("priority", string(priority)),
	)

	// Validate priority
	if !domain.IsValidPriority(priority) {
		s.logger.Debug("Invalid priority", zap.String("priority", string(priority)))
		return fmt.Errorf("invalid priority: %s", priority)
	}

	// Get issue
	issue, err := s.issueRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get issue for priority update",
			zap.Error(err),
			zap.String("issue_id", id.String()),
		)
		return fmt.Errorf("failed to get issue for priority update: %w", err)
	}

	// Update priority
	issue.Priority = priority

	if err := s.issueRepo.Update(ctx, issue); err != nil {
		s.logger.Error("Failed to update issue priority",
			zap.Error(err),
			zap.String("issue_id", id.String()),
		)
		return fmt.Errorf("failed to update issue priority: %w", err)
	}

	s.logger.Info("Issue priority updated successfully",
		zap.String("issue_id", id.String()),
		zap.String("priority", string(priority)),
	)

	return nil
}

// UpdateIssueStatus updates the status of an issue
func (s *issueService) UpdateIssueStatus(ctx context.Context, id uuid.UUID, status domain.Status) error {
	s.logger.Debug("Updating issue status",
		zap.String("issue_id", id.String()),
		zap.String("status", string(status)),
	)

	// Validate status
	if !domain.IsValidStatus(status) {
		s.logger.Debug("Invalid status", zap.String("status", string(status)))
		return fmt.Errorf("invalid status: %s", status)
	}

	// Get issue
	issue, err := s.issueRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get issue for status update",
			zap.Error(err),
			zap.String("issue_id", id.String()),
		)
		return fmt.Errorf("failed to get issue for status update: %w", err)
	}

	// Update status
	if status == domain.StatusClosed {
		issue.Close()
	} else {
		issue.Status = status
		if status == domain.StatusOpen {
			issue.Reopen()
		}
	}

	if err := s.issueRepo.Update(ctx, issue); err != nil {
		s.logger.Error("Failed to update issue status",
			zap.Error(err),
			zap.String("issue_id", id.String()),
		)
		return fmt.Errorf("failed to update issue status: %w", err)
	}

	s.logger.Info("Issue status updated successfully",
		zap.String("issue_id", id.String()),
		zap.String("status", string(status)),
	)

	return nil
}

// UpdateIssueThreadInfo updates the Discord thread information for an issue
func (s *issueService) UpdateIssueThreadInfo(ctx context.Context, id uuid.UUID, threadID, messageID string) error {
	s.logger.Debug("Updating issue thread info",
		zap.String("issue_id", id.String()),
		zap.String("thread_id", threadID),
		zap.String("message_id", messageID),
	)

	// Get issue
	issue, err := s.issueRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get issue for thread info update",
			zap.Error(err),
			zap.String("issue_id", id.String()),
		)
		return fmt.Errorf("failed to get issue for thread info update: %w", err)
	}

	// Update thread info
	issue.ThreadID = threadID
	issue.MessageID = messageID

	if err := s.issueRepo.Update(ctx, issue); err != nil {
		s.logger.Error("Failed to update issue thread info",
			zap.Error(err),
			zap.String("issue_id", id.String()),
		)
		return fmt.Errorf("failed to update issue thread info: %w", err)
	}

	s.logger.Info("Issue thread info updated successfully",
		zap.String("issue_id", id.String()),
		zap.String("thread_id", threadID),
	)

	return nil
}

// SearchIssuesByID searches for issues by partial ID match
func (s *issueService) SearchIssuesByID(ctx context.Context, partialID string) ([]*domain.Issue, error) {
	s.logger.Debug("Searching issues by partial ID", zap.String("partial_id", partialID))

	// For now, we'll implement a simple search by getting all issues and filtering
	// In a production system, you'd want to implement this at the database level
	// This is a placeholder implementation

	// Get all issues (in a real implementation, you'd add pagination and search to the repository)
	// For now, just return empty slice since this is a complex search operation
	// that would require additional repository methods

	s.logger.Debug("Issue search completed", zap.String("partial_id", partialID))
	return []*domain.Issue{}, nil
}

// GetOpenIssues retrieves all open issues
func (s *issueService) GetOpenIssues(ctx context.Context) ([]*domain.Issue, error) {
	s.logger.Debug("Getting open issues")

	issues, err := s.issueRepo.GetByStatus(ctx, domain.StatusOpen)
	if err != nil {
		s.logger.Error("Failed to get open issues", zap.Error(err))
		return nil, fmt.Errorf("failed to get open issues: %w", err)
	}

	s.logger.Debug("Open issues retrieved successfully", zap.Int("count", len(issues)))
	return issues, nil
}

// GetClosedIssues retrieves all closed issues
func (s *issueService) GetClosedIssues(ctx context.Context) ([]*domain.Issue, error) {
	s.logger.Debug("Getting closed issues")

	issues, err := s.issueRepo.GetByStatus(ctx, domain.StatusClosed)
	if err != nil {
		s.logger.Error("Failed to get closed issues", zap.Error(err))
		return nil, fmt.Errorf("failed to get closed issues: %w", err)
	}

	s.logger.Debug("Closed issues retrieved successfully", zap.Int("count", len(issues)))
	return issues, nil
}

// CloseIssue closes an issue
func (s *issueService) CloseIssue(ctx context.Context, id uuid.UUID) error {
	return s.UpdateIssueStatus(ctx, id, domain.StatusClosed)
}

// ListIssuesByChannel lists all issues for a specific channel (alias for GetIssuesByChannel)
func (s *issueService) ListIssuesByChannel(ctx context.Context, channelID string) ([]*domain.Issue, error) {
	return s.GetIssuesByChannel(ctx, channelID)
}

// ListOpenIssues lists all open issues (alias for GetOpenIssues)
func (s *issueService) ListOpenIssues(ctx context.Context) ([]*domain.Issue, error) {
	return s.GetOpenIssues(ctx)
}

// ReopenIssue reopens a closed issue
func (s *issueService) ReopenIssue(ctx context.Context, id uuid.UUID) error {
	return s.UpdateIssueStatus(ctx, id, domain.StatusOpen)
}

// SetThreadInfo sets the thread and message IDs for an issue (alias for UpdateIssueThreadInfo)
func (s *issueService) SetThreadInfo(ctx context.Context, id uuid.UUID, threadID, messageID string) error {
	return s.UpdateIssueThreadInfo(ctx, id, threadID, messageID)
}

// CreateWebIssue creates a new issue from web portal
func (s *issueService) CreateWebIssue(ctx context.Context, projectID uuid.UUID, title, description, imageURL string, reporterID uuid.UUID) (*domain.Issue, error) {
	s.logger.Debug("Creating web issue",
		zap.String("title", title),
		zap.String("project_id", projectID.String()),
		zap.String("reporter_id", reporterID.String()),
	)

	// Validate input
	if title == "" || description == "" || projectID == uuid.Nil || reporterID == uuid.Nil {
		s.logger.Debug("Invalid web issue input")
		return nil, fmt.Errorf("invalid issue input: title, description, project_id, and reporter_id are required")
	}

	// Create new web issue
	issue := &domain.Issue{
		ID:          uuid.New(),
		ProjectID:   projectID,
		ChannelID:   nil, // No channel for web issues
		ReporterID:  reporterID,
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(description),
		ImageURL:    strings.TrimSpace(imageURL),
		Priority:    domain.PriorityMedium,    // Default priority
		Status:      domain.StatusOpen,        // Default status
		Source:      string(domain.SourceWeb), // Mark as web issue
	}

	// Save to repository
	if err := s.issueRepo.Create(ctx, issue); err != nil {
		s.logger.Error("Failed to create web issue",
			zap.Error(err),
			zap.String("title", title),
		)
		return nil, fmt.Errorf("failed to create web issue: %w", err)
	}

	s.logger.Info("Web issue created successfully",
		zap.String("issue_id", issue.ID.String()),
		zap.String("title", title),
		zap.String("project_id", issue.ProjectID.String()),
	)

	return issue, nil
}
