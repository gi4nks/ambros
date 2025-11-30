#!/bin/bash
# Ambros Release Plugin
# Automates the release workflow: version bumping, tagging, and pushing
#
# Usage: ambros ambros-release <command> [args]
#
# Commands:
#   version              Show the current version
#   bump <type>          Bump version (major|minor|patch)
#   release <type>       Full release workflow
#   tag [--message msg]  Create and push a tag
#   push                 Push branch and tags to origin
#   status               Show release status

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration (can be overridden via AMBROS_PLUGIN_CONFIG)
VERSION_FILE="${VERSION_FILE:-cmd/commands/version.txt}"
REMOTE="${REMOTE:-origin}"
BRANCH_PREFIX="${BRANCH_PREFIX:-release/}"

# Parse plugin config if available
if [ -n "$AMBROS_PLUGIN_CONFIG" ]; then
    # Extract config values from JSON (requires jq or simple parsing)
    if command -v jq &> /dev/null; then
        VERSION_FILE=$(echo "$AMBROS_PLUGIN_CONFIG" | jq -r '.version_file // "cmd/commands/version.txt"')
        REMOTE=$(echo "$AMBROS_PLUGIN_CONFIG" | jq -r '.remote // "origin"')
        BRANCH_PREFIX=$(echo "$AMBROS_PLUGIN_CONFIG" | jq -r '.branch_prefix // "release/"')
    fi
fi

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1" >&2
}

# Get the repository root
get_repo_root() {
    git rev-parse --show-toplevel 2>/dev/null || {
        log_error "Not in a git repository"
        exit 1
    }
}

# Read current version from version file
get_current_version() {
    local repo_root=$(get_repo_root)
    local version_path="$repo_root/$VERSION_FILE"
    
    if [ ! -f "$version_path" ]; then
        log_error "Version file not found: $version_path"
        exit 1
    fi
    
    cat "$version_path" | tr -d '[:space:]'
}

# Parse version components (expects format vX.Y.Z)
parse_version() {
    local version="$1"
    # Remove 'v' prefix if present
    version="${version#v}"
    
    local major=$(echo "$version" | cut -d. -f1)
    local minor=$(echo "$version" | cut -d. -f2)
    local patch=$(echo "$version" | cut -d. -f3)
    
    echo "$major $minor $patch"
}

# Calculate new version based on bump type
calculate_new_version() {
    local current="$1"
    local bump_type="$2"
    
    local parts=($(parse_version "$current"))
    local major="${parts[0]}"
    local minor="${parts[1]}"
    local patch="${parts[2]}"
    
    case "$bump_type" in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
        *)
            log_error "Invalid bump type: $bump_type (use: major, minor, patch)"
            exit 1
            ;;
    esac
    
    echo "v${major}.${minor}.${patch}"
}

# Update version file
update_version_file() {
    local new_version="$1"
    local repo_root=$(get_repo_root)
    local version_path="$repo_root/$VERSION_FILE"
    
    echo "$new_version" > "$version_path"
    log_success "Updated $VERSION_FILE to $new_version"
}

# Check for uncommitted changes
check_clean_working_tree() {
    if ! git diff-index --quiet HEAD -- 2>/dev/null; then
        return 1
    fi
    return 0
}

# Check if tag exists
tag_exists() {
    local tag="$1"
    git rev-parse "$tag" &>/dev/null
}

# Command: version
cmd_version() {
    local current=$(get_current_version)
    echo -e "${CYAN}Current version:${NC} $current"
}

