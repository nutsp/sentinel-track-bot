package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// IssueRepository defines the interface for issue data operations
type IssueRepository interface {
	// Create creates a new issue in the repository
	Create(ctx context.Context, issue *Issue) error

	// GetByID retrieves an issue by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*Issue, error)

	// GetByChannelID retrieves all issues for a specific channel by UUID
	GetByChannelID(ctx context.Context, channelID uuid.UUID) ([]*Issue, error)

	// GetByDiscordChannelID retrieves all issues for a specific Discord channel by string ID
	GetByDiscordChannelID(ctx context.Context, discordChannelID string) ([]*Issue, error)

	// GetByStatus retrieves all issues with a specific status
	GetByStatus(ctx context.Context, status Status) ([]*Issue, error)

	// Update updates an existing issue
	Update(ctx context.Context, issue *Issue) error

	// Delete removes an issue from the repository
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves all issues with pagination
	List(ctx context.Context, offset, limit int) ([]*Issue, error)
}

// IssueService defines the interface for issue business logic
type IssueService interface {
	// CreateIssue creates a new issue with validation
	CreateIssue(ctx context.Context, title, description, imageURL, reporterID, channelID string) (*Issue, error)

	// GetIssue retrieves an issue by ID
	GetIssue(ctx context.Context, id uuid.UUID) (*Issue, error)

	// UpdateIssuePriority updates the priority of an issue
	UpdateIssuePriority(ctx context.Context, id uuid.UUID, priority Priority) error

	// CloseIssue closes an issue
	CloseIssue(ctx context.Context, id uuid.UUID) error

	// OpenIssue opens an issue
	OpenIssue(ctx context.Context, id uuid.UUID) error

	// InProgressIssue starts working on an issue
	InProgressIssue(ctx context.Context, id uuid.UUID) error

	// VerifiedIssue verifies an issue
	VerifiedIssue(ctx context.Context, id uuid.UUID) error

	// ReopenIssue reopens a closed issue
	ReopenIssue(ctx context.Context, id uuid.UUID) error

	// ListIssuesByChannel lists all issues for a specific channel
	ListIssuesByChannel(ctx context.Context, channelID string) ([]*Issue, error)

	// ListOpenIssues lists all open issues
	ListOpenIssues(ctx context.Context) ([]*Issue, error)

	// SetThreadInfo sets the thread and message IDs for an issue
	SetThreadInfo(ctx context.Context, id uuid.UUID, threadID, messageID string) error

	// UpdateIssueMessageID updates just the message ID for an issue
	UpdateIssueMessageID(ctx context.Context, id uuid.UUID, messageID string) error

	// UpdateIssueResolved updates the resolved information for an issue
	UpdateIssueResolved(ctx context.Context, id uuid.UUID, cause string, action string) error
}

// DiscordHandler defines the interface for Discord interaction handling
type DiscordHandler interface {
	// HandleSlashCommand handles Discord slash command interactions
	HandleSlashCommand(ctx context.Context, interactionData map[string]interface{}) error

	// HandleModalSubmit handles Discord modal submit interactions
	HandleModalSubmit(ctx context.Context, interactionData map[string]interface{}) error

	// HandleButtonClick handles Discord button click interactions
	HandleButtonClick(ctx context.Context, interactionData map[string]interface{}) error

	// HandleSelectMenu handles Discord select menu interactions
	HandleSelectMenu(ctx context.Context, interactionData map[string]interface{}) error
}

// ChannelRepository defines the interface for channel registration data operations
type ChannelRepository interface {
	// Create creates a new channel registration in the repository
	Create(ctx context.Context, channel *Channel) error

	// GetByChannelID retrieves a channel registration by its Discord channel ID
	GetByChannelID(ctx context.Context, channelID string) (*Channel, error)

	// GetByGuildID retrieves all channel registrations for a specific guild
	GetByGuildID(ctx context.Context, guildID string) ([]*Channel, error)

	// Update updates an existing channel registration
	Update(ctx context.Context, channel *Channel) error

	// Delete removes a channel registration from the repository
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves all channel registrations with pagination
	List(ctx context.Context, offset, limit int) ([]*Channel, error)

	// GetActiveChannels retrieves all active channel registrations
	GetActiveChannels(ctx context.Context) ([]*Channel, error)
}

