package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings" // Added for string manipulation
	"time"    // New import

	"github.com/spf13/viper" // Added for configuration management
	"go.uber.org/zap"
)

type Configuration struct {
	Logger              *zap.Logger
	RepositoryDirectory string
	RepositoryFile      string
	LastCountDefault    int
	DebugMode           bool
	Server              struct {
		Host string
		Port int
		Cors bool
	}
	Plugins struct {
		Directory   string
		AllowUnsafe bool
	}
	Scheduler struct {
		Enabled       bool
		CheckInterval string // e.g., "1m", "5s"
	}
	Analytics struct {
		Enabled       bool
		RetentionDays int
	}
}

func NewConfiguration(logger *zap.Logger) *Configuration {
	c := &Configuration{
		Logger: logger,
	}

	// Set sane defaults that can be overridden by config file or env vars
	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Failed to get home directory, using /tmp for default config dir", zap.Error(err))
		viper.SetDefault("repositoryDirectory", filepath.Join(os.TempDir(), ".ambros"))
		viper.SetDefault("plugins.directory", filepath.Join(os.TempDir(), ".ambros", "plugins"))
	} else {
		viper.SetDefault("repositoryDirectory", filepath.Join(home, ".ambros"))
		viper.SetDefault("plugins.directory", filepath.Join(home, ".ambros", "plugins"))
	}

	viper.SetDefault("repositoryFile", "ambros.db")
	viper.SetDefault("lastCountDefault", 10)
	viper.SetDefault("debugMode", false)
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.cors", true)
	viper.SetDefault("plugins.allowUnsafe", false)
	viper.SetDefault("scheduler.enabled", true)
	viper.SetDefault("scheduler.checkInterval", "1m")
	viper.SetDefault("analytics.enabled", true)
	viper.SetDefault("analytics.retentionDays", 365)

	// Environment variables
	viper.SetEnvPrefix("AMBROS")
	viper.AutomaticEnv()                                   // Read in ENV variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // AMBROS_SERVER_PORT instead of AMBROS_SERVER.PORT

	// Populate struct with defaults immediately so fields are usable before Load() is called
	_ = viper.Unmarshal(c)

	return c
}

// Load reads configuration from file and environment variables
func (c *Configuration) Load(cfgFile string) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile) // Use config file from the flag.
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			c.Logger.Warn("Cannot find home directory, will not search for default config file", zap.Error(err))
		} else {
			viper.AddConfigPath(filepath.Join(home, ".ambros")) // Path to look for the config file in
		}
		viper.AddConfigPath(".")       // Look for config in the current directory
		viper.SetConfigName(".ambros") // name of config file (without extension)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			c.Logger.Debug("No config file found, using defaults and environment variables", zap.Error(err))
		} else {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		c.Logger.Info("Using config file", zap.String("file", viper.ConfigFileUsed()))
	}

	// Unmarshal configuration into the struct
	if err := viper.Unmarshal(c); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Special handling for logger based on DebugMode, if not already set by root.go init
	if c.DebugMode {
		c.Logger.Debug("Debug mode enabled in configuration")
	}

	return nil
}

// Validate checks the loaded configuration for consistency and correctness
func (c *Configuration) Validate() error {
	if c.RepositoryDirectory == "" {
		return fmt.Errorf("repositoryDirectory cannot be empty")
	}
	if c.RepositoryFile == "" {
		return fmt.Errorf("repositoryFile cannot be empty")
	}
	if c.LastCountDefault <= 0 {
		return fmt.Errorf("lastCountDefault must be greater than 0")
	}
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be a valid port number (1-65535)")
	}
	if c.Plugins.Directory == "" {
		return fmt.Errorf("plugins.directory cannot be empty")
	}
	if c.Analytics.RetentionDays < 0 {
		return fmt.Errorf("analytics.retentionDays cannot be negative")
	}

	// Validate scheduler.checkInterval format (simple check)
	if _, err := time.ParseDuration(c.Scheduler.CheckInterval); err != nil {
		return fmt.Errorf("invalid scheduler.checkInterval format: %w", err)
	}

	// Ensure repository directory exists or can be created
	repoDir := filepath.Clean(c.RepositoryDirectory)
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		c.Logger.Warn("Repository directory does not exist, attempting to create", zap.String("path", repoDir))
		if err := os.MkdirAll(repoDir, 0755); err != nil {
			return fmt.Errorf("failed to create repository directory %s: %w", repoDir, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to access repository directory %s: %w", repoDir, err)
	}

	return nil
}

func (c *Configuration) RepositoryFullName() string {
	return filepath.Join(c.RepositoryDirectory, c.RepositoryFile)
}

func (c *Configuration) String() string {
	// Re-implement or remove this as the struct now has many fields.
	// For now, returning a basic string representation.
	return fmt.Sprintf("Repository: %s, DebugMode: %t, ServerPort: %d",
		c.RepositoryFullName(), c.DebugMode, c.Server.Port)
}
