package domain

import "errors"

var (
	// ErrIssueNotFound is returned when an issue is not found
	ErrIssueNotFound = errors.New("issue not found")

	// ErrInvalidPriority is returned when an invalid priority is provided
	ErrInvalidPriority = errors.New("invalid priority level")

	// ErrInvalidStatus is returned when an invalid status is provided
	ErrInvalidStatus = errors.New("invalid status")

	// ErrEmptyTitle is returned when an empty title is provided
	ErrEmptyTitle = errors.New("issue title cannot be empty")

	// ErrEmptyDescription is returned when an empty description is provided
	ErrEmptyDescription = errors.New("issue description cannot be empty")

	// ErrEmptyReporterID is returned when an empty reporter ID is provided
	ErrEmptyReporterID = errors.New("reporter ID cannot be empty")

	// ErrEmptyChannelID is returned when an empty channel ID is provided
	ErrEmptyChannelID = errors.New("channel ID cannot be empty")

	// ErrIssueAlreadyClosed is returned when trying to close an already closed issue
	ErrIssueAlreadyClosed = errors.New("issue is already closed")

	// ErrIssueAlreadyOpen is returned when trying to reopen an already open issue
	ErrIssueAlreadyOpen = errors.New("issue is already open")

	// Channel-related errors

	// ErrChannelNotFound is returned when a channel registration is not found
	ErrChannelNotFound = errors.New("channel registration not found")

	// ErrChannelAlreadyRegistered is returned when trying to register an already registered channel
	ErrChannelAlreadyRegistered = errors.New("channel is already registered")

	// ErrInvalidChannelRegistration is returned when channel registration data is invalid
	ErrInvalidChannelRegistration = errors.New("invalid channel registration data")

	// ErrEmptyCustomerName is returned when an empty customer name is provided
	ErrEmptyCustomerName = errors.New("customer name cannot be empty")

	// ErrEmptyProjectName is returned when an empty project name is provided
	ErrEmptyProjectName = errors.New("project name cannot be empty")

	// ErrEmptyGuildID is returned when an empty guild ID is provided
	ErrEmptyGuildID = errors.New("guild ID cannot be empty")

	// Customer-related errors

	// ErrCustomerNotFound is returned when a customer is not found
	ErrCustomerNotFound = errors.New("customer not found")

	// ErrCustomerAlreadyExists is returned when trying to create a duplicate customer
	ErrCustomerAlreadyExists = errors.New("customer already exists")

	// Project-related errors

	// ErrProjectNotFound is returned when a project is not found
	ErrProjectNotFound = errors.New("project not found")

	// ErrProjectAlreadyExists is returned when trying to create a duplicate project
	ErrProjectAlreadyExists = errors.New("project already exists")

	// User-related errors

	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists is returned when trying to create a duplicate user
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrInvalidUserRole is returned when an invalid user role is provided
	ErrInvalidUserRole = errors.New("invalid user role")

	// ErrInvalidDiscordID is returned when an invalid Discord ID is provided
	ErrInvalidDiscordID = errors.New("invalid Discord ID")

	// ErrUnauthorized is returned when a user lacks permission for an action
	ErrUnauthorized = errors.New("unauthorized access")

	// Issue assignee errors
	ErrAssigneeNotFound      = errors.New("assignee not found")
	ErrAssigneeAlreadyExists = errors.New("assignee already exists")
	ErrInvalidAssigneeRole   = errors.New("invalid assignee role")
)