// ChannelService defines the interface for channel registration business logic
type ChannelService interface {
	// RegisterChannel registers a new channel with customer and project information
	RegisterChannel(ctx context.Context, channelID, customerName, customerEmail, projectName, projectDescription, registeredBy, userName, guildID string) (*Channel, error)

	// GetChannelRegistration retrieves a channel registration by channel ID
	GetChannelRegistration(ctx context.Context, channelID string) (*Channel, error)

	// UpdateChannelRegistration updates the customer and project information for a channel
	UpdateChannelRegistration(ctx context.Context, channelID, customerName, projectName string) error

	// DeactivateChannel deactivates a channel registration
	DeactivateChannel(ctx context.Context, channelID string) error

	// ActivateChannel activates a channel registration
	ActivateChannel(ctx context.Context, channelID string) error

	// ListChannelsForGuild lists all channel registrations for a specific guild
	ListChannelsForGuild(ctx context.Context, guildID string) ([]*Channel, error)

	// IsChannelRegistered checks if a channel is already registered
	IsChannelRegistered(ctx context.Context, channelID string) (bool, error)
}

// CustomerRepository defines the interface for customer data operations
type CustomerRepository interface {
	// Create creates a new customer in the repository
	Create(ctx context.Context, customer *Customer) error

	// GetByID retrieves a customer by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*Customer, error)

	// GetByName retrieves a customer by name
	GetByName(ctx context.Context, name string) (*Customer, error)

	// Update updates an existing customer
	Update(ctx context.Context, customer *Customer) error

	// Delete removes a customer from the repository
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves all customers with pagination
	List(ctx context.Context, offset, limit int) ([]*Customer, error)
}

// ProjectRepository defines the interface for project data operations
type ProjectRepository interface {
	// Create creates a new project in the repository
	Create(ctx context.Context, project *Project) error

	// GetByID retrieves a project by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*Project, error)

	// GetByCustomerID retrieves all projects for a customer
	GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*Project, error)

	// GetByName retrieves a project by name and customer ID
	GetByName(ctx context.Context, customerID uuid.UUID, name string) (*Project, error)

	// Update updates an existing project
	Update(ctx context.Context, project *Project) error

	// Delete removes a project from the repository
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves all projects with pagination
	List(ctx context.Context, offset, limit int) ([]*Project, error)
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create creates a new user in the repository
	Create(ctx context.Context, user *User) error

	// GetByID retrieves a user by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)

	// GetByDiscordID retrieves a user by Discord ID
	GetByDiscordID(ctx context.Context, discordID string) (*User, error)

	// GetByCustomerID retrieves all users for a customer
	GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *User) error

	// Delete removes a user from the repository
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves all users with pagination
	List(ctx context.Context, offset, limit int) ([]*User, error)
}

// CustomerService defines the interface for customer business logic
type CustomerService interface {
	// CreateCustomer creates a new customer
	CreateCustomer(ctx context.Context, name, contactEmail string) (*Customer, error)

	// GetCustomer retrieves a customer by ID
	GetCustomer(ctx context.Context, id uuid.UUID) (*Customer, error)

	// GetCustomerByName retrieves a customer by name
	GetCustomerByName(ctx context.Context, name string) (*Customer, error)

	// UpdateCustomer updates customer information
	UpdateCustomer(ctx context.Context, id uuid.UUID, name, contactEmail string) error

	// ListCustomers lists all customers
	ListCustomers(ctx context.Context, offset, limit int) ([]*Customer, error)
}

// ProjectService defines the interface for project business logic
type ProjectService interface {
	// CreateProject creates a new project for a customer
	CreateProject(ctx context.Context, customerID uuid.UUID, name, description string) (*Project, error)

	// GetProject retrieves a project by ID
	GetProject(ctx context.Context, id uuid.UUID) (*Project, error)

	// GetProjectsByCustomer retrieves all projects for a customer
	GetProjectsByCustomer(ctx context.Context, customerID uuid.UUID) ([]*Project, error)

	// UpdateProject updates project information
	UpdateProject(ctx context.Context, id uuid.UUID, name, description string) error

	// ListProjects lists all projects
	ListProjects(ctx context.Context, offset, limit int) ([]*Project, error)
}

