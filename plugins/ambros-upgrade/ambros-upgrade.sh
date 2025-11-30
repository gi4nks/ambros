#!/bin/bash
# Ambros Upgrade Plugin
# Update and upgrade Ambros to the latest or a specific version
#
# Usage: ambros ambros-upgrade <command> [args]
#
# Commands:
#   check              Check for available updates
#   current            Show the currently installed version
#   list               List all available versions
#   upgrade            Upgrade to the latest version
#   install <version>  Install a specific version
#   rollback           Rollback to the previous version

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Configuration (can be overridden via environment or plugin config)
GITHUB_REPO="${GITHUB_REPO:-gi4nks/ambros}"
INSTALL_PATH="${INSTALL_PATH:-/usr/local/bin/ambros}"
BACKUP_DIR="${BACKUP_DIR:-$HOME/.ambros/backups}"
GITHUB_API="https://api.github.com/repos/${GITHUB_REPO}"
GITHUB_RELEASES="${GITHUB_API}/releases"

# Parse plugin config if available
if [ -n "$AMBROS_PLUGIN_CONFIG" ]; then
    if command -v jq &> /dev/null; then
        GITHUB_REPO=$(echo "$AMBROS_PLUGIN_CONFIG" | jq -r '.github_repo // "gi4nks/ambros"')
        INSTALL_PATH=$(echo "$AMBROS_PLUGIN_CONFIG" | jq -r '.install_path // "/usr/local/bin/ambros"')
        BACKUP_DIR=$(echo "$AMBROS_PLUGIN_CONFIG" | jq -r '.backup_dir // "~/.ambros/backups"')
        # Expand ~ in paths
        BACKUP_DIR="${BACKUP_DIR/#\~/$HOME}"
    fi
fi

# Ensure backup directory exists
mkdir -p "$BACKUP_DIR"

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

# Detect OS and architecture
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case "$os" in
        darwin) os="darwin" ;;
        linux) os="linux" ;;
        mingw*|msys*|cygwin*) os="windows" ;;
        *)
            log_error "Unsupported operating system: $os"
            exit 1
            ;;
    esac
    
    case "$arch" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        i386|i686) arch="386" ;;
        *)
            log_error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
    
    echo "${os}_${arch}"
}

# Get current installed version
get_current_version() {
    if [ -x "$INSTALL_PATH" ]; then
        "$INSTALL_PATH" version 2>/dev/null | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown"
    elif command -v ambros &> /dev/null; then
        ambros version 2>/dev/null | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown"
    else
        echo "not installed"
    fi
}

# Get latest release version from GitHub
get_latest_version() {
    local response
    response=$(curl -s "${GITHUB_RELEASES}/latest" 2>/dev/null)
    
    if [ $? -ne 0 ]; then
        log_error "Failed to fetch latest release information"
        exit 1
    fi
    
    echo "$response" | jq -r '.tag_name // empty'
}

# List available releases
list_releases() {
    local limit="${1:-10}"
    
    curl -s "${GITHUB_RELEASES}?per_page=${limit}" 2>/dev/null | \
        jq -r '.[] | "\(.tag_name)\t\(.published_at | split("T")[0])\t\(.prerelease | if . then "pre-release" else "stable" end)"'
}

# Compare versions (returns 0 if v1 > v2, 1 if v1 < v2, 2 if equal)
compare_versions() {
    local v1="${1#v}"
    local v2="${2#v}"
    
    if [ "$v1" = "$v2" ]; then
        return 2
    fi
    
    local IFS='.'
    local i
    local -a ver1=($v1)
    local -a ver2=($v2)
    
    for ((i=0; i<3; i++)); do
        local num1=${ver1[i]:-0}
        local num2=${ver2[i]:-0}
        
        if ((num1 > num2)); then
            return 0
        elif ((num1 < num2)); then
            return 1
        fi
    done
    
    return 2
}

# Download release asset
download_release() {
    local version="$1"
    local platform=$(detect_platform)
    local tmp_dir=$(mktemp -d)
    local asset_name="ambros_${version#v}_${platform}.tar.gz"
    local download_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${asset_name}"
    
    log_info "Downloading ${asset_name}..."
    
    if ! curl -sL -o "${tmp_dir}/${asset_name}" "$download_url" 2>/dev/null; then
        # Try alternative naming convention
        asset_name="ambros-${version#v}-${platform}.tar.gz"
        download_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${asset_name}"
        
        if ! curl -sL -o "${tmp_dir}/${asset_name}" "$download_url" 2>/dev/null; then
            # Try zip format for Windows
            if [[ "$platform" == *"windows"* ]]; then
                asset_name="ambros_${version#v}_${platform}.zip"
                download_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${asset_name}"
                
                if ! curl -sL -o "${tmp_dir}/${asset_name}" "$download_url" 2>/dev/null; then
                    log_error "Failed to download release asset"
                    rm -rf "$tmp_dir"
                    exit 1
                fi
            else
                log_error "Failed to download release asset"
                log_error "Tried: $download_url"
                rm -rf "$tmp_dir"
                exit 1
            fi
        fi
    fi
    
    echo "${tmp_dir}/${asset_name}"
}

