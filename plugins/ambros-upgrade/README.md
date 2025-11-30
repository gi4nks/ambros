# Ambros Upgrade Plugin

A shell-based plugin for updating and upgrading Ambros to the latest or a specific version.

## Overview

This plugin manages the Ambros binary lifecycle:
- Check for available updates
- Upgrade to the latest version
- Install specific versions
- Rollback to previous versions
- Automatic backup before upgrades

## Installation

The plugin is included in the `plugins/` directory of the Ambros repository.

```bash
# Copy plugin to Ambros plugins directory
cp -r plugins/ambros-upgrade ~/.ambros/plugins/

# Or create a symlink for development
ln -s $(pwd)/plugins/ambros-upgrade ~/.ambros/plugins/ambros-upgrade

# Make the script executable
chmod +x ~/.ambros/plugins/ambros-upgrade/ambros-upgrade.sh

# Enable the plugin
ambros plugin enable ambros-upgrade
```

## Commands

### `check`
Check if updates are available.

```bash
ambros ambros-upgrade check
```

**Output:**
```
═══════════════════════════════════════════
        Ambros Update Check
═══════════════════════════════════════════

  Current version: v3.2.0
  Latest version:  v3.3.0

  ✓ Update available!

  Run 'ambros ambros-upgrade upgrade' to update
  Or 'ambros ambros-upgrade install v3.3.0' to install specific version
```

### `current`
Show detailed information about the currently installed version.

```bash
ambros ambros-upgrade current
```

**Output:**
```
Ambros Version Information

  Installed version: v3.2.0
  Install path:      /usr/local/bin/ambros
  Backup directory:  /Users/you/.ambros/backups

  Binary size:       15M
  Modified:          Nov 30 10:00
```

### `list [--limit N]`
List available versions from GitHub releases.

```bash
# List 10 most recent releases (default)
ambros ambros-upgrade list

# List 20 releases
ambros ambros-upgrade list --limit 20
```

**Output:**
```
═══════════════════════════════════════════
        Available Ambros Versions
═══════════════════════════════════════════

  VERSION      DATE         TYPE
  ─────────────────────────────────────────
  v3.3.0       2024-12-01   stable
  v3.2.0       2024-11-30   stable       (installed)
  v3.1.4       2024-11-25   stable
  v3.1.2       2024-11-20   stable
  v3.1.0       2024-11-15   stable

  Showing 5 most recent releases
  Use --limit N to show more
```

### `upgrade [--force]`
Upgrade to the latest available version.

```bash
# Upgrade to latest
ambros ambros-upgrade upgrade

# Force reinstall even if already on latest
ambros ambros-upgrade upgrade --force
```

**Interactive Output:**
```
═══════════════════════════════════════════
        Ambros Upgrade
═══════════════════════════════════════════

  Current version: v3.2.0
  Target version:  v3.3.0
  Platform:        darwin_arm64

Proceed with upgrade? [y/N] y

ℹ Downloading ambros_3.3.0_darwin_arm64.tar.gz...
✓ Downloaded successfully
ℹ Extracting archive...
ℹ Backing up current version to /Users/you/.ambros/backups/ambros_v3.2.0_20241201_100000...
✓ Backup created
ℹ Installing to /usr/local/bin/ambros...
✓ Installation complete

═══════════════════════════════════════════
        Upgrade Complete!
═══════════════════════════════════════════

  Previous version: v3.2.0
  New version:      v3.3.0
```

### `install <version>`
Install a specific version of Ambros.

```bash
# Install specific version
ambros ambros-upgrade install v3.1.0

# Version prefix 'v' is optional
ambros ambros-upgrade install 3.1.0
```

### `rollback`
Rollback to a previously backed up version.

```bash
ambros ambros-upgrade rollback
```

**Interactive Output:**
```
═══════════════════════════════════════════
        Ambros Rollback
═══════════════════════════════════════════

  Current version: v3.3.0

  Available backups:
    1) v3.2.0 (backed up: 2024-12-01 10:00)
    2) v3.1.4 (backed up: 2024-11-28 14:30)
    3) v3.1.0 (backed up: 2024-11-20 09:15)

Select backup to restore (1-3) or 'q' to quit: 1

ℹ Restoring v3.2.0...
✓ Created pre-rollback backup
✓ Restored successfully

═══════════════════════════════════════════
        Rollback Complete!
═══════════════════════════════════════════

  Previous version: v3.3.0
  Restored version: v3.2.0
```

## Configuration

The plugin can be configured via `plugin.json`:

```json
{
  "config": {
    "github_repo": "gi4nks/ambros",
    "install_path": "/usr/local/bin/ambros",
    "backup_dir": "~/.ambros/backups"
  }
}
```

| Option | Default | Description |
|--------|---------|-------------|
| `github_repo` | `gi4nks/ambros` | GitHub repository to fetch releases from |
| `install_path` | `/usr/local/bin/ambros` | Where to install the Ambros binary |
| `backup_dir` | `~/.ambros/backups` | Directory to store version backups |

## Platform Support

The plugin automatically detects your platform and downloads the appropriate binary:

| OS | Architecture | Asset Name |
|----|--------------|------------|
| macOS | Intel | `ambros_X.Y.Z_darwin_amd64.tar.gz` |
| macOS | Apple Silicon | `ambros_X.Y.Z_darwin_arm64.tar.gz` |
| Linux | x86_64 | `ambros_X.Y.Z_linux_amd64.tar.gz` |
| Linux | ARM64 | `ambros_X.Y.Z_linux_arm64.tar.gz` |
| Windows | x86_64 | `ambros_X.Y.Z_windows_amd64.zip` |

## Backup System

The plugin automatically creates backups before any upgrade:

- Backups are stored in `~/.ambros/backups/`
- Naming format: `ambros_vX.Y.Z_YYYYMMDD_HHMMSS`
- Use `rollback` to restore any backup
- A pre-rollback backup is also created when rolling back

### Managing Backups

```bash
# List backups
ls -la ~/.ambros/backups/

# Manual cleanup (keep last 5)
ls -t ~/.ambros/backups/ambros_* | tail -n +6 | xargs rm -f
```

## Dependencies

- `curl` - For downloading releases
- `jq` - For parsing GitHub API responses
- `tar` - For extracting archives (Linux/macOS)
- `unzip` - For extracting archives (Windows)

Install dependencies:

```bash
# macOS
brew install curl jq

# Ubuntu/Debian
sudo apt-get install curl jq

# CentOS/RHEL
sudo yum install curl jq
```

## Troubleshooting

### "Could not determine latest version"
- Check your internet connection
- Verify the GitHub API is accessible: `curl -s https://api.github.com/repos/gi4nks/ambros/releases/latest`

### "Failed to download release asset"
- The release may not have pre-built binaries for your platform
- Check available assets at: https://github.com/gi4nks/ambros/releases

### "Permission denied"
- The plugin may need sudo to write to `/usr/local/bin`
- It will prompt for sudo automatically if needed

### "No backup versions found"
- No previous upgrades have been performed
- Backups are only created during upgrades

### Binary not working after upgrade
- Use `rollback` to restore the previous version
- Check if the binary is executable: `chmod +x /usr/local/bin/ambros`

## Security Notes

- Downloads are fetched from official GitHub releases only
- The plugin verifies downloads are from the configured `github_repo`
- Consider verifying checksums manually for critical installations

## License

Apache 2.0 - Same as Ambros
