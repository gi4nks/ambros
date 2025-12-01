package commands

import (
	"context"
	"encoding/json"
	"fmt"
	iofs "io/fs"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/sahilm/fuzzy"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/api"
	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"
	webstatic "github.com/gi4nks/ambros/v3/web/static"
)

// embedded static files are provided by the web/static package

// ServerCommand represents the web server command
type ServerCommand struct {
	*BaseCommand
	port    int
	host    string
	dev     bool
	cors    bool
	verbose bool
}

// NewServerCommand creates a new server command
func NewServerCommand(logger *zap.Logger, repo RepositoryInterface) *ServerCommand {
	sc := &ServerCommand{}

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the Ambros web dashboard server",
		Long: `Start a web server that provides a dashboard interface for Ambros.
The dashboard allows you to manage commands, view analytics, and configure
environments through a user-friendly web interface.

Features:
  - Command history browsing and search
  - Real-time analytics dashboard
  - Environment and template management
  - Interactive command execution
  - Export/import functionality

Examples:
  ambros server                          # Start server on default port 8080
  ambros server --port 3000              # Start on custom port
  ambros server --host 0.0.0.0          # Listen on all interfaces
  ambros server --dev                   # Development mode with hot reload
  ambros server --cors                  # Enable CORS headers`,
		RunE: sc.runE,
	}

	sc.BaseCommand = NewBaseCommand(cmd, logger, repo)
	sc.cmd = cmd
	sc.setupFlags(cmd)
	return sc
}

func (sc *ServerCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&sc.port, "port", "p", 8080, "Port to listen on")
	cmd.Flags().StringVarP(&sc.host, "host", "H", "localhost", "Host to bind to")
	cmd.Flags().BoolVarP(&sc.dev, "dev", "d", false, "Enable development mode")
	cmd.Flags().BoolVar(&sc.cors, "cors", false, "Enable CORS headers")
	cmd.Flags().BoolVarP(&sc.verbose, "verbose", "v", false, "Verbose logging")
}

func (sc *ServerCommand) runE(cmd *cobra.Command, args []string) error {
	sc.logger.Info("Starting Ambros web server",
		zap.String("host", sc.host),
		zap.Int("port", sc.port),
		zap.Bool("dev", sc.dev),
		zap.Bool("cors", sc.cors))

	// Create API server
	apiServer := api.NewServer(sc.repository, sc.logger)

	// Enhanced API with advanced dashboard features
	enhancedAPI := sc.createEnhancedAPI(apiServer)

	// Setup HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", sc.host, sc.port),
		Handler: enhancedAPI,
	}

	// Display startup information
	sc.displayStartupInfo()

	// Handle graceful shutdown
	go sc.handleShutdown(server)

	// Start server
	color.Green("🚀 Ambros Dashboard Server starting...")
	color.Cyan("📱 Web Interface: http://%s:%d", sc.host, sc.port)
	color.Cyan("🔗 API Endpoint: http://%s:%d/api", sc.host, sc.port)
	color.Yellow("⏹️  Press Ctrl+C to stop the server")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		sc.logger.Error("Server failed to start", zap.Error(err))
		return errors.NewError(errors.ErrInternalServer, "failed to start server", err)
	}

	return nil
}

func (sc *ServerCommand) createEnhancedAPI(apiServer *api.Server) http.Handler {
	mux := http.NewServeMux()

	// Setup CORS middleware if enabled
	var handler http.Handler = mux
	if sc.cors {
		handler = sc.corsMiddleware(mux)
	}

	// API routes
	apiMux := apiServer.SetupRoutes()
	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	// Enhanced API endpoints for advanced dashboard and integrations
	mux.HandleFunc("/api/dashboard", sc.handleDashboard)
	mux.HandleFunc("/api/analytics/advanced", sc.handleAdvancedAnalytics)
	mux.HandleFunc("/api/plugins", sc.handlePlugins)
	mux.HandleFunc("/api/search/smart", sc.handleSmartSearch)

	// Static files for web dashboard
	mux.HandleFunc("/", sc.handleWebApp)
	// Serve embedded static files at /static/ using Go embed
	sub, err := iofs.Sub(webstatic.StaticFiles, ".")
	if err == nil {
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))
	} else {
		// Fallback to on-disk static serving when embed FS is unavailable
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))
	}

	if sc.cors {
		return handler
	}
	return mux
}

func (sc *ServerCommand) handleDashboard(w http.ResponseWriter, r *http.Request) {
	sc.logger.Debug("Dashboard request", zap.String("method", r.Method))

	// Get summary statistics
	commands, err := sc.repository.GetAllCommands()
	if err != nil {
		http.Error(w, "Failed to get commands", http.StatusInternalServerError)
		return
	}

	// Calculate dashboard metrics
	totalCommands := len(commands)
	successCount := 0
	recentCommands := 0
	now := time.Now()
	last24h := now.Add(-24 * time.Hour)

	for _, cmd := range commands {
		if cmd.Status {
			successCount++
		}
		if cmd.CreatedAt.After(last24h) {
			recentCommands++
		}
	}

	dashboard := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_commands":  totalCommands,
			"success_rate":    float64(successCount) / float64(totalCommands) * 100,
			"recent_commands": recentCommands,
		},
		"recent_activity": commands[max(0, len(commands)-10):],
		"quick_stats": map[string]interface{}{
			"most_used_today": sc.getMostUsedCommands(commands, 1),
			"failure_rate":    float64(totalCommands-successCount) / float64(totalCommands) * 100,
			"avg_exec_time":   sc.calculateAverageExecutionTime(commands),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dashboard); err != nil {
		sc.logger.Error("Failed to encode dashboard", zap.Error(err))
	}
}

