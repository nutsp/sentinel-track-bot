package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleCustomer UserRole = "customer"
	UserRoleSupport  UserRole = "support"
	UserRoleAdmin    UserRole = "admin"
)

// User represents a system user
type User struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CustomerID *uuid.UUID `json:"customer_id,omitempty" gorm:"type:uuid"`
	Name       string     `json:"name,omitempty" gorm:"size:255"`
	Email      string     `json:"email,omitempty" gorm:"size:255"`
	DiscordID  string     `json:"discord_id,omitempty" gorm:"size:100;uniqueIndex"`
	Role       UserRole   `json:"role" gorm:"size:20;default:'customer'"`
	IsInternal bool       `json:"is_internal" gorm:"default:false"`
	CreatedAt  time.Time  `json:"created_at" gorm:"type:timestamptz;default:now()"`

	// Relationships
	Customer           *Customer `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	ReportedIssues     []Issue   `json:"reported_issues,omitempty" gorm:"foreignKey:ReporterID"`
	AssignedIssues     []Issue   `json:"assigned_issues,omitempty" gorm:"foreignKey:AssigneeID"`
	RegisteredChannels []Channel `json:"registered_channels,omitempty" gorm:"foreignKey:RegisteredBy"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// IsValidUserRole checks if the given role is valid
func IsValidUserRole(role UserRole) bool {
	return role == UserRoleCustomer || role == UserRoleSupport || role == UserRoleAdmin
}

// IsCustomerUser checks if the user belongs to a customer
func (u *User) IsCustomerUser() bool {
	return u.CustomerID != nil
}

// CanManageProject checks if user can manage a specific project
func (u *User) CanManageProject(projectCustomerID uuid.UUID) bool {
	switch u.Role {
	case UserRoleAdmin, UserRoleSupport:
		return true
	case UserRoleCustomer:
		return u.CustomerID != nil && *u.CustomerID == projectCustomerID
	default:
		return false
	}
}
