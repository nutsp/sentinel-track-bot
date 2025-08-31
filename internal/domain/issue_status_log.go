package domain

import (
	"time"

	"github.com/google/uuid"
)

// IssueStatusLog represents a status change log entry for an issue
type IssueStatusLog struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	IssueID   uuid.UUID  `json:"issue_id" gorm:"type:uuid;not null"`
	OldStatus *Status    `json:"old_status,omitempty" gorm:"size:20"`
	NewStatus Status     `json:"new_status" gorm:"size:20;not null"`
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
			StatusAssignedDev, // Assign to developer
			StatusClosed,      // Close without assignment
		},
		StatusAssignedDev: {
			StatusInProgress, // Developer starts working
			StatusOpen,       // Unassign developer
			StatusClosed,     // Close without work
		},
		StatusInProgress: {
			StatusResolved,    // Developer completes work
			StatusAssignedDev, // Back to assigned (pause work)
			StatusOpen,        // Unassign completely
		},
		StatusResolved: {
			StatusAssignedQA, // Assign to QA for testing
			StatusClosed,     // Close without QA (direct close)
			StatusInProgress, // Developer continues work
		},
		StatusAssignedQA: {
			StatusVerified, // QA approves
			StatusRejected, // QA rejects
			StatusResolved, // Unassign QA
		},
		StatusVerified: {
			StatusClosed,   // Close after QA approval
			StatusRejected, // Revert QA decision
		},
		StatusRejected: {
			StatusAssignedDev, // Reassign to developer
			StatusInProgress,  // Developer continues
			StatusOpen,        // Back to open
		},
		StatusClosed: {
			StatusReopened, // Reopen closed issue
		},
		StatusReopened: {
			StatusOpen,        // Back to open status
			StatusAssignedDev, // Direct assign to dev
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
	case StatusAssignedDev:
		return "Assigned to Developer"
	case StatusInProgress:
		return "In Progress"
	case StatusResolved:
		return "Resolved"
	case StatusAssignedQA:
		return "Assigned to QA"
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
	case StatusAssignedDev:
		return "#17a2b8" // Info blue
	case StatusInProgress:
		return "#ffc107" // Warning yellow
	case StatusResolved:
		return "#28a745" // Success green
	case StatusAssignedQA:
		return "#007bff" // Primary blue
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
		StatusOpen:        {StatusAssignedDev, StatusClosed},
		StatusAssignedDev: {StatusInProgress, StatusOpen, StatusClosed},
		StatusInProgress:  {StatusResolved, StatusAssignedDev, StatusOpen},
		StatusResolved:    {StatusAssignedQA, StatusClosed, StatusInProgress},
		StatusAssignedQA:  {StatusVerified, StatusRejected, StatusResolved},
		StatusVerified:    {StatusClosed, StatusRejected},
		StatusRejected:    {StatusAssignedDev, StatusInProgress, StatusOpen},
		StatusClosed:      {StatusReopened},
		StatusReopened:    {StatusOpen, StatusAssignedDev},
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
	case StatusAssignedDev:
		return 2
	case StatusInProgress:
		return 3
	case StatusResolved:
		return 4
	case StatusAssignedQA:
		return 5
	case StatusVerified:
		return 6
	case StatusClosed:
		return 7
	default:
		return 0 // Unknown/special status
	}
}
