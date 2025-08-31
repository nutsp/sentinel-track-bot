package domain

import (
	"time"

	"github.com/google/uuid"
)

// AssigneeRole represents the role of an assignee in an issue
type AssigneeRole string

const (
	AssigneeRoleDev      AssigneeRole = "dev"
	AssigneeRoleQA       AssigneeRole = "qa"
	AssigneeRoleReviewer AssigneeRole = "reviewer"
	AssigneeRoleOther    AssigneeRole = "other"
)

// IssueAssignee represents the assignment of a user to an issue with a specific role
type IssueAssignee struct {
	ID         uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	IssueID    uuid.UUID    `json:"issue_id" gorm:"type:uuid;not null"`
	UserID     uuid.UUID    `json:"user_id" gorm:"type:uuid;not null"`
	Role       AssigneeRole `json:"role" gorm:"size:20;not null"`
	AssignedAt time.Time    `json:"assigned_at" gorm:"type:timestamptz;default:now()"`

	// Relationships
	Issue Issue `json:"issue,omitempty" gorm:"foreignKey:IssueID"`
	User  User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for IssueAssignee
func (IssueAssignee) TableName() string {
	return "issue_assignees"
}

// IsValidRole checks if the role is valid
func (r AssigneeRole) IsValid() bool {
	switch r {
	case AssigneeRoleDev, AssigneeRoleQA, AssigneeRoleReviewer, AssigneeRoleOther:
		return true
	default:
		return false
	}
}

// String returns the string representation of the role
func (r AssigneeRole) String() string {
	return string(r)
}

// GetDisplayName returns a human-readable display name for the role
func (r AssigneeRole) GetDisplayName() string {
	switch r {
	case AssigneeRoleDev:
		return "Developer"
	case AssigneeRoleQA:
		return "QA Tester"
	case AssigneeRoleReviewer:
		return "Reviewer"
	case AssigneeRoleOther:
		return "Other"
	default:
		return "Unknown"
	}
}

// NewIssueAssignee creates a new issue assignee
func NewIssueAssignee(issueID, userID uuid.UUID, role AssigneeRole) *IssueAssignee {
	return &IssueAssignee{
		ID:         uuid.New(),
		IssueID:    issueID,
		UserID:     userID,
		Role:       role,
		AssignedAt: time.Now(),
	}
}

// IsAssignedToUser checks if the issue is assigned to a specific user with any role
func IsAssignedToUser(assignees []*IssueAssignee, userID uuid.UUID) bool {
	for _, assignee := range assignees {
		if assignee.UserID == userID {
			return true
		}
	}
	return false
}

// GetAssigneesByRole filters assignees by role
func GetAssigneesByRole(assignees []*IssueAssignee, role AssigneeRole) []*IssueAssignee {
	var filtered []*IssueAssignee
	for _, assignee := range assignees {
		if assignee.Role == role {
			filtered = append(filtered, assignee)
		}
	}
	return filtered
}

// GetUsersByRole extracts users with a specific role from assignees
func GetUsersByRole(assignees []*IssueAssignee, role AssigneeRole) []User {
	var users []User
	for _, assignee := range assignees {
		if assignee.Role == role {
			users = append(users, assignee.User)
		}
	}
	return users
}
