# Ambros Project Status - v4.0.0

## 🚀 Major Updates in v4.0.0

### 🖥️ TUI Dashboard
- Built a modern Terminal User Interface using **Bubble Tea** and **Lipgloss**.
- Interactive command browsing, real-time filtering (Fuzzy Search), and detailed view for output/metadata.
- Integrated quick actions: `r` for Rerun and `x` for Delete.
- Made TUI the default interface when running `ambros` without arguments.

### 🛡️ Security & Hardening (Production Ready)
- **Advanced Redacting**: Automatically masks sensitive data (API Keys, Passwords, Tokens) from command outputs.
- Supports OpenAI, GitHub, AWS, Slack, and Google Cloud patterns out of the box.
- User-configurable regex patterns via `.ambros.yaml`.

### 🧹 Data Management
- **Auto-Cleanup**: Implemented a retention policy to automatically delete commands older than X days.
- Configurable via `autoCleanup` settings in the configuration file.

### 🩺 Observability & Diagnostics
- **Persistent Logging**: Internal logs are now written to `~/.ambros/ambros.log`.
- New command `ambros logs`: View and follow internal logs.
- New command `ambros doctor`: Automated system health checks (DB, shell integration, config).

### 📦 CI/CD & Distribution
- **GoReleaser**: Full automation for cross-platform builds (Linux, macOS, Windows) and GitHub Releases.
- **Docker**: Provided an optimized multi-stage Dockerfile.
- **Auto-completion**: Support for Bash, Zsh, Fish, and PowerShell.

## 🛠 Tech Stack
- **Language**: Go 1.25+
- **TUI**: Charm.sh (Bubble Tea, Lipgloss, Bubbles)
- **Database**: BadgerDB (Key-Value Store)
- **CLI**: Cobra & Viper
- **CI/CD**: GitHub Actions + GoReleaser
