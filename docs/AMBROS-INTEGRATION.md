# Ambros transparent shell integration

This document explains how to enable "transparent" command tracking with Ambros using the helper script `scripts/.ambros-integration.sh`. When sourced into your shell profile, the script wraps a set of common developer commands and routes them through `ambros run --store --auto` so they are automatically captured in your Ambros history.

## Overview

The repository includes a helper script:
- `scripts/.ambros-integration.sh` — a shell script that creates functions for intercepted commands and forwards them to `ambros`.

When enabled, the script wraps commands like `git`, `docker`, `go`, `mvn`, `gradle`, etc.

By default the integration will try to call:

```
ambros run --store --auto -- <command> <args...>
```

However, older Ambros releases may not support the `--auto` flag. The integration script
performs runtime detection: if `ambros run --help` doesn't advertise `--auto`, the wrapper
falls back to calling:

```
ambros run --store -- <command> <args...>
```

This ensures the integration works across Ambros versions while preferring the newer `--auto`
behavior when supported. The `--` ensures arguments are passed to the wrapped command unchanged.

## Prerequisites

- `ambros` must be installed and reachable in your `PATH` (for example via `go install github.com/gi4nks/ambros/v3@latest` or your package of choice).
- A Bourne-like shell (bash, zsh). The script has minimal support for both.

## Installing (per-user)

1. Copy or symlink the script to a stable path (optional):
```bash
# example: copy into home config
cp scripts/.ambros-integration.sh ~/.ambros-integration.sh
chmod +x ~/.ambros-integration.sh
```

2. Source it from your shell profile:

- For zsh (`~/.zshrc`):
```bash
# add to ~/.zshrc
source ~/.ambros-integration.sh
```

- For bash (`~/.bashrc` or `~/.bash_profile`):
```bash
# add to ~/.bashrc
source ~/.ambros-integration.sh
```

Reload your shell or run `source ~/.zshrc` (or `source ~/.bashrc`).

## Quick test

After sourcing, try a simple command such as:
```bash
git status
```
You should see normal output and the wrapper will call `ambros run --store --auto -- git status` under the hood. To bypass the wrapper for a single invocation you can use the `command` builtin or an absolute path, e.g.:
```bash
command ls -la
# or
/usr/bin/ls -la
```

## Customization

- Change which commands are intercepted by editing the `AMBROS_INTERCEPTED_COMMANDS` array in `scripts/.ambros-integration.sh`.
- The script saves original executables in variables like `ORIGINAL_GIT` so you can modify the wrapper to call \${ORIGINAL_GIT} directly if needed.
- To temporarily disable tracking in your session:
```bash
unset AMBROS_INTERCEPTED_COMMANDS
# or open a new shell without sourcing the script
```

## Security & privacy notes

- The script executes wrapped commands through `ambros`, which will store metadata and the command/arguments per your Ambros configuration.
- Do not enable this on multi-user or shared shells unless you trust the environment.
- The wrapper may expose sensitive command arguments (passwords, tokens) if you run them in plain arguments. Prefer environment variables or safe input methods for secrets.

## Caveats & known limitations

- `sudo <command>` typically runs the program directly and will not go through the wrapper. If you try `sudo git`, the wrapper won't intercept it (and it may bypass User PATH). Use care when combining `sudo`.
- Interactive programs that expect direct TTY control (some `ssh` usage, editors) may behave differently when routed through another program. Test carefully.
- Completion support: the script attempts to preserve shell completion for intercepted commands but some completions may break depending on how your shell is configured.

## Troubleshooting

- If commands stopped working after sourcing:
  - Ensure `ambros` is available in `PATH`.
  - Try `command <cmd>` (bypasses wrapper) to see if native executable still works.
  - Check for errors when you source the script: `source ~/.ambros-integration.sh` and watch STDOUT/STDERR.
- If completions are broken, re-run `compinit` (zsh) or re-enable completion for bash.

## Uninstall / disable

- Remove the `source` line from your shell profile and restart the shell session.
- If you copied the script to `~/.ambros-integration.sh`, you can delete it:
```bash
rm ~/.ambros-integration.sh
```

## Example: minimal `~/.zshrc` snippet
```bash
# load ambros integration (ensure correct path)
if [ -f "$HOME/.ambros-integration.sh" ]; then
  source "$HOME/.ambros-integration.sh"
fi
```

If you'd like, I can:
- Commit this file to `docs/` for you, or
- Add an `ambros integrate` subcommand that installs and wires this script safely into user profiles.
