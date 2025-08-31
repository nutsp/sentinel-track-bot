package domain

import (
	"time"

	"github.com/google/uuid"
)

// Priority represents the priority level of an issue
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// Status represents the status of an issue
type Status string

const (
	StatusOpen        Status = "open"         // 1. Initial status when issue is created
	StatusAssignedDev Status = "assigned_dev" // 2. Assigned to developer
	StatusInProgress  Status = "in_progress"  // 3. Developer is working on it
	StatusResolved    Status = "resolved"     // 4. Developer marked as resolved
	StatusAssignedQA  Status = "assigned_qa"  // 5. Assigned to QA for testing
	StatusVerified    Status = "verified"     // 6. QA verified the fix
	StatusClosed      Status = "closed"       // 7. Issue is closed
	StatusRejected    Status = "rejected"     // QA rejected the fix (back to dev)
	StatusReopened    Status = "reopened"     // Issue was reopened
)

// Source represents the source of an issue
type Source string

const (
	SourceWeb     Source = "web"
	SourceDiscord Source = "discord"
)

// Issue represents a bug report or feature request
type Issue struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID        uuid.UUID  `json:"project_id" gorm:"type:uuid;not null"`  // Always required - main relationship
	ChannelID        *uuid.UUID `json:"channel_id,omitempty" gorm:"type:uuid"` // Optional - only for Discord issues
	ReporterID       uuid.UUID  `json:"reporter_id" gorm:"type:uuid;not null"`
	AssigneeID       *uuid.UUID `json:"assignee_id,omitempty" gorm:"type:uuid"`
	Title            string     `json:"title" gorm:"not null;size:255"`
	Description      string     `json:"description" gorm:"not null;type:text"`
	ImageURL         string     `json:"image_url,omitempty" gorm:"size:500"`
	Priority         Priority   `json:"priority" gorm:"size:10;default:'medium'"`
	Status           Status     `json:"status" gorm:"size:10;default:'open'"`
	Source           string     `json:"source" gorm:"size:20;default:'web'"`               // 'discord' or 'web'
	ThreadID         string     `json:"thread_id,omitempty" gorm:"size:100"`               // Discord thread ID (optional)
	MessageID        string     `json:"message_id,omitempty" gorm:"size:100"`              // Discord message ID (optional)
	PublicHash       string     `json:"public_hash,omitempty" gorm:"size:100;uniqueIndex"` // For public links
	ResolutionCause  string     `json:"resolution_cause,omitempty" gorm:"size:255"`        // For resolution cause
	ResolutionAction string     `json:"resolution_action,omitempty" gorm:"size:255"`       // For resolution action
	CreatedAt        time.Time  `json:"created_at" gorm:"type:timestamptz;default:now()"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"type:timestamptz;default:now()"`
	ClosedAt         *time.Time `json:"closed_at,omitempty" gorm:"type:timestamptz"`

	// Relationships
	Project    Project          `json:"project,omitempty" gorm:"foreignKey:ProjectID"` // Main relationship
	Channel    *Channel         `json:"channel,omitempty" gorm:"foreignKey:ChannelID"` // Optional Discord channel (UUID → channels.id)
	Reporter   User             `json:"reporter,omitempty" gorm:"foreignKey:ReporterID"`
	Assignee   *User            `json:"assignee,omitempty" gorm:"foreignKey:AssigneeID"` // Legacy single assignee (deprecated)
	Assignees  []IssueAssignee  `json:"assignees,omitempty" gorm:"foreignKey:IssueID"`   // New multi-assignee with roles
	StatusLogs []IssueStatusLog `json:"status_logs,omitempty" gorm:"foreignKey:IssueID"` // Status change history
}

// TableName specifies the table name for Issue
func (Issue) TableName() string {
	return "issues"
}

// IsValidPriority checks if the given priority is valid
func IsValidPriority(p Priority) bool {
	return p == PriorityLow || p == PriorityMedium || p == PriorityHigh
}

// IsValidStatus checks if the given status is valid
func IsValidStatus(s Status) bool {
	return s == StatusOpen || s == StatusClosed
}

// IsValidSource checks if the given source is valid
func IsValidSource(s Source) bool {
	return s == SourceWeb || s == SourceDiscord
}

// IsDiscordIssue checks if the issue was created from Discord
func (i *Issue) IsDiscordIssue() bool {
	return i.Source == string(SourceDiscord) && i.ChannelID != nil
}

// IsWebIssue checks if the issue was created from web
func (i *Issue) IsWebIssue() bool {
	return i.Source == string(SourceWeb)
}

// Close marks the issue as closed
func (i *Issue) Close() {
	i.Status = StatusClosed
	now := time.Now()
	i.ClosedAt = &now
}

// Reopen marks the issue as open
func (i *Issue) Reopen() {
	i.Status = StatusOpen
	i.ClosedAt = nil
}

// GenerateIssueID generates a human-readable issue ID
func GenerateIssueID() string {
	now := time.Now()
	return "ISS-" + now.Format("2006") + "-" + uuid.New().String()[:8]
}

