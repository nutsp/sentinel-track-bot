package domain

import (
	"time"

	"github.com/google/uuid"
)

// Customer represents a customer organization
type Customer struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name         string    `json:"name" gorm:"not null;size:255"`
	ContactEmail string    `json:"contact_email,omitempty" gorm:"size:255"`
	CreatedAt    time.Time `json:"created_at" gorm:"type:timestamptz;default:now()"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"type:timestamptz;default:now()"`

	// Relationships
	Projects []Project `json:"projects,omitempty" gorm:"foreignKey:CustomerID"`
	Users    []User    `json:"users,omitempty" gorm:"foreignKey:CustomerID"`
}

// TableName specifies the table name for Customer
func (Customer) TableName() string {
	return "customers"
}

// IsValidCustomer validates customer data
func IsValidCustomer(name string) bool {
	return name != ""
}
