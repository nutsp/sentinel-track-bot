package service

import (
	"context"
	"fmt"

	"fix-track-bot/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// projectService implements the ProjectService interface
type projectService struct {
	projectRepo  domain.ProjectRepository
	customerRepo domain.CustomerRepository
	logger       *zap.Logger
}

// NewProjectService creates a new instance of project service
func NewProjectService(projectRepo domain.ProjectRepository, customerRepo domain.CustomerRepository, logger *zap.Logger) domain.ProjectService {
	return &projectService{
		projectRepo:  projectRepo,
		customerRepo: customerRepo,
		logger:       logger,
	}
}

// CreateProject creates a new project for a customer
func (s *projectService) CreateProject(ctx context.Context, customerID uuid.UUID, name, description string) (*domain.Project, error) {
	s.logger.Debug("Creating project",
		zap.String("customer_id", customerID.String()),
		zap.String("name", name),
	)

	// Validate input
	if !domain.IsValidProject(name, customerID) {
		s.logger.Debug("Invalid project data",
			zap.String("customer_id", customerID.String()),
			zap.String("name", name),
		)
		return nil, domain.ErrEmptyProjectName
	}

	// Verify customer exists
	_, err := s.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		s.logger.Error("Failed to verify customer exists",
			zap.Error(err),
			zap.String("customer_id", customerID.String()),
		)
		return nil, fmt.Errorf("failed to verify customer exists: %w", err)
	}

	// Check if project already exists for this customer
	existing, err := s.projectRepo.GetByName(ctx, customerID, name)
	if err != nil && err != domain.ErrProjectNotFound {
		s.logger.Error("Failed to check existing project",
			zap.Error(err),
			zap.String("customer_id", customerID.String()),
			zap.String("name", name),
		)
		return nil, fmt.Errorf("failed to check existing project: %w", err)
	}

	if existing != nil {
		s.logger.Debug("Project already exists",
			zap.String("customer_id", customerID.String()),
			zap.String("name", name),
		)
		return nil, domain.ErrProjectAlreadyExists
	}

	// Create new project
	project := &domain.Project{
		ID:          uuid.New(),
		CustomerID:  customerID,
		Name:        name,
		Description: description,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		s.logger.Error("Failed to create project",
			zap.Error(err),
			zap.String("customer_id", customerID.String()),
			zap.String("name", name),
		)
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	s.logger.Info("Project created successfully",
		zap.String("project_id", project.ID.String()),
		zap.String("customer_id", customerID.String()),
		zap.String("name", name),
	)

	return project, nil
}

// GetProject retrieves a project by ID
func (s *projectService) GetProject(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	s.logger.Debug("Retrieving project", zap.String("project_id", id.String()))

	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to retrieve project",
			zap.Error(err),
			zap.String("project_id", id.String()),
		)
		return nil, fmt.Errorf("failed to retrieve project: %w", err)
	}

	s.logger.Debug("Project retrieved successfully", zap.String("project_id", id.String()))
	return project, nil
}

// GetProjectsByCustomer retrieves all projects for a customer
func (s *projectService) GetProjectsByCustomer(ctx context.Context, customerID uuid.UUID) ([]*domain.Project, error) {
	s.logger.Debug("Retrieving projects by customer", zap.String("customer_id", customerID.String()))

	projects, err := s.projectRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		s.logger.Error("Failed to retrieve projects by customer",
			zap.Error(err),
			zap.String("customer_id", customerID.String()),
		)
		return nil, fmt.Errorf("failed to retrieve projects by customer: %w", err)
	}

	s.logger.Debug("Projects retrieved successfully",
		zap.String("customer_id", customerID.String()),
		zap.Int("count", len(projects)),
	)

	return projects, nil
}

// UpdateProject updates project information
func (s *projectService) UpdateProject(ctx context.Context, id uuid.UUID, name, description string) error {
	s.logger.Debug("Updating project",
		zap.String("project_id", id.String()),
		zap.String("name", name),
	)

	// Get existing project
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to retrieve project for update",
			zap.Error(err),
			zap.String("project_id", id.String()),
		)
		return fmt.Errorf("failed to retrieve project for update: %w", err)
	}

	// Validate input
	if !domain.IsValidProject(name, project.CustomerID) {
		s.logger.Debug("Invalid project data",
			zap.String("project_id", id.String()),
			zap.String("name", name),
		)
		return domain.ErrEmptyProjectName
	}

	// Update fields
	project.Name = name
	project.Description = description

	if err := s.projectRepo.Update(ctx, project); err != nil {
		s.logger.Error("Failed to update project",
			zap.Error(err),
			zap.String("project_id", id.String()),
		)
		return fmt.Errorf("failed to update project: %w", err)
	}

	s.logger.Info("Project updated successfully",
		zap.String("project_id", id.String()),
		zap.String("name", name),
	)

	return nil
}

// ListProjects lists all projects
func (s *projectService) ListProjects(ctx context.Context, offset, limit int) ([]*domain.Project, error) {
	s.logger.Debug("Listing projects",
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	projects, err := s.projectRepo.List(ctx, offset, limit)
	if err != nil {
		s.logger.Error("Failed to list projects",
			zap.Error(err),
			zap.Int("offset", offset),
			zap.Int("limit", limit),
		)
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	s.logger.Debug("Projects listed successfully",
		zap.Int("count", len(projects)),
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)

	return projects, nil
}
