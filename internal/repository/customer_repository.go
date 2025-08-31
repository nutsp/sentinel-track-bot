package repository

import (
	"context"
	"fmt"

	"fix-track-bot/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// customerRepository implements the CustomerRepository interface
type customerRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewCustomerRepository creates a new instance of customer repository
func NewCustomerRepository(db *gorm.DB, logger *zap.Logger) domain.CustomerRepository {
	return &customerRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new customer in the database
func (r *customerRepository) Create(ctx context.Context, customer *domain.Customer) error {
	r.logger.Debug("Creating new customer",
		zap.String("name", customer.Name),
		zap.String("contact_email", customer.ContactEmail),
	)

	if err := r.db.WithContext(ctx).Create(customer).Error; err != nil {
		r.logger.Error("Failed to create customer",
			zap.Error(err),
			zap.String("name", customer.Name),
		)
		return fmt.Errorf("failed to create customer: %w", err)
	}

	r.logger.Info("Customer created successfully",
		zap.String("customer_id", customer.ID.String()),
		zap.String("name", customer.Name),
	)

	return nil
}

// GetByID retrieves a customer by its ID
func (r *customerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	r.logger.Debug("Retrieving customer by ID", zap.String("customer_id", id.String()))

	var customer domain.Customer
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&customer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("Customer not found", zap.String("customer_id", id.String()))
			return nil, domain.ErrCustomerNotFound
		}
		r.logger.Error("Failed to retrieve customer",
			zap.Error(err),
			zap.String("customer_id", id.String()),
		)
		return nil, fmt.Errorf("failed to retrieve customer: %w", err)
	}

	r.logger.Debug("Customer retrieved successfully", zap.String("customer_id", id.String()))
	return &customer, nil
}

// GetByName retrieves a customer by name
func (r *customerRepository) GetByName(ctx context.Context, name string) (*domain.Customer, error) {
	r.logger.Debug("Retrieving customer by name", zap.String("name", name))

	var customer domain.Customer
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&customer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("Customer not found", zap.String("name", name))
			return nil, domain.ErrCustomerNotFound
		}
		r.logger.Error("Failed to retrieve customer by name",
			zap.Error(err),
			zap.String("name", name),
		)
		return nil, fmt.Errorf("failed to retrieve customer by name: %w", err)
	}

	r.logger.Debug("Customer retrieved successfully", zap.String("name", name))
	return &customer, nil
}

// Update updates an existing customer
func (r *customerRepository) Update(ctx context.Context, customer *domain.Customer) error {
	r.logger.Debug("Updating customer", zap.String("customer_id", customer.ID.String()))

	result := r.db.WithContext(ctx).Save(customer)
	if result.Error != nil {
		r.logger.Error("Failed to update customer",
			zap.Error(result.Error),
			zap.String("customer_id", customer.ID.String()),
		)
		return fmt.Errorf("failed to update customer: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("Customer not found for update", zap.String("customer_id", customer.ID.String()))
		return domain.ErrCustomerNotFound
	}

	r.logger.Info("Customer updated successfully",
		zap.String("customer_id", customer.ID.String()),
		zap.String("name", customer.Name),
	)

	return nil
}

// Delete removes a customer from the database
func (r *customerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug("Deleting customer", zap.String("customer_id", id.String()))

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Customer{})
	if result.Error != nil {
		r.logger.Error("Failed to delete customer",
			zap.Error(result.Error),
			zap.String("customer_id", id.String()),
		)
		return fmt.Errorf("failed to delete customer: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("Customer not found for deletion", zap.String("customer_id", id.String()))
		return domain.ErrCustomerNotFound
	}

	r.logger.Info("Customer deleted successfully", zap.String("customer_id", id.String()))
	return nil
}

// List retrieves all customers with pagination
func (r *customerRepository) List(ctx context.Context, offset, limit int) ([]*domain.Customer, error) {
	r.logger.Debug("Listing customers",
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	var customers []*domain.Customer
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("created_at DESC").Find(&customers).Error; err != nil {
		r.logger.Error("Failed to list customers",
			zap.Error(err),
			zap.Int("offset", offset),
			zap.Int("limit", limit),
		)
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}

	r.logger.Debug("Customers listed successfully",
		zap.Int("count", len(customers)),
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	return customers, nil
}
