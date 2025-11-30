# Ambros: Potential Improvements and New Features

This document outlines potential improvements and new features for the Ambros project, based on an analysis of the codebase and its existing features.

**Last Updated**: November 30, 2025

## Status Legend
- ✅ **Completed** - Feature fully implemented and working
- 🟡 **Partial** - Feature partially implemented, needs more work
- ❌ **Not Started** - Feature not yet implemented

---

## Core & Usability

*   ✅ **Configuration Validation**: [Completed] Implemented a `config validate` command to check the syntax and validity of the `.ambros.yaml` file, helping users avoid runtime errors due to misconfiguration. This involved integrating `viper` for robust configuration loading and adding a validation mechanism.

*   ✅ **Shell Alias/Widget**: [Completed] Transparent shell integration implemented in `scripts/.ambros-integration.sh`. Supports bash and zsh with auto-detection. Install via `ambros integrate install`. Automatically intercepts common commands (ls, git, curl, docker, etc.) and logs them through Ambros with `--auto` flag support.

*   ❌ **Command History Sync**: Add a feature to sync the command history across multiple devices. This could be implemented using a cloud storage backend (e.g., S3, Google Drive, or a self-hosted server).
    - *Status*: No cloud/remote storage backends implemented yet.

*   🟡 **Interactive Command Builder**: An interactive `ambros run --interactive` mode that walks the user through building a command, selecting arguments, adding tags, etc. This would be especially helpful for new users.
    - *Status*: `ambros interactive` command exists for browsing/executing stored commands, but `ambros run --interactive` (command builder wizard) is NOT implemented.

*   ❌ **AI-Powered `fix` command**: A new `ambros fix` command that, when executed after a failed command (`ambros run ...`), uses an LLM to analyze the error and suggest a fix. The user could then approve and execute the suggested command.
    - *Status*: Not started. No LLM integration exists.

---

## Plugin System

*   ✅ **Plugin API (CoreAPI)**: [Completed] Exposed a comprehensive Go API for plugins, allowing them to interact with the Ambros core in a structured way.
    - *Status*:
      - ✅ `CoreAPI` interface defined in `internal/plugins/api.go`
      - ✅ `GoPlugin` interface for Go-based internal plugins
      - ✅ `InternalPluginRegistry` for plugin management
      - ✅ `coreAPIImpl` with full implementations:
        - `ExecuteCommand` - with store, tag, category, dry-run, and auto mode options
        - `RegisterCommand` - register new Cobra commands dynamically
        - `TriggerHook` - dispatches hooks to all subscribed Go internal plugins
      - ✅ Go internal plugin execution via `plugin_runner.go`
      - ✅ Example Go plugin (`internal/plugins/example_plugin.go`) with hello/info commands
      - ✅ Unit tests for plugin system (`internal/plugins/plugins_test.go`)
      - 🟡 Plugin lifecycle management (start/stop/restart) - not yet implemented

*   ❌ **Official Plugin Repository**: Create an official, curated repository of high-quality plugins. This would make it easier for users to discover and install useful plugins.
    - *Status*: `plugin install` supports local paths and simple URL registry, but no official curated repository exists.

*   ❌ **Plugin Sandboxing**: The `README.md` mentions "Security Sandboxing" for plugins, but the implementation in `plugin.go` mainly focuses on path safety. True sandboxing (e.g., using containers or a restricted execution environment like gVisor) would be a significant improvement for security.
    - *Status*: Only path validation implemented. No container/gVisor isolation.

*   ❌ **Webhooks for Plugins**: Allow plugins to register webhooks to be notified of events (e.g., command execution, a new plugin is installed, etc.). This would enable more reactive and integrated plugins.
    - *Status*: Plugin config supports `webhook.url` but it's just configuration storage. No event-driven webhook system implemented.

---

## Web Dashboard

*   ❌ **Real-time Log Streaming**: Stream the output of running commands to the web dashboard in real-time using WebSockets. This would provide a much better user experience than having to wait for the command to finish.
    - *Status*: Dashboard uses REST API only. No WebSocket support.

