# Ambros User Manual

**Version 3.2.7**

Ambros is a powerful command-line tool for storing, managing, and analyzing shell command executions. It captures command outputs, tracks success/failure status, and provides advanced analytics to help you understand your command-line workflow.

---

## Table of Contents

1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [Core Concepts](#core-concepts)
4. [Command Reference](#command-reference)
   - [Running Commands](#running-commands)
   - [Viewing History](#viewing-history)
   - [Searching Commands](#searching-commands)
   - [Plugins](#plugins)
   - [Analytics](#analytics)
   - [Import/Export](#importexport)
   - [Database Cleanup](#database-cleanup)
   - [Configuration](#configuration)
5. [Web Interface](#web-interface)
6. [Advanced Usage](#advanced-usage)
7. [Troubleshooting](#troubleshooting)

---

## Installation

### Using Go Install (Recommended)

```bash
go install github.com/gi4nks/ambros/v3@latest
```

### From Source

```bash
git clone https://github.com/gi4nks/ambros.git
cd ambros
go build -o ambros .
sudo mv ambros /usr/local/bin/
```

### Verify Installation

```bash
ambros version
```

---

## Quick Start

### 1. Run Your First Command

```bash
# Run a command and store it
ambros run -- ls -la

# Run with tags for organization
ambros run --tag project-x --tag deployment -- kubectl get pods
```

### 2. View Recent Commands

```bash
# Show last 10 commands
ambros last 10

# Search for specific commands
ambros search "docker"
```

### 3. Recall and Rerun

```bash
# Recall a command by ID
ambros recall <command-id>

# Rerun a command
ambros rerun <command-id>
```

---

## Core Concepts

### Commands
Every command executed through Ambros is stored with:
- **ID**: Unique identifier
- **Name**: The command name (e.g., `docker`, `kubectl`)
- **Arguments**: Command arguments
- **Output**: Captured stdout/stderr
- **Status**: Success (exit code 0) or failure
- **Timestamps**: Creation and termination time
- **Tags**: Optional labels for organization
- **Category**: Optional grouping

### Plugins
Plugins extend Ambros functionality with custom scripts:
```bash
# List plugins
ambros plugin list

# Run a plugin command
ambros plugin run ambros-release version
```

---

## Command Reference

### Running Commands

#### `ambros run`
Execute a command and store it in the database.

```bash
# Basic usage (use -- to separate ambros flags from command)
ambros run -- <command> [arguments...]

# With tags
ambros run --tag <tag1> --tag <tag2> -- <command>

# With category
ambros run --category deployment -- kubectl apply -f app.yaml

# Auto mode (interactive TTY)
ambros run --auto -- vim file.txt

# Dry run (don't execute)
ambros run --dry-run -- rm -rf /important

# Without storing
ambros run --store=false -- curl https://api.example.com
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--tag` | `-t` | Add tags to the command (repeatable) |
| `--category` | `-c` | Assign a category |
| `--auto` | | Enable TTY mode for interactive commands |
| `--dry-run` | | Preview without executing |
| `--store` | | Store command execution (default: true) |

> **Important:** Use `--` to separate Ambros flags from the command and its arguments.
> This prevents flags like `-la` from being interpreted as Ambros flags.

---

### Viewing History

#### `ambros last`
Show the most recent commands.

```bash
# Show last 10 commands (default)
ambros last

# Show last N commands
ambros last 20

# Show with full output
ambros last 5 --output
```

#### `ambros recall`
Retrieve details of a specific command.

```bash
# By ID
ambros recall <command-id>

# Show full output
ambros recall <command-id> --output
```

#### `ambros rerun`
Re-execute a previously stored command.

```bash
# Rerun by ID
ambros rerun <command-id>

# Rerun with new tag
ambros rerun <command-id> --tag rerun-test
```

---

### Searching Commands

#### `ambros search`
Search for commands by text, tags, or filters.

```bash
# Text search
ambros search "docker"

# Search by tag
ambros search --tag deployment

# Search by category
ambros search --category kubernetes

# Search by status
ambros search --status success
ambros search --status failed

# Combine filters
ambros search "kubectl" --tag prod --status success

# Limit results
ambros search "git" --limit 50

# Search with date range
ambros search --from "2024-01-01" --to "2024-12-31"
```

**Flags:**
| Flag | Description |
|------|-------------|
| `--tag` | Filter by tag |
| `--category` | Filter by category |
| `--status` | Filter by status (success/failed) |
| `--limit` | Maximum results |
| `--from` | Start date |
| `--to` | End date |
| `--output` | Include command output |

---

### Plugins

#### `ambros plugin`
Manage and run plugins.

```bash
# List installed plugins
ambros plugin list

# Install a plugin from local path
ambros plugin install /path/to/plugin

# Install from registry
ambros plugin install <plugin-name>

# Run a plugin command (stored in history)
ambros plugin run <plugin-name> <command> [args...]

# Examples
ambros plugin run ambros-release version
ambros plugin run ambros-release status
ambros plugin run ambros-upgrade check

# Enable/disable plugins
ambros plugin enable <name>
ambros plugin disable <name>

# Show plugin info
ambros plugin info <name>

# Plugin configuration
ambros plugin config <name> list
ambros plugin config <name> get <key>
ambros plugin config <name> set <key> <value>

# Create a new plugin template
ambros plugin create my-plugin

# Uninstall a plugin
ambros plugin uninstall <name>

# Manage plugin registries
ambros plugin registry list
ambros plugin registry add <name> <url>
ambros plugin registry remove <name>
```

**Built-in Plugins:**

| Plugin | Description |
|--------|-------------|
| `ambros-release` | Automate release workflow (version, tag, push) |
| `ambros-upgrade` | Update Ambros to newer versions |

---

### Analytics

#### `ambros analytics`
View command usage statistics and insights.

```bash
# Show basic analytics
ambros analytics

# Show detailed analytics
ambros analytics --detailed

# Analytics for specific time period
ambros analytics --from "2024-01-01" --to "2024-12-31"

# Export analytics as JSON
ambros analytics --format json
```

**Analytics includes:**
- Total commands executed
- Success/failure rates
- Most used commands
- Command frequency by hour/day
- Tag distribution
- Category breakdown
- Alias suggestions for frequent long commands
- Command sequence patterns
- Workflow detection (Git, Docker, K8s, etc.)

---

### Import/Export

#### `ambros export`
Export command history to a file.

```bash
# Export all commands to JSON
ambros export commands.json

# Export with filters
ambros export --tag important backup.json

# Export specific format
ambros export --format json commands.json
ambros export --format csv commands.csv
```

#### `ambros import`
Import commands from a file.

```bash
# Import from JSON
ambros import commands.json

# Import with tag
ambros import --tag imported backup.json
```

---

### Database Cleanup

#### `ambros interactive cleanup`
Interactive database cleanup with TUI menu.

```bash
# Start interactive cleanup
ambros interactive cleanup

# Or just
ambros interactive
```

**Cleanup options:**
1. Remove failed commands
2. Remove commands older than 30 days
3. Remove duplicate commands
4. Full cleanup (all above)

Features dry-run mode and confirmation prompts.

---

### Configuration

#### `ambros db`
Database management commands.

```bash
# Show database info
ambros db info

# Compact/optimize database
ambros db compact

# Backup database
ambros db backup /path/to/backup.db
```

---

### Server & Web Interface

#### `ambros server`
Start the web server for the UI and API.

```bash
# Start server on default port (8080)
ambros server

# Start on specific port
ambros server --port 3000

# Start with specific host
ambros server --host 0.0.0.0 --port 8080
```

**API Endpoints:**

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/commands` | GET | List commands |
| `/api/commands/:id` | GET | Get command by ID |
| `/api/commands` | POST | Execute command |
| `/api/search` | GET | Search commands |
| `/api/analytics` | GET | Basic analytics |
| `/api/analytics/advanced` | GET | Advanced analytics |

---

## Web Interface

Access the web interface at `http://localhost:8080` after starting the server.

### Features:
- **Dashboard**: Overview of recent commands and statistics
- **Command History**: Browse and search all commands
- **Search**: Advanced search with filters
- **Analytics**: Visual charts and insights

---

## Advanced Usage

### Shell Integration

Ambros provides transparent shell integration that automatically tracks your commands without needing to type `ambros run` every time.

#### Installation

```bash
# Interactive installation (updates ~/.bashrc and ~/.zshrc)
ambros integrate install

# Non-interactive (for scripts/CI)
ambros integrate install --yes

# Target specific shell
ambros integrate install --shell ~/.zshrc
```

#### Integration Modes

**Whitelist Mode (Default)** - Only tracks specific commands:

```bash
# Uses default list: ls, git, curl, wget, npm, yarn, docker, ssh, 
# kubectl, helm, terraform, go, python, node, java, mvn, gradle, etc.
source ~/.ambros-integration.sh
```

**All Commands Mode** - Tracks everything except blacklisted commands:

```bash
# Enable before sourcing the script
export AMBROS_INTEGRATION_MODE="all"
source ~/.ambros-integration.sh
```

#### Customizing the Whitelist

```bash
# In your .zshrc or .bashrc, BEFORE sourcing the integration:
AMBROS_INTERCEPTED_COMMANDS=(
    "git" "docker" "kubectl" "terraform" "npm"
    # Add your own commands here
)
source ~/.ambros-integration.sh
```

#### Customizing the Blacklist

Commands in the blacklist are NEVER tracked (prevents issues with shell builtins):

```bash
# Add to blacklist (in your .zshrc/.bashrc before sourcing)
AMBROS_BLACKLIST=(
    # Defaults include: cd, pwd, echo, exit, source, ambros, vim, less, sudo...
    "my-sensitive-command"
    "internal-tool"
)
source ~/.ambros-integration.sh
```

#### Runtime Controls

```bash
# Check integration status
ambros_status

# Temporarily disable for current session
ambros_disable

# Re-enable (requires shell restart or re-source)
ambros_enable

# Add command to blacklist (session only)
ambros_blacklist_add mycommand

# Remove from blacklist (session only)
ambros_blacklist_remove mycommand
```

#### Bypass for Single Commands

```bash
# Use 'command' builtin to bypass the wrapper
command git status

# Or use full path
/usr/bin/git status
```

#### Uninstallation

```bash
ambros integrate uninstall
```

---

### Manual Shell Aliases

If you prefer manual control instead of full integration, add to your `.bashrc` or `.zshrc`:

```bash
# Alias for quick command execution (note the --)
alias ar="ambros run --"

# Function to run and tag
art() {
    ambros run --tag "$1" -- "${@:2}"
}

# Function to run with category
arc() {
    ambros run --category "$1" -- "${@:2}"
}
```

### Using with Scripts

```bash
#!/bin/bash
# Example: Deployment script with Ambros tracking

# Run deployment with tracking
ambros run --tag deployment --category production -- \
    kubectl apply -f deployment.yaml

# Check last deployment status
ambros search --tag deployment --limit 1
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `AMBROS_DB_PATH` | Database file location | `~/.ambros/ambros.db` |
| `AMBROS_LOG_LEVEL` | Log verbosity | `info` |
| `AMBROS_PLUGIN_DIR` | Plugin directory | `~/.ambros/plugins` |

### Debug Mode

```bash
# Enable debug logging
ambros --debug run -- ls -la

# Or via environment
AMBROS_LOG_LEVEL=debug ambros run -- ls -la
```

---

## Troubleshooting

### Common Issues

#### "Command not found: ambros"
```bash
# Check if Go bin is in PATH
echo $PATH | grep -q "go/bin" || export PATH=$PATH:$(go env GOPATH)/bin

# Or reinstall
go install github.com/gi4nks/ambros/v3@latest
```

#### "Database locked"
```bash
# Only one Ambros instance can write at a time
# Close other terminals running Ambros or wait

# If stuck, find and kill processes
ps aux | grep ambros
```

#### "Plugin not found"
```bash
# Check plugin is installed
ambros plugin list

# Reinstall plugin
ambros plugin uninstall <name>
ambros plugin install /path/to/plugin
```

#### "Permission denied" for plugins
```bash
# Make plugin executable
chmod +x ~/.ambros/plugins/<plugin-name>/<script>.sh
```

### Getting Help

```bash
# General help
ambros --help

# Command-specific help
ambros run --help
ambros plugin --help

# Version info
ambros version
```

### Reporting Issues

Report bugs at: https://github.com/gi4nks/ambros/issues

Include:
- Ambros version (`ambros version`)
- Operating system
- Steps to reproduce
- Error messages

---

## Quick Reference Card

```
┌─────────────────────────────────────────────────────────────────┐
│                    AMBROS QUICK REFERENCE                       │
├─────────────────────────────────────────────────────────────────┤
│ RUNNING COMMANDS                                                │
│   ambros run -- <cmd>           Execute and store command       │
│   ambros run -t tag -- <cmd>    Run with tag                    │
│   ambros run --auto -- <cmd>    Run in auto/TTY mode            │
│                                                                 │
│ VIEWING HISTORY                                                 │
│   ambros last [n]               Show last n commands            │
│   ambros recall <id>            Show command details            │
│   ambros rerun <id>             Re-execute command              │
│                                                                 │
│ SEARCHING                                                       │
│   ambros search "text"          Search by text                  │
│   ambros search --tag <tag>     Search by tag                   │
│   ambros search --status fail   Search failed commands          │
│                                                                 │
│ PLUGINS                                                         │
│   ambros plugin list            List plugins                    │
│   ambros plugin run <p> <cmd>   Run plugin command              │
│   ambros plugin install <path>  Install plugin                  │
│                                                                 │
│ DATABASE                                                        │
│   ambros interactive cleanup    Interactive cleanup wizard      │
│   ambros export <file>          Export history                  │
│   ambros import <file>          Import history                  │
│                                                                 │
│ OTHER                                                           │
│   ambros analytics              Show statistics                 │
│   ambros server                 Start web interface             │
│   ambros --help                 Show help                       │
└─────────────────────────────────────────────────────────────────┘
```

---

## License

Apache 2.0 - See [LICENSE](../LICENSE) for details.

---

*Last updated: December 2025 - Ambros v3.2.7*