# Extract and install binary
install_binary() {
    local archive="$1"
    local tmp_dir=$(dirname "$archive")
    local binary_name="ambros"
    
    log_info "Extracting archive..."
    
    if [[ "$archive" == *.tar.gz ]]; then
        tar -xzf "$archive" -C "$tmp_dir"
    elif [[ "$archive" == *.zip ]]; then
        unzip -q "$archive" -d "$tmp_dir"
        binary_name="ambros.exe"
    else
        log_error "Unknown archive format"
        rm -rf "$tmp_dir"
        exit 1
    fi
    
    # Find the binary
    local binary_path=$(find "$tmp_dir" -name "$binary_name" -type f | head -1)
    
    if [ -z "$binary_path" ]; then
        log_error "Binary not found in archive"
        rm -rf "$tmp_dir"
        exit 1
    fi
    
    # Make it executable
    chmod +x "$binary_path"
    
    # Backup current version if exists
    if [ -f "$INSTALL_PATH" ]; then
        local current_version=$(get_current_version)
        local backup_path="${BACKUP_DIR}/ambros_${current_version}_$(date +%Y%m%d_%H%M%S)"
        log_info "Backing up current version to ${backup_path}..."
        cp "$INSTALL_PATH" "$backup_path"
        log_success "Backup created"
    fi
    
    # Install new binary
    log_info "Installing to ${INSTALL_PATH}..."
    
    if [ -w "$(dirname "$INSTALL_PATH")" ]; then
        mv "$binary_path" "$INSTALL_PATH"
    else
        log_warning "Elevated permissions required for installation"
        sudo mv "$binary_path" "$INSTALL_PATH"
    fi
    
    # Cleanup
    rm -rf "$tmp_dir"
    
    log_success "Installation complete"
}

# Command: current
cmd_current() {
    local version=$(get_current_version)
    
    echo ""
    echo -e "${CYAN}Ambros Version Information${NC}"
    echo ""
    echo -e "  ${BLUE}Installed version:${NC} $version"
    echo -e "  ${BLUE}Install path:${NC}      $INSTALL_PATH"
    echo -e "  ${BLUE}Backup directory:${NC}  $BACKUP_DIR"
    echo ""
    
    # Show binary info if available
    if [ -x "$INSTALL_PATH" ]; then
        local binary_size=$(ls -lh "$INSTALL_PATH" | awk '{print $5}')
        local binary_date=$(ls -l "$INSTALL_PATH" | awk '{print $6, $7, $8}')
        echo -e "  ${BLUE}Binary size:${NC}       $binary_size"
        echo -e "  ${BLUE}Modified:${NC}          $binary_date"
    fi
    echo ""
}

# Command: check
cmd_check() {
    log_info "Checking for updates..."
    
    local current=$(get_current_version)
    local latest=$(get_latest_version)
    
    if [ -z "$latest" ]; then
        log_error "Could not determine latest version"
        exit 1
    fi
    
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo -e "${CYAN}        Ambros Update Check${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo ""
    echo -e "  ${BLUE}Current version:${NC} $current"
    echo -e "  ${BLUE}Latest version:${NC}  $latest"
    echo ""
    
    compare_versions "$latest" "$current"
    local cmp_result=$?
    
    if [ $cmp_result -eq 0 ]; then
        echo -e "  ${GREEN}✓${NC} ${BOLD}Update available!${NC}"
        echo ""
        echo "  Run 'ambros ambros-upgrade upgrade' to update"
        echo "  Or 'ambros ambros-upgrade install $latest' to install specific version"
    elif [ $cmp_result -eq 2 ]; then
        echo -e "  ${GREEN}✓${NC} You are running the latest version"
    else
        echo -e "  ${YELLOW}⚠${NC} You are running a newer version than the latest release"
        echo "    (possibly a development or pre-release version)"
    fi
    echo ""
}

