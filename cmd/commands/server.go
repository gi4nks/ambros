package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/api"
	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/models"
)

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
  - Scheduling management
  - Export/import functionality

Examples:
  ambros server                          # Start server on default port 8080
  ambros server --port 3000              # Start on custom port
  ambros server --host 0.0.0.0          # Listen on all interfaces
  ambros server --dev                   # Development mode with hot reload
  ambros server --cors                  # Enable CORS for development`,
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
	mux.HandleFunc("/api/environments", sc.handleEnvironments)
	mux.HandleFunc("/api/templates", sc.handleTemplates)
	mux.HandleFunc("/api/scheduler", sc.handleScheduler)
	mux.HandleFunc("/api/chains", sc.handleChains)
	mux.HandleFunc("/api/plugins", sc.handlePlugins)
	mux.HandleFunc("/api/search/smart", sc.handleSmartSearch)

	// Static files for web dashboard
	mux.HandleFunc("/", sc.handleWebApp)
	// Serve files under ./web/static at the /static/ URL path
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

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
	templateCount := 0
	scheduledCount := 0

	now := time.Now()
	last24h := now.Add(-24 * time.Hour)

	for _, cmd := range commands {
		if cmd.Status {
			successCount++
		}
		if cmd.CreatedAt.After(last24h) {
			recentCommands++
		}
		for _, tag := range cmd.Tags {
			if tag == "template" {
				templateCount++
				break
			}
		}
		if cmd.Schedule != nil {
			scheduledCount++
		}
	}

	dashboard := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_commands":  totalCommands,
			"success_rate":    float64(successCount) / float64(totalCommands) * 100,
			"recent_commands": recentCommands,
			"template_count":  templateCount,
			"scheduled_count": scheduledCount,
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

	analytics := map[string]interface{}{
		"command_patterns":    sc.analyzeCommandPatterns(commands),
		"execution_trends":    sc.analyzeExecutionTrends(commands),
		"failure_analysis":    sc.analyzeFailures(commands),
		"performance_metrics": sc.analyzePerformance(commands),
		"usage_predictions":   sc.generateUsagePredictions(commands),
		"recommendations":     sc.generateRecommendations(commands),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(analytics); err != nil {
		sc.logger.Error("Failed to encode analytics", zap.Error(err))
	}
}

func (sc *ServerCommand) handleEnvironments(w http.ResponseWriter, r *http.Request) {
	// Handle environment API requests
	switch r.Method {
	case http.MethodGet:
		envs, err := sc.repository.SearchByTag("environment")
		if err != nil {
			http.Error(w, "Failed to get environments", http.StatusInternalServerError)
			return
		}

		// Group by environment name
		envMap := make(map[string]interface{})
		for _, env := range envs {
			if env.Category == "environment" {
				envName := sc.extractEnvName(env.Name)
				if envName != "" {
					if _, exists := envMap[envName]; !exists {
						envMap[envName] = map[string]interface{}{
							"name":       envName,
							"variables":  []map[string]string{},
							"created_at": env.CreatedAt,
						}
					}

					if env.Variables != nil {
						if key, exists := env.Variables["var_key"]; exists {
							value := env.Variables["var_value"]
							envData := envMap[envName].(map[string]interface{})
							vars := envData["variables"].([]map[string]string)
							vars = append(vars, map[string]string{
								"key":   key,
								"value": value,
							})
							envData["variables"] = vars
						}
					}
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(envMap); err != nil {
			sc.logger.Error("Failed to encode environment map", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (sc *ServerCommand) handleTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := sc.repository.SearchByTag("template")
	if err != nil {
		http.Error(w, "Failed to get templates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(templates); err != nil {
		sc.logger.Error("Failed to encode templates", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (sc *ServerCommand) handleScheduler(w http.ResponseWriter, r *http.Request) {
	commands, err := sc.repository.GetAllCommands()
	if err != nil {
		http.Error(w, "Failed to get commands", http.StatusInternalServerError)
		return
	}

	var scheduled []models.Command
	for _, cmd := range commands {
		if cmd.Schedule != nil {
			scheduled = append(scheduled, cmd)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(scheduled); err != nil {
		sc.logger.Error("Failed to encode scheduled commands", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (sc *ServerCommand) handleChains(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement chain storage and retrieval
	chains := []map[string]interface{}{
		{
			"id":          "chain-1",
			"name":        "Deployment Chain",
			"description": "Build, test, and deploy",
			"commands":    []string{"build", "test", "deploy"},
			"created_at":  time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(chains); err != nil {
		sc.logger.Error("Failed to encode chains", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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
            <a href="#environments" class="nav-item" onclick="loadSection('environments')">🌍 Environments</a>
            <a href="#templates" class="nav-item" onclick="loadSection('templates')">🎯 Templates</a>
            <a href="#scheduler" class="nav-item" onclick="loadSection('scheduler')">📅 Scheduler</a>
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

func (sc *ServerCommand) handleStatic(w http.ResponseWriter, r *http.Request) {
	// Serve static files (CSS, JS, images)
	http.NotFound(w, r)
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
	color.Cyan("║  • 🌍 Environment Management                                 ║")
	color.Cyan("║  • 🎯 Template Management                                    ║")
	color.Cyan("║  • 📅 Scheduler Management                                   ║")
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
	return []string{
		"Consider creating a template for frequently used Git commands",
		"Schedule regular backup commands during off-peak hours",
		"Use environment variables for API endpoints",
	}
}

func (sc *ServerCommand) extractEnvName(cmdName string) string {
	parts := strings.Split(cmdName, ":")
	if len(parts) >= 2 && parts[0] == "env" {
		return parts[1]
	}
	return ""
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
	return []string{
		"Did you mean: " + query + "?",
		"Try searching for: git, docker, npm",
		"Use quotes for exact matches",
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (sc *ServerCommand) Command() *cobra.Command {
	return sc.cmd
}