func (sc *ServerCommand) handleAdvancedAnalytics(w http.ResponseWriter, r *http.Request) {
	sc.logger.Debug("Advanced analytics request")

	commands, err := sc.repository.GetAllCommands()
	if err != nil {
		http.Error(w, "Failed to get commands", http.StatusInternalServerError)
		return
	}

	// Generate deep analytics with alias suggestions and pattern recognition
	deepAnalytics := sc.generateDeepAnalytics(commands)

	analytics := map[string]interface{}{
		"command_patterns":    sc.analyzeCommandPatterns(commands),
		"execution_trends":    sc.analyzeExecutionTrends(commands),
		"failure_analysis":    sc.analyzeFailures(commands),
		"performance_metrics": sc.analyzePerformance(commands),
		"usage_predictions":   sc.generateUsagePredictions(commands),
		"recommendations":     sc.generateRecommendations(commands),
		// New deep analytics features
		"alias_suggestions":  deepAnalytics.AliasSuggestions,
		"sequence_patterns":  deepAnalytics.SequencePatterns,
		"workflow_insights":  deepAnalytics.WorkflowInsights,
		"command_complexity": deepAnalytics.CommandComplexity,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(analytics); err != nil {
		sc.logger.Error("Failed to encode analytics", zap.Error(err))
	}
}

func (sc *ServerCommand) handlePlugins(w http.ResponseWriter, r *http.Request) {
	// Plugin system placeholder
	plugins := map[string]interface{}{
		"available": []map[string]interface{}{
			{
				"name":        "Docker Integration",
				"version":     "1.0.0",
				"description": "Execute commands in Docker containers",
				"enabled":     false,
			},
			{
				"name":        "Slack Notifications",
				"version":     "1.0.0",
				"description": "Send notifications to Slack",
				"enabled":     false,
			},
			{
				"name":        "Git Integration",
				"version":     "1.0.0",
				"description": "Git repository integration",
				"enabled":     true,
			},
		},
		"installed": []string{"Git Integration"},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(plugins); err != nil {
		sc.logger.Error("Failed to encode plugins", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (sc *ServerCommand) handleSmartSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter required", http.StatusBadRequest)
		return
	}

	// Enhanced search with ML-like features
	commands, err := sc.repository.GetAllCommands()
	if err != nil {
		http.Error(w, "Failed to search commands", http.StatusInternalServerError)
		return
	}

	results := sc.performSmartSearch(commands, query)

	response := map[string]interface{}{
		"query":       query,
		"results":     results,
		"suggestions": sc.generateSearchSuggestions(query, commands),
		"total":       len(results),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		sc.logger.Error("Failed to encode smart search response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (sc *ServerCommand) handleWebApp(w http.ResponseWriter, r *http.Request) {
	// Serve the React web app
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Ambros Dashboard</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif; margin: 0; padding: 0; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 2rem; border-radius: 8px; margin-bottom: 2rem; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin-bottom: 2rem; }
        .stat-card { background: white; padding: 1.5rem; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .stat-value { font-size: 2rem; font-weight: bold; color: #667eea; }
        .stat-label { color: #666; font-size: 0.9rem; }
        .content { background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .nav { display: flex; gap: 1rem; margin-bottom: 2rem; }
        .nav-item { padding: 0.5rem 1rem; background: #667eea; color: white; border-radius: 4px; text-decoration: none; }
        .loading { text-align: center; padding: 2rem; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎯 Ambros Dashboard</h1>
            <p>Command Management & Analytics Platform</p>
        </div>
        
        <div class="nav">
            <a href="#dashboard" class="nav-item" onclick="loadSection('dashboard')">📊 Dashboard</a>
            <a href="#commands" class="nav-item" onclick="loadSection('commands')">💻 Commands</a>
            <a href="#analytics" class="nav-item" onclick="loadSection('analytics')">📈 Analytics</a>
        </div>
        
        <div id="main-content" class="content">
            <div class="loading">Loading dashboard...</div>
        </div>
    </div>

	<script src="/static/app.js"></script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		sc.logger.Error("Failed to write HTML response", zap.Error(err))
	}
}

func (sc *ServerCommand) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (sc *ServerCommand) displayStartupInfo() {
	color.Cyan("╔══════════════════════════════════════════════════════════════╗")
	color.Cyan("║                    🎯 AMBROS DASHBOARD                       ║")
	color.Cyan("╠══════════════════════════════════════════════════════════════╣")
	color.Cyan("║  Features:                                                  ║")
	color.Cyan("║  • 📱 Web Dashboard Interface                                ║")
	color.Cyan("║  • 🔍 Smart Search & Analytics                               ║")
	color.Cyan("║  • 📊 Command History & Statistics                           ║")
	color.Cyan("║  • 🔗 API Endpoints                                         ║")
	color.Cyan("╚══════════════════════════════════════════════════════════════╝")
}

func (sc *ServerCommand) handleShutdown(server *http.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	color.Yellow("\n🛑 Shutdown signal received...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		sc.logger.Error("Server shutdown error", zap.Error(err))
	} else {
		color.Green("✅ Server shutdown completed")
	}
}

// Helper functions for analytics
func (sc *ServerCommand) getMostUsedCommands(commands []models.Command, days int) []string {
	// Return empty slice if no commands
	if len(commands) == 0 {
		return []string{}
	}

	// Compute cutoff time
	cutoff := time.Now().AddDate(0, 0, -days)

	freq := make(map[string]int)
	for _, c := range commands {
		if !c.CreatedAt.After(cutoff) {
			continue
		}
		name := c.Name
		if name == "" && len(c.Arguments) > 0 {
			name = c.Arguments[0]
		}
		if name == "" {
			continue
		}
		freq[name]++
	}

	type kv struct {
		k string
		v int
	}
	var pairs []kv
	for k, v := range freq {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].v == pairs[j].v {
			return pairs[i].k < pairs[j].k
		}
		return pairs[i].v > pairs[j].v
	})

	limit := 10
	if len(pairs) < limit {
		limit = len(pairs)
	}

	res := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		res = append(res, fmt.Sprintf("%s (%d)", pairs[i].k, pairs[i].v))
	}
	return res
}

func (sc *ServerCommand) calculateAverageExecutionTime(commands []models.Command) time.Duration {
	if len(commands) == 0 {
		return 0
	}

	total := time.Duration(0)
	for _, cmd := range commands {
		duration := cmd.TerminatedAt.Sub(cmd.CreatedAt)
		total += duration
	}

	return total / time.Duration(len(commands))
}

// Typed analytics outputs
type CommandPatterns struct {
	MostCommon []string `json:"most_common"`
	Patterns   []string `json:"patterns"`
}

type ExecutionTrends struct {
	Trend     string         `json:"trend"`
	PeakHours []int          `json:"peak_hours"`
	ByDay     map[string]int `json:"by_day"`
}

type FailureAnalysis struct {
	TotalFailures int      `json:"total_failures"`
	FailureRate   float64  `json:"failure_rate"`
	CommonCauses  []string `json:"common_causes"`
}

type SlowCommand struct {
	Command  string `json:"command"`
	Duration string `json:"duration"`
}

type PerformanceMetrics struct {
	AvgDuration     string        `json:"avg_duration"`
	SlowestCommands []SlowCommand `json:"slowest_commands"`
}

type UsagePredictions struct {
	PredictedPeak    string   `json:"predicted_peak"`
	TrendingCommands []string `json:"trending_commands"`
}

// AliasSuggestion represents a suggested shell alias for a frequently used command
type AliasSuggestion struct {
	Alias       string `json:"alias"`
	Command     string `json:"command"`
	FullCommand string `json:"full_command"`
	UsageCount  int    `json:"usage_count"`
	Reason      string `json:"reason"`
}

// CommandSequencePattern represents a detected sequence of commands often run together
type CommandSequencePattern struct {
	Sequence    []string `json:"sequence"`
	Occurrences int      `json:"occurrences"`
	AvgInterval string   `json:"avg_interval"`
	Suggestion  string   `json:"suggestion"`
}

// WorkflowInsight represents an identified workflow pattern with automation opportunities
type WorkflowInsight struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Commands    []string `json:"commands"`
	Frequency   int      `json:"frequency"`
	Suggestion  string   `json:"suggestion"`
}

// DeepAnalytics contains advanced analytics with alias suggestions and pattern recognition
type DeepAnalytics struct {
	AliasSuggestions  []AliasSuggestion        `json:"alias_suggestions"`
	SequencePatterns  []CommandSequencePattern `json:"sequence_patterns"`
	WorkflowInsights  []WorkflowInsight        `json:"workflow_insights"`
	CommandComplexity map[string]int           `json:"command_complexity"`
}

func (sc *ServerCommand) analyzeCommandPatterns(commands []models.Command) CommandPatterns {
	// Count by command name (fallback to first argument)
	freq := make(map[string]int)
	for _, c := range commands {
		name := c.Name
		if name == "" && len(c.Arguments) > 0 {
			name = c.Arguments[0]
		}
		if name == "" {
			continue
		}
		freq[name]++
	}

	type kv struct {
		k string
		v int
	}
	var pairs []kv
	for k, v := range freq {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].v == pairs[j].v {
			return pairs[i].k < pairs[j].k
		}
		return pairs[i].v > pairs[j].v
	})

	top := make([]string, 0, min(len(pairs), 10))
	for i := 0; i < min(len(pairs), 10); i++ {
		top = append(top, pairs[i].k)
	}

	patternsSet := make(map[string]struct{})
	for name := range freq {
		lower := strings.ToLower(name)
		switch {
		case strings.HasPrefix(lower, "git") || strings.HasPrefix(lower, "gh"):
			patternsSet["version control"] = struct{}{}
		case strings.HasPrefix(lower, "ls") || strings.HasPrefix(lower, "cp") || strings.HasPrefix(lower, "mv") || strings.HasPrefix(lower, "rm"):
			patternsSet["file operations"] = struct{}{}
		case strings.HasPrefix(lower, "docker") || strings.HasPrefix(lower, "kubectl"):
			patternsSet["container/orchestration"] = struct{}{}
		case strings.HasPrefix(lower, "npm") || strings.HasPrefix(lower, "yarn"):
			patternsSet["package management"] = struct{}{}
		default:
			patternsSet["system/other"] = struct{}{}
		}
	}
	patterns := make([]string, 0, len(patternsSet))
	for p := range patternsSet {
		patterns = append(patterns, p)
	}

	return CommandPatterns{MostCommon: top, Patterns: patterns}
}

func (sc *ServerCommand) analyzeExecutionTrends(commands []models.Command) ExecutionTrends {
	if len(commands) == 0 {
		return ExecutionTrends{Trend: "stable", PeakHours: []int{}, ByDay: map[string]int{}}
	}

	now := time.Now()
	dayCounts := make(map[string]int)
	hourCounts := make(map[int]int)
	for _, c := range commands {
		day := c.CreatedAt.Format("2006-01-02")
		dayCounts[day]++
		hourCounts[c.CreatedAt.Hour()]++
	}

	days := 14
	var daySeq []int
	for i := days - 1; i >= 0; i-- {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		daySeq = append(daySeq, dayCounts[d])
	}
	firstHalf := 0
	secondHalf := 0
	for i, v := range daySeq {
		if i < len(daySeq)/2 {
			firstHalf += v
		} else {
			secondHalf += v
		}
	}
	trend := "stable"
	if secondHalf > firstHalf {
		trend = "increasing"
	}
	if secondHalf < firstHalf {
		trend = "decreasing"
	}

	type hv struct {
		h int
		v int
	}
	var hvs []hv
	for h, v := range hourCounts {
		hvs = append(hvs, hv{h, v})
	}
	sort.Slice(hvs, func(i, j int) bool {
		if hvs[i].v == hvs[j].v {
			return hvs[i].h < hvs[j].h
		}
		return hvs[i].v > hvs[j].v
	})
	peak := make([]int, 0, min(3, len(hvs)))
	for i := 0; i < min(3, len(hvs)); i++ {
		peak = append(peak, hvs[i].h)
	}

	return ExecutionTrends{Trend: trend, PeakHours: peak, ByDay: dayCounts}
}

func (sc *ServerCommand) analyzeFailures(commands []models.Command) FailureAnalysis {
	failures := 0
	for _, cmd := range commands {
		if !cmd.Status {
			failures++
		}
	}
	total := len(commands)
	failureRate := 0.0
	if total > 0 {
		failureRate = float64(failures) / float64(total) * 100.0
	}

	causes := make(map[string]int)
	for _, cmd := range commands {
		if cmd.Status {
			continue
		}
		txt := strings.ToLower(cmd.Command)
		if strings.Contains(txt, "permission denied") || strings.Contains(txt, "permission") {
			causes["permission denied"]++
		} else if strings.Contains(txt, "not found") || strings.Contains(txt, "command not found") {
			causes["command not found"]++
		} else if strings.Contains(txt, "timeout") {
			causes["timeout"]++
		} else {
			causes["other"]++
		}
	}
	type kvs struct {
		k string
		v int
	}
	var pairs []kvs
	for k, v := range causes {
		pairs = append(pairs, kvs{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].v == pairs[j].v {
			return pairs[i].k < pairs[j].k
		}
		return pairs[i].v > pairs[j].v
	})
	common := make([]string, 0, len(pairs))
	for _, p := range pairs {
		common = append(common, p.k)
	}

	return FailureAnalysis{TotalFailures: failures, FailureRate: failureRate, CommonCauses: common}
}

func (sc *ServerCommand) analyzePerformance(commands []models.Command) PerformanceMetrics {
	type sv struct {
		cmd string
		d   time.Duration
	}
	var slow []sv
	for _, c := range commands {
		d := c.TerminatedAt.Sub(c.CreatedAt)
		slow = append(slow, sv{c.Command, d})
	}
	sort.Slice(slow, func(i, j int) bool { return slow[i].d > slow[j].d })
	slowest := make([]SlowCommand, 0, min(5, len(slow)))
	for i := 0; i < min(5, len(slow)); i++ {
		slowest = append(slowest, SlowCommand{Command: slow[i].cmd, Duration: slow[i].d.String()})
	}
	return PerformanceMetrics{AvgDuration: sc.calculateAverageExecutionTime(commands).String(), SlowestCommands: slowest}
}

func (sc *ServerCommand) generateUsagePredictions(commands []models.Command) UsagePredictions {
	now := time.Now()
	cutoffRecent := now.AddDate(0, 0, -7)
	cutoffPrev := now.AddDate(0, 0, -14)

	recent := make(map[string]int)
	prev := make(map[string]int)
	for _, c := range commands {
		name := c.Name
		if name == "" && len(c.Arguments) > 0 {
			name = c.Arguments[0]
		}
		if name == "" {
			continue
		}
		if c.CreatedAt.After(cutoffRecent) {
			recent[name]++
		} else if c.CreatedAt.After(cutoffPrev) {
			prev[name]++
		}
	}
	type gv struct {
		k string
		d int
	}
	var growth []gv
	for k, rv := range recent {
		pv := prev[k]
		growth = append(growth, gv{k, rv - pv})
	}
	sort.Slice(growth, func(i, j int) bool {
		if growth[i].d == growth[j].d {
			return growth[i].k < growth[j].k
		}
		return growth[i].d > growth[j].d
	})
	trending := make([]string, 0, min(5, len(growth)))
	for i := 0; i < min(5, len(growth)); i++ {
		trending = append(trending, growth[i].k)
	}

	hourCounts := make(map[int]int)
	for _, c := range commands {
		hourCounts[c.CreatedAt.Hour()]++
	}
	peakHour := 0
	peakVal := 0
	for h, v := range hourCounts {
		if v > peakVal {
			peakVal = v
			peakHour = h
		}
	}

	return UsagePredictions{PredictedPeak: fmt.Sprintf("%02d:00", peakHour), TrendingCommands: trending}
}

func (sc *ServerCommand) generateRecommendations(commands []models.Command) []string {
	// heuristic-based recommendations from historical commands
	// Count top command usage
	counts := make(map[string]int)
	for _, c := range commands {
		cmdName := strings.Fields(c.Command)[0]
		counts[cmdName]++
	}

	// find top 3 commands
	type kv struct {
		k string
		v int
	}
	top := make([]kv, 0, len(counts))
	for k, v := range counts {
		top = append(top, kv{k, v})
	}
	sort.Slice(top, func(i, j int) bool { return top[i].v > top[j].v })

	suggestions := []string{}
	if len(top) > 0 {
		if top[0].v >= 3 { // frequent command threshold
			suggestions = append(suggestions, fmt.Sprintf("Consider creating a template for frequently used command: %s (used %d times)", top[0].k, top[0].v))
		}

		// detect common consecutive command pairs (chain candidates)
		pairCounts := make(map[string]int)
		for i := 1; i < len(commands); i++ {
			prev := strings.Fields(commands[i-1].Command)
			curr := strings.Fields(commands[i].Command)
			if len(prev) == 0 || len(curr) == 0 {
				continue
			}
			pair := prev[0] + " -> " + curr[0]
			pairCounts[pair]++
		}

		type kvp struct {
			k string
			v int
		}
		pairs := make([]kvp, 0, len(pairCounts))
		for k, v := range pairCounts {
			pairs = append(pairs, kvp{k, v})
		}
		sort.Slice(pairs, func(i, j int) bool { return pairs[i].v > pairs[j].v })
		if len(pairs) > 0 && pairs[0].v >= 2 {
			suggestions = append([]string{fmt.Sprintf("Consider creating a chain for frequently occurring sequence: %s (seen %d times)", pairs[0].k, pairs[0].v)}, suggestions...)
		}

		// detect frequently failing commands
		failCounts := make(map[string]int)
		for _, c := range commands {
			if !c.Status {
				base := strings.Fields(c.Command)[0]
				failCounts[base]++
			}
		}
		for k, v := range failCounts {
			if v >= 2 {
				suggestions = append(suggestions, fmt.Sprintf("Investigate frequent failures for command '%s' (failed %d times)", k, v))
			}
		}
		if len(top) > 1 && top[1].v >= 3 {
			suggestions = append(suggestions, fmt.Sprintf("Consider creating a template for frequently used command: %s (used %d times)", top[1].k, top[1].v))
		}
	}

	// generic recommendations
	suggestions = append(suggestions, "Schedule regular backup commands during off-peak hours")
	suggestions = append(suggestions, "Use environment variables for API endpoints")

	if len(suggestions) == 0 {
		// fallback hints
		suggestions = append(suggestions, "Review frequent commands and convert repeating flows to chains or templates")
	}

	return suggestions
}

// generateDeepAnalytics performs advanced analytics including alias suggestions, sequence patterns, and workflow insights
func (sc *ServerCommand) generateDeepAnalytics(commands []models.Command) DeepAnalytics {
	return DeepAnalytics{
		AliasSuggestions:  sc.generateAliasSuggestions(commands),
		SequencePatterns:  sc.analyzeSequencePatterns(commands),
		WorkflowInsights:  sc.identifyWorkflowInsights(commands),
		CommandComplexity: sc.analyzeCommandComplexity(commands),
	}
}

// generateAliasSuggestions suggests shell aliases for frequently used long commands
func (sc *ServerCommand) generateAliasSuggestions(commands []models.Command) []AliasSuggestion {
	suggestions := []AliasSuggestion{}

	// Track full command strings and their counts
	commandCounts := make(map[string]int)
	for _, cmd := range commands {
		fullCmd := cmd.Command
		if fullCmd == "" {
			fullCmd = cmd.Name
			if len(cmd.Arguments) > 0 {
				fullCmd = fullCmd + " " + strings.Join(cmd.Arguments, " ")
			}
		}
		if fullCmd != "" {
			commandCounts[fullCmd]++
		}
	}

	// Find long commands that are used frequently
	type cmdCount struct {
		cmd   string
		count int
	}
	var sortedCmds []cmdCount
	for cmd, count := range commandCounts {
		// Only consider commands longer than 15 chars and used at least 3 times
		if len(cmd) > 15 && count >= 3 {
			sortedCmds = append(sortedCmds, cmdCount{cmd, count})
		}
	}

	// Sort by count descending
	sort.Slice(sortedCmds, func(i, j int) bool {
		return sortedCmds[i].count > sortedCmds[j].count
	})

	// Generate alias suggestions for top commands
	for i := 0; i < min(5, len(sortedCmds)); i++ {
		cmd := sortedCmds[i]
		alias := sc.generateAliasName(cmd.cmd)
		suggestions = append(suggestions, AliasSuggestion{
			Alias:       alias,
			Command:     strings.Fields(cmd.cmd)[0],
			FullCommand: cmd.cmd,
			UsageCount:  cmd.count,
			Reason:      fmt.Sprintf("Used %d times, saves %d characters per invocation", cmd.count, len(cmd.cmd)-len(alias)),
		})
	}

	// Also suggest aliases for complex flag combinations
	flagPatterns := sc.detectComplexFlagPatterns(commands)
	for _, fp := range flagPatterns {
		suggestions = append(suggestions, fp)
	}

	return suggestions
}

// generateAliasName generates a suggested alias name from a command
func (sc *ServerCommand) generateAliasName(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "alias"
	}

	baseName := parts[0]
	// Remove common prefixes/paths
	if idx := strings.LastIndex(baseName, "/"); idx >= 0 {
		baseName = baseName[idx+1:]
	}

	// Extract first letter of each significant part
	alias := ""
	for _, part := range parts {
		if len(part) == 0 || strings.HasPrefix(part, "-") || strings.HasPrefix(part, "/") {
			continue
		}
		// Skip common words
		lower := strings.ToLower(part)
		if lower == "the" || lower == "a" || lower == "an" || lower == "to" || lower == "for" {
			continue
		}
		alias += string(part[0])
		if len(alias) >= 4 {
			break
		}
	}

	if alias == "" {
		alias = baseName[:min(3, len(baseName))]
	}

	return strings.ToLower(alias)
}

// detectComplexFlagPatterns finds commands with frequently used flag combinations
func (sc *ServerCommand) detectComplexFlagPatterns(commands []models.Command) []AliasSuggestion {
	suggestions := []AliasSuggestion{}

	// Track base command + flags patterns
	patternCounts := make(map[string]struct {
		fullCmd string
		count   int
	})

	for _, cmd := range commands {
		fullCmd := cmd.Command
		if fullCmd == "" {
			fullCmd = cmd.Name
			if len(cmd.Arguments) > 0 {
				fullCmd = fullCmd + " " + strings.Join(cmd.Arguments, " ")
			}
		}

		parts := strings.Fields(fullCmd)
		if len(parts) < 2 {
			continue
		}

		// Extract base command and flags
		base := parts[0]
		var flags []string
		for _, p := range parts[1:] {
			if strings.HasPrefix(p, "-") {
				flags = append(flags, p)
			}
		}

		if len(flags) >= 2 {
			pattern := base + " " + strings.Join(flags, " ")
			if existing, ok := patternCounts[pattern]; ok {
				patternCounts[pattern] = struct {
					fullCmd string
					count   int
				}{existing.fullCmd, existing.count + 1}
			} else {
				patternCounts[pattern] = struct {
					fullCmd string
					count   int
				}{fullCmd, 1}
			}
		}
	}

	// Find patterns used at least 3 times
	for pattern, data := range patternCounts {
		if data.count >= 3 && len(pattern) > 20 {
			parts := strings.Fields(pattern)
			alias := parts[0][:min(2, len(parts[0]))]
			for _, p := range parts[1:] {
				if strings.HasPrefix(p, "--") && len(p) > 2 {
					alias += string(p[2])
				} else if strings.HasPrefix(p, "-") && len(p) > 1 {
					alias += string(p[1])
				}
			}
			suggestions = append(suggestions, AliasSuggestion{
				Alias:       strings.ToLower(alias),
				Command:     parts[0],
				FullCommand: pattern,
				UsageCount:  data.count,
				Reason:      fmt.Sprintf("Complex flag pattern used %d times", data.count),
			})
		}
	}

	return suggestions
}

// analyzeSequencePatterns detects sequences of commands that are often run together
func (sc *ServerCommand) analyzeSequencePatterns(commands []models.Command) []CommandSequencePattern {
	patterns := []CommandSequencePattern{}

	if len(commands) < 2 {
		return patterns
	}

	// Sort commands by creation time
	sortedCmds := make([]models.Command, len(commands))
	copy(sortedCmds, commands)
	sort.Slice(sortedCmds, func(i, j int) bool {
		return sortedCmds[i].CreatedAt.Before(sortedCmds[j].CreatedAt)
	})

	// Detect 2-command sequences
	pairCounts := make(map[string]struct {
		intervals []time.Duration
		count     int
	})

	for i := 1; i < len(sortedCmds); i++ {
		prev := sc.getCommandBase(sortedCmds[i-1])
		curr := sc.getCommandBase(sortedCmds[i])
		if prev == "" || curr == "" {
			continue
		}

		// Only count if within 5 minutes of each other
		interval := sortedCmds[i].CreatedAt.Sub(sortedCmds[i-1].CreatedAt)
		if interval > 5*time.Minute {
			continue
		}

		pair := prev + " → " + curr
		if existing, ok := pairCounts[pair]; ok {
			pairCounts[pair] = struct {
				intervals []time.Duration
				count     int
			}{append(existing.intervals, interval), existing.count + 1}
		} else {
			pairCounts[pair] = struct {
				intervals []time.Duration
				count     int
			}{[]time.Duration{interval}, 1}
		}
	}

	// Detect 3-command sequences
	tripleCounts := make(map[string]int)
	for i := 2; i < len(sortedCmds); i++ {
		c1 := sc.getCommandBase(sortedCmds[i-2])
		c2 := sc.getCommandBase(sortedCmds[i-1])
		c3 := sc.getCommandBase(sortedCmds[i])
		if c1 == "" || c2 == "" || c3 == "" {
			continue
		}

		// Only count if all three within 5 minutes
		if sortedCmds[i].CreatedAt.Sub(sortedCmds[i-2].CreatedAt) > 5*time.Minute {
			continue
		}

		triple := c1 + " → " + c2 + " → " + c3
		tripleCounts[triple]++
	}

	// Convert pairs to patterns (threshold: 3 occurrences)
	for pair, data := range pairCounts {
		if data.count >= 3 {
			avgInterval := time.Duration(0)
			for _, d := range data.intervals {
				avgInterval += d
			}
			avgInterval /= time.Duration(len(data.intervals))

			parts := strings.Split(pair, " → ")
			patterns = append(patterns, CommandSequencePattern{
				Sequence:    parts,
				Occurrences: data.count,
				AvgInterval: avgInterval.Round(time.Second).String(),
				Suggestion:  fmt.Sprintf("Consider creating a chain command for '%s' and '%s'", parts[0], parts[1]),
			})
		}
	}

	// Convert triples to patterns (threshold: 2 occurrences)
	for triple, count := range tripleCounts {
		if count >= 2 {
			parts := strings.Split(triple, " → ")
			patterns = append(patterns, CommandSequencePattern{
				Sequence:    parts,
				Occurrences: count,
				AvgInterval: "< 5m",
				Suggestion:  "Consider creating a workflow for this 3-step sequence",
			})
		}
	}

	// Sort by occurrences descending
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Occurrences > patterns[j].Occurrences
	})

	// Return top 10 patterns
	if len(patterns) > 10 {
		patterns = patterns[:10]
	}

	return patterns
}

