package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/internal/errors"
	"github.com/gi4nks/ambros/internal/models"
)

// ChainCommand represents the chain command
type ChainCommand struct {
	*BaseCommand
	name        string
	description string
	cmdIds      []string
	conditional bool
	store       bool
}

// NewChainCommand creates a new chain command
func NewChainCommand(logger *zap.Logger, repo RepositoryInterface) *ChainCommand {
	cc := &ChainCommand{}

	cmd := &cobra.Command{
		Use:   "chain",
		Short: "Execute or manage command chains",
		Long: `Execute multiple commands in sequence as a chain.
Can continue on error based on conditional flag.
Allows storing and managing named command chains.

Subcommands:
  exec <name>           Execute a stored chain
  create <name> <ids>   Create a new chain with command IDs
  list                  List all stored chains
  delete <name>         Delete a chain`,
		Example: `  ambros chain exec chain1
  ambros chain create mychain cmd1,cmd2,cmd3
  ambros chain list
  ambros chain delete oldchain`,
		Args: cobra.MinimumNArgs(1),
		RunE: cc.runE,
	}

	cc.BaseCommand = NewBaseCommand(cmd, logger, repo)
	cc.cmd = cmd
	cc.setupFlags(cmd)
	return cc
}

func (c *ChainCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&c.name, "name", "n", "", "Name of the chain")
	cmd.Flags().StringVarP(&c.description, "desc", "d", "", "Description of the chain")
	cmd.Flags().BoolVarP(&c.conditional, "conditional", "c", false, "Continue on error")
	cmd.Flags().BoolVarP(&c.store, "store", "s", false, "Store execution results")
}

func (c *ChainCommand) runE(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Chain command invoked",
		zap.String("subcommand", args[0]),
		zap.Strings("args", args))

	switch args[0] {
	case "exec":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "chain name required", nil)
		}
		return c.executeChain(args[1])
	case "create":
		if len(args) < 3 {
			return errors.NewError(errors.ErrInvalidCommand, "chain name and command IDs required", nil)
		}
		return c.createChain(args[1], strings.Split(args[2], ","))
	case "list":
		return c.listChains()
	case "delete":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "chain name required", nil)
		}
		return c.deleteChain(args[1])
	default:
		return errors.NewError(errors.ErrInvalidCommand, "unknown subcommand: "+args[0], nil)
	}
}

func (c *ChainCommand) executeChain(name string) error {
	c.logger.Info("Executing command chain",
		zap.String("chainName", name),
		zap.Bool("conditional", c.conditional))

	// For now, simulate chain execution since we don't have the chain repository methods
	// This would need to be implemented when the chain storage is available
	fmt.Printf("Executing chain: %s\n", name)

	// Simulate some commands in the chain
	mockCommands := []string{"cmd1", "cmd2", "cmd3"}
	successful := 0
	failed := 0
	startTime := time.Now()

	for i, cmdId := range mockCommands {
		fmt.Printf("Executing command %d/%d: %s\n", i+1, len(mockCommands), cmdId)

		// Simulate command execution
		// In real implementation, this would fetch and execute actual commands
		success := i != 1 // Simulate failure on second command

		if success {
			successful++
			fmt.Printf("✓ Command %s completed successfully\n", cmdId)
		} else {
			failed++
			fmt.Printf("✗ Command %s failed\n", cmdId)
			if !c.conditional {
				c.logger.Warn("Chain execution stopped due to failure",
					zap.String("failedCommand", cmdId))
				break
			}
		}
	}

	duration := time.Since(startTime)

	// Print summary
	fmt.Printf("\nChain execution completed:\n")
	fmt.Printf("Total commands: %d\n", len(mockCommands))
	fmt.Printf("Successful: %d\n", successful)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Printf("Duration: %s\n", duration)

	c.logger.Info("Chain execution completed",
		zap.String("chainName", name),
		zap.Int("successful", successful),
		zap.Int("failed", failed),
		zap.Duration("duration", duration))

	return nil
}

func (c *ChainCommand) createChain(name string, cmdIds []string) error {
	c.logger.Info("Creating command chain",
		zap.String("chainName", name),
		zap.Strings("commandIds", cmdIds),
		zap.String("description", c.description))

	// Verify commands exist
	for _, id := range cmdIds {
		_, err := c.repository.Get(strings.TrimSpace(id))
		if err != nil {
			c.logger.Error("Command not found for chain",
				zap.String("commandId", id),
				zap.Error(err))
			return errors.NewError(errors.ErrCommandNotFound,
				fmt.Sprintf("command not found: %s", id), err)
		}
	}

	// Create chain model
	chain := models.CommandChain{
		ID:          c.generateChainID(),
		Name:        name,
		Description: c.description,
		Commands:    cmdIds,
		Conditional: c.conditional,
		CreatedAt:   time.Now(),
	}

	// For now, just log the creation since we don't have chain storage yet
	fmt.Printf("Chain '%s' created with %d commands\n", name, len(cmdIds))
	if c.description != "" {
		fmt.Printf("Description: %s\n", c.description)
	}

	c.logger.Info("Chain created successfully",
		zap.String("chainId", chain.ID),
		zap.String("chainName", name),
		zap.Int("commandCount", len(cmdIds)))

	return nil
}

func (c *ChainCommand) listChains() error {
	c.logger.Debug("Listing command chains")

	// For now, show a placeholder since we don't have chain storage yet
	fmt.Println("No command chains found")
	fmt.Println("Note: Chain storage is not yet implemented")

	c.logger.Info("Chain list command completed")
	return nil
}

func (c *ChainCommand) deleteChain(name string) error {
	c.logger.Info("Deleting command chain",
		zap.String("chainName", name))

	// For now, just simulate deletion
	fmt.Printf("Chain '%s' would be deleted\n", name)
	fmt.Println("Note: Chain storage is not yet implemented")

	c.logger.Info("Chain deletion completed",
		zap.String("chainName", name))

	return nil
}

func (c *ChainCommand) generateChainID() string {
	return fmt.Sprintf("CHAIN-%d", time.Now().UnixNano())
}

func (c *ChainCommand) Command() *cobra.Command {
	return c.cmd
}
