# Ambros Release Plugin

A shell-based plugin for automating the Ambros release workflow.

## Overview

This plugin streamlines the release process for Ambros by automating:
- Version bumping (major, minor, patch)
- Creating release branches
- Creating annotated tags
- Pushing to remote repository

## Installation

The plugin is included in the `plugins/` directory of the Ambros repository. To install it:

```bash
# Copy plugin to Ambros plugins directory
cp -r plugins/ambros-release ~/.ambros/plugins/

# Or create a symlink for development
ln -s $(pwd)/plugins/ambros-release ~/.ambros/plugins/ambros-release

# Enable the plugin
ambros plugin enable ambros-release
```

## Commands

### `version`
Show the current version of Ambros.

```bash
ambros ambros-release version
```

**Output:**
```
Current version: v3.2.0
```

### `status`
Show comprehensive release status including current version, branch, uncommitted changes, and recent tags.

```bash
ambros ambros-release status
```

**Output:**
```
═══════════════════════════════════════════
        Ambros Release Status
═══════════════════════════════════════════

  Version:     v3.2.0
  Branch:      release/v3.2.0
  Remote:      origin (https://github.com/gi4nks/ambros.git)
  Version file: cmd/commands/version.txt

  ✓ Working tree is clean
  ✓ Tag v3.2.0 exists

  Recent tags:
    v3.2.0
    v3.1.4
    v3.1.2
    v3.1.1
    v3.1.0
```

### `bump <type>`
Bump the version without committing. Use this when you want more control over the release process.

```bash
# Bump patch version (v3.2.0 → v3.2.1)
ambros ambros-release bump patch

# Bump minor version (v3.2.0 → v3.3.0)
ambros ambros-release bump minor

# Bump major version (v3.2.0 → v4.0.0)
ambros ambros-release bump major
```

### `tag [--message "msg"]`
Create and push a tag for the current version.

```bash
# Create tag with default message
ambros ambros-release tag

# Create tag with custom message
ambros ambros-release tag --message "Hotfix for issue #123"
```

### `push`
Push the current branch and all tags to the remote.

```bash
ambros ambros-release push
```

### `create-branch`
Create a release branch for the current version without the full release workflow.

```bash
ambros ambros-release create-branch
```

This will:
1. Create a branch named `release/vX.Y.Z` from current HEAD
2. Optionally push it to the remote

### `changelog [from-tag]`
Generate a changelog from the previous tag (or a specified tag) to the current HEAD.

```bash
# Generate changelog from previous tag
ambros ambros-release changelog

# Generate changelog from a specific tag
ambros ambros-release changelog v3.1.0
```

**Output:**
```
═══════════════════════════════════════════
        Changelog: v3.1.4 → v3.2.0
═══════════════════════════════════════════

## Features
- feat: add deep analytics ... (abc123)

## Bug Fixes
- fix: resolve memory leak ... (def456)

## Documentation
- docs: update README ... (ghi789)

## Other Changes
- chore: bump dependencies ... (jkl012)

All commits:
  - feat: add deep analytics (abc123) by Developer
  - fix: resolve memory leak (def456) by Developer
  ...

Statistics:
  Commits: 15
  Contributors: 2
  25 files changed, 1500 insertions(+), 300 deletions(-)
```

### `release <type> [--message "msg"]`
**The main command** - performs the complete release workflow:

1. Checks working tree status
2. Updates version file
3. Commits all changes
4. Creates release branch (`release/vX.Y.Z`)
5. Creates annotated tag
6. Pushes branch and tag to origin

```bash
# Patch release (v3.2.0 → v3.2.1)
ambros ambros-release release patch

# Minor release with message (v3.2.0 → v3.3.0)
ambros ambros-release release minor --message "Added new analytics features"

# Major release (v3.2.0 → v4.0.0)
ambros ambros-release release major --message "Breaking changes: New plugin API"
```