// getCommandBase extracts the base command name
func (sc *ServerCommand) getCommandBase(cmd models.Command) string {
	fullCmd := cmd.Command
	if fullCmd == "" {
		fullCmd = cmd.Name
	}
	parts := strings.Fields(fullCmd)
	if len(parts) == 0 {
		return ""
	}
	base := parts[0]
	// Remove path prefix
	if idx := strings.LastIndex(base, "/"); idx >= 0 {
		base = base[idx+1:]
	}
	return base
}

// identifyWorkflowInsights identifies common workflow patterns
func (sc *ServerCommand) identifyWorkflowInsights(commands []models.Command) []WorkflowInsight {
	insights := []WorkflowInsight{}

	// Count command usage by base name
	cmdCounts := make(map[string]int)
	for _, cmd := range commands {
		base := sc.getCommandBase(cmd)
		if base != "" {
			cmdCounts[base]++
		}
	}

	// Predefined workflow patterns to detect
	workflowPatterns := []struct {
		name     string
		desc     string
		commands []string
		minCount int
	}{
		{
			name:     "Git Development Flow",
			desc:     "Standard git development workflow detected",
			commands: []string{"git", "make", "go"},
			minCount: 5,
		},
		{
			name:     "Docker Development",
			desc:     "Container-based development workflow",
			commands: []string{"docker", "docker-compose"},
			minCount: 3,
		},
		{
			name:     "Kubernetes Operations",
			desc:     "Kubernetes cluster management workflow",
			commands: []string{"kubectl", "helm"},
			minCount: 3,
		},
		{
			name:     "Node.js Development",
			desc:     "Node.js/npm-based development workflow",
			commands: []string{"npm", "node", "yarn"},
			minCount: 3,
		},
		{
			name:     "Python Development",
			desc:     "Python development workflow",
			commands: []string{"python", "pip", "pytest"},
			minCount: 3,
		},
		{
			name:     "Go Development",
			desc:     "Go development workflow",
			commands: []string{"go", "make"},
			minCount: 3,
		},
		{
			name:     "CI/CD Pipeline",
			desc:     "Continuous integration/deployment commands",
			commands: []string{"git", "docker", "kubectl"},
			minCount: 3,
		},
	}

	// Check each workflow pattern
	for _, pattern := range workflowPatterns {
		matchCount := 0
		totalUsage := 0
		matchedCmds := []string{}

		for _, cmd := range pattern.commands {
			if count, ok := cmdCounts[cmd]; ok && count >= pattern.minCount {
				matchCount++
				totalUsage += count
				matchedCmds = append(matchedCmds, cmd)
			}
		}

		// If at least 2 commands from the pattern are frequently used
		if matchCount >= 2 {
			suggestion := "Consider creating command chains to automate common sequences"
			if pattern.name == "Git Development Flow" {
				suggestion = "Try: `ambros chain create dev-flow 'git pull' 'make test' 'git push'`"
			} else if pattern.name == "Docker Development" {
				suggestion = "Try: `ambros chain create docker-dev 'docker-compose build' 'docker-compose up'`"
			} else if pattern.name == "CI/CD Pipeline" {
				suggestion = "Consider setting up scheduled tasks for deployment automation"
			}

			insights = append(insights, WorkflowInsight{
				Name:        pattern.name,
				Description: pattern.desc,
				Commands:    matchedCmds,
				Frequency:   totalUsage,
				Suggestion:  suggestion,
			})
		}
	}

	// Sort by frequency
	sort.Slice(insights, func(i, j int) bool {
		return insights[i].Frequency > insights[j].Frequency
	})

	return insights
}