# Command: status
cmd_status() {
    local repo_root=$(get_repo_root)
    local current=$(get_current_version)
    local branch=$(git branch --show-current)
    local remote_url=$(git remote get-url "$REMOTE" 2>/dev/null || echo "Not configured")
    
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo -e "${CYAN}        Ambros Release Status${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo ""
    echo -e "  ${BLUE}Version:${NC}     $current"
    echo -e "  ${BLUE}Branch:${NC}      $branch"
    echo -e "  ${BLUE}Remote:${NC}      $REMOTE ($remote_url)"
    echo -e "  ${BLUE}Version file:${NC} $VERSION_FILE"
    echo ""
    
    # Check working tree status
    if check_clean_working_tree; then
        echo -e "  ${GREEN}✓${NC} Working tree is clean"
    else
        echo -e "  ${YELLOW}⚠${NC} Uncommitted changes detected"
        echo ""
        echo -e "  ${YELLOW}Changed files:${NC}"
        git status --short | head -10 | sed 's/^/    /'
        local total=$(git status --short | wc -l | tr -d ' ')
        if [ "$total" -gt 10 ]; then
            echo "    ... and $((total - 10)) more files"
        fi
    fi
    
    # Check if current version tag exists
    if tag_exists "$current"; then
        echo -e "  ${GREEN}✓${NC} Tag $current exists"
    else
        echo -e "  ${YELLOW}⚠${NC} Tag $current does not exist"
    fi
    
    # Show recent tags
    echo ""
    echo -e "  ${BLUE}Recent tags:${NC}"
    git tag -l --sort=-version:refname | head -5 | sed 's/^/    /'
    
    echo ""
}

# Command: bump
cmd_bump() {
    local bump_type="$1"
    
    if [ -z "$bump_type" ]; then
        log_error "Usage: ambros ambros-release bump <major|minor|patch>"
        exit 1
    fi
    
    local current=$(get_current_version)
    local new_version=$(calculate_new_version "$current" "$bump_type")
    
    log_info "Bumping version: $current → $new_version"
    update_version_file "$new_version"
    
    echo ""
    echo -e "${GREEN}Version bumped to $new_version${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Commit the version change:"
    echo "     git add $VERSION_FILE && git commit -m \"chore: Bump version to $new_version\""
    echo "  2. Create a tag:"
    echo "     ambros ambros-release tag"
    echo "  3. Or run the full release:"
    echo "     ambros ambros-release release $bump_type"
}

# Command: tag
cmd_tag() {
    local message="$1"
    local current=$(get_current_version)
    
    if tag_exists "$current"; then
        log_error "Tag $current already exists"
        exit 1
    fi
    
    if [ -z "$message" ]; then
        message="Release $current"
    fi
    
    log_info "Creating tag $current..."
    git tag -a "$current" -m "$message"
    log_success "Created tag $current"
    
    log_info "Pushing tag to $REMOTE..."
    git push "$REMOTE" "$current"
    log_success "Pushed tag $current to $REMOTE"
}

# Command: push
cmd_push() {
    local branch=$(git branch --show-current)
    
    log_info "Pushing branch $branch to $REMOTE..."
    git push "$REMOTE" "$branch"
    log_success "Pushed branch $branch"
    
    log_info "Pushing tags to $REMOTE..."
    git push "$REMOTE" --tags
    log_success "Pushed all tags"
}

# Command: create-branch
cmd_create_branch() {
    local current=$(get_current_version)
    local release_branch="${BRANCH_PREFIX}${current}"
    
    if git show-ref --verify --quiet "refs/heads/$release_branch"; then
        log_error "Branch $release_branch already exists locally"
        exit 1
    fi
    
    log_info "Creating release branch $release_branch..."
    git checkout -b "$release_branch"
    log_success "Created and switched to branch $release_branch"
    
    read -p "Push branch to $REMOTE? [y/N] " confirm
    if [[ "$confirm" =~ ^[Yy]$ ]]; then
        git push -u "$REMOTE" "$release_branch"
        log_success "Pushed branch $release_branch to $REMOTE"
    fi
}