**Interactive Output:**
```
═══════════════════════════════════════════
        Ambros Release Workflow
═══════════════════════════════════════════

  Current version: v3.2.0
  New version:     v3.2.1
  Release branch: release/v3.2.1
  Remote:          origin

Proceed with release? [y/N] y

ℹ Step 1/6: Checking working tree...
✓ Working tree is clean
ℹ Step 2/6: Updating version to v3.2.1...
✓ Updated cmd/commands/version.txt to v3.2.1
ℹ Step 3/6: Committing changes...
✓ Committed release changes
ℹ Step 4/6: Creating release branch release/v3.2.1...
✓ Created and switched to branch release/v3.2.1
ℹ Step 5/6: Creating tag v3.2.1...
✓ Created tag v3.2.1
ℹ Step 6/6: Pushing to origin...
✓ Pushed branch release/v3.2.1
✓ Pushed tag v3.2.1

═══════════════════════════════════════════
        Release v3.2.1 Complete!
═══════════════════════════════════════════

  Release branch: release/v3.2.1
  Tag: v3.2.1

  GitHub Release URL:
  https://github.com/gi4nks/ambros/releases/tag/v3.2.1
```

## Configuration

The plugin can be configured via `plugin.json`:

```json
{
  "config": {
    "version_file": "cmd/commands/version.txt",
    "remote": "origin",
    "branch_prefix": "release/"
  }
}
```

| Option | Default | Description |
|--------|---------|-------------|
| `version_file` | `cmd/commands/version.txt` | Path to the version file relative to repo root |
| `remote` | `origin` | Git remote to push to |
| `branch_prefix` | `release/` | Prefix for release branches |

## Versioning Scheme

Ambros uses [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality in a backwards compatible manner
- **PATCH** version for backwards compatible bug fixes

Version format: `vMAJOR.MINOR.PATCH` (e.g., `v3.2.0`)

## Release Workflow

### Standard Release Process

1. **Complete your feature work** on your development branch
2. **Run tests** to ensure everything passes:
   ```bash
   go test ./... -count=1
   ```
3. **Check release status**:
   ```bash
   ambros ambros-release status
   ```
4. **Create the release**:
   ```bash
   ambros ambros-release release minor --message "Description of changes"
   ```
5. **Create a GitHub Release** (optional):
   - Go to the GitHub releases page
   - The tag is already pushed
   - Add release notes and publish

### Hotfix Release Process

For urgent fixes to production:

1. **Checkout the production branch**:
   ```bash
   git checkout release/v3.2.0
   ```
2. **Make your fix**
3. **Create patch release**:
   ```bash
   ambros ambros-release release patch --message "Hotfix: Description"
   ```

### Manual Release Process

If you need more control:

1. **Bump version only**:
   ```bash
   ambros ambros-release bump minor
   ```
2. **Make additional changes**
3. **Commit manually**:
   ```bash
   git add -A
   git commit -m "Release v3.3.0: Description"
   ```
4. **Create branch and tag**:
   ```bash
   git checkout -b release/v3.3.0
   ambros ambros-release tag --message "Release description"
   ```
5. **Push**:
   ```bash
   ambros ambros-release push
   ```

## Dependencies

- `git` - For version control operations
- `sed` - For text processing (usually pre-installed)
- `jq` - Optional, for parsing plugin config JSON

## Troubleshooting

### "Tag already exists"
The tag for this version was already created. Either:
- Delete the existing tag: `git tag -d vX.Y.Z && git push origin :refs/tags/vX.Y.Z`
- Or bump to a new version

### "Not in a git repository"
Make sure you're running the command from within the Ambros repository.

### "Version file not found"
Check that `cmd/commands/version.txt` exists. If your version file is in a different location, update `plugin.json` config.

### "Permission denied"
Make sure the script is executable:
```bash
chmod +x ~/.ambros/plugins/ambros-release/ambros-release.sh
```

## License

Apache 2.0 - Same as Ambros
