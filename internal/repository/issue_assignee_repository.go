package repository

import (
	"context"
	"fmt"

	"fix-track-bot/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type issueAssigneeRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewIssueAssigneeRepository creates a new issue assignee repository
func NewIssueAssigneeRepository(db *gorm.DB, logger *zap.Logger) domain.IssueAssigneeRepository {
	return &issueAssigneeRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new issue assignee in the database
func (r *issueAssigneeRepository) Create(ctx context.Context, assignee *domain.IssueAssignee) error {
	r.logger.Debug("Creating new issue assignee",
		zap.String("assignee_id", assignee.ID.String()),
		zap.String("issue_id", assignee.IssueID.String()),
		zap.String("user_id", assignee.UserID.String()),
		zap.String("role", assignee.Role.String()),
	)

	if err := r.db.WithContext(ctx).Create(assignee).Error; err != nil {
		r.logger.Error("Failed to create issue assignee",
			zap.Error(err),
			zap.String("issue_id", assignee.IssueID.String()),
			zap.String("user_id", assignee.UserID.String()),
			zap.String("role", assignee.Role.String()),
		)
		return fmt.Errorf("failed to create issue assignee: %w", err)
	}

	r.logger.Info("Issue assignee created successfully",
		zap.String("assignee_id", assignee.ID.String()),
		zap.String("issue_id", assignee.IssueID.String()),
		zap.String("user_id", assignee.UserID.String()),
		zap.String("role", assignee.Role.String()),
	)

	return nil
}

// GetByID retrieves an issue assignee by its ID
func (r *issueAssigneeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.IssueAssignee, error) {
	r.logger.Debug("Retrieving issue assignee by ID", zap.String("assignee_id", id.String()))

	var assignee domain.IssueAssignee
	if err := r.db.WithContext(ctx).
		Preload("Issue").
		Preload("User").
		First(&assignee, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("Issue assignee not found", zap.String("assignee_id", id.String()))
			return nil, domain.ErrAssigneeNotFound
		}
		r.logger.Error("Failed to retrieve issue assignee by ID",
			zap.Error(err),
			zap.String("assignee_id", id.String()),
		)
		return nil, fmt.Errorf("failed to retrieve issue assignee: %w", err)
	}

	r.logger.Debug("Issue assignee retrieved successfully", zap.String("assignee_id", id.String()))
	return &assignee, nil
}

// GetByIssueID retrieves all assignees for a specific issue
func (r *issueAssigneeRepository) GetByIssueID(ctx context.Context, issueID uuid.UUID) ([]*domain.IssueAssignee, error) {
	r.logger.Debug("Retrieving assignees by issue ID", zap.String("issue_id", issueID.String()))

	var assignees []*domain.IssueAssignee
	if err := r.db.WithContext(ctx).
		Preload("Issue").
		Preload("User").
		Where("issue_id = ?", issueID).
		Order("assigned_at ASC").
		Find(&assignees).Error; err != nil {
		r.logger.Error("Failed to retrieve assignees by issue ID",
			zap.Error(err),
			zap.String("issue_id", issueID.String()),
		)
		return nil, fmt.Errorf("failed to retrieve assignees by issue ID: %w", err)
	}

	r.logger.Debug("Assignees retrieved successfully",
		zap.String("issue_id", issueID.String()),
		zap.Int("count", len(assignees)),
	)

	return assignees, nil
}

// GetByUserID retrieves all issue assignments for a specific user
func (r *issueAssigneeRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.IssueAssignee, error) {
	r.logger.Debug("Retrieving assignments by user ID", zap.String("user_id", userID.String()))

	var assignees []*domain.IssueAssignee
	if err := r.db.WithContext(ctx).
		Preload("Issue").
		Preload("Issue.Project").
		Preload("User").
		Where("user_id = ?", userID).
		Order("assigned_at DESC").
		Find(&assignees).Error; err != nil {
		r.logger.Error("Failed to retrieve assignments by user ID",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return nil, fmt.Errorf("failed to retrieve assignments by user ID: %w", err)
	}

	r.logger.Debug("Assignments retrieved successfully",
		zap.String("user_id", userID.String()),
		zap.Int("count", len(assignees)),
	)

	return assignees, nil
}

// GetByIssueAndUser retrieves all assignments for a specific issue and user
func (r *issueAssigneeRepository) GetByIssueAndUser(ctx context.Context, issueID, userID uuid.UUID) ([]*domain.IssueAssignee, error) {
	r.logger.Debug("Retrieving assignments by issue and user",
		zap.String("issue_id", issueID.String()),
		zap.String("user_id", userID.String()),
	)

	var assignees []*domain.IssueAssignee
	if err := r.db.WithContext(ctx).
		Preload("Issue").
		Preload("User").
		Where("issue_id = ? AND user_id = ?", issueID, userID).
		Order("assigned_at ASC").
		Find(&assignees).Error; err != nil {
		r.logger.Error("Failed to retrieve assignments by issue and user",
			zap.Error(err),
			zap.String("issue_id", issueID.String()),
			zap.String("user_id", userID.String()),
		)
		return nil, fmt.Errorf("failed to retrieve assignments by issue and user: %w", err)
	}

	r.logger.Debug("Assignments retrieved successfully",
		zap.String("issue_id", issueID.String()),
		zap.String("user_id", userID.String()),
		zap.Int("count", len(assignees)),
	)

	return assignees, nil
}

// GetByIssueAndRole retrieves all assignees for a specific issue with a specific role
func (r *issueAssigneeRepository) GetByIssueAndRole(ctx context.Context, issueID uuid.UUID, role domain.AssigneeRole) ([]*domain.IssueAssignee, error) {
	r.logger.Debug("Retrieving assignees by issue and role",
		zap.String("issue_id", issueID.String()),
		zap.String("role", role.String()),
	)

	var assignees []*domain.IssueAssignee
	if err := r.db.WithContext(ctx).
		Preload("Issue").
		Preload("User").
		Where("issue_id = ? AND role = ?", issueID, role).
		Order("assigned_at ASC").
		Find(&assignees).Error; err != nil {
		r.logger.Error("Failed to retrieve assignees by issue and role",
			zap.Error(err),
			zap.String("issue_id", issueID.String()),
			zap.String("role", role.String()),
		)
		return nil, fmt.Errorf("failed to retrieve assignees by issue and role: %w", err)
	}

	r.logger.Debug("Assignees retrieved successfully",
		zap.String("issue_id", issueID.String()),
		zap.String("role", role.String()),
		zap.Int("count", len(assignees)),
	)

	return assignees, nil
}

// Update updates an issue assignee in the database
func (r *issueAssigneeRepository) Update(ctx context.Context, assignee *domain.IssueAssignee) error {
	r.logger.Debug("Updating issue assignee",
		zap.String("assignee_id", assignee.ID.String()),
		zap.String("role", assignee.Role.String()),
	)

	result := r.db.WithContext(ctx).Save(assignee)
	if result.Error != nil {
		r.logger.Error("Failed to update issue assignee",
			zap.Error(result.Error),
			zap.String("assignee_id", assignee.ID.String()),
		)
		return fmt.Errorf("failed to update issue assignee: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("Issue assignee not found for update", zap.String("assignee_id", assignee.ID.String()))
		return domain.ErrAssigneeNotFound
	}

	r.logger.Info("Issue assignee updated successfully",
		zap.String("assignee_id", assignee.ID.String()),
		zap.String("role", assignee.Role.String()),
	)

	return nil
}

// Delete removes an issue assignee from the database
func (r *issueAssigneeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug("Deleting issue assignee", zap.String("assignee_id", id.String()))

	result := r.db.WithContext(ctx).Delete(&domain.IssueAssignee{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("Failed to delete issue assignee",
			zap.Error(result.Error),
			zap.String("assignee_id", id.String()),
		)
		return fmt.Errorf("failed to delete issue assignee: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("Issue assignee not found for deletion", zap.String("assignee_id", id.String()))
		return domain.ErrAssigneeNotFound
	}

	r.logger.Info("Issue assignee deleted successfully", zap.String("assignee_id", id.String()))
	return nil
}

// DeleteByIssueAndUser removes all assignments for a specific issue and user
func (r *issueAssigneeRepository) DeleteByIssueAndUser(ctx context.Context, issueID, userID uuid.UUID) error {
	r.logger.Debug("Deleting assignments by issue and user",
		zap.String("issue_id", issueID.String()),
		zap.String("user_id", userID.String()),
	)

	result := r.db.WithContext(ctx).Delete(&domain.IssueAssignee{}, "issue_id = ? AND user_id = ?", issueID, userID)
	if result.Error != nil {
		r.logger.Error("Failed to delete assignments by issue and user",
			zap.Error(result.Error),
			zap.String("issue_id", issueID.String()),
			zap.String("user_id", userID.String()),
		)
		return fmt.Errorf("failed to delete assignments by issue and user: %w", result.Error)
	}

	r.logger.Info("Assignments deleted successfully",
		zap.String("issue_id", issueID.String()),
		zap.String("user_id", userID.String()),
		zap.Int64("count", result.RowsAffected),
	)

	return nil
}

// DeleteByIssueAndUserAndRole removes a specific assignment for an issue, user, and role
func (r *issueAssigneeRepository) DeleteByIssueAndUserAndRole(ctx context.Context, issueID, userID uuid.UUID, role domain.AssigneeRole) error {
	r.logger.Debug("Deleting assignment by issue, user, and role",
		zap.String("issue_id", issueID.String()),
		zap.String("user_id", userID.String()),
		zap.String("role", role.String()),
	)

	result := r.db.WithContext(ctx).Delete(&domain.IssueAssignee{}, "issue_id = ? AND user_id = ? AND role = ?", issueID, userID, role)
	if result.Error != nil {
		r.logger.Error("Failed to delete assignment by issue, user, and role",
			zap.Error(result.Error),
			zap.String("issue_id", issueID.String()),
			zap.String("user_id", userID.String()),
			zap.String("role", role.String()),
		)
		return fmt.Errorf("failed to delete assignment by issue, user, and role: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("Assignment not found for deletion",
			zap.String("issue_id", issueID.String()),
			zap.String("user_id", userID.String()),
			zap.String("role", role.String()),
		)
		return domain.ErrAssigneeNotFound
	}

	r.logger.Info("Assignment deleted successfully",
		zap.String("issue_id", issueID.String()),
		zap.String("user_id", userID.String()),
		zap.String("role", role.String()),
	)

	return nil
}
