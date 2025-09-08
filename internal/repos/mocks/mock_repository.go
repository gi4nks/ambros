package mocks

import (
	"context"

	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of the repository interface for testing
type MockRepository struct {
	mock.Mock
}

// NewMockRepository creates a new mock repository
func NewMockRepository() *MockRepository {
	return &MockRepository{}
}

// Put implements Repository.Put
func (m *MockRepository) Put(ctx context.Context, command models.Command) error {
	args := m.Called(ctx, command)
	return args.Error(0)
}

// Get implements Repository.Get
func (m *MockRepository) Get(id string) (*models.Command, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Command), args.Error(1)
}

// FindById implements Repository.FindById
func (m *MockRepository) FindById(id string) (models.Command, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return models.Command{}, args.Error(1)
	}
	return args.Get(0).(models.Command), args.Error(1)
}

// GetAllCommands implements Repository.GetAllCommands
func (m *MockRepository) GetAllCommands() ([]models.Command, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Command), args.Error(1)
}

// GetLimitCommands implements Repository.GetLimitCommands
func (m *MockRepository) GetLimitCommands(limit int) ([]models.Command, error) {
	args := m.Called(limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Command), args.Error(1)
}

// SearchByTag implements Repository.SearchByTag
func (m *MockRepository) SearchByTag(tag string) ([]models.Command, error) {
	args := m.Called(tag)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Command), args.Error(1)
}

// SearchByStatus implements Repository.SearchByStatus
func (m *MockRepository) SearchByStatus(success bool) ([]models.Command, error) {
	args := m.Called(success)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Command), args.Error(1)
}

// GetTemplate implements Repository.GetTemplate
func (m *MockRepository) GetTemplate(name string) (*models.Template, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Template), args.Error(1)
}

// Push implements Repository.Push
func (m *MockRepository) Push(command models.Command) error {
	args := m.Called(command)
	return args.Error(0)
}

// Delete implements Repository.Delete
func (m *MockRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
