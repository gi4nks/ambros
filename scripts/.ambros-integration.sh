#!/bin/bash

# Ambros transparent integration
# Source this file in your ~/.bashrc or ~/.zshrc

# ============================================================================
# CONFIGURATION
# ============================================================================
# 
# AMBROS_INTEGRATION_MODE: Controls which commands are tracked
#   - "whitelist" (default): Only track commands in AMBROS_INTERCEPTED_COMMANDS
#   - "all": Track ALL commands, except those in AMBROS_BLACKLIST
#
# AMBROS_INTERCEPTED_COMMANDS: Commands to track (whitelist mode only)
#   Default: ls git curl wget npm yarn docker ssh scp rsync tar zip unzip 
#            make cargo go python node java mvn gradle gradlew
#
# AMBROS_BLACKLIST: Commands to NEVER track (used in "all" mode, or to 
#                   exclude from whitelist)
#   Default: cd pwd echo printf test [ true false exit return source .
#            export unset alias unalias type which whereis history fc
#            jobs fg bg kill wait suspend ambros
#
# Set these in your shell profile BEFORE sourcing this script:
#   export AMBROS_INTEGRATION_MODE="all"
#   export AMBROS_BLACKLIST="cd pwd echo ambros vim nano"
# ============================================================================

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

# Integration mode: "whitelist" (default) or "all"
AMBROS_INTEGRATION_MODE="${AMBROS_INTEGRATION_MODE:-whitelist}"

# Default whitelist of commands to intercept (used in whitelist mode)
if [[ -z "${AMBROS_INTERCEPTED_COMMANDS[*]}" ]]; then
    AMBROS_INTERCEPTED_COMMANDS=(
        "ls" "git" "curl" "wget" "npm" "yarn" "docker" 
        "ssh" "scp" "rsync" "tar" "zip" "unzip" "make" "cargo"
        "go" "python" "node" "java" "mvn" "gradle" "gradlew"
        "kubectl" "helm" "terraform" "ansible" "pip" "brew"
    )
fi

# Default blacklist - commands that should NEVER be wrapped
# These are shell builtins, navigation commands, or could cause recursion
if [[ -z "${AMBROS_BLACKLIST[*]}" ]]; then
    AMBROS_BLACKLIST=(
        # Shell builtins and navigation
        "cd" "pwd" "pushd" "popd" "dirs"
        # Output/test builtins  
        "echo" "printf" "test" "[" "[[" "true" "false"
        # Control flow
        "exit" "return" "break" "continue"
        # Sourcing and environment
        "source" "." "export" "unset" "local" "declare" "typeset" "readonly"
        # Aliases and functions
        "alias" "unalias" "function" "builtin" "command"
        # Type inspection
        "type" "which" "whereis" "whence" "hash"
        # History and job control
        "history" "fc" "jobs" "fg" "bg" "kill" "wait" "suspend" "disown"
        # Ambros itself (prevent recursion!)
        "ambros"
        # Common interactive editors (use --auto explicitly for these)
        "vim" "vi" "nvim" "nano" "emacs" "code" "subl"
        # Pagers
        "less" "more" "most" "head" "tail" "cat"
        # Dangerous to wrap
        "sudo" "su" "exec" "eval" "time"
    )
fi

# Convert blacklist array to associative array for O(1) lookup
declare -A AMBROS_BLACKLIST_MAP
for cmd in "${AMBROS_BLACKLIST[@]}"; do
    AMBROS_BLACKLIST_MAP["$cmd"]=1
done

# Check if a command is blacklisted
is_blacklisted() {
    [[ -n "${AMBROS_BLACKLIST_MAP[$1]}" ]]
}

# Detect whether installed 'ambros' supports the --auto flag.
AMBROS_SUPPORTS_AUTO=false
if command -v ambros >/dev/null 2>&1; then
    if command ambros run --help 2>&1 | grep -q -- "--auto"; then
        AMBROS_SUPPORTS_AUTO=true
    fi