// analyzeCommandComplexity measures the complexity of commands (args, pipes, redirects)
func (sc *ServerCommand) analyzeCommandComplexity(commands []models.Command) map[string]int {
	complexity := make(map[string]int)

	for _, cmd := range commands {
		fullCmd := cmd.Command
		if fullCmd == "" {
			fullCmd = cmd.Name
			if len(cmd.Arguments) > 0 {
				fullCmd = fullCmd + " " + strings.Join(cmd.Arguments, " ")
			}
		}

		score := sc.calculateComplexityScore(fullCmd)
		base := sc.getCommandBase(cmd)
		if base != "" {
			// Keep track of max complexity for each base command
			if current, ok := complexity[base]; !ok || score > current {
				complexity[base] = score
			}
		}
	}

	return complexity
}

// calculateComplexityScore calculates a complexity score for a command
func (sc *ServerCommand) calculateComplexityScore(cmd string) int {
	score := 0

	// Base score from length
	score += len(cmd) / 20

	// Count arguments
	parts := strings.Fields(cmd)
	score += len(parts) - 1

	// Count flags
	for _, p := range parts {
		if strings.HasPrefix(p, "--") {
			score += 2
		} else if strings.HasPrefix(p, "-") {
			score++
		}
	}

	// Pipes add complexity
	score += strings.Count(cmd, "|") * 3

	// Redirects add complexity
	score += strings.Count(cmd, ">") * 2
	score += strings.Count(cmd, "<") * 2

	// Subshells/command substitution
	score += strings.Count(cmd, "$(") * 3
	score += strings.Count(cmd, "`") * 3

	// Logical operators
	score += strings.Count(cmd, "&&") * 2
	score += strings.Count(cmd, "||") * 2

	return score
}

