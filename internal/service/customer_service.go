package service

import (
	"context"
	"fmt"

	"fix-track-bot/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// customerService implements the CustomerService interface
type customerService struct {
	customerRepo domain.CustomerRepository
	logger       *zap.Logger
}

// NewCustomerService creates a new instance of customer service
func NewCustomerService(customerRepo domain.CustomerRepository, logger *zap.Logger) domain.CustomerService {
	return &customerService{
		customerRepo: customerRepo,
		logger:       logger,
	}
}

// CreateCustomer creates a new customer
func (s *customerService) CreateCustomer(ctx context.Context, name, contactEmail string) (*domain.Customer, error) {
	s.logger.Debug("Creating customer",
		zap.String("name", name),
		zap.String("contact_email", contactEmail),
	)

	// Validate input
	if !domain.IsValidCustomer(name) {
		s.logger.Debug("Invalid customer data", zap.String("name", name))
		return nil, domain.ErrEmptyCustomerName
	}

	// Check if customer already exists
	existing, err := s.customerRepo.GetByName(ctx, name)
	if err != nil && err != domain.ErrCustomerNotFound {
		s.logger.Error("Failed to check existing customer",
			zap.Error(err),
			zap.String("name", name),
		)
		return nil, fmt.Errorf("failed to check existing customer: %w", err)
	}

	if existing != nil {
		s.logger.Debug("Customer already exists", zap.String("name", name))
		return nil, domain.ErrCustomerAlreadyExists
	}

	// Create new customer
	customer := &domain.Customer{
		ID:           uuid.New(),
		Name:         name,
		ContactEmail: contactEmail,
	}

	if err := s.customerRepo.Create(ctx, customer); err != nil {
		s.logger.Error("Failed to create customer",
			zap.Error(err),
			zap.String("name", name),
		)
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	s.logger.Info("Customer created successfully",
		zap.String("customer_id", customer.ID.String()),
		zap.String("name", name),
	)

	return customer, nil
}

// GetCustomer retrieves a customer by ID
func (s *customerService) GetCustomer(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	s.logger.Debug("Retrieving customer", zap.String("customer_id", id.String()))

	customer, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to retrieve customer",
			zap.Error(err),
			zap.String("customer_id", id.String()),
		)
		return nil, fmt.Errorf("failed to retrieve customer: %w", err)
	}

	s.logger.Debug("Customer retrieved successfully", zap.String("customer_id", id.String()))
	return customer, nil
}

// GetCustomerByName retrieves a customer by name
func (s *customerService) GetCustomerByName(ctx context.Context, name string) (*domain.Customer, error) {
	s.logger.Debug("Retrieving customer by name", zap.String("name", name))

	customer, err := s.customerRepo.GetByName(ctx, name)
	if err != nil {
		s.logger.Error("Failed to retrieve customer by name",
			zap.Error(err),
			zap.String("name", name),
		)
		return nil, fmt.Errorf("failed to retrieve customer by name: %w", err)
	}

	s.logger.Debug("Customer retrieved successfully", zap.String("name", name))
	return customer, nil
}

// UpdateCustomer updates customer information
func (s *customerService) UpdateCustomer(ctx context.Context, id uuid.UUID, name, contactEmail string) error {
	s.logger.Debug("Updating customer",
		zap.String("customer_id", id.String()),
		zap.String("name", name),
	)

	// Validate input
	if !domain.IsValidCustomer(name) {
		s.logger.Debug("Invalid customer data", zap.String("name", name))
		return domain.ErrEmptyCustomerName
	}

	// Get existing customer
	customer, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to retrieve customer for update",
			zap.Error(err),
			zap.String("customer_id", id.String()),
		)
		return fmt.Errorf("failed to retrieve customer for update: %w", err)
	}

	// Update fields
	customer.Name = name
	customer.ContactEmail = contactEmail

	if err := s.customerRepo.Update(ctx, customer); err != nil {
		s.logger.Error("Failed to update customer",
			zap.Error(err),
			zap.String("customer_id", id.String()),
		)
		return fmt.Errorf("failed to update customer: %w", err)
	}

	s.logger.Info("Customer updated successfully",
		zap.String("customer_id", id.String()),
		zap.String("name", name),
	)

	return nil
}

// ListCustomers lists all customers
func (s *customerService) ListCustomers(ctx context.Context, offset, limit int) ([]*domain.Customer, error) {
	s.logger.Debug("Listing customers",
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	customers, err := s.customerRepo.List(ctx, offset, limit)
	if err != nil {
		s.logger.Error("Failed to list customers",
			zap.Error(err),
			zap.Int("offset", offset),
			zap.Int("limit", limit),
		)
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}

	s.logger.Debug("Customers listed successfully",
		zap.Int("count", len(customers)),
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	return customers, nil
}
