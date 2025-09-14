# Ambros Shell Integration

This document explains the transparent shell integration script shipped in `scripts/.ambros-integration.sh` and how to safely enable, disable, and troubleshoot it.

## Quick enable

Source the script from your shell profile (e.g., `~/.bashrc` or `~/.zshrc`):

```bash
source ~/path/to/ambros/scripts/.ambros-integration.sh
```

You should see a confirmation message: `✅ Ambros transparent mode activated`.

## What it does

- Creates shell functions for a curated list of common developer tools (git, npm, docker, gradle/gradlew, mvn, go, etc.).
- When you run one of those commands, the wrapper transparently calls `ambros run --store --auto -- <cmd>` (or falls back to `ambros run --store -- <cmd>` when `--auto` is not supported by your installed Ambros).
- The integration attempts to preserve original behaviour and exit codes. For most CLI programs the behavior will be identical.

## Opt-out and automatic bypass

- Per-shell opt-out: set `AMBROS_INTEGRATION=off` before sourcing the script to disable it for that shell/session.
# Ambros Shell Integration

This document explains the transparent shell integration helper and the `ambros integrate` installer. It covers the integration script, how to enable it, installation via the `ambros` CLI, and CI/sync guidance.

## Quick enable

If you prefer manual installation, copy the helper script and source it from your shell profile:

```bash
cp scripts/.ambros-integration.sh ~/.ambros-integration.sh
chmod +x ~/.ambros-integration.sh
echo "source ~/.ambros-integration.sh" >> ~/.zshrc  # or ~/.bashrc
source ~/.zshrc
```

Or use the built-in installer (recommended):

```bash
ambros integrate install        # interactive
ambros integrate install --yes  # non-interactive (CI/scripts)
ambros integrate install --shell ~/.zshrc  # target a specific rc file
```

You should see: `✅ Ambros transparent mode activated` when the script is sourced in a new shell.

## What it does

- Wraps a curated list of developer tools (git, npm, docker, gradle/gradlew, mvn, go, etc.) with shell functions that route executions through Ambros.
- At runtime the wrapper prefers calling:

```
ambros run --store --auto -- <command> <args...>
```

  If the installed Ambros doesn't support `--auto`, the script falls back to:

```
ambros run --store -- <command> <args...>
```

This guarantees backward compatibility while preferring the newer transparent mode when available.

## Installer: `ambros integrate`

The CLI includes an `integrate` command to automate installation and removal:

- `ambros integrate install [--shell <path>] [--yes]`
  - Writes `~/.ambros-integration.sh` (idempotent — it only overwrites when content changes).
  - Adds `source ~/.ambros-integration.sh` to `~/.bashrc` and `~/.zshrc` by default (or to the path provided with `--shell`).
  - Prompts for confirmation; pass `--yes` to skip prompts (useful for scripts/CI).

- `ambros integrate uninstall [--shell <path>] [--yes]`
  - Removes `~/.ambros-integration.sh` and removes the `source` line from the specified rc file(s).
  - Prompts for confirmation unless `--yes` is provided.

Examples:

```bash
# interactive install that updates both ~/.bashrc and ~/.zshrc
ambros integrate install

# unattended install for CI or scripted setups
ambros integrate install --yes

# install and only update zshrc
ambros integrate install --shell ~/.zshrc
```

## Idempotency & safety

- The installer checks the installed script content and will not rewrite an identical file.
- When adding `source` lines the installer checks for existing lines and appends only if missing.
- Uninstall removes the first matching `source ~/.ambros-integration.sh` line to avoid touching unrelated content aggressively.

## go generate / build-time sync

To keep the embedded installer script in the CLI up-to-date, the project uses a `go:generate` helper in `cmd/commands`:

```bash
go generate ./cmd/commands
```

This copies `scripts/.ambros-integration.sh` into `cmd/commands/scripts/` so the CLI can embed it with `//go:embed` and ship the installer with the binary. Add `go generate ./cmd/commands` to your CI before `go build` to ensure the embedded copy is synced during automated builds.

## CI guidance

- For automated installation in CI environments, use `ambros integrate install --yes` and avoid sourcing the script in non-interactive jobs unless you explicitly want command capturing in that environment.
- The integration script already auto-bypasses non-interactive shells and common hook variables (e.g., `CI`, `GIT_PARAMS`) to avoid interfering with CI or git hooks.

- ## PTY support (interactive programs)

- Modern Ambros builds include PTY-backed interactive support when you run commands with `ambros run --auto`.
  - The CLI will allocate a pseudo-terminal when your shell session is a TTY and forward window-size changes and signals to the child process. This enables interactive/full-screen programs (e.g., `vim`, `ssh`, `docker -it`, and other TUIs) to behave correctly when run through Ambros.
  - If you rely on interactive programs through the integration, make sure you have an Ambros build that includes PTY support (recent releases include this feature).
  - The integration script will continue to detect whether Ambros supports `--auto` at runtime; older Ambros versions will fall back to `ambros run --store --` which does not allocate a PTY.

## Limitations

- While PTY support covers the vast majority of interactive cases, there are edge cases around advanced terminal modes and job-control signals that may need additional tuning. If you encounter issues with a specific program, please open an issue with reproduction steps.

## Interactive examples

Here are two short examples showing how to run interactive programs through Ambros. These rely on the PTY-backed support in modern Ambros builds.

- Open an editor (vim) through Ambros and preserve the interactive TTY:

```bash
# Run vim and have it behave normally inside your terminal
ambros run --auto -- vim README.md
```

When you exit vim, Ambros will preserve vim's exit code and store the run if configured.

- Run a Docker container interactively (`-it`) through Ambros:

```bash
ambros run --auto -- docker run -it --rm ubuntu bash
```

You should get an interactive shell inside the container, resize events and signals will be proxied correctly thanks to PTY allocation.

Notes:

- If your installed Ambros is older and `--auto` isn't available, the integration script will fall back to `ambros run --store --` and PTY behavior will not be active — update Ambros or use a binary built with PTY support for interactive runs.
- For CI or headless environments where no TTY is present, don't use `--auto`; the non-PTY path will be used instead.

### Local PTY smoke test

You can run a quick local smoke test (requires `python3` and an Ambros binary in your PATH):

```bash
chmod +x scripts/pty-smoke.sh
./scripts/pty-smoke.sh
```

The script spawns a small command inside a pty and looks for the `PTY_OK` marker in output. It should print `PTY smoke test: OK` when the PTY path is working correctly.
- The installer performs simple text edits to rc files; if users heavily customize their rc files, manual review may be needed.

## Troubleshooting

- If things go wrong, temporarily disable integration for the session:

```bash
export AMBROS_INTEGRATION=off
```

- To bypass the wrapper for a single command:

```bash
command git status
# or
/usr/bin/git status
```

## Where to look next

- `scripts/.ambros-integration.sh` — master copy
- `cmd/commands/integrate.go` — CLI installer implementation
- `go generate ./cmd/commands` — sync step for embedding the script

If you'd like, I can add a CI check that fails the build when the embedded copy and `scripts/.ambros-integration.sh` diverge.