fi

# Function to execute command through Ambros.
ambros_exec() {
    local cmd="$1"
    shift

    if [ "$AMBROS_SUPPORTS_AUTO" = true ]; then
        command ambros run --store --auto -- "$cmd" "$@"
    else
        command ambros run --store -- "$cmd" "$@"
    fi
}

# ============================================================================
# MODE: WHITELIST - Only wrap specific commands
# ============================================================================
if [[ "$AMBROS_INTEGRATION_MODE" == "whitelist" ]]; then
    AMBROS_WRAPPED_COMMANDS=()
    for cmd in "${AMBROS_INTERCEPTED_COMMANDS[@]}"; do
        # Skip if blacklisted
        if is_blacklisted "$cmd"; then
            continue
        fi
        # Only wrap if command exists
        if command -v "$cmd" >/dev/null 2>&1; then
            eval "function $cmd() { ambros_exec '$cmd' \"\$@\"; }"
            AMBROS_WRAPPED_COMMANDS+=("$cmd")
        fi
    done
    
    echo "✅ Ambros transparent mode activated (whitelist)"
    echo "📊 Tracking ${#AMBROS_WRAPPED_COMMANDS[@]} commands: ${AMBROS_WRAPPED_COMMANDS[*]}"

# ============================================================================
# MODE: ALL - Wrap all commands via preexec/precmd hooks
# ============================================================================
elif [[ "$AMBROS_INTEGRATION_MODE" == "all" ]]; then
    
    if [[ -n "$ZSH_VERSION" ]]; then
        # ZSH: Use preexec hook to intercept commands BEFORE they run
        autoload -Uz add-zsh-hook
        
        # Override command execution using preexec
        ambros_preexec() {
            local cmd_line="$1"
            local cmd_name="${cmd_line%% *}"
            
            # Skip blacklisted commands
            if is_blacklisted "$cmd_name"; then
                return
            fi
            
            # Skip if command starts with space (privacy feature)
            if [[ "$cmd_line" =~ ^[[:space:]] ]]; then
                return
            fi
            
            # Skip empty commands
            [[ -z "$cmd_line" ]] && return
            
            # Mark that we want to track this command
            export AMBROS_TRACK_CMD="$cmd_line"
        }
        
        ambros_precmd() {
            local exit_code=$?
            
            # If we were tracking a command, log it now (after execution)
            if [[ -n "$AMBROS_TRACK_CMD" ]]; then
                # Log asynchronously to not block the prompt
                (
                    if [ "$AMBROS_SUPPORTS_AUTO" = true ]; then
                        # Use a marker file approach to log post-execution
                        command ambros run --store --tag "shell-integration" --tag "exit:$exit_code" -- echo "[tracked] $AMBROS_TRACK_CMD" >/dev/null 2>&1
                    fi
                ) &
                unset AMBROS_TRACK_CMD
            fi
        }
        
        add-zsh-hook preexec ambros_preexec
        add-zsh-hook precmd ambros_precmd
        
        echo "✅ Ambros transparent mode activated (all commands - zsh hooks)"
        echo "🚫 Blacklisted: ${AMBROS_BLACKLIST[*]:0:10}... (${#AMBROS_BLACKLIST[@]} total)"
        echo "💡 Note: 'all' mode logs commands after execution. For full output capture, use whitelist mode."
        
    elif [[ -n "$BASH_VERSION" ]]; then
        # BASH: Use DEBUG trap and PROMPT_COMMAND
        
        ambros_bash_preexec() {
            local cmd_line="$BASH_COMMAND"
            local cmd_name="${cmd_line%% *}"
            
            # Skip our own hooks and prompt command
            [[ "$cmd_name" == "ambros_bash_"* ]] && return
            [[ "$BASH_COMMAND" == "$PROMPT_COMMAND" ]] && return
            
            # Skip blacklisted
            if is_blacklisted "$cmd_name"; then
                return
            fi
            
            export AMBROS_TRACK_CMD="$cmd_line"
        }
        
        ambros_bash_precmd() {
            local exit_code=$?
            
            if [[ -n "$AMBROS_TRACK_CMD" ]]; then
                (
                    command ambros run --store --tag "shell-integration" --tag "exit:$exit_code" -- echo "[tracked] $AMBROS_TRACK_CMD" >/dev/null 2>&1
                ) &
                unset AMBROS_TRACK_CMD
            fi
        }
        
        trap 'ambros_bash_preexec' DEBUG
        PROMPT_COMMAND="ambros_bash_precmd${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
        
        echo "✅ Ambros transparent mode activated (all commands - bash hooks)"
        echo "🚫 Blacklisted: ${AMBROS_BLACKLIST[*]:0:10}... (${#AMBROS_BLACKLIST[@]} total)"
        echo "💡 Note: 'all' mode logs commands after execution. For full output capture, use whitelist mode."
    else
        echo "⚠️  Ambros: 'all' mode requires zsh or bash. Falling back to whitelist mode."
        AMBROS_INTEGRATION_MODE="whitelist"
        # Re-source to use whitelist mode
        source "${BASH_SOURCE[0]:-$0}"
    fi
