package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/models"
	"github.com/gi4nks/ambros/v3/internal/plugins"
)

type MCPCommand struct {
	*BaseCommand
}

func NewMCPCommand(logger *zap.Logger, repo RepositoryInterface, api plugins.CoreAPI) *MCPCommand {
	mc := &MCPCommand{}

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP server exposing Ambros tools",
		Long: `Start a Model Context Protocol (MCP) server that exposes Ambros functionality as tools.

The MCP server runs over stdio and provides the following tools:
  - ambros_last: Get recent commands from history
  - ambros_search: Search command history
  - ambros_analytics: Get analytics and statistics
  - ambros_output: Get output of a specific command
  - ambros_command: Get details of a specific command

This allows AI assistants and other MCP clients to interact with your command history.

Example:
  ambros mcp

Configure in Claude Desktop (~/.config/claude/claude_desktop_config.json):
  {
    "mcpServers": {
      "ambros": {
        "command": "ambros",
        "args": ["mcp"]
      }
    }
  }`,
		RunE: mc.runE,
	}

	mc.BaseCommand = NewBaseCommand(cmd, logger, repo, api)
	mc.cmd = cmd
	return mc
}

func (mc *MCPCommand) runE(cmd *cobra.Command, args []string) error {
	mc.logger.Debug("Starting MCP server")

	// Create MCP server
	s := server.NewMCPServer(
		"ambros",
		"3.3.0",
		server.WithToolCapabilities(true),
	)

	// Register tools
	mc.registerTools(s)

	// Start stdio server
	if err := server.ServeStdio(s); err != nil {
		mc.logger.Error("MCP server error", zap.Error(err))
		return fmt.Errorf("MCP server error: %w", err)
	}

	return nil
}

func (mc *MCPCommand) registerTools(s *server.MCPServer) {
	// Tool: ambros_last - Get recent commands
	s.AddTool(
		mcp.NewTool("ambros_last",
			mcp.WithDescription("Get the most recent commands from Ambros history. Returns command details including ID, name, arguments, status, tags, and timestamps."),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of commands to return (default: 10, max: 100)"),
			),
			mcp.WithBoolean("failed_only",
				mcp.Description("If true, only return failed commands"),
			),
		),
		mc.handleLast,
	)

	// Tool: ambros_search - Search command history
	s.AddTool(
		mcp.NewTool("ambros_search",
			mcp.WithDescription("Search through Ambros command history using text queries and filters. Returns matching commands with details."),
			mcp.WithString("query",
				mcp.Description("Text to search for in command names and arguments"),
			),
			mcp.WithString("tag",
				mcp.Description("Filter by tag"),
			),
			mcp.WithString("category",
				mcp.Description("Filter by category"),
			),
			mcp.WithString("status",
				mcp.Description("Filter by status: 'success' or 'failed'"),
			),
			mcp.WithString("since",
				mcp.Description("Show commands since duration (e.g., '24h', '7d', '30d')"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of results (default: 10, max: 100)"),
			),
		),
		mc.handleSearch,
	)

	// Tool: ambros_analytics - Get analytics
	s.AddTool(
		mcp.NewTool("ambros_analytics",
			mcp.WithDescription("Get analytics and statistics about command usage. Provides summary stats, most used commands, slowest commands, or failure analysis."),
			mcp.WithString("action",
				mcp.Description("Type of analytics: 'summary' (default), 'most-used', 'slowest', or 'failures'"),
			),
		),
		mc.handleAnalytics,
	)

	// Tool: ambros_output - Get command output
	s.AddTool(
		mcp.NewTool("ambros_output",
			mcp.WithDescription("Get the full output (stdout/stderr) of a specific command by its ID."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("The command ID to retrieve output for"),
			),
		),
		mc.handleOutput,
	)

	// Tool: ambros_command - Get command details
	s.AddTool(
		mcp.NewTool("ambros_command",
			mcp.WithDescription("Get detailed information about a specific command by its ID, including full command line, output, timestamps, and metadata."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("The command ID to retrieve"),
			),
		),
		mc.handleCommand,
	)
}

// handleLast returns recent commands
func (mc *MCPCommand) handleLast(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := request.GetInt("limit", 10)
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 1
	}

	failedOnly := request.GetBool("failed_only", false)

	commands, err := mc.repository.GetAllCommands()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to retrieve commands: %v", err)), nil
	}

	// Sort by creation time (most recent first)
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].CreatedAt.After(commands[j].CreatedAt)
	})

	// Filter and limit
	var result []models.Command
	for _, cmd := range commands {
		if failedOnly && cmd.Status {
			continue
		}
		result = append(result, cmd)
		if len(result) >= limit {
			break
		}
	}

	// Format response
	response := mc.formatCommandsResponse(result)
	return mcp.NewToolResultText(response), nil
}

