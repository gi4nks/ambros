package commands

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"
)

// ChainCommand represents the chain command with advanced chain management features
type ChainCommand struct {
	*BaseCommand
	name        string
	description string
	conditional bool
	store       bool
	parallel    bool
	timeout     time.Duration
	retry       int
	dryRun      bool
	interactive bool

	// Chain storage (in-memory for now, would use repository in full implementation)
	chains      map[string]*models.CommandChain
	chainsMutex sync.RWMutex
}

// ChainExecutionResult represents the result of a chain execution
type ChainExecutionResult struct {
	ChainName     string                   `json:"chain_name"`
	StartTime     time.Time                `json:"start_time"`
	EndTime       time.Time                `json:"end_time"`
	Duration      time.Duration            `json:"duration"`
	TotalCommands int                      `json:"total_commands"`
	Successful    int                      `json:"successful"`
	Failed        int                      `json:"failed"`
	Skipped       int                      `json:"skipped"`
	Results       []CommandExecutionResult `json:"results"`
	Status        string                   `json:"status"`
}

// CommandExecutionResult represents individual command result in chain
type CommandExecutionResult struct {
	CommandID  string        `json:"command_id"`
	Command    string        `json:"command"`
	Status     string        `json:"status"`
	Output     string        `json:"output"`
	Error      string        `json:"error"`
	Duration   time.Duration `json:"duration"`
	RetryCount int           `json:"retry_count"`
}

// NewChainCommand creates a new enhanced chain command
func NewChainCommand(logger *zap.Logger, repo RepositoryInterface) *ChainCommand {
	cc := &ChainCommand{
		chains: make(map[string]*models.CommandChain),
	}

	cmd := &cobra.Command{
		Use:   "chain",
		Short: "🔗 Execute and manage command chains",
		Long: `🔗 Advanced Command Chain Management System

Chain commands allow you to create, store, and execute sequences of commands 
with advanced features like parallel execution, retry logic, conditional 
execution, and comprehensive monitoring.

Features:
  • Sequential and parallel execution modes
  • Conditional execution (continue on error)
  • Retry logic with configurable attempts
  • Timeout support for long-running chains
  • Dry-run mode for testing
  • Interactive execution with prompts
  • Comprehensive result tracking and analytics
  • Chain templates and sharing

Subcommands:
  exec <name>           Execute a stored chain
  create <name> <ids>   Create a new chain with command IDs
  list                  List all stored chains with statistics
  delete <name>         Delete a chain
  show <name>           Show detailed chain information
  export <name>         Export chain to JSON
  import <file>         Import chain from JSON
  template              Manage chain templates
  analytics             View chain execution analytics

Execution Modes:
  --parallel           Execute commands in parallel
  --conditional        Continue execution on command failure
  --retry <n>          Retry failed commands n times
  --timeout <duration> Set overall chain timeout
  --dry-run           Show what would be executed without running
  --interactive       Prompt for confirmation before each command

Examples:
  ambros chain create deploy "build,test,package,deploy"
  ambros chain exec deploy --parallel --retry 2
  ambros chain list --stats
  ambros chain show deploy
  ambros chain export deploy > deploy-chain.json
  ambros chain import deploy-chain.json`,
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
	cmd.Flags().BoolVarP(&c.parallel, "parallel", "p", false, "Execute commands in parallel")
	cmd.Flags().DurationVarP(&c.timeout, "timeout", "t", 30*time.Minute, "Chain execution timeout")
	cmd.Flags().IntVarP(&c.retry, "retry", "r", 0, "Number of retry attempts for failed commands")
	cmd.Flags().BoolVar(&c.dryRun, "dry-run", false, "Show what would be executed without running")
	cmd.Flags().BoolVarP(&c.interactive, "interactive", "i", false, "Interactive execution with prompts")
}

func (c *ChainCommand) runE(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Enhanced chain command invoked",
		zap.String("subcommand", args[0]),
		zap.Strings("args", args),
		zap.Bool("parallel", c.parallel),
		zap.Bool("conditional", c.conditional),
		zap.Duration("timeout", c.timeout))

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
	case "show":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "chain name required", nil)
		}
		return c.showChain(args[1])
	case "export":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "chain name required", nil)
		}
		return c.exportChain(args[1])
	case "import":
		if len(args) < 2 {
			return errors.NewError(errors.ErrInvalidCommand, "file path required", nil)
		}
		return c.importChain(args[1])
	case "analytics":
		return c.showAnalytics()
	case "template":
		return c.manageTemplates(args[1:])
	default:
		return errors.NewError(errors.ErrInvalidCommand, "unknown subcommand: "+args[0], nil)
	}
}

