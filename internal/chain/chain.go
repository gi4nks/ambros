package chain

import (
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/models"
)

// Chain represents a command chain
type Chain struct {
	logger *zap.Logger
}

// NewChain creates a new chain instance
func NewChain(logger *zap.Logger) *Chain {
	return &Chain{
		logger: logger,
	}
}

// ExecuteChain executes a command chain
func (c *Chain) ExecuteChain(chain *models.CommandChain) error {
	c.logger.Info("Executing command chain", zap.String("name", chain.Name))

	// Implementation would execute commands in sequence
	// For now, just return success
	return nil
}

// ValidateChain validates a command chain
func (c *Chain) ValidateChain(chain *models.CommandChain) error {
	if chain.Name == "" {
		return &ChainError{Message: "chain name is required"}
	}

	if len(chain.Commands) == 0 {
		return &ChainError{Message: "chain must contain at least one command"}
	}

	return nil
}

// ChainError represents a chain-related error
type ChainError struct {
	Message string
}

func (e *ChainError) Error() string {
	return e.Message
}