// handleSearch searches command history
func (mc *MCPCommand) handleSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := request.GetString("query", "")
	tag := request.GetString("tag", "")
	category := request.GetString("category", "")
	status := request.GetString("status", "")
	since := request.GetString("since", "")
	limit := request.GetInt("limit", 10)
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 1
	}

	commands, err := mc.repository.GetAllCommands()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to retrieve commands: %v", err)), nil
	}

	// Parse since duration
	var sinceTime time.Time
	if since != "" {
		duration, err := time.ParseDuration(since)
		if err == nil {
			sinceTime = time.Now().Add(-duration)
		}
	}

	// Filter commands
	var filtered []models.Command
	for _, cmd := range commands {
		// Text query filter
		if query != "" {
			if !strings.Contains(strings.ToLower(cmd.Name), strings.ToLower(query)) &&
				!strings.Contains(strings.ToLower(strings.Join(cmd.Arguments, " ")), strings.ToLower(query)) {
				continue
			}
		}

		// Tag filter
		if tag != "" {
			tagMatch := false
			for _, cmdTag := range cmd.Tags {
				if strings.EqualFold(cmdTag, tag) {
					tagMatch = true
					break
				}
			}
			if !tagMatch {
				continue
			}
		}

		// Category filter
		if category != "" && !strings.EqualFold(cmd.Category, category) {
			continue
		}

		// Status filter
		if status != "" {
			if (status == "success" && !cmd.Status) || (status == "failed" && cmd.Status) {
				continue
			}
		}

		// Since filter
		if !sinceTime.IsZero() && cmd.CreatedAt.Before(sinceTime) {
			continue
		}

		filtered = append(filtered, cmd)
	}

	// Apply limit
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	// Format response
	response := mc.formatCommandsResponse(filtered)
	return mcp.NewToolResultText(response), nil
}

// handleAnalytics returns analytics data
func (mc *MCPCommand) handleAnalytics(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action := request.GetString("action", "summary")

	commands, err := mc.repository.GetAllCommands()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to retrieve commands: %v", err)), nil
	}

	var response string

	switch action {
	case "summary":
		response = mc.formatSummaryAnalytics(commands)
	case "most-used":
		response = mc.formatMostUsedAnalytics(commands)
	case "slowest":
		response = mc.formatSlowestAnalytics(commands)
	case "failures":
		response = mc.formatFailuresAnalytics(commands)
	default:
		return mcp.NewToolResultError(fmt.Sprintf("Unknown action: %s. Use 'summary', 'most-used', 'slowest', or 'failures'", action)), nil
	}

	return mcp.NewToolResultText(response), nil
}

