# Ambros - The Command Butler 🤖

[![Build Status](https://github.com/gi4nks/ambros/workflows/CI/badge.svg)](https://github.com/gi4nks/ambros/actions)
[![Go Version](https://img.shields.io/badge/go-1.23.0-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)

**Ambros** is a comprehensive command management and execution platform that transforms how you interact with command-line tools. It provides intelligent command tracking, workflow automation, analytics, and a complete plugin ecosystem - all wrapped in a sleek web dashboard and powerful CLI.

## ✨ Key Features

### 🚀 **Command Execution & Management**
- **Smart Command Storage**: Automatically track command execution with detailed metadata
- **Advanced Search**: Find commands by name, tags, status, date range, and more
- **Command Templates**: Save and reuse command patterns with variable substitution
- **Dry Run Mode**: Preview commands before execution
- **Interactive Execution**: Step-by-step command execution with user prompts

### 🔗 **Workflow Automation**
- **Command Chains**: Create complex workflows with sequential or parallel execution
- **Conditional Logic**: Continue on errors, retry failed commands, timeout controls
- **Chain Templates**: Save and share workflow configurations
- **Import/Export**: Backup and restore chains in JSON format

### 🌐 **Web Dashboard**
- **Modern UI**: Complete web interface for all operations
- **Real-time Analytics**: Live command execution monitoring
- **Visual Workflow Builder**: Create chains through the web interface
- **Multi-environment Support**: Manage different execution contexts

### 🔌 **Plugin System**
- **Extensible Architecture**: Custom plugins for specialized functionality
- **Plugin Templates**: Quick scaffolding for new plugins
- **Lifecycle Management**: Install, enable, disable, and configure plugins
- **Security Sandboxing**: Safe plugin execution environment

### 📊 **Advanced Analytics**
- **ML-Powered Insights**: Command usage predictions and recommendations
- **Performance Metrics**: Execution time analysis and trending
- **Failure Analysis**: Smart suggestions for failed commands
- **Export Capabilities**: Data export for external analysis

### ⏰ **Scheduling & Automation**
- **Cron Integration**: Schedule commands with flexible timing
- **Interval Scheduling**: Run commands at regular intervals
- **Conditional Execution**: Schedule based on system states
- **Monitoring Dashboard**: Track scheduled command execution

## 🚀 Installation

### Quick Install (Recommended)

```bash
# Using Go (requires Go 1.23.0+)
go install github.com/gi4nks/ambros/v3/cmd@latest

# Verify installation
ambros version
```

### From Source

```bash
# Clone the repository
git clone https://github.com/gi4nks/ambros.git
cd ambros

# Build the project
make build

# Install binary
sudo cp bin/ambros /usr/local/bin/
```

### Platform-Specific Builds

```bash
# Build for all platforms
make build-all

# Available binaries:
# - bin/ambros-darwin (macOS)
# - bin/ambros-linux (Linux)
# - bin/ambros-windows.exe (Windows)
```

## 📖 Usage Guide

### Basic Command Execution

```bash
# Run and store a command
ambros run -- ls -la
ambros run --store -- echo "hello world"

# Add metadata
ambros run --tag backup,sync --category file-ops -- rsync -av src/ dest/

# Dry run (preview without execution)
ambros run --dry-run -- rm -rf /important/files
```

### Command History & Search

```bash
# View recent commands
ambros last
ambros last --limit 20 --failed

# Search command history
ambros search "docker"
ambros search --tag backup --since 24h --status success
ambros search --category deployment --format json

# View command output
ambros output --id CMD-123456
```

### Template Management

```bash
# Save a command as template
ambros template save backup-db "pg_dump -h localhost -U user dbname"

# List templates
ambros template list

# Run template with parameters
ambros template run backup-db --args production-db

# Show template details
ambros template show backup-db

# Delete template
ambros template delete backup-db
```

### Command Chains (Workflows)

```bash
# Create a deployment chain
ambros chain create deploy "build,test,package,deploy" \
  --desc "Full deployment pipeline"

# Execute chain
ambros chain exec deploy
ambros chain exec deploy --parallel --retry 2 --timeout 10m

# Interactive execution
ambros chain exec deploy --interactive

# Dry run chain
ambros chain exec deploy --dry-run

# List chains
ambros chain list

# Export/Import chains
ambros chain export deploy > deployment-chain.json
ambros chain import deployment-chain.json
```

### Plugin Management

```bash
# List installed plugins
ambros plugin list

# Create a custom plugin
ambros plugin create docker-integration

# Install plugin template
ambros plugin install slack-notifications

# Enable/disable plugins
ambros plugin enable docker-integration
ambros plugin disable old-plugin

# Show plugin info
ambros plugin info docker-integration

# Configure plugin
ambros plugin config docker-integration
```

### Scheduling

```bash
# Schedule a command
ambros scheduler add CMD-123 "0 9 * * 1-5"  # Weekdays at 9 AM
ambros scheduler add CMD-456 --cron "*/15 * * * *"  # Every 15 minutes
ambros scheduler add CMD-789 --interval 1h  # Every hour

# List scheduled commands
ambros scheduler list

# Enable/disable schedules
ambros scheduler enable CMD-123
ambros scheduler disable CMD-456

# Remove schedule
ambros scheduler remove CMD-123

# View scheduler status
ambros scheduler status
```

### Analytics & Insights

```bash
# View command analytics
ambros analytics
ambros analytics summary
ambros analytics most-used
ambros analytics slowest
ambros analytics failures

# Advanced analytics via web dashboard
ambros server --port 8080
# Visit: http://localhost:8080/#analytics
```

### Web Dashboard

```bash
# Start web server
ambros server --port 8080 --host localhost

# Development mode with CORS
ambros server --dev --cors --verbose

# Access dashboard at http://localhost:8080
```

### Data Management

```bash
# Export command history
ambros export --output commands.json
ambros export --output backup.yaml --format yaml --tag backup

# Import commands
ambros import --input backup.json --format json
ambros import --input data.yaml --format yaml --merge

# Store arbitrary commands
ambros store "echo hello" --name greeting --tag demo
```

### Interactive Mode

```bash
# Enter interactive mode
ambros interactive

# Features available in interactive mode:
# - Command execution with real-time feedback
# - Template management
# - Chain execution
# - Environment switching
# - Analytics viewing
```

## 🏗️ Project Structure

```
ambros/
├── cmd/
│   ├── main.go                 # Application entry point
│   ├── commands/               # CLI command implementations
│   │   ├── run.go             # Core command execution
│   │   ├── search.go          # Command search & filtering
│   │   ├── template.go        # Template management
│   │   ├── chain.go           # Workflow orchestration
│   │   ├── plugin.go          # Plugin system
│   │   ├── scheduler.go       # Command scheduling
│   │   ├── server.go          # Web dashboard server
│   │   ├── analytics.go       # Usage analytics
│   │   └── ...
│   └── migrate/               # Database migration tools
├── internal/
│   ├── models/                # Data structures
│   ├── repos/                 # Data access layer
│   ├── utils/                 # Shared utilities
│   ├── chain/                 # Chain execution engine
│   ├── scheduler/             # Scheduling logic
│   └── analytics/             # Analytics engine
├── web/                       # Web dashboard assets
├── plugins/                   # Plugin templates
└── docs/                      # Documentation
```

## 🛠️ Development

### Prerequisites

- **Go 1.23.0+** (required)
- **Make** (for build automation)
- **Git** (for version control)

### Development Workflow

```bash
# Clone repository
git clone https://github.com/gi4nks/ambros.git
cd ambros

# Install development dependencies
make dev-deps

# Run tests
make test
make test-integration
make test-all

# Check code coverage
make coverage

# Lint code
make lint

# Format code
make fmt

# Build binary
make build

# Build for all platforms
make build-all
```

### Available Make Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the binary |
| `make build-all` | Build for all platforms |
| `make test` | Run unit tests |
| `make test-integration` | Run integration tests |
| `make coverage` | Run tests with coverage |
| `make lint` | Run linters |
| `make fmt` | Format code |
| `make clean` | Clean build artifacts |
| `make deps` | Download dependencies |
| `make dev-deps` | Install development tools |
| `make help` | Show all available commands |

### Testing

```bash
# Run all tests
make test

# Run with coverage
make coverage

# Run benchmarks
make benchmark

# Run specific test files
go test ./cmd/commands -run TestPluginCommand -v
go test ./internal/repos -v
```

### Contributing

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Write tests** for your changes
4. **Ensure** all tests pass (`make test`)
5. **Format** your code (`make fmt`)
6. **Commit** your changes (`git commit -m 'Add amazing feature'`)
7. **Push** to the branch (`git push origin feature/amazing-feature`)
8. **Open** a Pull Request

### Code Style

- Follow standard Go conventions
- Write comprehensive tests for new features
- Add documentation for public APIs
- Use meaningful commit messages
- Ensure all linters pass

## ⚙️ Configuration

Ambros stores configuration and data in `~/.ambros/` by default.

### Configuration File

Create `~/.ambros/config.yaml`:

```yaml
# Database settings
repositoryDirectory: "~/.ambros"
repositoryFile: "ambros.db"

# Default settings
lastCountDefault: 10
debugMode: false

# Web server settings
server:
  host: "localhost"
  port: 8080
  cors: true

# Plugin settings
plugins:
  directory: "~/.ambros/plugins"
  allowUnsafe: false

# Scheduling settings
scheduler:
  enabled: true
  checkInterval: "1m"

# Analytics settings
analytics:
  enabled: true
  retentionDays: 365
```

### Environment Variables

```bash
export AMBROS_CONFIG_DIR="$HOME/.ambros"
export AMBROS_DEBUG=true
export AMBROS_LOG_LEVEL=debug
```

## 🔄 Migration from Previous Versions

If upgrading from BoltDB to BadgerDB:

```bash
# Backup existing database
make migrate-backup

# Run migration
make migrate

# Or do both at once
make migrate-all

# Verify migration
ambros last --limit 5
```

## 🌐 Web Dashboard Features

The web dashboard provides a complete interface for all Ambros features:

### Dashboard Sections

- **📊 Overview**: Real-time metrics and recent activity
- **💻 Commands**: Execute and manage commands
- **📋 Templates**: Template library and editor
- **🔗 Chains**: Workflow builder and execution
- **🔌 Plugins**: Plugin marketplace and management
- **⏰ Scheduler**: Scheduled task management
- **📈 Analytics**: Advanced usage insights
- **⚙️ Settings**: Configuration and preferences

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/dashboard` | GET | Dashboard metrics |
| `/api/commands` | GET/POST | Command management |
| `/api/templates` | GET/POST/PUT/DELETE | Template operations |
| `/api/chains` | GET/POST/PUT/DELETE | Chain management |
| `/api/plugins` | GET/POST | Plugin management |
| `/api/scheduler` | GET/POST | Scheduled tasks |
| `/api/analytics/advanced` | GET | Advanced analytics |
| `/api/search/smart` | GET | Enhanced search |

## 🔌 Plugin Development

### Creating a Plugin

```bash
# Generate plugin template
ambros plugin create my-awesome-plugin

# Plugin structure created:
# ~/.ambros/plugins/my-awesome-plugin/
# ├── plugin.json          # Plugin manifest
# ├── my-awesome-plugin.sh # Main executable
# └── README.md           # Documentation
```

### Plugin Manifest (`plugin.json`)

```json
{
  "name": "my-awesome-plugin",
  "version": "1.0.0",
  "description": "An awesome plugin for Ambros",
  "author": "Your Name",
  "enabled": true,
  "executable": "./my-awesome-plugin.sh",
  "commands": [
    {
      "name": "awesome-cmd",
      "description": "Does awesome things",
      "usage": "awesome-cmd [options]",
      "args": ["option1", "option2"]
    }
  ],
  "hooks": ["pre-run", "post-run"],
  "config": {
    "setting1": "value1",
    "setting2": "value2"
  },
  "dependencies": ["curl", "jq"]
}
```

### Plugin API

Plugins can access Ambros data through environment variables:

```bash
# Available in plugin environment
$AMBROS_COMMAND_ID      # Current command ID
$AMBROS_COMMAND_NAME    # Command name
$AMBROS_CONFIG_DIR      # Ambros config directory
$AMBROS_PLUGIN_CONFIG   # Plugin configuration (JSON)
```

## 🚀 Production Deployment

### Single Binary Deployment

```bash
# Build optimized binary
make build

# Deploy to server
scp bin/ambros user@server:/usr/local/bin/

# Start as systemd service
sudo systemctl enable ambros-server
sudo systemctl start ambros-server
```

### Docker Deployment

```bash
# Build Docker image
make docker-build

# Run container
docker run -d \
  --name ambros \
  -p 8080:8080 \
  -v ambros-data:/root/.ambros \
  ambros:latest server --host 0.0.0.0
```

### Environment-Specific Configurations

```bash
# Development
ambros server --dev --cors --verbose

# Production
ambros server --host 0.0.0.0 --port 8080 --secure

# Enterprise
ambros server --auth --audit-log --metrics
```

## 📊 Use Cases

### Individual Developers
- **Command History**: Never lose track of useful commands
- **Template Library**: Reuse complex command patterns
- **Analytics**: Understand your workflow patterns
- **Automation**: Chain related commands together

### Development Teams
- **Shared Templates**: Team-wide command standardization
- **Workflow Automation**: Deployment and testing pipelines
- **Knowledge Sharing**: Document and share command expertise
- **Onboarding**: New team members learn from command history

### DevOps & SRE Teams
- **Infrastructure Automation**: Complex deployment workflows
- **Monitoring Integration**: Command-based health checks
- **Incident Response**: Pre-defined troubleshooting workflows
- **Compliance**: Audit trail of all executed commands

### Enterprise Users
- **Web Dashboard**: No terminal required for basic operations
- **Plugin Ecosystem**: Custom integrations with enterprise tools
- **Analytics & Reporting**: Command usage insights and optimization
- **Security & Audit**: Complete execution tracking and analysis

## 🤝 Community & Support

### Getting Help

- **📖 Documentation**: [GitHub Wiki](https://github.com/gi4nks/ambros/wiki)
- **🐛 Bug Reports**: [GitHub Issues](https://github.com/gi4nks/ambros/issues)
- **💡 Feature Requests**: [GitHub Discussions](https://github.com/gi4nks/ambros/discussions)
- **💬 Community**: [Discord Server](https://discord.gg/ambros) (coming soon)

### Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Roadmap

- **Phase 4**: Advanced ML features and predictive analytics
- **Phase 5**: Multi-user support and team collaboration
- **Phase 6**: Cloud integration and hosted service
- **Phase 7**: Enterprise security and compliance features

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **[Cobra](https://github.com/spf13/cobra)** - CLI framework
- **[BadgerDB](https://github.com/dgraph-io/badger)** - Embedded database
- **[Zap](https://github.com/uber-go/zap)** - Logging framework
- **[Testify](https://github.com/stretchr/testify)** - Testing toolkit

---

**Made with ❤️ by the Ambros community**

*Transform your command-line experience with Ambros - The Command Butler!*