// trigramTokens returns the set of trigrams for a string
func trigramTokens(s string) map[string]struct{} {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "_")
	tokens := make(map[string]struct{})
	r := []rune(s)
	if len(r) < 3 {
		if len(r) > 0 {
			tokens[string(r)] = struct{}{}
		}
		return tokens
	}
	for i := 0; i < len(r)-2; i++ {
		t := string(r[i : i+3])
		tokens[t] = struct{}{}
	}
	return tokens
}

// jaccardSimilarity computes the Jaccard index between two strings using trigram tokens
func jaccardSimilarity(a, b string) float64 {
	ta := trigramTokens(a)
	tb := trigramTokens(b)
	if len(ta) == 0 && len(tb) == 0 {
		return 1.0
	}
	inter := 0
	for k := range ta {
		if _, ok := tb[k]; ok {
			inter++
		}
	}
	union := len(ta) + len(tb) - inter
	if union == 0 {
		return 0.0
	}
	return float64(inter) / float64(union)
}

func (sc *ServerCommand) performSmartSearch(commands []models.Command, query string) []models.Command {
	var results []models.Command
	query = strings.ToLower(query)

	for _, cmd := range commands {
		// Simple matching - can be enhanced with fuzzy search, NLP, etc.
		if strings.Contains(strings.ToLower(cmd.Command), query) ||
			strings.Contains(strings.ToLower(cmd.Name), query) {
			results = append(results, cmd)
		}
	}

	return results
}