else
    echo "⚠️  Ambros: Unknown AMBROS_INTEGRATION_MODE='$AMBROS_INTEGRATION_MODE'. Use 'whitelist' or 'all'."
fi

# ============================================================================
# HELPER FUNCTIONS
# ============================================================================

# Temporarily disable Ambros integration for the current session
ambros_disable() {
    export AMBROS_INTEGRATION=off
    echo "🔕 Ambros integration disabled for this session"
}

# Re-enable Ambros integration
ambros_enable() {
    unset AMBROS_INTEGRATION
    echo "🔔 Ambros integration enabled - restart shell or re-source script"
}

# Add a command to the blacklist (session only)
ambros_blacklist_add() {
    local cmd="$1"
    if [[ -z "$cmd" ]]; then
        echo "Usage: ambros_blacklist_add <command>"
        return 1
    fi
    AMBROS_BLACKLIST+=("$cmd")
    AMBROS_BLACKLIST_MAP["$cmd"]=1
    echo "Added '$cmd' to blacklist"
}

# Remove a command from the blacklist (session only)
ambros_blacklist_remove() {
    local cmd="$1"
    if [[ -z "$cmd" ]]; then
        echo "Usage: ambros_blacklist_remove <command>"
        return 1
    fi
    unset "AMBROS_BLACKLIST_MAP[$cmd]"
    echo "Removed '$cmd' from blacklist (if present)"
}

# Show current integration status
ambros_status() {
    echo "Ambros Shell Integration Status:"
    echo "  Mode: $AMBROS_INTEGRATION_MODE"
    echo "  Auto flag support: $AMBROS_SUPPORTS_AUTO"
    echo "  Integration: ${AMBROS_INTEGRATION:-on}"
    if [[ "$AMBROS_INTEGRATION_MODE" == "whitelist" ]]; then
        echo "  Tracked commands: ${AMBROS_WRAPPED_COMMANDS[*]}"
    fi
    echo "  Blacklisted (${#AMBROS_BLACKLIST[@]}): ${AMBROS_BLACKLIST[*]:0:15}..."
}

# ============================================================================
# COMPLETION SUPPORT  
# ============================================================================
if [[ "$AMBROS_INTEGRATION_MODE" == "whitelist" ]]; then
    if [[ -n "$BASH_VERSION" ]]; then
        for cmd in "${AMBROS_WRAPPED_COMMANDS[@]}"; do
            complete -F _command "$cmd" 2>/dev/null || true
        done
    elif [[ -n "$ZSH_VERSION" ]]; then
        autoload -U compinit && compinit 2>/dev/null
    fi
fi