*   ❌ **Visual Workflow Builder**: The `README.md` mentions a "Visual Workflow Builder" for command chains. If this is not yet implemented, it would be a fantastic feature that would make it much easier to create and manage complex workflows.
    - *Status*: Chain commands exist (CLI only). No visual/drag-drop UI.

*   ❌ **User Authentication**: Add user authentication to the web dashboard to control access in a multi-user environment. This is essential for teams and enterprise users.
    - *Status*: Server has no auth. `ErrUnauthorized` error code exists but not used. `--auth` flag mentioned in README but not implemented.

*   ❌ **Team Collaboration**: Introduce features for teams, such as shared command history, templates, and dashboards. This would make Ambros a more powerful tool for collaboration.
    - *Status*: Not started. No multi-user or sharing features.

---

## Analytics

*   ✅ **Deeper Analytics**: [Completed] Provide more in-depth analytics, including identifying user's most common command patterns and suggesting aliases for frequently used long commands.
    - *Status*: Fully implemented in `server.go` with `/api/analytics/advanced` endpoint:
      - ✅ **Alias Suggestions**: Automatically detects long commands used frequently (>15 chars, ≥3 uses) and suggests shell aliases with character savings calculation
      - ✅ **Complex Flag Pattern Detection**: Identifies frequently used flag combinations and suggests aliases
      - ✅ **Command Sequence Patterns**: Detects 2-command and 3-command sequences that are often run together within 5 minutes, with occurrence counts and average intervals
      - ✅ **Workflow Insights**: Identifies common workflows (Git, Docker, Kubernetes, Node.js, Python, Go, CI/CD) and provides specific automation suggestions
      - ✅ **Command Complexity Analysis**: Measures command complexity based on arguments, flags, pipes, redirects, subshells, and logical operators
      - ✅ Unit tests for all deep analytics functions (8 new tests)

*   ❌ **Integration with Monitoring Tools**: Integrate with tools like Prometheus or Grafana to export metrics. This would allow users to monitor the performance and usage of Ambros in their existing monitoring infrastructure.
    - *Status*: Internal metrics exist in server analytics but not exported. No `/metrics` endpoint for Prometheus scraping.

---

## Refactoring and Optimization

*   ✅ **Code Duplication**: [Completed] Refactored command execution logic into a shared `executor.go` to improve maintainability and reduce redundancy across `run.go`, `interactive.go`, `chain.go`, and `env.go`.

*   ✅ **Error Handling**: [Completed] Standardized error handling across the application by consistently using `internal/errors.NewError` and specific error codes, replacing ad-hoc `fmt.Errorf` calls.

*   🟡 **Testing**: Increase test coverage, especially for the web dashboard and the plugin system.
    - *Status*: Current coverage is **41.6%** for `cmd/commands`. Target is 80%+ per CONTRIBUTING.md.
    - ✅ Plugin system tests added (`internal/plugins/plugins_test.go` - 12 tests)
    - Missing test coverage for:
      - Web server endpoints (`server.go`)
      - Interactive command flows
      - Chain execution edge cases

---

## Priority Recommendations

Based on the current state, here are recommended priorities for completion:

### High Priority (Foundation)
1. ~~**Complete CoreAPI/Plugin API**~~ ✅ Completed - `TriggerHook`, `ExecuteCommand`, example plugin implemented
2. ~~**Deeper Analytics**~~ ✅ Completed - Alias suggestions, sequence patterns, workflow insights, complexity analysis
3. **Increase Test Coverage** - Target 80%+ before adding new features
4. **User Authentication** - Essential for production/team use
5. **Plugin Lifecycle Management** - Add start/stop/restart capabilities

### Medium Priority (User Experience)
6. **Interactive Command Builder** - `ambros run --interactive` wizard
7. **Real-time Log Streaming** - WebSocket support for dashboard
8. **Prometheus Metrics Export** - `/metrics` endpoint

### Lower Priority (Nice to Have)
9. **Command History Sync** - Cloud storage backends
10. **Visual Workflow Builder** - Drag-drop chain builder
11. **AI-Powered `fix` command** - LLM integration
12. **Plugin Sandboxing** - Container isolation