# Command: list
cmd_list() {
    local limit="${1:-10}"
    
    # Parse --limit flag
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --limit|-l)
                limit="$2"
                shift 2
                ;;
            *)
                if [[ "$1" =~ ^[0-9]+$ ]]; then
                    limit="$1"
                fi
                shift
                ;;
        esac
    done
    
    log_info "Fetching available releases..."
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo -e "${CYAN}        Available Ambros Versions${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo ""
    
    local current=$(get_current_version)
    
    printf "  ${BOLD}%-12s %-12s %-12s${NC}\n" "VERSION" "DATE" "TYPE"
    echo "  ─────────────────────────────────────────"
    
    while IFS=$'\t' read -r version date type; do
        if [ "$version" = "$current" ]; then
            printf "  ${GREEN}%-12s${NC} %-12s %-12s ${GREEN}(installed)${NC}\n" "$version" "$date" "$type"
        else
            printf "  %-12s %-12s %-12s\n" "$version" "$date" "$type"
        fi
    done < <(list_releases "$limit")
    
    echo ""
    echo "  Showing $limit most recent releases"
    echo "  Use --limit N to show more"
    echo ""
}

# Command: upgrade
cmd_upgrade() {
    local force=false
    
    # Parse --force flag
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --force|-f)
                force=true
                shift
                ;;
            *)
                shift
                ;;
        esac
    done
    
    local current=$(get_current_version)
    local latest=$(get_latest_version)
    
    if [ -z "$latest" ]; then
        log_error "Could not determine latest version"
        exit 1
    fi
    
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo -e "${CYAN}        Ambros Upgrade${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo ""
    echo -e "  ${BLUE}Current version:${NC} $current"
    echo -e "  ${BLUE}Target version:${NC}  $latest"
    echo -e "  ${BLUE}Platform:${NC}        $(detect_platform)"
    echo ""
    
    compare_versions "$latest" "$current"
    local cmp_result=$?
    
    if [ $cmp_result -eq 2 ]; then
        if [ "$force" = false ]; then
            log_warning "You are already running the latest version ($current)"
            echo "  Use --force to reinstall"
            exit 0
        fi
        log_info "Reinstalling current version (--force)"
    elif [ $cmp_result -eq 1 ]; then
        if [ "$force" = false ]; then
            log_warning "You are running a newer version than latest release"
            echo "  Use --force to downgrade"
            exit 0
        fi
        log_warning "Downgrading to $latest (--force)"
    fi
    
    # Confirm
    read -p "Proceed with upgrade? [y/N] " confirm
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        log_warning "Upgrade cancelled"
        exit 0
    fi
    
    echo ""
    
    # Download and install
    local archive=$(download_release "$latest")
    log_success "Downloaded successfully"
    
    install_binary "$archive"
    
    # Verify installation
    local new_version=$(get_current_version)
    
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════${NC}"
    echo -e "${GREEN}        Upgrade Complete!${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════${NC}"
    echo ""
    echo -e "  ${BLUE}Previous version:${NC} $current"
    echo -e "  ${BLUE}New version:${NC}      $new_version"
    echo ""
}

# Command: install
cmd_install() {
    local version="$1"
    
    if [ -z "$version" ]; then
        log_error "Usage: ambros ambros-upgrade install <version>"
        echo "  Example: ambros ambros-upgrade install v3.2.0"
        exit 1
    fi
    
    # Ensure version starts with 'v'
    if [[ ! "$version" =~ ^v ]]; then
        version="v${version}"
    fi
    
    local current=$(get_current_version)
    
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo -e "${CYAN}        Ambros Install${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo ""
    echo -e "  ${BLUE}Current version:${NC} $current"
    echo -e "  ${BLUE}Target version:${NC}  $version"
    echo -e "  ${BLUE}Platform:${NC}        $(detect_platform)"
    echo ""
    
    # Confirm
    read -p "Proceed with installation? [y/N] " confirm
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        log_warning "Installation cancelled"
        exit 0
    fi
    
    echo ""
    
    # Download and install
    local archive=$(download_release "$version")
    log_success "Downloaded successfully"
    
    install_binary "$archive"
    
    # Verify installation
    local new_version=$(get_current_version)
    
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════${NC}"
    echo -e "${GREEN}        Installation Complete!${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════${NC}"
    echo ""
    echo -e "  ${BLUE}Previous version:${NC} $current"
    echo -e "  ${BLUE}Installed version:${NC} $new_version"
    echo ""
}

