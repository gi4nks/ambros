# ambros
The command butler!! 

[![Build Status](https://github.com/gi4nks/ambros/workflows/CI/badge.svg)](https://github.com/gi4nks/ambros/actions)

Ambros creates a local history of executed commands, keeping track also of the output. At the moment it does not work with interactive commands.

## Installation

### Using Go

```bash
go install github.com/gi4nks/ambros@latest
```

### From Source

1. Clone the repository
```bash
git clone https://github.com/gi4nks/ambros.git
cd ambros
```

2. Build the project
```bash
make build
```

## Usage

### Run and store a command
```bash
ambros run -- ls -la
ambros run --store -- echo "hello world"
ambros run --tag backup,sync -- rsync -av src/ dest/
```

### Search command history
```bash
ambros search "ls"
ambros search --tag backup --since 24h
ambros search --status failed --format json
```

### View last commands
```bash
ambros last
ambros last --limit 20
ambros last --failed
```

### Show command output
```bash
ambros output --id CMD-123456
```

### Export command history
```bash
ambros export --output commands.json
ambros export --output backup.yaml --format yaml --tag backup
```

### Get version information
```bash
ambros version
ambros version --short
```

### Get help
```bash
ambros help
ambros run --help
```

## Features

- **Command Execution**: Run commands and automatically store execution details
- **History Management**: Keep track of all executed commands with timestamps
- **Search & Filter**: Search through command history with various filters
- **Export**: Export command history in JSON or YAML formats
- **Tagging**: Organize commands with custom tags
- **Templates**: Use command templates for repeated operations
- **Dry Run**: Preview commands before execution

## Development

### Prerequisites

- Go 1.22 or later
- Make

### Available Make Commands

- `make build` - Build the binary
- `make build-all` - Build for all platforms (darwin, linux, windows)
- `make test` - Run tests
- `make coverage` - Run tests with coverage report
- `make lint` - Run linters
- `make clean` - Clean build artifacts
- `make deps` - Download dependencies
- `make dev-deps` - Install development dependencies
- `make help` - Show available commands

### Running Tests

```bash
make test
```

### Running with Coverage

```bash
make coverage
```

## Configuration

Ambros stores its configuration and database in `~/.ambros/` by default. You can customize this with a configuration file.

Example configuration file (`.ambros.yaml`):
```yaml
repositoryDirectory: "~/.ambros"
repositoryFile: "ambros.db"
lastCountDefault: 10
debugMode: false
```

# Migration to BadgerDB

If you're upgrading from a previous version that used BoltDB, you'll need to migrate your data. The project includes a migration tool to help with this process.

## Migration Steps

1. Backup your current database:
```bash
make migrate-backup
```

2. Run the migration:
```bash
make migrate
```

Or run both steps at once:
```bash
make migrate-all
```

By default, the migration tool will:
- Look for your old database in `$HOME/.ambros/ambros.db`
- Create the new database in `$HOME/.ambros/ambros_new.db`

To specify custom paths:
```bash
make migrate OLD_DB=/path/to/old.db NEW_DB=/path/to/new.db
```

## Post-Migration

After successful migration:
1. Verify your data in the new database
2. Update your configuration to point to the new database location
3. Keep the old database as backup until you've verified everything works

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.