// GetAssigneesByRole returns assignees with a specific role
func (i *Issue) GetAssigneesByRole(role AssigneeRole) []IssueAssignee {
	var assignees []IssueAssignee
	for _, assignee := range i.Assignees {
		if assignee.Role == role {
			assignees = append(assignees, assignee)
		}
	}
	return assignees
}

// GetDevelopers returns all developers assigned to this issue
func (i *Issue) GetDevelopers() []IssueAssignee {
	return i.GetAssigneesByRole(AssigneeRoleDev)
}

// GetQATesters returns all QA testers assigned to this issue
func (i *Issue) GetQATesters() []IssueAssignee {
	return i.GetAssigneesByRole(AssigneeRoleQA)
}

// GetReviewers returns all reviewers assigned to this issue
func (i *Issue) GetReviewers() []IssueAssignee {
	return i.GetAssigneesByRole(AssigneeRoleReviewer)
}

// IsAssignedToUser checks if a user is assigned to this issue with any role
func (i *Issue) IsAssignedToUser(userID uuid.UUID) bool {
	for _, assignee := range i.Assignees {
		if assignee.UserID == userID {
			return true
		}
	}
	return false
}

// IsAssignedToUserWithRole checks if a user is assigned to this issue with a specific role
func (i *Issue) IsAssignedToUserWithRole(userID uuid.UUID, role AssigneeRole) bool {
	for _, assignee := range i.Assignees {
		if assignee.UserID == userID && assignee.Role == role {
			return true
		}
	}
	return false
}

// GetAssigneeCount returns the total number of assignees
func (i *Issue) GetAssigneeCount() int {
	return len(i.Assignees)
}

// GetAssigneeCountByRole returns the number of assignees with a specific role
func (i *Issue) GetAssigneeCountByRole(role AssigneeRole) int {
	count := 0
	for _, assignee := range i.Assignees {
		if assignee.Role == role {
			count++
		}
	}
	return count
}

// ChangeStatus changes the issue status and creates a status log entry
func (i *Issue) ChangeStatus(newStatus Status, changedBy *uuid.UUID) *IssueStatusLog {
	oldStatus := &i.Status
	if i.Status == "" {
		oldStatus = nil
	}

	i.Status = newStatus
	i.UpdatedAt = time.Now()

	// Set ClosedAt when closing
	if newStatus == StatusClosed {
		now := time.Now()
		i.ClosedAt = &now
	} else if i.ClosedAt != nil {
		// Clear ClosedAt if reopening
		i.ClosedAt = nil
	}

	return NewIssueStatusLog(i.ID, oldStatus, newStatus, changedBy)
}

// CanTransitionTo checks if the issue can transition to the given status
func (i *Issue) CanTransitionTo(newStatus Status) bool {
	currentStatus := &i.Status
	if i.Status == "" {
		currentStatus = nil
	}
	return IsStatusTransitionValid(currentStatus, newStatus)
}

// GetWorkflowStage returns the current workflow stage (1-7)
func (i *Issue) GetWorkflowStage() int {
	return GetWorkflowStage(i.Status)
}

// GetNextPossibleStatuses returns possible next statuses for this issue
func (i *Issue) GetNextPossibleStatuses() []Status {
	return GetNextPossibleStatuses(i.Status)
}

// IsInDevPhase checks if the issue is in development phase
func (i *Issue) IsInDevPhase() bool {
	return i.Status == StatusAssignedDev || i.Status == StatusInProgress || i.Status == StatusResolved
}

// IsInQAPhase checks if the issue is in QA phase
func (i *Issue) IsInQAPhase() bool {
	return i.Status == StatusAssignedQA || i.Status == StatusVerified || i.Status == StatusRejected
}

// IsActive checks if the issue is in an active state
func (i *Issue) IsActive() bool {
	return IsActiveStatus(i.Status)
}

// IsClosed checks if the issue is closed
func (i *Issue) IsClosed() bool {
	return i.Status == StatusClosed
}

// GetStatusDisplayName returns human-readable status name
func (i *Issue) GetStatusDisplayName() string {
	return GetStatusDisplayName(i.Status)
}

// GetStatusColor returns color code for UI display
func (i *Issue) GetStatusColor() string {
	return GetStatusColor(i.Status)
}

// GetLatestStatusLog returns the most recent status change log
func (i *Issue) GetLatestStatusLog() *IssueStatusLog {
	if len(i.StatusLogs) == 0 {
		return nil
	}

	latest := &i.StatusLogs[0]
	for _, log := range i.StatusLogs {
		if log.ChangedAt.After(latest.ChangedAt) {
			latest = &log
		}
	}
	return latest
}

// GetStatusHistory returns status logs ordered by time (newest first)
func (i *Issue) GetStatusHistory() []IssueStatusLog {
	logs := make([]IssueStatusLog, len(i.StatusLogs))
	copy(logs, i.StatusLogs)

	// Sort by ChangedAt descending (newest first)
	for i := 0; i < len(logs)-1; i++ {
		for j := i + 1; j < len(logs); j++ {
			if logs[i].ChangedAt.Before(logs[j].ChangedAt) {
				logs[i], logs[j] = logs[j], logs[i]
			}
		}
	}

	return logs
}