# Command: rollback
cmd_rollback() {
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo -e "${CYAN}        Ambros Rollback${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo ""
    
    # List available backups
    local backups=($(ls -t "$BACKUP_DIR"/ambros_* 2>/dev/null))
    
    if [ ${#backups[@]} -eq 0 ]; then
        log_error "No backup versions found in $BACKUP_DIR"
        exit 1
    fi
    
    local current=$(get_current_version)
    echo -e "  ${BLUE}Current version:${NC} $current"
    echo ""
    echo -e "  ${BLUE}Available backups:${NC}"
    
    local i=1
    for backup in "${backups[@]}"; do
        local backup_name=$(basename "$backup")
        local backup_version=$(echo "$backup_name" | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' || echo "unknown")
        local backup_date=$(echo "$backup_name" | grep -oE '[0-9]{8}_[0-9]{6}' || echo "")
        
        if [ -n "$backup_date" ]; then
            # Format date nicely
            local formatted_date="${backup_date:0:4}-${backup_date:4:2}-${backup_date:6:2} ${backup_date:9:2}:${backup_date:11:2}"
            printf "    ${YELLOW}%d)${NC} %s (backed up: %s)\n" "$i" "$backup_version" "$formatted_date"
        else
            printf "    ${YELLOW}%d)${NC} %s\n" "$i" "$backup_name"
        fi
        ((i++))
    done
    
    echo ""
    read -p "Select backup to restore (1-${#backups[@]}) or 'q' to quit: " selection
    
    if [[ "$selection" = "q" || "$selection" = "Q" ]]; then
        log_warning "Rollback cancelled"
        exit 0
    fi
    
    if ! [[ "$selection" =~ ^[0-9]+$ ]] || [ "$selection" -lt 1 ] || [ "$selection" -gt ${#backups[@]} ]; then
        log_error "Invalid selection"
        exit 1
    fi
    
    local selected_backup="${backups[$((selection-1))]}"
    local selected_version=$(basename "$selected_backup" | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' || echo "backup")
    
    echo ""
    log_info "Restoring $selected_version..."
    
    # Backup current before rollback
    if [ -f "$INSTALL_PATH" ]; then
        local rollback_backup="${BACKUP_DIR}/ambros_${current}_pre_rollback_$(date +%Y%m%d_%H%M%S)"
        cp "$INSTALL_PATH" "$rollback_backup"
        log_success "Created pre-rollback backup"
    fi
    
    # Restore selected backup
    if [ -w "$(dirname "$INSTALL_PATH")" ]; then
        cp "$selected_backup" "$INSTALL_PATH"
    else
        log_warning "Elevated permissions required"
        sudo cp "$selected_backup" "$INSTALL_PATH"
    fi
    
    chmod +x "$INSTALL_PATH"
    
    local new_version=$(get_current_version)
    
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════${NC}"
    echo -e "${GREEN}        Rollback Complete!${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════${NC}"
    echo ""
    echo -e "  ${BLUE}Previous version:${NC} $current"
    echo -e "  ${BLUE}Restored version:${NC} $new_version"
    echo ""
}

# Show help
cmd_help() {
    echo ""
    echo -e "${CYAN}Ambros Upgrade Plugin${NC}"
    echo ""
    echo "Update and upgrade Ambros to the latest or a specific version."
    echo ""
    echo -e "${YELLOW}Usage:${NC}"
    echo "  ambros ambros-upgrade <command> [arguments]"
    echo ""
    echo -e "${YELLOW}Commands:${NC}"
    echo "  check              Check for available updates"
    echo "  current            Show the currently installed version"
    echo "  list [--limit N]   List available versions (default: 10)"
    echo "  upgrade [--force]  Upgrade to the latest version"
    echo "  install <version>  Install a specific version"
    echo "  rollback           Rollback to a previous backup"
    echo ""
    echo -e "${YELLOW}Examples:${NC}"
    echo "  ambros ambros-upgrade check"
    echo "  ambros ambros-upgrade current"
    echo "  ambros ambros-upgrade list --limit 20"
    echo "  ambros ambros-upgrade upgrade"
    echo "  ambros ambros-upgrade upgrade --force"
    echo "  ambros ambros-upgrade install v3.1.0"
    echo "  ambros ambros-upgrade rollback"
    echo ""
    echo -e "${YELLOW}Configuration:${NC}"
    echo "  The plugin reads from plugin.json config:"
    echo "  - github_repo:  GitHub repository (default: gi4nks/ambros)"
    echo "  - install_path: Where to install binary (default: /usr/local/bin/ambros)"
    echo "  - backup_dir:   Where to store backups (default: ~/.ambros/backups)"
    echo ""
    echo -e "${YELLOW}Notes:${NC}"
    echo "  - Backups are automatically created before each upgrade"
    echo "  - Use 'rollback' to restore any previous version"
    echo "  - Requires: curl, jq, tar"
    echo ""
}

# Main entry point
main() {
    local command="${1:-help}"
    shift || true
    
    case "$command" in
        check)
            cmd_check
            ;;
        current)
            cmd_current
            ;;
        list)
            cmd_list "$@"
            ;;
        upgrade)
            cmd_upgrade "$@"
            ;;
        install)
            cmd_install "$@"
            ;;
        rollback)
            cmd_rollback
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
