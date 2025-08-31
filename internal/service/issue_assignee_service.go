package service

import (
	"context"
	"fix-track-bot/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type issueAssigneeService struct {
	issueAssigneeRepo domain.IssueAssigneeRepository
	userRepo          domain.UserRepository
	logger            *zap.Logger
}

// NewIssueAssigneeService creates a new issue assignee service
func NewIssueAssigneeService(issueAssigneeRepo domain.IssueAssigneeRepository, userRepo domain.UserRepository, logger *zap.Logger) domain.IssueAssigneeService {
	return &issueAssigneeService{
		issueAssigneeRepo: issueAssigneeRepo,
		userRepo:          userRepo,
		logger:            logger,
	}
}

// AssignUserToIssue assigns a user to an issue with a specific role (using Discord ID)
func (s *issueAssigneeService) AssignUserToIssue(ctx context.Context, issueID uuid.UUID, discordID string, role domain.AssigneeRole) (*domain.IssueAssignee, error) {
	s.logger.Debug("Assigning user to issue by Discord ID",
		zap.String("issue_id", issueID.String()),
		zap.String("discord_id", discordID),
		zap.String("role", role.String()),
	)

	// Validate role
	if !role.IsValid() {
		s.logger.Error("Invalid assignee role",
			zap.String("role", role.String()),
		)
		return nil, domain.ErrInvalidAssigneeRole
	}

	// Get user by Discord ID
	user, err := s.userRepo.GetByDiscordID(ctx, discordID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			s.logger.Error("User not found by Discord ID",
				zap.String("discord_id", discordID),
			)
			return nil, domain.ErrUserNotFound
		}
		s.logger.Error("Failed to get user by Discord ID",
			zap.Error(err),
			zap.String("discord_id", discordID),
		)
		return nil, err
	}

	// Check if this assignment already exists
	existing, err := s.issueAssigneeRepo.GetByIssueAndUser(ctx, issueID, user.ID)
	if err != nil {
		s.logger.Error("Failed to check existing assignments",
			zap.Error(err),
			zap.String("issue_id", issueID.String()),
			zap.String("user_id", user.ID.String()),
		)
		return nil, err
	}

	// Check if user is already assigned with this role
	for _, assignment := range existing {
		if assignment.Role == role {
			s.logger.Debug("User already assigned with this role",
				zap.String("issue_id", issueID.String()),
				zap.String("user_id", user.ID.String()),
				zap.String("role", role.String()),
			)
			return assignment, nil // Return existing assignment
		}
	}

	// Create new assignment
	assignee := domain.NewIssueAssignee(issueID, user.ID, role)

	if err := s.issueAssigneeRepo.Create(ctx, assignee); err != nil {
		s.logger.Error("Failed to create issue assignment",
			zap.Error(err),
			zap.String("issue_id", issueID.String()),
			zap.String("user_id", user.ID.String()),
			zap.String("role", role.String()),
		)
		return nil, err
	}

	s.logger.Info("User assigned to issue successfully",
		zap.String("assignment_id", assignee.ID.String()),
		zap.String("issue_id", issueID.String()),
		zap.String("user_id", user.ID.String()),
		zap.String("role", role.String()),
	)

	return assignee, nil
}

// UnassignUserFromIssue removes a specific user role assignment from an issue
func (s *issueAssigneeService) UnassignUserFromIssue(ctx context.Context, issueID, userID uuid.UUID, role domain.AssigneeRole) error {
	s.logger.Debug("Unassigning user from issue",
		zap.String("issue_id", issueID.String()),
		zap.String("user_id", userID.String()),
		zap.String("role", role.String()),
	)

	if err := s.issueAssigneeRepo.DeleteByIssueAndUserAndRole(ctx, issueID, userID, role); err != nil {
		if err == domain.ErrAssigneeNotFound {
			s.logger.Debug("Assignment not found for removal",
				zap.String("issue_id", issueID.String()),
				zap.String("user_id", userID.String()),
				zap.String("role", role.String()),
			)
			return domain.ErrAssigneeNotFound
		}
		s.logger.Error("Failed to unassign user from issue",
			zap.Error(err),
			zap.String("issue_id", issueID.String()),
			zap.String("user_id", userID.String()),
			zap.String("role", role.String()),
		)
		return err
	}

	s.logger.Info("User unassigned from issue successfully",
		zap.String("issue_id", issueID.String()),
		zap.String("user_id", userID.String()),
		zap.String("role", role.String()),
	)

	return nil
}

// UnassignAllUsersFromIssue removes all assignments from an issue
func (s *issueAssigneeService) UnassignAllUsersFromIssue(ctx context.Context, issueID uuid.UUID) error {
	s.logger.Debug("Unassigning all users from issue",
		zap.String("issue_id", issueID.String()),
	)

	// Get all assignments for the issue
	assignments, err := s.issueAssigneeRepo.GetByIssueID(ctx, issueID)
	if err != nil {
		s.logger.Error("Failed to get issue assignments",
			zap.Error(err),
			zap.String("issue_id", issueID.String()),
		)
		return err
	}

	// Delete each assignment
	for _, assignment := range assignments {
		if err := s.issueAssigneeRepo.Delete(ctx, assignment.ID); err != nil {
			s.logger.Error("Failed to delete assignment",
				zap.Error(err),
				zap.String("assignment_id", assignment.ID.String()),
			)
			// Continue with other assignments even if one fails
		}
	}

	s.logger.Info("All users unassigned from issue",
		zap.String("issue_id", issueID.String()),
		zap.Int("count", len(assignments)),
	)

	return nil
}

