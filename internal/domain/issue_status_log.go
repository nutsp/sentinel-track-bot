package domain

import (
	"time"

	"github.com/google/uuid"
)

// IssueStatusLog represents a status change log entry for an issue
type IssueStatusLog struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	IssueID   uuid.UUID  `json:"issue_id" gorm:"type:uuid;not null"`
	OldStatus *Status    `json:"old_status,omitempty" gorm:"size:40"`
	NewStatus Status     `json:"new_status" gorm:"size:40;not null"`
	ChangedBy *uuid.UUID `json:"changed_by,omitempty" gorm:"type:uuid"`
	ChangedAt time.Time  `json:"changed_at" gorm:"type:timestamptz;default:now()"`

	// Relationships
	Issue         Issue `json:"issue,omitempty" gorm:"foreignKey:IssueID"`
	ChangedByUser *User `json:"changed_by_user,omitempty" gorm:"foreignKey:ChangedBy"`
}

// TableName specifies the table name for IssueStatusLog
func (IssueStatusLog) TableName() string {
	return "issue_status_logs"
}

// NewIssueStatusLog creates a new status log entry
func NewIssueStatusLog(issueID uuid.UUID, oldStatus *Status, newStatus Status, changedBy *uuid.UUID) *IssueStatusLog {
	return &IssueStatusLog{
		ID:        uuid.New(),
		IssueID:   issueID,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		ChangedBy: changedBy,
		ChangedAt: time.Now(),
	}
}

// IsStatusTransitionValid checks if a status transition is valid according to workflow
func IsStatusTransitionValid(from *Status, to Status) bool {
	// If no previous status (new issue), only allow "open"
	if from == nil {
		return to == StatusOpen
	}

	// Define valid transitions based on workflow
	validTransitions := map[Status][]Status{
		StatusOpen: {
			StatusClosed, // Close without assignment
		},
		StatusInProgress: {
			StatusResolved, // Developer starts working
			StatusOpen,     // Unassign developer
			StatusClosed,   // Close without work
		},
		StatusResolved: {
			StatusVerified, // Developer completes work
			StatusOpen,     // Back to assigned (pause work)
		},
		StatusVerified: {
			StatusClosed,   // QA approves
			StatusRejected, // QA rejects
			StatusResolved, // Unassign QA
		},
		StatusRejected: {
			StatusInProgress, // Reassign to developer
			StatusInProgress, // Developer continues
			StatusOpen,       // Back to open
		},
		StatusClosed: {
			StatusReopened, // Reopen closed issue
		},
		StatusReopened: {
			StatusOpen,       // Back to open status
			StatusInProgress, // Direct assign to dev
		},
	}

	allowedStatuses, exists := validTransitions[*from]
	if !exists {
		return false
	}

	for _, allowedStatus := range allowedStatuses {
		if allowedStatus == to {
			return true
		}
	}

	return false
}

// GetStatusDisplayName returns a human-readable display name for the status
func GetStatusDisplayName(status Status) string {
	switch status {
	case StatusOpen:
		return "Open"
	case StatusInProgress:
		return "In Progress"
	case StatusResolved:
		return "Resolved"
	case StatusVerified:
		return "Verified"
	case StatusClosed:
		return "Closed"
	case StatusRejected:
		return "Rejected by QA"
	case StatusReopened:
		return "Reopened"
	default:
		return string(status)
	}
}

// GetStatusColor returns a color code for UI display
func GetStatusColor(status Status) string {
	switch status {
	case StatusOpen:
		return "#6c757d" // Gray
	case StatusInProgress:
		return "#ffc107" // Warning yellow
	case StatusResolved:
		return "#28a745" // Success green
	case StatusVerified:
		return "#20c997" // Teal
	case StatusClosed:
		return "#6f42c1" // Purple
	case StatusRejected:
		return "#dc3545" // Danger red
	case StatusReopened:
		return "#fd7e14" // Orange
	default:
		return "#6c757d" // Default gray
	}
}

// IsTerminalStatus checks if the status is a terminal state
func IsTerminalStatus(status Status) bool {
	return status == StatusClosed
}

// IsActiveStatus checks if the status represents an active issue
func IsActiveStatus(status Status) bool {
	return !IsTerminalStatus(status)
}

// GetNextPossibleStatuses returns the list of possible next statuses
func GetNextPossibleStatuses(currentStatus Status) []Status {
	validTransitions := map[Status][]Status{
		StatusDraft:      {StatusOpen},
		StatusOpen:       {StatusInProgress, StatusClosed},
		StatusInProgress: {StatusResolved},
		StatusResolved:   {StatusVerified},
		StatusVerified:   {StatusClosed, StatusRejected},
		StatusRejected:   {StatusInProgress, StatusOpen},
		StatusClosed:     {StatusReopened},
		StatusReopened:   {StatusOpen},
	}

	if transitions, exists := validTransitions[currentStatus]; exists {
		return transitions
	}
	return []Status{}
}

// GetWorkflowStage returns the workflow stage number (1-7)
func GetWorkflowStage(status Status) int {
	switch status {
	case StatusOpen:
		return 1
	case StatusInProgress:
		return 3
	case StatusResolved:
		return 4
	case StatusVerified:
		return 6
	case StatusClosed:
		return 7
	default:
		return 0 // Unknown/special status
	}
}
