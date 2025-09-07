package repos

import (
	"context"

	"github.com/gi4nks/ambros/internal/models"
)

// RepositoryInterface defines the methods that a repository must implement
type RepositoryInterface interface {
	Put(ctx context.Context, command models.Command) error
	Get(id string) (*models.Command, error)
	FindById(id string) (models.Command, error)
	GetLimitCommands(limit int) ([]models.Command, error)
	GetAllCommands() ([]models.Command, error)
	SearchByTag(tag string) ([]models.Command, error)
	SearchByStatus(success bool) ([]models.Command, error)
	GetTemplate(name string) (*models.Template, error)
	Push(command models.Command) error
}