// GetIssueAssignees returns all assignees for a specific issue
func (s *issueAssigneeService) GetIssueAssignees(ctx context.Context, issueID uuid.UUID) ([]*domain.IssueAssignee, error) {
	s.logger.Debug("Getting issue assignees",
		zap.String("issue_id", issueID.String()),
	)

	assignees, err := s.issueAssigneeRepo.GetByIssueID(ctx, issueID)
	if err != nil {
		s.logger.Error("Failed to get issue assignees",
			zap.Error(err),
			zap.String("issue_id", issueID.String()),
		)
		return nil, err
	}

	s.logger.Debug("Retrieved issue assignees",
		zap.String("issue_id", issueID.String()),
		zap.Int("count", len(assignees)),
	)

	return assignees, nil
}

// GetUserAssignments returns all assignments for a specific user
func (s *issueAssigneeService) GetUserAssignments(ctx context.Context, userID uuid.UUID) ([]*domain.IssueAssignee, error) {
	s.logger.Debug("Getting user assignments",
		zap.String("user_id", userID.String()),
	)

	assignments, err := s.issueAssigneeRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user assignments",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return nil, err
	}

	s.logger.Debug("Retrieved user assignments",
		zap.String("user_id", userID.String()),
		zap.Int("count", len(assignments)),
	)

	return assignments, nil
}

// GetAssigneesByRole returns all assignees for a specific issue with a specific role
func (s *issueAssigneeService) GetAssigneesByRole(ctx context.Context, issueID uuid.UUID, role domain.AssigneeRole) ([]*domain.IssueAssignee, error) {
	s.logger.Debug("Getting assignees by role",
		zap.String("issue_id", issueID.String()),
		zap.String("role", role.String()),
	)

	assignees, err := s.issueAssigneeRepo.GetByIssueAndRole(ctx, issueID, role)
	if err != nil {
		s.logger.Error("Failed to get assignees by role",
			zap.Error(err),
			zap.String("issue_id", issueID.String()),
			zap.String("role", role.String()),
		)
		return nil, err
	}

	s.logger.Debug("Retrieved assignees by role",
		zap.String("issue_id", issueID.String()),
		zap.String("role", role.String()),
		zap.Int("count", len(assignees)),
	)

	return assignees, nil
}

// IsUserAssignedToIssue checks if a user is assigned to an issue with any role
func (s *issueAssigneeService) IsUserAssignedToIssue(ctx context.Context, issueID, userID uuid.UUID) (bool, error) {
	s.logger.Debug("Checking if user is assigned to issue",
		zap.String("issue_id", issueID.String()),
		zap.String("user_id", userID.String()),
	)

	assignments, err := s.issueAssigneeRepo.GetByIssueAndUser(ctx, issueID, userID)
	if err != nil {
		s.logger.Error("Failed to check user assignment",
			zap.Error(err),
			zap.String("issue_id", issueID.String()),
			zap.String("user_id", userID.String()),
		)
		return false, err
	}

	isAssigned := len(assignments) > 0
	s.logger.Debug("User assignment check result",
		zap.String("issue_id", issueID.String()),
		zap.String("user_id", userID.String()),
		zap.Bool("is_assigned", isAssigned),
	)

	return isAssigned, nil
}

// IsUserAssignedWithRole checks if a user is assigned to an issue with a specific role
func (s *issueAssigneeService) IsUserAssignedWithRole(ctx context.Context, issueID, userID uuid.UUID, role domain.AssigneeRole) (bool, error) {
	s.logger.Debug("Checking if user is assigned with role",
		zap.String("issue_id", issueID.String()),
		zap.String("user_id", userID.String()),
		zap.String("role", role.String()),
	)

	assignments, err := s.issueAssigneeRepo.GetByIssueAndUser(ctx, issueID, userID)
	if err != nil {
		s.logger.Error("Failed to check user role assignment",
			zap.Error(err),
			zap.String("issue_id", issueID.String()),
			zap.String("user_id", userID.String()),
			zap.String("role", role.String()),
		)
		return false, err
	}

	// Check if any assignment has the specified role
	for _, assignment := range assignments {
		if assignment.Role == role {
			s.logger.Debug("User assigned with role",
				zap.String("issue_id", issueID.String()),
				zap.String("user_id", userID.String()),
				zap.String("role", role.String()),
			)
			return true, nil
		}
	}

	s.logger.Debug("User not assigned with role",
		zap.String("issue_id", issueID.String()),
		zap.String("user_id", userID.String()),
		zap.String("role", role.String()),
	)

	return false, nil
}
