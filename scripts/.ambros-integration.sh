#!/bin/bash

# Ambros transparent integration
# Source this file in your ~/.bashrc or ~/.zshrc

# Store original command executables
ORIGINAL_LS=$(which ls)
ORIGINAL_GIT=$(which git)
ORIGINAL_CURL=$(which curl)
ORIGINAL_NPM=$(which npm)
ORIGINAL_DOCKER=$(which docker)

# List of commands to intercept (add more as needed)
AMBROS_INTERCEPTED_COMMANDS=(
    "ls" "git" "curl" "wget" "npm" "yarn" "docker" 
    "ssh" "scp" "rsync" "tar" "zip" "unzip" "make" "cargo"
    "go" "python" "node" "java" "mvn" "gradle"
)

// ...existing code...
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
        # Execute normally for non-tracked commands
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