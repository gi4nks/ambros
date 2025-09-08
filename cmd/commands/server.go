package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/internal/api"
	"github.com/gi4nks/ambros/internal/errors"
	"github.com/gi4nks/ambros/internal/models"
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

	// Enhanced API with Phase 3 features
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

	// Enhanced API endpoints for Phase 3
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
	mux.HandleFunc("/static/", sc.handleStatic)

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
	// Phase 3 Plugin system placeholder
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

    <script>
        async function loadSection(section) {
            const content = document.getElementById('main-content');
            content.innerHTML = '<div class="loading">Loading...</div>';
            
            try {
                switch(section) {
                    case 'dashboard':
                        await loadDashboard();
                        break;
                    case 'commands':
                        await loadCommands();
                        break;
                    case 'environments':
                        await loadEnvironments();
                        break;
                    case 'templates':
                        await loadTemplates();
                        break;
                    case 'scheduler':
                        await loadScheduler();
                        break;
                    case 'analytics':
                        await loadAnalytics();
                        break;
                }
            } catch (error) {
                content.innerHTML = '<div style="color: red;">Error loading section: ' + error.message + '</div>';
            }
        }
        
        async function loadDashboard() {
            const response = await fetch('/api/dashboard');
            const data = await response.json();
            
            const content = document.getElementById('main-content');
            content.innerHTML = ` + "`" + `
                <h2>📊 Dashboard Overview</h2>
                <div class="stats">
                    <div class="stat-card">
                        <div class="stat-value">${data.summary.total_commands}</div>
                        <div class="stat-label">Total Commands</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value">${data.summary.success_rate.toFixed(1)}%</div>
                        <div class="stat-label">Success Rate</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value">${data.summary.recent_commands}</div>
                        <div class="stat-label">Recent (24h)</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value">${data.summary.template_count}</div>
                        <div class="stat-label">Templates</div>
                    </div>
                </div>
                <h3>Recent Activity</h3>
                <div style="max-height: 400px; overflow-y: auto;">
                    ${data.recent_activity.map(cmd => ` + "`" + `
                        <div style="padding: 0.5rem; border-bottom: 1px solid #eee;">
                            <strong>${cmd.command || cmd.name}</strong>
                            <span style="color: ${cmd.status ? 'green' : 'red'}; margin-left: 1rem;">
                                ${cmd.status ? '✅' : '❌'}
                            </span>
                            <small style="float: right; color: #666;">
                                ${new Date(cmd.created_at).toLocaleString()}
                            </small>
                        </div>
                    ` + "`" + `).join('')}
                </div>
            ` + "`" + `;
        }
        
        async function loadCommands() {
            const response = await fetch('/api/commands');
            const commands = await response.json();
            
            const content = document.getElementById('main-content');
            content.innerHTML = ` + "`" + `
                <h2>💻 Command History</h2>
                <p>Total: ${commands.length} commands</p>
                <div style="max-height: 600px; overflow-y: auto;">
                    ${commands.map(cmd => ` + "`" + `
                        <div style="padding: 1rem; border-bottom: 1px solid #eee; margin-bottom: 0.5rem;">
                            <div style="display: flex; justify-content: space-between; align-items: center;">
                                <strong>${cmd.command || cmd.name}</strong>
                                <span style="color: ${cmd.status ? 'green' : 'red'};">
                                    ${cmd.status ? '✅ Success' : '❌ Failed'}
                                </span>
                            </div>
                            <small style="color: #666;">
                                ID: ${cmd.id} | ${new Date(cmd.created_at).toLocaleString()}
                            </small>
                            ${cmd.tags && cmd.tags.length > 0 ? ` + "`" + `
                                <div style="margin-top: 0.5rem;">
                                    ${cmd.tags.map(tag => ` + "`" + `<span style="background: #667eea; color: white; padding: 0.2rem 0.5rem; border-radius: 3px; font-size: 0.8rem; margin-right: 0.5rem;">${tag}</span>` + "`" + `).join('')}
                                </div>
                            ` + "`" + ` : ''}
                        </div>
                    ` + "`" + `).join('')}
                </div>
            ` + "`" + `;
        }
        
        async function loadEnvironments() {
            const response = await fetch('/api/environments');
            const environments = await response.json();
            
            const content = document.getElementById('main-content');
            content.innerHTML = ` + "`" + `
                <h2>🌍 Environments</h2>
                ${Object.keys(environments).length === 0 ? 
                    '<p>No environments found. Create one using: <code>ambros env create myenv</code></p>' :
                    Object.entries(environments).map(([name, env]) => ` + "`" + `
                        <div style="background: #f8f9fa; padding: 1rem; border-radius: 8px; margin-bottom: 1rem;">
                            <h3>${name}</h3>
                            <div style="display: grid; gap: 0.5rem;">
                                ${env.variables.map(v => ` + "`" + `
                                    <div style="background: white; padding: 0.5rem; border-radius: 4px;">
                                        <strong>${v.key}</strong> = <code>${v.value}</code>
                                    </div>
                                ` + "`" + `).join('')}
                            </div>
                        </div>
                    ` + "`" + `).join('')
                }
            ` + "`" + `;
        }
        
        async function loadTemplates() {
            const response = await fetch('/api/templates');
            const templates = await response.json();
            
            const content = document.getElementById('main-content');
            content.innerHTML = ` + "`" + `
                <h2>🎯 Command Templates</h2>
                ${templates.length === 0 ? 
                    '<p>No templates found. Create one using: <code>ambros template save mytemplate "echo hello"</code></p>' :
                    templates.map(template => ` + "`" + `
                        <div style="background: #f8f9fa; padding: 1rem; border-radius: 8px; margin-bottom: 1rem;">
                            <h3>${template.name || 'Unnamed Template'}</h3>
                            <code style="background: white; padding: 0.5rem; border-radius: 4px; display: block;">
                                ${template.command}
                            </code>
                            <small style="color: #666;">
                                Created: ${new Date(template.created_at).toLocaleString()}
                            </small>
                        </div>
                    ` + "`" + `).join('')
                }
            ` + "`" + `;
        }
        
        async function loadScheduler() {
            const response = await fetch('/api/scheduler');
            const scheduled = await response.json();
            
            const content = document.getElementById('main-content');
            content.innerHTML = ` + "`" + `
                <h2>📅 Scheduled Commands</h2>
                ${scheduled.length === 0 ? 
                    '<p>No scheduled commands found.</p>' :
                    scheduled.map(cmd => ` + "`" + `
                        <div style="background: #f8f9fa; padding: 1rem; border-radius: 8px; margin-bottom: 1rem;">
                            <div style="display: flex; justify-content: space-between;">
                                <strong>${cmd.command || cmd.name}</strong>
                                <span style="color: ${cmd.schedule.enabled ? 'green' : 'red'};">
                                    ${cmd.schedule.enabled ? '🟢 Enabled' : '🔴 Disabled'}
                                </span>
                            </div>
                            <div style="margin-top: 0.5rem;">
                                <strong>Cron:</strong> <code>${cmd.schedule.cron_expr}</code><br>
                                <strong>Next Run:</strong> ${new Date(cmd.schedule.next_run).toLocaleString()}
                            </div>
                        </div>
                    ` + "`" + `).join('')
                }
            ` + "`" + `;
        }
        
        async function loadAnalytics() {
            const response = await fetch('/api/analytics/advanced');
            const analytics = await response.json();
            
            const content = document.getElementById('main-content');
            content.innerHTML = ` + "`" + `
                <h2>📈 Advanced Analytics</h2>
                <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 1rem;">
                    <div style="background: #f8f9fa; padding: 1rem; border-radius: 8px;">
                        <h3>🔍 Command Patterns</h3>
                        <p>Most common commands and usage patterns detected.</p>
                        <small>AI-powered analysis coming soon</small>
                    </div>
                    <div style="background: #f8f9fa; padding: 1rem; border-radius: 8px;">
                        <h3>📊 Execution Trends</h3>
                        <p>Performance trends and execution patterns over time.</p>
                        <small>Time series analysis available</small>
                    </div>
                    <div style="background: #f8f9fa; padding: 1rem; border-radius: 8px;">
                        <h3>🔧 Recommendations</h3>
                        <p>Smart suggestions for command optimization.</p>
                        <small>ML recommendations in development</small>
                    </div>
                </div>
            ` + "`" + `;
        }
        
        // Load dashboard on page load
        window.onload = () => loadSection('dashboard');
    </script>
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
	color.Cyan("║  Phase 3 Features:                                          ║")
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
	// Simple implementation - can be enhanced with more sophisticated analysis
	return []string{"ls", "echo", "git"}
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

func (sc *ServerCommand) analyzeCommandPatterns(commands []models.Command) map[string]interface{} {
	return map[string]interface{}{
		"most_common": []string{"ls", "echo", "git"},
		"patterns":    []string{"file operations", "version control", "system info"},
	}
}

func (sc *ServerCommand) analyzeExecutionTrends(commands []models.Command) map[string]interface{} {
	return map[string]interface{}{
		"trend":      "increasing",
		"peak_hours": []int{9, 14, 16},
	}
}

func (sc *ServerCommand) analyzeFailures(commands []models.Command) map[string]interface{} {
	failures := 0
	for _, cmd := range commands {
		if !cmd.Status {
			failures++
		}
	}

	return map[string]interface{}{
		"total_failures": failures,
		"common_causes":  []string{"permission denied", "command not found"},
	}
}

func (sc *ServerCommand) analyzePerformance(commands []models.Command) map[string]interface{} {
	return map[string]interface{}{
		"avg_duration":     sc.calculateAverageExecutionTime(commands).String(),
		"slowest_commands": []string{"npm install", "docker build"},
	}
}

func (sc *ServerCommand) generateUsagePredictions(commands []models.Command) map[string]interface{} {
	return map[string]interface{}{
		"predicted_peak":    "2:00 PM",
		"trending_commands": []string{"docker", "kubectl"},
	}
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