func (sc *ServerCommand) generateSearchSuggestions(query string, commands []models.Command) []string {
	// build simple suggestions based on historical commands
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return []string{"Add a search term to get suggestions, e.g., 'git' or 'docker'"}
	}

	matches := make(map[string]int)
	candidates := []string{}
	candidateMap := make(map[string]string)
	for _, c := range commands {
		nameLower := strings.ToLower(c.Name)
		cmdLower := strings.ToLower(c.Command)
		// map candidate to a human label
		if nameLower != "" {
			candidates = append(candidates, nameLower)
			candidateMap[nameLower] = c.Name
		}
		if cmdLower != "" {
			candidates = append(candidates, cmdLower)
			candidateMap[cmdLower] = c.Name
		}
	}

	// run fuzzy matching on candidates using sahilm/fuzzy (Find returns ordered matches)
	if len(candidates) > 0 {
		fr := fuzzy.Find(q, candidates)
		if len(fr) > 0 {
			// Add top fuzzy results by translating back to original command names
			for i := 0; i < min(3, len(fr)); i++ {
				cand := candidates[fr[i].Index]
				matches[candidateMap[cand]]++
			}
		}
	}

	// exact and jaccard matches after fuzzy
	for _, c := range commands {
		nameLower := strings.ToLower(c.Name)
		cmdLower := strings.ToLower(c.Command)
		// exact contains
		if strings.Contains(cmdLower, q) || strings.Contains(nameLower, q) {
			matches[c.Name]++
			continue
		}
		// fuzzy match via Jaccard similarity
		if jaccardSimilarity(q, cmdLower) >= 0.25 || jaccardSimilarity(q, nameLower) >= 0.25 {
			matches[c.Name]++
		}
	}

	// get best match
	type kv struct {
		k string
		v int
	}
	list := make([]kv, 0, len(matches))
	for k, v := range matches {
		list = append(list, kv{k, v})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].v > list[j].v })

	suggestions := []string{}
	if len(list) > 0 {
		suggestions = append(suggestions, "Did you mean: "+list[0].k+"?")
	}

	suggestions = append(suggestions, "Try searching for: git, docker, npm")
	suggestions = append(suggestions, "Use quotes for exact matches")
	return suggestions
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the smaller of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (sc *ServerCommand) Command() *cobra.Command {
	return sc.cmd
}