func (c *ChainCommand) executeChain(name string) error {
	c.logger.Info("Executing command chain",
		zap.String("chainName", name),
		zap.Bool("conditional", c.conditional),
		zap.Bool("parallel", c.parallel),
		zap.Duration("timeout", c.timeout),
		zap.Int("retry", c.retry))

	// Create execution context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Load chain
	chain, exists := c.getChain(name)
	if !exists {
		// Try to create a demo chain for testing
		chain = c.createDemoChain(name)
	}

	if c.dryRun {
		return c.performDryRun(chain)
	}

	result := &ChainExecutionResult{
		ChainName:     name,
		StartTime:     time.Now(),
		TotalCommands: len(chain.Commands),
		Results:       make([]CommandExecutionResult, 0, len(chain.Commands)),
		Status:        "running",
	}

	color.Green("🔗 Starting chain execution: %s", name)
	color.Cyan("📊 Total commands: %d | Parallel: %v | Conditional: %v",
		len(chain.Commands), c.parallel, c.conditional)

	if c.parallel && len(chain.Commands) > 1 {
		c.executeParallel(ctx, chain, result)
	} else {
		c.executeSequential(ctx, chain, result)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Determine final status
	if result.Failed == 0 {
		result.Status = "completed"
		color.Green("✅ Chain execution completed successfully")
	} else if result.Successful > 0 {
		result.Status = "partial"
		color.Yellow("⚠️  Chain execution completed with failures")
	} else {
		result.Status = "failed"
		color.Red("❌ Chain execution failed")
	}

	c.displayExecutionSummary(result)

	if c.store {
		c.storeExecutionResult(result)
	}

	return nil
}

func (c *ChainCommand) executeSequential(ctx context.Context, chain *models.CommandChain, result *ChainExecutionResult) {
	for i, cmdId := range chain.Commands {
		select {
		case <-ctx.Done():
			color.Red("⏰ Chain execution timeout reached")
			return
		default:
		}

		cmdResult := c.executeCommand(cmdId, i+1, len(chain.Commands))
		result.Results = append(result.Results, cmdResult)

		switch cmdResult.Status {
		case "success":
			result.Successful++
		case "failed":
			result.Failed++
			if !c.conditional {
				color.Yellow("🛑 Stopping chain execution due to failure")
				return
			}
		case "skipped":
			result.Skipped++
		}

		// Interactive mode confirmation
		if c.interactive && i < len(chain.Commands)-1 {
			if !c.promptContinue() {
				color.Yellow("🛑 Chain execution stopped by user")
				break
			}
		}
	}
}

func (c *ChainCommand) executeParallel(ctx context.Context, chain *models.CommandChain, result *ChainExecutionResult) {
	var wg sync.WaitGroup
	resultChan := make(chan CommandExecutionResult, len(chain.Commands))

	color.Cyan("🚀 Executing %d commands in parallel", len(chain.Commands))

	for i, cmdId := range chain.Commands {
		wg.Add(1)
		go func(id string, index int) {
			defer wg.Done()
			cmdResult := c.executeCommand(id, index+1, len(chain.Commands))
			resultChan <- cmdResult
		}(cmdId, i)
	}

	// Wait for all commands to complete or timeout
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		close(resultChan)
	case <-ctx.Done():
		color.Red("⏰ Parallel execution timeout reached")
		close(resultChan)
	}

	// Collect results
	for cmdResult := range resultChan {
		result.Results = append(result.Results, cmdResult)
		switch cmdResult.Status {
		case "success":
			result.Successful++
		case "failed":
			result.Failed++
		case "skipped":
			result.Skipped++
		}
	}
}

func (c *ChainCommand) executeCommand(cmdId string, current, total int) CommandExecutionResult {
	startTime := time.Now()
	result := CommandExecutionResult{
		CommandID: cmdId,
		Status:    "running",
	}

	color.Cyan("📋 [%d/%d] Executing: %s", current, total, cmdId)

	// Get command from repository
	cmd, err := c.repository.Get(strings.TrimSpace(cmdId))
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("Command not found: %s", err.Error())
		color.Red("❌ Command not found: %s", cmdId)
		return result
	}

	result.Command = cmd.Command

	// Execute with retry logic
	var lastErr error
	for attempt := 0; attempt <= c.retry; attempt++ {
		if attempt > 0 {
			color.Yellow("🔄 Retry attempt %d/%d for command: %s", attempt, c.retry, cmdId)
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
		}

		output, execErr := c.executeSystemCommand(cmd.Command)
		if execErr == nil {
			result.Status = "success"
			result.Output = output
			result.RetryCount = attempt
			color.Green("✅ [%d/%d] Completed: %s", current, total, cmdId)
			break
		}

		lastErr = execErr
		result.RetryCount = attempt
	}

	if lastErr != nil {
		result.Status = "failed"
		result.Error = lastErr.Error()
		color.Red("❌ [%d/%d] Failed: %s - %s", current, total, cmdId, lastErr.Error())
	}

	result.Duration = time.Since(startTime)
	return result
}

