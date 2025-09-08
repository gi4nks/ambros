# Contributing to Ambros 🤝

Thank you for your interest in contributing to **Ambros - The Command Butler**! We welcome contributions from developers of all skill levels and backgrounds. This guide will help you get started with contributing to the project.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)
- [Community](#community)

## 📜 Code of Conduct

This project adheres to a code of conduct that we expect all contributors to follow. Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md) to help us maintain a welcoming and inclusive community.

## 🚀 Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.23.0 or later** - [Download Go](https://golang.org/dl/)
- **Git** - [Install Git](https://git-scm.com/downloads)
- **Make** (optional but recommended) - Usually pre-installed on Unix systems

### Fork and Clone

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/ambros.git
   cd ambros
   ```

3. **Add the original repository** as a remote:
   ```bash
   git remote add upstream https://github.com/gi4nks/ambros.git
   ```

## 🛠️ Development Setup

### Initial Setup

1. **Install dependencies**:
   ```bash
   make deps
   # or manually:
   go mod download
   ```

2. **Build the project**:
   ```bash
   make build
   # or manually:
   go build -o bin/ambros ./cmd
   ```

3. **Run tests** to ensure everything works:
   ```bash
   make test
   # or manually:
   go test ./...
   ```

### Development Workflow

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following our coding standards

3. **Test your changes**:
   ```bash
   make test
   make lint
   ```

4. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add awesome new feature"
   ```

### Available Make Targets

```bash
make build        # Build the project
make test         # Run all tests
make test-short   # Run tests in short mode
make test-cover   # Run tests with coverage
make lint         # Run linting checks
make format       # Format code with gofmt
make clean        # Clean build artifacts
make deps         # Install dependencies
make build-all    # Build for all platforms
```

## 🤝 How to Contribute

### Types of Contributions

We welcome various types of contributions:

#### 🐛 **Bug Reports**
- Use our [bug report template](.github/ISSUE_TEMPLATE/bug_report.md)
- Include steps to reproduce, expected behavior, and actual behavior
- Provide system information (OS, Go version, Ambros version)

#### ✨ **Feature Requests**
- Use our [feature request template](.github/ISSUE_TEMPLATE/feature_request.md)
- Explain the problem you're trying to solve
- Describe your proposed solution
- Consider backwards compatibility

#### 🔧 **Code Contributions**
- Bug fixes
- New features
- Performance improvements
- Documentation improvements
- Test coverage improvements

#### 📚 **Documentation**
- README improvements
- API documentation
- Code comments
- Usage examples
- Tutorials and guides

#### 🔌 **Plugin Development**
- Create new plugins for the Ambros ecosystem
- Improve existing plugins
- Plugin documentation and examples

### Areas for Contribution

#### Core Features
- **Command execution engine** improvements
- **Database layer** optimizations
- **CLI interface** enhancements
- **Error handling** improvements

#### Web Dashboard
- **Frontend UI/UX** improvements
- **API endpoints** and functionality
- **Real-time features** and WebSocket support
- **Mobile responsiveness**

#### Plugin System
- **Plugin framework** enhancements
- **Security sandboxing** improvements
- **Plugin templates** and generators
- **Plugin registry** development

#### Analytics & Intelligence
- **ML algorithms** for command recommendations
- **Performance analytics** features
- **Data visualization** components
- **Export/import** capabilities

#### Testing & Quality
- **Unit tests** for new features
- **Integration tests** for workflows
- **Performance benchmarks**
- **Cross-platform testing**

## 🔄 Pull Request Process

### Before Submitting

1. **Sync with upstream**:
   ```bash
   git checkout master
   git pull upstream master
   git checkout your-feature-branch
   git rebase master
   ```

2. **Run the complete test suite**:
   ```bash
   make test
   make lint
   ```

3. **Update documentation** if needed

4. **Add tests** for new functionality

### PR Guidelines

1. **Use descriptive titles** following [Conventional Commits](https://conventionalcommits.org/):
   - `feat: add new command chain execution mode`
   - `fix: resolve database connection timeout issue`
   - `docs: update plugin development guide`
   - `test: add integration tests for scheduler`

2. **Provide detailed descriptions**:
   - What problem does this solve?
   - What changes were made?
   - How was it tested?
   - Any breaking changes?

3. **Keep PRs focused** - one feature or fix per PR

4. **Update CHANGELOG.md** if applicable

### PR Checklist

- [ ] Code follows project coding standards
- [ ] Tests pass locally (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] Documentation updated if needed
- [ ] Commits follow conventional commit format
- [ ] PR description is clear and detailed
- [ ] Breaking changes are documented

## 📏 Coding Standards

### Go Code Style

We follow standard Go conventions with some project-specific guidelines:

#### General Guidelines
- **Follow `gofmt`** - all code must be formatted with `gofmt`
- **Follow `golint`** - address all linting issues
- **Use meaningful names** for variables, functions, and types
- **Write self-documenting code** with clear intent
- **Add comments** for exported functions and complex logic

#### Project Structure
```
cmd/                    # Main applications
├── commands/          # CLI command implementations
├── migrate/           # Database migration tools
internal/              # Private application code
├── models/            # Data models and structures
├── repos/             # Repository pattern implementation
├── utils/             # Utility functions
├── analytics/         # Analytics and ML components
├── chain/             # Command chaining logic
├── scheduler/         # Scheduling functionality
└── api/               # Web API implementation
```

#### Code Organization
- **Keep functions small** and focused on single responsibility
- **Use interfaces** for testing and modularity
- **Handle errors explicitly** - no silent failures
- **Use structured logging** with Zap logger
- **Follow repository pattern** for data access

#### Error Handling
```go
// Good: Explicit error handling
result, err := someOperation()
if err != nil {
    logger.Error("Operation failed", zap.Error(err))
    return errors.NewError(errors.ErrInternalServer, "operation failed", err)
}

// Good: Custom error types
return errors.NewError(errors.ErrInvalidCommand, "invalid input", nil)
```

#### Logging
```go
// Use structured logging
logger.Debug("Command executed", 
    zap.String("command", cmd.Name),
    zap.Duration("duration", elapsed),
    zap.Bool("success", result.Success))
```

### Command Implementation

When adding new commands, follow this pattern:

```go
type YourCommand struct {
    *BaseCommand
    // command-specific fields
}

func NewYourCommand(logger *zap.Logger, repo RepositoryInterface) *YourCommand {
    cmd := &YourCommand{}
    
    cobraCmd := &cobra.Command{
        Use:   "your-command",
        Short: "Brief description",
        Long:  "Detailed description",
        RunE:  cmd.runE,
    }
    
    cmd.BaseCommand = NewBaseCommand(cobraCmd, logger, repo)
    cmd.setupFlags(cobraCmd)
    return cmd
}

func (yc *YourCommand) runE(cmd *cobra.Command, args []string) error {
    // Implementation
}
```

## 🧪 Testing Guidelines

### Test Organization

- **Unit tests**: Test individual functions and methods
- **Integration tests**: Test component interactions
- **End-to-end tests**: Test complete workflows

### Test File Structure
```
package_test.go        # Unit tests for package
integration_test.go    # Integration tests
*_test.go             # Specific feature tests
```

### Writing Tests

1. **Use table-driven tests** for multiple scenarios:
```go
func TestCommandExecution(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"valid command", "echo hello", "hello\n", false},
        {"invalid command", "nonexistent", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

2. **Use testify for assertions**:
```go
assert.Equal(t, expected, actual)
assert.NoError(t, err)
require.NotNil(t, result)
```

3. **Mock external dependencies**:
```go
mockRepo := &mocks.MockRepository{}
mockRepo.On("Get", "test-id").Return(&models.Command{}, nil)
```

### Test Coverage

- Aim for **80%+ coverage** for new code
- **100% coverage** for critical paths
- Use `make test-cover` to check coverage

### Running Tests

```bash
# All tests
make test

# Short tests only
make test-short

# With coverage
make test-cover

# Specific package
go test ./cmd/commands

# Specific test
go test -run TestCommandExecution ./cmd/commands
```

## 📖 Documentation

### Code Documentation

- **Document all exported functions** and types
- **Use GoDoc format** for documentation comments
- **Include examples** in documentation when helpful
- **Keep comments up-to-date** with code changes

### User Documentation

- **Update README.md** for new features
- **Add usage examples** for new commands
- **Update API documentation** for new endpoints
- **Create tutorials** for complex features

### Documentation Standards

```go
// Package documentation
// Package commands provides CLI command implementations for Ambros.

// Function documentation
// ExecuteCommand runs the specified command and returns the result.
// It handles command validation, execution, and result formatting.
//
// Example:
//   result, err := ExecuteCommand("ls -la")
//   if err != nil {
//       log.Fatal(err)
//   }
func ExecuteCommand(cmd string) (*CommandResult, error) {
    // Implementation
}
```

## 🔄 Release Process

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features, backwards compatible
- **PATCH**: Bug fixes, backwards compatible

### Release Checklist

- [ ] All tests pass
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version bumped in relevant files
- [ ] Release notes prepared
- [ ] Tagged and pushed

## 🌟 Recognition

### Contributors

We recognize contributions in several ways:
- **Contributors list** in README.md
- **Changelog mentions** for significant contributions
- **Special recognition** for major features or improvements

### Becoming a Maintainer

Regular contributors may be invited to become maintainers with:
- **Commit access** to the repository
- **Review responsibilities** for pull requests
- **Release management** participation

## 💬 Community

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and community discussion
- **Pull Request Reviews**: Code review and feedback

### Getting Help

- **Documentation**: Check README.md and wiki first
- **Search Issues**: Look for existing issues or discussions
- **Ask Questions**: Open a new discussion for help

### Community Guidelines

- **Be respectful** and inclusive
- **Help others** when you can
- **Share knowledge** and experiences
- **Follow our code of conduct**

## 🎯 Roadmap

### Current Priorities

1. **Plugin system enhancement**
2. **Web dashboard improvements**
3. **Performance optimizations**
4. **Security hardening**
5. **Mobile application development**

### Future Goals

- Multi-node deployment support
- Cloud-hosted SaaS version
- Enterprise features and SSO
- Advanced AI/ML capabilities
- Kubernetes operator

## 📞 Contact

- **Project Maintainer**: [@gi4nks](https://github.com/gi4nks)
- **Issues**: [GitHub Issues](https://github.com/gi4nks/ambros/issues)
- **Discussions**: [GitHub Discussions](https://github.com/gi4nks/ambros/discussions)

---

Thank you for contributing to Ambros! Your contributions help make command management better for everyone. 🚀

For questions about contributing, feel free to open a discussion or reach out to the maintainers.