# Command: changelog
cmd_changelog() {
    local from_tag="$1"
    local to_ref="${2:-HEAD}"
    
    # If no from_tag specified, find the previous tag
    if [ -z "$from_tag" ]; then
        # Get the second most recent tag (skip the current one)
        from_tag=$(git tag -l --sort=-version:refname | head -2 | tail -1)
        if [ -z "$from_tag" ]; then
            log_error "No previous tag found to generate changelog from"
            exit 1
        fi
    fi
    
    local current=$(get_current_version)
    
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo -e "${CYAN}        Changelog: $from_tag → $current${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo ""
    
    # Group commits by type
    echo -e "${YELLOW}## Features${NC}"
    git log "$from_tag..$to_ref" --pretty=format:"- %s (%h)" --grep="^feat" 2>/dev/null || echo "  No features"
    echo ""
    
    echo -e "${YELLOW}## Bug Fixes${NC}"
    git log "$from_tag..$to_ref" --pretty=format:"- %s (%h)" --grep="^fix" 2>/dev/null || echo "  No bug fixes"
    echo ""
    
    echo -e "${YELLOW}## Documentation${NC}"
    git log "$from_tag..$to_ref" --pretty=format:"- %s (%h)" --grep="^docs" 2>/dev/null || echo "  No documentation changes"
    echo ""
    
    echo -e "${YELLOW}## Other Changes${NC}"
    git log "$from_tag..$to_ref" --pretty=format:"- %s (%h)" --grep="^chore\|^refactor\|^test\|^perf" 2>/dev/null || echo "  No other changes"
    echo ""
    
    echo -e "${BLUE}All commits:${NC}"
    git log "$from_tag..$to_ref" --pretty=format:"  - %s (%h) by %an" 2>/dev/null
    echo ""
    
    # Statistics
    echo ""
    echo -e "${BLUE}Statistics:${NC}"
    local commit_count=$(git rev-list --count "$from_tag..$to_ref" 2>/dev/null || echo "0")
    local contributors=$(git log "$from_tag..$to_ref" --pretty=format:"%an" 2>/dev/null | sort -u | wc -l | tr -d ' ')
    local files_changed=$(git diff --stat "$from_tag..$to_ref" 2>/dev/null | tail -1 || echo "0 files changed")
    
    echo "  Commits: $commit_count"
    echo "  Contributors: $contributors"
    echo "  $files_changed"
    echo ""
}

# Command: release (full workflow)
cmd_release() {
    local bump_type="$1"
    shift
    local message=""
    
    # Parse optional --message argument
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --message|-m)
                message="$2"
                shift 2
                ;;
            *)
                shift
                ;;
        esac
    done
    
    if [ -z "$bump_type" ]; then
        log_error "Usage: ambros ambros-release release <major|minor|patch> [--message \"msg\"]"
        exit 1
    fi
    
    local current=$(get_current_version)
    local new_version=$(calculate_new_version "$current" "$bump_type")
    local release_branch="${BRANCH_PREFIX}${new_version}"
    
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo -e "${CYAN}        Ambros Release Workflow${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo ""
    echo -e "  ${BLUE}Current version:${NC} $current"
    echo -e "  ${BLUE}New version:${NC}     $new_version"
    echo -e "  ${BLUE}Release branch:${NC} $release_branch"
    echo -e "  ${BLUE}Remote:${NC}          $REMOTE"
    echo ""
    
    # Confirm
    read -p "Proceed with release? [y/N] " confirm
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        log_warning "Release cancelled"
        exit 0
    fi
    
    echo ""
    
    # Step 1: Check for uncommitted changes (warn but allow)
    log_info "Step 1/6: Checking working tree..."
    if ! check_clean_working_tree; then
        log_warning "Uncommitted changes detected. They will be included in the release commit."
        read -p "Continue? [y/N] " confirm
        if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
            log_warning "Release cancelled"
            exit 0
        fi
    else
        log_success "Working tree is clean"
    fi
    
    # Step 2: Update version file
    log_info "Step 2/6: Updating version to $new_version..."
    update_version_file "$new_version"
    
    # Step 3: Stage and commit
    log_info "Step 3/6: Committing changes..."
    git add -A
    
    local commit_message="chore: Release $new_version"
    if [ -n "$message" ]; then
        commit_message="$commit_message

