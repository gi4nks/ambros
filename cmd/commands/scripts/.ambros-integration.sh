#!/bin/bash

# Ambros transparent integration
# Source this file in your ~/.bashrc or ~/.zshrc

# Opt-out: if AMBROS_INTEGRATION=off, do nothing (useful for CI or to disable)
if [[ "${AMBROS_INTEGRATION:-}" = "off" ]]; then
    # If the script is being sourced, return; if executed, exit.
    (return 0 2>/dev/null) || exit 0
fi

# Detect non-interactive shells / CI or git hooks and opt out automatically.
# If stdin is not a terminal or common hook variables are set, skip integration.
if [[ ! -t 0 ]] || [[ -n "${GIT_PARAMS:-}${GIT_DIR:-}${CI:-}" ]]; then
    (return 0 2>/dev/null) || exit 0
fi

# Store original command executables (use command -v for portability)
ORIGINAL_LS=$(command -v ls || true)
ORIGINAL_GIT=$(command -v git || true)
ORIGINAL_CURL=$(command -v curl || true)
ORIGINAL_NPM=$(command -v npm || true)
ORIGINAL_DOCKER=$(command -v docker || true)

# List of commands to intercept (add more as needed)
AMBROS_INTERCEPTED_COMMANDS=(
    "ls" "git" "curl" "wget" "npm" "yarn" "docker" 
    "ssh" "scp" "rsync" "tar" "zip" "unzip" "make" "cargo"
# Add common wrapper and build tool wrappers
    "go" "python" "node" "java" "mvn" "gradle" "gradlew"
)
# Detect whether installed 'ambros' supports the --auto flag.
# We'll call 'ambros run --help' and look for the string '--auto'.
AMBROS_SUPPORTS_AUTO=false
if command -v ambros >/dev/null 2>&1; then
    if ambros run --help 2>&1 | grep -q -- "--auto"; then
        AMBROS_SUPPORTS_AUTO=true
    fi
fi

# Function to execute command through Ambros
ambros_exec() {
    local cmd="$1"
    shift

    # Check if Ambros should track this command
    if [[ " ${AMBROS_INTERCEPTED_COMMANDS[@]} " =~ " ${cmd} " ]]; then
        # Build the ambros invocation with supported flags only
        if [ "$AMBROS_SUPPORTS_AUTO" = true ]; then
            ambros run --store --auto -- "$cmd" "$@"
        else
            # Fallback: use --store only (supported in older versions)
            ambros run --store -- "$cmd" "$@"
        fi
    else
        # Execute normally for non-tracked commands.
        # Prefer the saved absolute path (ORIGINAL_...) if available to avoid
        # recursive trapping when the user has the integration enabled.
        case "$cmd" in
            ls)
                if [[ -n "$ORIGINAL_LS" ]]; then
                    "$ORIGINAL_LS" "$@"; return
                fi
                ;;
            git)
                if [[ -n "$ORIGINAL_GIT" ]]; then
                    "$ORIGINAL_GIT" "$@"; return
                fi
                ;;
            curl)
                if [[ -n "$ORIGINAL_CURL" ]]; then
                    "$ORIGINAL_CURL" "$@"; return
                fi
                ;;
            npm)
                if [[ -n "$ORIGINAL_NPM" ]]; then
                    "$ORIGINAL_NPM" "$@"; return
                fi
                ;;
            docker)
                if [[ -n "$ORIGINAL_DOCKER" ]]; then
                    "$ORIGINAL_DOCKER" "$@"; return
                fi
                ;;
        esac
        # Fallback to command which will execute the external binary.
        command "$cmd" "$@"
    fi
}

# Create transparent aliases for intercepted commands
for cmd in "${AMBROS_INTERCEPTED_COMMANDS[@]}"; do
    eval "function $cmd() { ambros_exec '$cmd' \"\$@\"; }"
done

# Optional: Add completion support
if [[ -n "$BASH_VERSION" ]]; then
    # Bash completion
    for cmd in "${AMBROS_INTERCEPTED_COMMANDS[@]}"; do
        complete -F _command "$cmd" 2>/dev/null || true
    done
elif [[ -n "$ZSH_VERSION" ]]; then
    # Zsh completion
    autoload -U compinit && compinit
fi

echo "✅ Ambros transparent mode activated"
echo "📊 Tracking commands: ${AMBROS_INTERCEPTED_COMMANDS[*]}"

# To enable, add the following line to your ~/.bashrc or ~/.zshrc:
# source ~/path/to/.ambros-integration.sh
# Make it executable
#chmod +x ~/.ambros-integration.sh
