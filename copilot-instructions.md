# Copilot Instructions for Ambros

This file provides context and coding guidelines for GitHub Copilot when working on the Ambros command-line tool project.

## Project Overview

Ambros is a Go-based CLI tool for running, storing, and managing command executions. It provides command history, search functionality, and execution analytics.

### Core Functionality
- **Command Execution**: Run shell commands with optional storage and categorization
- **Command History**: Store and retrieve previously executed commands
- **Search & Analytics**: Find commands by various criteria and analyze execution patterns
- **Export/Import**: Backup and restore command history
- **Scheduling**: Plan and manage command executions

## Architecture

### Project Structure
```
cmd/
├── main.go                 # Application entry point
├── commands/               # Cobra CLI command implementations
│   ├── run.go             # Core command execution
│   ├── last.go            # Command history display
│   ├── search.go          # Command search functionality
│   ├── command-wrapper.go # Command execution utilities
│   └── ...
└── migrate/               # Database migration tools

internal/
├── models/                # Data structures and domain models
├── repos/                 # Repository pattern for data access
├── utils/                 # Shared utilities and configuration
├── chain/                 # Command chaining functionality
├── scheduler/             # Command scheduling
└── analytics/             # Execution analytics

commands/                  # Legacy command structure (being phased out)
```

### Key Dependencies
- **Cobra**: CLI framework (`github.com/spf13/cobra`)
- **BadgerDB**: Embedded key-value database (`github.com/dgraph-io/badger/v4`)
- **Zap**: Structured logging (`go.uber.org/zap`)
- **Color**: Terminal colored output (`github.com/fatih/color`)
- **YAML**: Configuration and export (`gopkg.in/yaml.v2`, `gopkg.in/yaml.v3`)

## Coding Standards

### Go Conventions
- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Implement proper error handling with custom error types
- Include comprehensive logging with Zap logger
- Write unit tests for all new functionality

### Command Structure Pattern
```go
type XyzCommand struct {
    logger   *zap.Logger
    repo     RepositoryInterface
    opts     XyzOptions
}

type XyzOptions struct {
    // Command-specific options
}

func NewXyzCommand(logger *zap.Logger, repo RepositoryInterface) *XyzCommand {
    cmd := &XyzCommand{
        logger: logger,
        repo:   repo,
    }
    
    cobraCmd := &cobra.Command{
        Use:   "xyz [flags]",
        Short: "Brief description",
        Long:  `Detailed description...`,
        RunE:  cmd.runE,
    }
    
    // Add flags
    cobraCmd.Flags().StringVar(&cmd.opts.field, "flag", "default", "description")
    
    cmd.cmd = cobraCmd
    return cmd
}

func (c *XyzCommand) runE(cmd *cobra.Command, args []string) error {
    c.logger.Debug("Command invoked", zap.Strings("args", args))
    
    // Implementation
    
    return nil
}
```

### Error Handling
- Use custom error types from `internal/errors`
- Provide meaningful error messages
- Log errors with appropriate levels
- Return errors to caller for proper handling

### Database Operations
- Use repository pattern for data access
- Implement proper transaction handling
- Include error handling for database operations
- Use structured logging for database queries

### Testing
- Write unit tests in `*_test.go` files
- Use testify for assertions
- Mock external dependencies
- Test both success and error cases

## Key Patterns

### Command Execution Flow
1. Parse command flags and arguments
2. Validate input parameters
3. Execute command through CommandWrapper
4. Store execution details (if enabled)
5. Display results with colored output
6. Handle errors gracefully

### Repository Pattern
```go
type RepositoryInterface interface {
    StoreCommand(command *models.Command) error
    GetCommand(id string) (*models.Command, error)
    GetAllCommands() ([]*models.Command, error)
    // ... other methods
}
```

### Color Output Pattern
```go
// Success case
color.Green("[%s]", commandID)

// Failure case  
color.Red("[%s]", commandID)

// Status display
if command.Status {
    color.Green("   ID: %s", command.ID)
    color.Green("Success")
} else {
    color.Red("   ID: %s", command.ID)
    color.Red("Failed")
}
```

### Command Argument Handling
- Use `--` separator for commands with flags: `ambros run -- ls -la`
- Support simple commands without separator: `ambros run echo "hello"`
- Provide clear error messages for flag conflicts

## Database Schema

### Command Model
```go
type Command struct {
    ID          string
    Command     string
    Args        []string
    Output      string
    Error       string
    ExitCode    int
    Status      bool
    Duration    time.Duration
    CreatedAt   time.Time
    UpdatedAt   time.Time
    Tags        []string
    Category    string
    Environment map[string]string
}
```

### Storage Keys
- Commands: `cmd:<timestamp>:<id>`
- Indexes: Various indexing strategies for search
- Configuration: `config:*`

## Configuration

### YAML Configuration
```yaml
database:
  path: ~/.ambros/data
logging:
  level: info
  file: ~/.ambros/logs/ambros.log
defaults:
  store: true
  timeout: 30s
```

## Development Guidelines

### Adding New Commands
1. Create command file in `cmd/commands/`
2. Implement the command pattern above
3. Add command to root command registration
4. Write comprehensive tests
5. Update documentation

### Modifying Existing Commands
1. Maintain backward compatibility
2. Update tests for new functionality
3. Update help text and examples
4. Consider migration needs for stored data

### Database Changes
1. Create migration scripts if needed
2. Test with existing data
3. Provide rollback capability
4. Update repository interface if needed

## Common Operations

### Logging
```go
logger.Debug("Operation started", zap.String("operation", "example"))
logger.Info("Command executed", zap.String("command", cmd), zap.Duration("duration", elapsed))
logger.Error("Operation failed", zap.Error(err))
```

### Command Storage
```go
wrapper := NewCommandWrapper(logger, repo)
result := wrapper.ExecuteCommand(command, args, options)
err := wrapper.FinalizeCommand(result)
```

### Flag Definition
```go
cmd.Flags().StringVarP(&opts.category, "category", "c", "", "Command category")
cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Show what would be executed")
cmd.Flags().StringSliceVarP(&opts.tags, "tag", "t", []string{}, "Command tags")
```

## Testing Strategy

### Unit Tests
- Test individual functions and methods
- Mock external dependencies (database, filesystem)
- Test error conditions
- Verify logging output

### Integration Tests
- Test command execution end-to-end
- Verify database operations
- Test configuration loading
- Validate CLI argument parsing

### Example Test Structure
```go
func TestXyzCommand(t *testing.T) {
    logger := zap.NewNop()
    mockRepo := &mocks.MockRepository{}
    
    cmd := NewXyzCommand(logger, mockRepo)
    
    // Test cases
    t.Run("success case", func(t *testing.T) {
        // Test implementation
    })
    
    t.Run("error case", func(t *testing.T) {
        // Test implementation
    })
}
```

## Performance Considerations

- Use BadgerDB efficiently with proper key design
- Implement pagination for large result sets
- Cache frequently accessed data
- Use background processing for heavy operations
- Monitor memory usage for long-running operations

## Security Considerations

- Sanitize command arguments
- Validate user input
- Secure database file permissions
- Don't log sensitive information
- Implement proper access controls

This guide should help GitHub Copilot generate code that follows the project's patterns and conventions.