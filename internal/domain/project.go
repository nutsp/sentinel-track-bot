package domain

import (
	"time"

	"github.com/google/uuid"
)

// Project represents a customer project
type Project struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CustomerID  uuid.UUID `json:"customer_id" gorm:"type:uuid;not null"`
	Name        string    `json:"name" gorm:"not null;size:255"`
	Description string    `json:"description,omitempty" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"type:timestamptz;default:now()"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"type:timestamptz;default:now()"`

	// Relationships
	Customer Customer  `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Channels []Channel `json:"channels,omitempty" gorm:"foreignKey:ProjectID"`
	Issues   []Issue   `json:"issues,omitempty" gorm:"foreignKey:ProjectID"`
}

// TableName specifies the table name for Project
func (Project) TableName() string {
	return "projects"
}

// IsValidProject validates project data
func IsValidProject(name string, customerID uuid.UUID) bool {
	return name != "" && customerID != uuid.Nil
}