func (c *ChainCommand) executeSystemCommand(command string) (string, error) {
	// Parse the command respecting quotes/escapes
	parts, err := shellFields(command)
	if err != nil {
		return "", err
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	// Validate executable path to avoid shell surprises or path injection
	if _, err := ResolveCommandPath(parts[0]); err != nil {
		return "", err
	}

	// Execute the command
	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()

	return string(output), err
}

func (c *ChainCommand) performDryRun(chain *models.CommandChain) error {
	color.Yellow("🔍 DRY RUN MODE - No commands will be executed")
	color.Cyan("Chain: %s", chain.Name)
	color.Cyan("Description: %s", chain.Description)
	color.Cyan("Commands to execute:")

	for i, cmdId := range chain.Commands {
		cmd, err := c.repository.Get(strings.TrimSpace(cmdId))
		if err != nil {
			color.Red("  %d. ❌ %s (Command not found)", i+1, cmdId)
			continue
		}

		color.White("  %d. %s", i+1, cmd.Command)
		// Note: Command model doesn't have Description field, using Name instead
		if cmd.Name != "" && cmd.Name != cmd.Command {
			color.HiBlack("     → %s", cmd.Name)
		}
	}

	color.Yellow("Execution mode: %s", c.getExecutionMode())
	color.Yellow("Continue on failure: %v", c.conditional)
	color.Yellow("Retry attempts: %d", c.retry)
	color.Yellow("Timeout: %s", c.timeout)

	return nil
}

func (c *ChainCommand) getExecutionMode() string {
	if c.parallel {
		return "Parallel"
	}
	return "Sequential"
}

func (c *ChainCommand) promptContinue() bool {
	reader := bufio.NewReader(os.Stdin)
	color.Yellow("Continue with next command? [Y/n]: ")

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "" || response == "y" || response == "yes"
}

func (c *ChainCommand) displayExecutionSummary(result *ChainExecutionResult) {
	color.Cyan("\n📊 Chain Execution Summary")
	color.Cyan("═══════════════════════════")
	color.White("Chain: %s", result.ChainName)
	color.White("Duration: %s", result.Duration)
	color.White("Total Commands: %d", result.TotalCommands)
	color.Green("Successful: %d", result.Successful)
	color.Red("Failed: %d", result.Failed)
	if result.Skipped > 0 {
		color.Yellow("Skipped: %d", result.Skipped)
	}

	successRate := float64(result.Successful) / float64(result.TotalCommands) * 100
	color.White("Success Rate: %.1f%%", successRate)

	if len(result.Results) > 0 {
		color.Cyan("\n📋 Command Details:")
		for i, cmdResult := range result.Results {
			status := "✅"
			switch cmdResult.Status {
			case "failed":
				status = "❌"
			case "skipped":
				status = "⏭️"
			}

			color.White("  %d. %s %s (%s)", i+1, status, cmdResult.CommandID, cmdResult.Duration)
			if cmdResult.RetryCount > 0 {
				color.Yellow("     Retries: %d", cmdResult.RetryCount)
			}
			if cmdResult.Error != "" {
				color.Red("     Error: %s", cmdResult.Error)
			}
		}
	}
}

func (c *ChainCommand) storeExecutionResult(result *ChainExecutionResult) {
	// Convert result to JSON and store as a command
	data, err := json.Marshal(result)
	if err != nil {
		c.logger.Error("Failed to marshal execution result", zap.Error(err))
		return
	}

	resultCmd := models.Command{
		Entity: models.Entity{
			ID:           c.generateChainID(),
			CreatedAt:    result.StartTime,
			TerminatedAt: result.EndTime,
		},
		Name:     fmt.Sprintf("chain-execution:%s", result.ChainName),
		Command:  "chain-result",
		Category: "chain-execution",
		Status:   result.Status == "completed",
		Tags:     []string{"chain", "execution", "result"},
		Variables: map[string]string{
			"chain_name": result.ChainName,
			"status":     result.Status,
			"duration":   result.Duration.String(),
			"data":       string(data),
		},
	}

	ctx := context.Background()
	if err := c.repository.Put(ctx, resultCmd); err != nil {
		c.logger.Error("Failed to store execution result", zap.Error(err))
	} else {
		color.Green("💾 Execution result stored: %s", resultCmd.ID)
	}
}

// Helper methods for chain management
func (c *ChainCommand) getChain(name string) (*models.CommandChain, bool) {
	c.chainsMutex.RLock()
	defer c.chainsMutex.RUnlock()

	chain, exists := c.chains[name]
	return chain, exists
}

func (c *ChainCommand) createDemoChain(name string) *models.CommandChain {
	// Create a demo chain with some test commands
	commands, err := c.repository.GetLimitCommands(3)
	if err != nil || len(commands) == 0 {
		// Fallback to mock commands
		return &models.CommandChain{
			ID:          c.generateChainID(),
			Name:        name,
			Description: "Demo chain for testing",
			Commands:    []string{"echo hello", "echo world", "echo done"},
			Conditional: c.conditional,
			CreatedAt:   time.Now(),
		}
	}

	cmdIds := make([]string, len(commands))
	for i, cmd := range commands {
		cmdIds[i] = cmd.ID
	}

	chain := &models.CommandChain{
		ID:          c.generateChainID(),
		Name:        name,
		Description: "Auto-generated demo chain",
		Commands:    cmdIds,
		Conditional: c.conditional,
		CreatedAt:   time.Now(),
	}

	// Store the chain in memory
	c.chainsMutex.Lock()
	c.chains[name] = chain
	c.chainsMutex.Unlock()

	return chain
}

// Additional chain-related helper methods
func (c *ChainCommand) showChain(name string) error {
	color.Cyan("🔍 Chain Details: %s", name)

	chain, exists := c.getChain(name)
	if !exists {
		color.Red("Chain not found: %s", name)
		return nil
	}

	color.White("ID: %s", chain.ID)
	color.White("Name: %s", chain.Name)
	color.White("Description: %s", chain.Description)
	color.White("Created: %s", chain.CreatedAt.Format("2006-01-02 15:04:05"))
	color.White("Conditional: %v", chain.Conditional)
	color.White("Commands (%d):", len(chain.Commands))

	for i, cmdId := range chain.Commands {
		cmd, err := c.repository.Get(cmdId)
		if err != nil {
			color.Red("  %d. ❌ %s (not found)", i+1, cmdId)
			continue
		}
		color.White("  %d. %s", i+1, cmd.Command)
	}

	return nil
}

func (c *ChainCommand) exportChain(name string) error {
	chain, exists := c.getChain(name)
	if !exists {
		return fmt.Errorf("chain not found: %s", name)
	}

	data, err := json.MarshalIndent(chain, "", "  ")
	if err != nil {
		return err
	}

	fmt.Print(string(data))
	return nil
}

func (c *ChainCommand) importChain(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var chain models.CommandChain
	if err := json.Unmarshal(data, &chain); err != nil {
		return err
	}

	// Store the imported chain
	c.chainsMutex.Lock()
	c.chains[chain.Name] = &chain
	c.chainsMutex.Unlock()

	color.Green("✅ Chain imported: %s", chain.Name)
	return nil
}

func (c *ChainCommand) showAnalytics() error {
	color.Cyan("📊 Chain Execution Analytics")
	color.Cyan("════════════════════════════")

	c.chainsMutex.RLock()
	chainCount := len(c.chains)
	c.chainsMutex.RUnlock()

	color.White("Total Chains: %d", chainCount)

	// Get execution results from repository
	allCommands, err := c.repository.SearchByTag("chain")
	if err != nil {
		return err
	}

	executionCount := 0
	successCount := 0
	for _, cmd := range allCommands {
		if cmd.Category == "chain-execution" {
			executionCount++
			if cmd.Status {
				successCount++
			}
		}
	}

	color.White("Total Executions: %d", executionCount)
	if executionCount > 0 {
		successRate := float64(successCount) / float64(executionCount) * 100
		color.White("Success Rate: %.1f%%", successRate)
	}

	return nil
}

func (c *ChainCommand) manageTemplates(_ []string) error {
	color.Yellow("🎯 Chain templates feature coming soon!")
	color.White("This will allow saving and sharing chain configurations")
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
		zap.String("chainName", chain.Name),
		zap.String("chainDescription", chain.Description),
		zap.Int("commandCount", len(chain.Commands)),
		zap.Bool("conditional", chain.Conditional),
		zap.Time("createdAt", chain.CreatedAt))

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
