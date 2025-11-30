# Ambros Release Workflow

This document describes the release workflow for Ambros, including versioning strategy, branching model, and release procedures.

## Table of Contents

- [Versioning Strategy](#versioning-strategy)
- [Branching Model](#branching-model)
- [Release Process](#release-process)
- [Using the Release Plugin](#using-the-release-plugin)
- [Manual Release Steps](#manual-release-steps)
- [Hotfix Releases](#hotfix-releases)
- [Best Practices](#best-practices)

## Versioning Strategy

Ambros follows [Semantic Versioning 2.0.0](https://semver.org/):

```
MAJOR.MINOR.PATCH
```

- **MAJOR**: Incompatible API changes, breaking changes
- **MINOR**: New functionality in a backwards compatible manner
- **PATCH**: Backwards compatible bug fixes

### Version File Location

The version is stored in `cmd/commands/version.txt` as a simple string:

```
v3.2.0
```

## Branching Model

Ambros uses a simplified GitFlow-inspired branching model:

```
main (stable releases)
  │
  ├── release/v3.2.0 (release preparation)
  │
  └── feature/* (development work)
```

### Branch Types

| Branch | Purpose | Naming Convention |
|--------|---------|-------------------|
| `main` | Stable releases, production-ready code | `main` |
| `release/*` | Release preparation and stabilization | `release/v{VERSION}` |
| `feature/*` | New features and enhancements | `feature/{description}` |
| `bugfix/*` | Bug fixes | `bugfix/{description}` |
| `hotfix/*` | Critical production fixes | `hotfix/v{VERSION}` |

## Release Process

### Standard Release Workflow

1. **Prepare for Release**
   - Ensure all features are merged to main
   - Run full test suite
   - Review and update documentation

2. **Check Current Status**
   ```bash
   ambros plugin run ambros-release status
   ```

3. **Bump Version**
   ```bash
   # For new features (backwards compatible)
   ambros plugin run ambros-release bump minor
   
   # For bug fixes only
   ambros plugin run ambros-release bump patch
   
   # For breaking changes
   ambros plugin run ambros-release bump major
   ```

4. **Create Release**
   ```bash
   # This creates branch, commits, tags, and pushes
   ambros plugin run ambros-release release
   ```

5. **Generate Changelog (Optional)**
   ```bash
   ambros plugin run ambros-release changelog
   ```

6. **Create GitHub Release**
   - Go to GitHub Releases page
   - Create release from the new tag
   - Add release notes from RELEASE_NOTES.md

### Release Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     RELEASE WORKFLOW                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Check Status          2. Bump Version      3. Create Release│
│  ┌──────────────┐        ┌──────────────┐     ┌──────────────┐  │
│  │ git status   │───────►│ Update       │────►│ Create branch│  │
│  │ Show version │        │ version.txt  │     │ Commit       │  │
│  │ Check remote │        │ Validate     │     │ Tag          │  │
│  └──────────────┘        └──────────────┘     │ Push         │  │
│                                               └──────────────┘  │
│                                                      │          │
│                                                      ▼          │
│                                            ┌──────────────────┐ │
│                                            │  GitHub Release  │ │
│                                            │  (Manual Step)   │ │
│                                            └──────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Using the Release Plugin

The `ambros-release` plugin automates most of the release workflow.

### Installation

```bash
# The plugin is included in the plugins directory
# No additional installation required
```

### Available Commands

| Command | Description | Example |
|---------|-------------|---------|
| `version` | Show current version | `ambros plugin run ambros-release version` |
| `bump` | Increment version | `ambros plugin run ambros-release bump minor` |
| `release` | Full release workflow | `ambros plugin run ambros-release release` |
| `create-branch` | Create release branch only | `ambros plugin run ambros-release create-branch` |
| `status` | Show release status | `ambros plugin run ambros-release status` |
| `changelog` | Generate changelog | `ambros plugin run ambros-release changelog` |

### Environment Variables

The plugin uses these environment variables when run through Ambros:

- `AMBROS_PLUGIN_NAME` - Plugin name
- `AMBROS_PLUGIN_DIR` - Plugin directory path
- `AMBROS_PLUGIN_COMMAND` - Command being executed
- `AMBROS_WORKING_DIR` - Working directory

### Dry Run Mode

Test releases without making changes:

```bash
# Preview what would happen
./plugins/ambros-release/ambros-release.sh release --dry-run
```

## Manual Release Steps

If you prefer to release manually or need more control:

### 1. Check Prerequisites

```bash
# Ensure you're on main branch
git checkout main

# Pull latest changes
git pull origin main

# Run tests
make test

# Check for uncommitted changes
git status
```

### 2. Update Version

```bash
# Edit version file
echo "v3.3.0" > cmd/commands/version.txt

# Verify
cat cmd/commands/version.txt
```

### 3. Commit Version Change

```bash
git add cmd/commands/version.txt
git commit -m "chore: bump version to v3.3.0"
```

### 4. Create Release Branch

```bash
git checkout -b release/v3.3.0
git push -u origin release/v3.3.0
```

### 5. Create Tag

```bash
# Create annotated tag
git tag -a v3.3.0 -m "Release v3.3.0"

# Push tag
git push origin v3.3.0
```

### 6. Update RELEASE_NOTES.md

Document the changes in the release:

```markdown
# Release Notes v3.3.0

## New Features
- Feature 1 description
- Feature 2 description

## Bug Fixes
- Bug fix 1 description

## Breaking Changes
- None
```

## Hotfix Releases

For critical production fixes:

### Hotfix Process

1. **Create Hotfix Branch from Tag**
   ```bash
   git checkout v3.2.0
   git checkout -b hotfix/v3.2.1
   ```

2. **Apply Fix**
   ```bash
   # Make necessary changes
   git add .
   git commit -m "fix: critical bug description"
   ```

3. **Bump Patch Version**
   ```bash
   echo "v3.2.1" > cmd/commands/version.txt
   git add cmd/commands/version.txt
   git commit -m "chore: bump version to v3.2.1"
   ```

4. **Tag and Push**
   ```bash
   git tag -a v3.2.1 -m "Hotfix v3.2.1"
   git push origin hotfix/v3.2.1
   git push origin v3.2.1
   ```

5. **Merge Back to Main**
   ```bash
   git checkout main
   git merge hotfix/v3.2.1
   git push origin main
   ```

## Best Practices

### Before Releasing

- [ ] All tests passing (`make test`)
- [ ] No uncommitted changes
- [ ] Documentation updated
- [ ] RELEASE_NOTES.md updated
- [ ] Version number follows semver

### Version Guidelines

| Change Type | Version Bump | Example |
|-------------|--------------|---------|
| Bug fixes, patches | PATCH | v3.2.0 → v3.2.1 |
| New features (backwards compatible) | MINOR | v3.2.0 → v3.3.0 |
| Breaking changes | MAJOR | v3.2.0 → v4.0.0 |

### Commit Message Format

Follow conventional commits:

```
<type>: <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `chore`: Maintenance tasks
- `refactor`: Code refactoring
- `test`: Adding tests
- `perf`: Performance improvements

### Tag Naming

- Always prefix with `v`: `v3.2.0`
- Use annotated tags: `git tag -a v3.2.0 -m "message"`
- Never delete or modify published tags

## Automation Integration

### GitHub Actions (Optional)

You can integrate the release plugin with GitHub Actions:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build
        run: make build
      
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: bin/*
```

### Makefile Integration

The project Makefile includes release targets:

```bash
# Build release binary
make build

# Run tests before release
make test

# Full release (if implemented)
make release VERSION=v3.3.0
```

## Troubleshooting

### Common Issues

**"Not on main branch"**
```bash
git checkout main
git pull origin main
```

**"Uncommitted changes"**
```bash
git stash  # Save changes temporarily
# ... do release ...
git stash pop  # Restore changes
```

**"Tag already exists"**
```bash
# Delete local tag
git tag -d v3.2.0

# Delete remote tag (use with caution!)
git push origin :refs/tags/v3.2.0
```

**"Push rejected"**
```bash
git pull --rebase origin main
git push origin main
```

## Related Documentation

- [CONTRIBUTING.md](../CONTRIBUTING.md) - How to contribute
- [README.md](../README.md) - Project overview
- [RELEASE_NOTES.md](../RELEASE_NOTES.md) - Current release notes
- [plugins/ambros-release/README.md](../plugins/ambros-release/README.md) - Plugin documentation
