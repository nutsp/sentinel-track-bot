package repository

import (
	"context"
	"fmt"

	"fix-track-bot/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// projectRepository implements the ProjectRepository interface
type projectRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewProjectRepository creates a new instance of project repository
func NewProjectRepository(db *gorm.DB, logger *zap.Logger) domain.ProjectRepository {
	return &projectRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new project in the database
func (r *projectRepository) Create(ctx context.Context, project *domain.Project) error {
	r.logger.Debug("Creating new project",
		zap.String("name", project.Name),
		zap.String("customer_id", project.CustomerID.String()),
	)

	if err := r.db.WithContext(ctx).Create(project).Error; err != nil {
		r.logger.Error("Failed to create project",
			zap.Error(err),
			zap.String("name", project.Name),
		)
		return fmt.Errorf("failed to create project: %w", err)
	}

	r.logger.Info("Project created successfully",
		zap.String("project_id", project.ID.String()),
		zap.String("name", project.Name),
	)

	return nil
}

// GetByID retrieves a project by its ID
func (r *projectRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	r.logger.Debug("Retrieving project by ID", zap.String("project_id", id.String()))

	var project domain.Project
	if err := r.db.WithContext(ctx).Preload("Customer").Where("id = ?", id).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("Project not found", zap.String("project_id", id.String()))
			return nil, domain.ErrProjectNotFound
		}
		r.logger.Error("Failed to retrieve project",
			zap.Error(err),
			zap.String("project_id", id.String()),
		)
		return nil, fmt.Errorf("failed to retrieve project: %w", err)
	}

	r.logger.Debug("Project retrieved successfully", zap.String("project_id", id.String()))
	return &project, nil
}

// GetByCustomerID retrieves all projects for a customer
func (r *projectRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*domain.Project, error) {
	r.logger.Debug("Retrieving projects by customer ID", zap.String("customer_id", customerID.String()))

	var projects []*domain.Project
	if err := r.db.WithContext(ctx).Preload("Customer").Where("customer_id = ?", customerID).Order("created_at DESC").Find(&projects).Error; err != nil {
		r.logger.Error("Failed to retrieve projects by customer ID",
			zap.Error(err),
			zap.String("customer_id", customerID.String()),
		)
		return nil, fmt.Errorf("failed to retrieve projects by customer ID: %w", err)
	}

	r.logger.Debug("Projects retrieved successfully",
		zap.String("customer_id", customerID.String()),
		zap.Int("count", len(projects)),
	)

	return projects, nil
}

// GetByName retrieves a project by name and customer ID
func (r *projectRepository) GetByName(ctx context.Context, customerID uuid.UUID, name string) (*domain.Project, error) {
	r.logger.Debug("Retrieving project by name",
		zap.String("customer_id", customerID.String()),
		zap.String("name", name),
	)

	var project domain.Project
	if err := r.db.WithContext(ctx).Preload("Customer").Where("customer_id = ? AND name = ?", customerID, name).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("Project not found",
				zap.String("customer_id", customerID.String()),
				zap.String("name", name),
			)
			return nil, domain.ErrProjectNotFound
		}
		r.logger.Error("Failed to retrieve project by name",
			zap.Error(err),
			zap.String("customer_id", customerID.String()),
			zap.String("name", name),
		)
		return nil, fmt.Errorf("failed to retrieve project by name: %w", err)
	}

	r.logger.Debug("Project retrieved successfully",
		zap.String("customer_id", customerID.String()),
		zap.String("name", name),
	)
	return &project, nil
}

// Update updates an existing project
func (r *projectRepository) Update(ctx context.Context, project *domain.Project) error {
	r.logger.Debug("Updating project", zap.String("project_id", project.ID.String()))

	result := r.db.WithContext(ctx).Save(project)
	if result.Error != nil {
		r.logger.Error("Failed to update project",
			zap.Error(result.Error),
			zap.String("project_id", project.ID.String()),
		)
		return fmt.Errorf("failed to update project: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("Project not found for update", zap.String("project_id", project.ID.String()))
		return domain.ErrProjectNotFound
	}

	r.logger.Info("Project updated successfully",
		zap.String("project_id", project.ID.String()),
		zap.String("name", project.Name),
	)

	return nil
}

// Delete removes a project from the database
func (r *projectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug("Deleting project", zap.String("project_id", id.String()))

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Project{})
	if result.Error != nil {
		r.logger.Error("Failed to delete project",
			zap.Error(result.Error),
			zap.String("project_id", id.String()),
		)
		return fmt.Errorf("failed to delete project: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		r.logger.Debug("Project not found for deletion", zap.String("project_id", id.String()))
		return domain.ErrProjectNotFound
	}

	r.logger.Info("Project deleted successfully", zap.String("project_id", id.String()))
	return nil
}

// List retrieves all projects with pagination
func (r *projectRepository) List(ctx context.Context, offset, limit int) ([]*domain.Project, error) {
	r.logger.Debug("Listing projects",
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	var projects []*domain.Project
	if err := r.db.WithContext(ctx).Preload("Customer").Offset(offset).Limit(limit).Order("created_at DESC").Find(&projects).Error; err != nil {
		r.logger.Error("Failed to list projects",
			zap.Error(err),
			zap.Int("offset", offset),
			zap.Int("limit", limit),
		)
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	r.logger.Debug("Projects listed successfully",
		zap.Int("count", len(projects)),
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	return projects, nil
}