// UserService defines the interface for user business logic
type UserService interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, discordID, name, email string, customerID *uuid.UUID, role UserRole) (*User, error)

	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, id uuid.UUID) (*User, error)

	// GetUserByDiscordID retrieves a user by Discord ID
	GetUserByDiscordID(ctx context.Context, discordID string) (*User, error)

	// GetOrCreateUserByDiscordID gets existing user or creates new one
	GetOrCreateUserByDiscordID(ctx context.Context, discordID, name string) (*User, error)

	// UpdateUser updates user information
	UpdateUser(ctx context.Context, id uuid.UUID, name, email string, role UserRole) error

	// AssignUserToCustomer assigns a user to a customer
	AssignUserToCustomer(ctx context.Context, userID, customerID uuid.UUID) error

	// ListUsers lists all users
	ListUsers(ctx context.Context, offset, limit int) ([]*User, error)
}

// IssueAssigneeRepository defines the interface for issue assignee data access
type IssueAssigneeRepository interface {
	Create(ctx context.Context, assignee *IssueAssignee) error
	GetByID(ctx context.Context, id uuid.UUID) (*IssueAssignee, error)
	GetByIssueID(ctx context.Context, issueID uuid.UUID) ([]*IssueAssignee, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*IssueAssignee, error)
	GetByIssueAndUser(ctx context.Context, issueID, userID uuid.UUID) ([]*IssueAssignee, error)
	GetByIssueAndRole(ctx context.Context, issueID uuid.UUID, role AssigneeRole) ([]*IssueAssignee, error)
	Update(ctx context.Context, assignee *IssueAssignee) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByIssueAndUser(ctx context.Context, issueID, userID uuid.UUID) error
	DeleteByIssueAndUserAndRole(ctx context.Context, issueID, userID uuid.UUID, role AssigneeRole) error
}

// IssueAssigneeService defines the interface for issue assignee business logic
type IssueAssigneeService interface {
	AssignUserToIssue(ctx context.Context, issueID uuid.UUID, discordID string, role AssigneeRole) (*IssueAssignee, error)
	UnassignUserFromIssue(ctx context.Context, issueID, userID uuid.UUID, role AssigneeRole) error
	UnassignAllUsersFromIssue(ctx context.Context, issueID uuid.UUID) error
	GetIssueAssignees(ctx context.Context, issueID uuid.UUID) ([]*IssueAssignee, error)
	GetUserAssignments(ctx context.Context, userID uuid.UUID) ([]*IssueAssignee, error)
	GetAssigneesByRole(ctx context.Context, issueID uuid.UUID, role AssigneeRole) ([]*IssueAssignee, error)
	IsUserAssignedToIssue(ctx context.Context, issueID, userID uuid.UUID) (bool, error)
	IsUserAssignedWithRole(ctx context.Context, issueID, userID uuid.UUID, role AssigneeRole) (bool, error)
}

// IssueStatusLogRepository defines the interface for issue status log data access
type IssueStatusLogRepository interface {
	Create(ctx context.Context, log *IssueStatusLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*IssueStatusLog, error)
	GetByIssueID(ctx context.Context, issueID uuid.UUID) ([]*IssueStatusLog, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*IssueStatusLog, error)
	GetRecentLogs(ctx context.Context, limit int) ([]*IssueStatusLog, error)
	GetLogsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*IssueStatusLog, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// IssueStatusLogService defines the interface for issue status log business logic
type IssueStatusLogService interface {
	LogStatusChange(ctx context.Context, issueID uuid.UUID, oldStatus *Status, newStatus Status, changedBy *uuid.UUID) (*IssueStatusLog, error)
	GetIssueStatusHistory(ctx context.Context, issueID uuid.UUID) ([]*IssueStatusLog, error)
	GetUserStatusChanges(ctx context.Context, userID uuid.UUID) ([]*IssueStatusLog, error)
	GetRecentStatusChanges(ctx context.Context, limit int) ([]*IssueStatusLog, error)
	ValidateStatusTransition(ctx context.Context, issueID uuid.UUID, newStatus Status) error
}