// handleOutput returns command output
func (mc *MCPCommand) handleOutput(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := request.RequireString("id")
	if err != nil || id == "" {
		return mcp.NewToolResultError("Command ID is required"), nil
	}

	command, err := mc.repository.Get(id)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Command not found: %s", id)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Command: %s %s\n", command.Name, strings.Join(command.Arguments, " ")))
	sb.WriteString(fmt.Sprintf("Status: %s\n", mc.formatStatus(command.Status)))
	sb.WriteString(fmt.Sprintf("Executed: %s\n\n", command.CreatedAt.Format(time.RFC3339)))

	if command.Output != "" {
		sb.WriteString("=== STDOUT ===\n")
		sb.WriteString(command.Output)
		sb.WriteString("\n")
	}

	if command.Error != "" {
		sb.WriteString("\n=== STDERR ===\n")
		sb.WriteString(command.Error)
		sb.WriteString("\n")
	}

	if command.Output == "" && command.Error == "" {
		sb.WriteString("No output captured for this command.\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// handleCommand returns full command details
func (mc *MCPCommand) handleCommand(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := request.RequireString("id")
	if err != nil || id == "" {
		return mcp.NewToolResultError("Command ID is required"), nil
	}

	command, err := mc.repository.Get(id)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Command not found: %s", id)), nil
	}

	// Return as JSON for structured data
	data, err := json.MarshalIndent(command, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format command: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

// Helper functions for formatting responses

func (mc *MCPCommand) formatCommandsResponse(commands []models.Command) string {
	if len(commands) == 0 {
		return "No commands found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d command(s):\n\n", len(commands)))

	for i, cmd := range commands {
		sb.WriteString(fmt.Sprintf("%d. %s %s\n", i+1, cmd.Name, strings.Join(cmd.Arguments, " ")))
		sb.WriteString(fmt.Sprintf("   ID: %s\n", cmd.ID))
		sb.WriteString(fmt.Sprintf("   Status: %s\n", mc.formatStatus(cmd.Status)))
		sb.WriteString(fmt.Sprintf("   Created: %s\n", cmd.CreatedAt.Format("2006-01-02 15:04:05")))

		if len(cmd.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("   Tags: %s\n", strings.Join(cmd.Tags, ", ")))
		}
		if cmd.Category != "" {
			sb.WriteString(fmt.Sprintf("   Category: %s\n", cmd.Category))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (mc *MCPCommand) formatStatus(status bool) string {
	if status {
		return "Success"
	}
	return "Failed"
}

func (mc *MCPCommand) formatSummaryAnalytics(commands []models.Command) string {
	if len(commands) == 0 {
		return "No commands found in history."
	}

	totalCommands := len(commands)
	successCount := 0
	totalDuration := time.Duration(0)

	for _, cmd := range commands {
		if cmd.Status {
			successCount++
		}
		duration := cmd.TerminatedAt.Sub(cmd.CreatedAt)
		totalDuration += duration
	}

	successRate := float64(successCount) / float64(totalCommands) * 100
	avgDuration := totalDuration / time.Duration(totalCommands)

	var sb strings.Builder
	sb.WriteString("=== Command Analytics Summary ===\n\n")
	sb.WriteString(fmt.Sprintf("Total Commands: %d\n", totalCommands))
	sb.WriteString(fmt.Sprintf("Successful: %d\n", successCount))
	sb.WriteString(fmt.Sprintf("Failed: %d\n", totalCommands-successCount))
	sb.WriteString(fmt.Sprintf("Success Rate: %.1f%%\n", successRate))
	sb.WriteString(fmt.Sprintf("Average Duration: %v\n", avgDuration.Round(time.Millisecond)))

	return sb.String()
}

func (mc *MCPCommand) formatMostUsedAnalytics(commands []models.Command) string {
	cmdCount := make(map[string]int)
	for _, cmd := range commands {
		cmdCount[cmd.Name]++
	}

	type cmdFreq struct {
		name  string
		count int
	}

	var sorted []cmdFreq
	for name, count := range cmdCount {
		sorted = append(sorted, cmdFreq{name, count})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	var sb strings.Builder
	sb.WriteString("=== Most Used Commands ===\n\n")

	limit := 10
	if len(sorted) < limit {
		limit = len(sorted)
	}

	for i := 0; i < limit; i++ {
		sb.WriteString(fmt.Sprintf("%d. %s - %d times\n", i+1, sorted[i].name, sorted[i].count))
	}

	return sb.String()
}

func (mc *MCPCommand) formatSlowestAnalytics(commands []models.Command) string {
	type cmdDuration struct {
		cmd      models.Command
		duration time.Duration
	}

	var withDuration []cmdDuration
	for _, cmd := range commands {
		duration := cmd.TerminatedAt.Sub(cmd.CreatedAt)
		if duration > time.Millisecond*10 {
			withDuration = append(withDuration, cmdDuration{cmd, duration})
		}
	}

	sort.Slice(withDuration, func(i, j int) bool {
		return withDuration[i].duration > withDuration[j].duration
	})

	var sb strings.Builder
	sb.WriteString("=== Slowest Commands ===\n\n")

	limit := 10
	if len(withDuration) < limit {
		limit = len(withDuration)
	}

	for i := 0; i < limit; i++ {
		item := withDuration[i]
		sb.WriteString(fmt.Sprintf("%d. %s %s - %v\n",
			i+1,
			item.cmd.Name,
			strings.Join(item.cmd.Arguments, " "),
			item.duration.Round(time.Millisecond)))
	}

	return sb.String()
}

func (mc *MCPCommand) formatFailuresAnalytics(commands []models.Command) string {
	failCount := make(map[string]int)
	for _, cmd := range commands {
		if !cmd.Status {
			failCount[cmd.Name]++
		}
	}

	if len(failCount) == 0 {
		return "No command failures found! 🎉"
	}

	type failureStats struct {
		name     string
		failures int
	}

	var failures []failureStats
	for name, count := range failCount {
		failures = append(failures, failureStats{name, count})
	}

	sort.Slice(failures, func(i, j int) bool {
		return failures[i].failures > failures[j].failures
	})

	var sb strings.Builder
	sb.WriteString("=== Commands with Most Failures ===\n\n")

	limit := 10
	if len(failures) < limit {
		limit = len(failures)
	}

	for i := 0; i < limit; i++ {
		sb.WriteString(fmt.Sprintf("%d. %s - %d failures\n", i+1, failures[i].name, failures[i].failures))
	}

	return sb.String()
}

func (mc *MCPCommand) Command() *cobra.Command {
	return mc.cmd
}
