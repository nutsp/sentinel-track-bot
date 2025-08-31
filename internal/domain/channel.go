package domain

import (
	"time"

	"github.com/google/uuid"
)

// Channel represents a registered Discord channel with customer and project information
type Channel struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID        uuid.UUID `json:"project_id" gorm:"type:uuid;not null"`
	DiscordChannelID string    `json:"discord_channel_id" gorm:"column:discord_channel_id;not null;size:100;uniqueIndex:unique_channel"`
	GuildID          string    `json:"guild_id" gorm:"not null;size:100"`
	RegisteredBy     uuid.UUID `json:"registered_by" gorm:"type:uuid;not null"`
	IsActive         bool      `json:"is_active" gorm:"default:true"`
	ChannelType      string    `json:"channel_type" gorm:"size:100"`
	CreatedAt        time.Time `json:"created_at" gorm:"type:timestamptz;default:now()"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"type:timestamptz;default:now()"`

	// Relationships
	Project          Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	RegisteredByUser User    `json:"registered_by_user,omitempty" gorm:"foreignKey:RegisteredBy"`
}

// TableName specifies the table name for Channel
func (Channel) TableName() string {
	return "channels"
}

// IsValidChannelRegistration validates the channel registration data
func IsValidChannelRegistration(projectID uuid.UUID, discordChannelID, guildID string, registeredBy uuid.UUID) bool {
	return projectID != uuid.Nil && discordChannelID != "" && guildID != "" && registeredBy != uuid.Nil
}

// Deactivate marks the channel registration as inactive
func (c *Channel) Deactivate() {
	c.IsActive = false
}

// Activate marks the channel registration as active
func (c *Channel) Activate() {
	c.IsActive = true
}

// Update updates the channel registration information
func (c *Channel) Update(projectID uuid.UUID) {
	c.ProjectID = projectID
}