$message"
    fi
    
    git commit -m "$commit_message"
    log_success "Committed release changes"
    
    # Step 4: Create release branch
    log_info "Step 4/6: Creating release branch $release_branch..."
    git checkout -b "$release_branch"
    log_success "Created and switched to branch $release_branch"
    
    # Step 5: Create tag
    log_info "Step 5/6: Creating tag $new_version..."
    local tag_message="Release $new_version"
    if [ -n "$message" ]; then
        tag_message="$tag_message

$message"
    fi
    git tag -a "$new_version" -m "$tag_message"
    log_success "Created tag $new_version"
    
    # Step 6: Push to remote
    log_info "Step 6/6: Pushing to $REMOTE..."
    git push "$REMOTE" "$release_branch"
    log_success "Pushed branch $release_branch"
    
    git push "$REMOTE" "$new_version"
    log_success "Pushed tag $new_version"
    
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════${NC}"
    echo -e "${GREEN}        Release $new_version Complete!${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════${NC}"
    echo ""
    echo "  Release branch: $release_branch"
    echo "  Tag: $new_version"
    echo ""
    echo "  GitHub Release URL:"
    echo "  https://github.com/gi4nks/ambros/releases/tag/$new_version"
    echo ""
}

# Show help
cmd_help() {
    echo ""
    echo -e "${CYAN}Ambros Release Plugin${NC}"
    echo ""
    echo "Automates the release workflow for Ambros."
    echo ""
    echo -e "${YELLOW}Usage:${NC}"
    echo "  ambros ambros-release <command> [arguments]"
    echo ""
    echo -e "${YELLOW}Commands:${NC}"
    echo "  version              Show the current version"
    echo "  status               Show release status (version, branch, changes)"
    echo "  bump <type>          Bump version (major|minor|patch)"
    echo "  tag [--message msg]  Create and push a tag for current version"
    echo "  push                 Push current branch and tags to origin"
    echo "  create-branch        Create release branch for current version"
    echo "  changelog [from]     Generate changelog from previous tag (or specified tag)"
    echo "  release <type>       Full release workflow:"
    echo "                         1. Bump version"
    echo "                         2. Commit changes"
    echo "                         3. Create release branch"
    echo "                         4. Create annotated tag"
    echo "                         5. Push branch and tag to origin"
    echo ""
    echo -e "${YELLOW}Examples:${NC}"
    echo "  ambros ambros-release version"
    echo "  ambros ambros-release status"
    echo "  ambros ambros-release bump patch"
    echo "  ambros ambros-release release minor --message \"New features\""
    echo "  ambros ambros-release tag --message \"Hotfix release\""
    echo "  ambros ambros-release changelog v3.1.0"
    echo "  ambros ambros-release create-branch"
    echo ""
    echo -e "${YELLOW}Configuration:${NC}"
    echo "  The plugin reads from plugin.json config:"
    echo "  - version_file:  Path to version file (default: cmd/commands/version.txt)"
    echo "  - remote:        Git remote name (default: origin)"
    echo "  - branch_prefix: Release branch prefix (default: release/)"
    echo ""
}

# Main entry point
main() {
    local command="${1:-help}"
    shift || true
    
    case "$command" in
        version)
            cmd_version
            ;;
        status)
            cmd_status
            ;;
        bump)
            cmd_bump "$@"
            ;;
        tag)
            cmd_tag "$@"
            ;;
        push)
            cmd_push
            ;;
        create-branch)
            cmd_create_branch
            ;;
        changelog)
            cmd_changelog "$@"
            ;;
        release)
            cmd_release "$@"
            ;;
        help|--help|-h)
            cmd_help
            ;;
        *)
            log_error "Unknown command: $command"
            cmd_help
            exit 1
            ;;
    esac
}

main "$@"
