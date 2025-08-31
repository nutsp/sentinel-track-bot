package repository

import (
	"context"
	"fmt"

	"fix-track-bot/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// userRepository implements the UserRepository interface
type userRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserRepository creates a new instance of user repository
func NewUserRepository(db *gorm.DB, logger *zap.Logger) domain.UserRepository {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new user in the database
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	r.logger.Debug("Creating new user",
		zap.String("discord_id", user.DiscordID),
		zap.String("name", user.Name),
		zap.String("role", string(user.Role)),
	)

	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		r.logger.Error("Failed to create user",
			zap.Error(err),
			zap.String("discord_id", user.DiscordID),
		)
		return fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.Info("User created successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("discord_id", user.DiscordID),
	)

	return nil
}

// GetByID retrieves a user by its ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	r.logger.Debug("Retrieving user by ID", zap.String("user_id", id.String()))

	var user domain.User
	if err := r.db.WithContext(ctx).Preload("Customer").Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("User not found", zap.String("user_id", id.String()))
			return nil, domain.ErrUserNotFound
		}
		r.logger.Error("Failed to retrieve user",
			zap.Error(err),
			zap.String("user_id", id.String()),
		)
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}

	r.logger.Debug("User retrieved successfully", zap.String("user_id", id.String()))
	return &user, nil
}

// GetByDiscordID retrieves a user by Discord ID
func (r *userRepository) GetByDiscordID(ctx context.Context, discordID string) (*domain.User, error) {
	r.logger.Debug("Retrieving user by Discord ID", zap.String("discord_id", discordID))

	var user domain.User
	if err := r.db.WithContext(ctx).Preload("Customer").Where("discord_id = ?", discordID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("User not found", zap.String("discord_id", discordID))
			return nil, domain.ErrUserNotFound
		}
		r.logger.Error("Failed to retrieve user by Discord ID",
			zap.Error(err),
			zap.String("discord_id", discordID),
		)
		return nil, fmt.Errorf("failed to retrieve user by Discord ID: %w", err)
	}

	r.logger.Debug("User retrieved successfully", zap.String("discord_id", discordID))
	return &user, nil
}

// GetByCustomerID retrieves all users for a customer
func (r *userRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*domain.User, error) {
	r.logger.Debug("Retrieving users by customer ID", zap.String("customer_id", customerID.String()))

	var users []*domain.User
	if err := r.db.WithContext(ctx).Preload("Customer").Where("customer_id = ?", customerID).Order("created_at DESC").Find(&users).Error; err != nil {
		r.logger.Error("Failed to retrieve users by customer ID",
			zap.Error(err),
			zap.String("customer_id", customerID.String()),
		)
		return nil, fmt.Errorf("failed to retrieve users by customer ID: %w", err)
	}

	r.logger.Debug("Users retrieved successfully",
		zap.String("customer_id", customerID.String()),
		zap.Int("count", len(users)),
	)

	return users, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	r.logger.Debug("Updating user", zap.String("user_id", user.ID.String()))

	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		r.logger.Error("Failed to update user",
			zap.Error(result.Error),
			zap.String("user_id", user.ID.String()),
		)
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("User not found for update", zap.String("user_id", user.ID.String()))
		return domain.ErrUserNotFound
	}

	r.logger.Info("User updated successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("discord_id", user.DiscordID),
	)

	return nil
}

// Delete removes a user from the database
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug("Deleting user", zap.String("user_id", id.String()))

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.User{})
	if result.Error != nil {
		r.logger.Error("Failed to delete user",
			zap.Error(result.Error),
			zap.String("user_id", id.String()),
		)
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("User not found for deletion", zap.String("user_id", id.String()))
		return domain.ErrUserNotFound
	}

	r.logger.Info("User deleted successfully", zap.String("user_id", id.String()))
	return nil
}

// List retrieves all users with pagination
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	r.logger.Debug("Listing users",
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	var users []*domain.User
	if err := r.db.WithContext(ctx).Preload("Customer").Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		r.logger.Error("Failed to list users",
			zap.Error(err),
			zap.Int("offset", offset),
			zap.Int("limit", limit),
		)
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	r.logger.Debug("Users listed successfully",
		zap.Int("count", len(users)),
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	return users, nil
}
