package service

import (
	"context"
	"fmt"

	"fix-track-bot/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// userService implements the UserService interface
type userService struct {
	userRepo     domain.UserRepository
	customerRepo domain.CustomerRepository
	logger       *zap.Logger
}

// NewUserService creates a new instance of user service
func NewUserService(userRepo domain.UserRepository, customerRepo domain.CustomerRepository, logger *zap.Logger) domain.UserService {
	return &userService{
		userRepo:     userRepo,
		customerRepo: customerRepo,
		logger:       logger,
	}
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, discordID, name, email string, customerID *uuid.UUID, role domain.UserRole) (*domain.User, error) {
	s.logger.Debug("Creating user",
		zap.String("discord_id", discordID),
		zap.String("name", name),
		zap.String("role", string(role)),
	)

	// Validate input
	if discordID == "" {
		s.logger.Debug("Invalid Discord ID", zap.String("discord_id", discordID))
		return nil, domain.ErrInvalidDiscordID
	}

	if !domain.IsValidUserRole(role) {
		s.logger.Debug("Invalid user role", zap.String("role", string(role)))
		return nil, domain.ErrInvalidUserRole
	}

	// Verify customer exists if provided
	if customerID != nil {
		_, err := s.customerRepo.GetByID(ctx, *customerID)
		if err != nil {
			s.logger.Error("Failed to verify customer exists",
				zap.Error(err),
				zap.String("customer_id", customerID.String()),
			)
			return nil, fmt.Errorf("failed to verify customer exists: %w", err)
		}
	}

	// Check if user already exists
	existing, err := s.userRepo.GetByDiscordID(ctx, discordID)
	if err != nil && err != domain.ErrUserNotFound {
		s.logger.Error("Failed to check existing user",
			zap.Error(err),
			zap.String("discord_id", discordID),
		)
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existing != nil {
		s.logger.Debug("User already exists", zap.String("discord_id", discordID))
		return nil, domain.ErrUserAlreadyExists
	}

	// Create new user
	user := &domain.User{
		ID:         uuid.New(),
		CustomerID: customerID,
		Name:       name,
		Email:      email,
		DiscordID:  discordID,
		Role:       role,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("Failed to create user",
			zap.Error(err),
			zap.String("discord_id", discordID),
		)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("User created successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("discord_id", discordID),
	)

	return user, nil
}

// GetUser retrieves a user by ID
func (s *userService) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	s.logger.Debug("Retrieving user", zap.String("user_id", id.String()))

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to retrieve user",
			zap.Error(err),
			zap.String("user_id", id.String()),
		)
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}

	s.logger.Debug("User retrieved successfully", zap.String("user_id", id.String()))
	return user, nil
}

// GetUserByDiscordID retrieves a user by Discord ID
func (s *userService) GetUserByDiscordID(ctx context.Context, discordID string) (*domain.User, error) {
	s.logger.Debug("Retrieving user by Discord ID", zap.String("discord_id", discordID))

	user, err := s.userRepo.GetByDiscordID(ctx, discordID)
	if err != nil {
		s.logger.Error("Failed to retrieve user by Discord ID",
			zap.Error(err),
			zap.String("discord_id", discordID),
		)
		return nil, fmt.Errorf("failed to retrieve user by Discord ID: %w", err)
	}

	s.logger.Debug("User retrieved successfully", zap.String("discord_id", discordID))
	return user, nil
}

// GetOrCreateUserByDiscordID gets existing user or creates new one
func (s *userService) GetOrCreateUserByDiscordID(ctx context.Context, discordID, name string) (*domain.User, error) {
	s.logger.Debug("Getting or creating user by Discord ID",
		zap.String("discord_id", discordID),
		zap.String("name", name),
	)

	// Try to get existing user
	user, err := s.userRepo.GetByDiscordID(ctx, discordID)
	if err != nil && err != domain.ErrUserNotFound {
		s.logger.Error("Failed to check existing user",
			zap.Error(err),
			zap.String("discord_id", discordID),
		)
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// If user exists, return it
	if user != nil {
		s.logger.Debug("User found", zap.String("discord_id", discordID))
		return user, nil
	}

	// Create new user with default role
	newUser := &domain.User{
		ID:        uuid.New(),
		Name:      name,
		DiscordID: discordID,
		Role:      domain.UserRoleCustomer, // Default role
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		s.logger.Error("Failed to create new user",
			zap.Error(err),
			zap.String("discord_id", discordID),
		)
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	s.logger.Info("New user created",
		zap.String("user_id", newUser.ID.String()),
		zap.String("discord_id", discordID),
	)

	return newUser, nil
}

// UpdateUser updates user information
func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, name, email string, role domain.UserRole) error {
	s.logger.Debug("Updating user",
		zap.String("user_id", id.String()),
		zap.String("name", name),
		zap.String("role", string(role)),
	)

	// Validate role
	if !domain.IsValidUserRole(role) {
		s.logger.Debug("Invalid user role", zap.String("role", string(role)))
		return domain.ErrInvalidUserRole
	}

	// Get existing user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to retrieve user for update",
			zap.Error(err),
			zap.String("user_id", id.String()),
		)
		return fmt.Errorf("failed to retrieve user for update: %w", err)
	}

	// Update fields
	user.Name = name
	user.Email = email
	user.Role = role

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update user",
			zap.Error(err),
			zap.String("user_id", id.String()),
		)
		return fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.Info("User updated successfully",
		zap.String("user_id", id.String()),
		zap.String("name", name),
	)

	return nil
}

// AssignUserToCustomer assigns a user to a customer
func (s *userService) AssignUserToCustomer(ctx context.Context, userID, customerID uuid.UUID) error {
	s.logger.Debug("Assigning user to customer",
		zap.String("user_id", userID.String()),
		zap.String("customer_id", customerID.String()),
	)

	// Verify customer exists
	_, err := s.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		s.logger.Error("Failed to verify customer exists",
			zap.Error(err),
			zap.String("customer_id", customerID.String()),
		)
		return fmt.Errorf("failed to verify customer exists: %w", err)
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to retrieve user",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return fmt.Errorf("failed to retrieve user: %w", err)
	}

	// Update user's customer assignment
	user.CustomerID = &customerID

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to assign user to customer",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("customer_id", customerID.String()),
		)
		return fmt.Errorf("failed to assign user to customer: %w", err)
	}

	s.logger.Info("User assigned to customer successfully",
		zap.String("user_id", userID.String()),
		zap.String("customer_id", customerID.String()),
	)

	return nil
}

// ListUsers lists all users
func (s *userService) ListUsers(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	s.logger.Debug("Listing users",
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	users, err := s.userRepo.List(ctx, offset, limit)
	if err != nil {
		s.logger.Error("Failed to list users",
			zap.Error(err),
			zap.Int("offset", offset),
			zap.Int("limit", limit),
		)
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	s.logger.Debug("Users listed successfully",
		zap.Int("count", len(users)),
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	return users, nil
}
