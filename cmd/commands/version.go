package commands

import (
	_ "embed"
	"fmt"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/plugins" // New import
)

//go:embed version.txt
var embeddedVersion string

// Version information - these can be set at build time
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
	GoVersion = runtime.Version()
)

// getVersion returns the actual version, preferring build-time injection over embedded
func getVersion() string {
	if Version != "dev" && Version != "" {
		return Version
	}
	return strings.TrimSpace(embeddedVersion)
}

// VersionCommand represents the version command
type VersionCommand struct {
	*BaseCommand
	short bool
}

// NewVersionCommand creates a new version command
func NewVersionCommand(logger *zap.Logger, api plugins.CoreAPI) *VersionCommand {
	vc := &VersionCommand{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display version information including build details and runtime information.`,
		RunE:  vc.runE,
	}

	// Version command doesn't need repository, so pass nil for repo
	vc.BaseCommand = NewBaseCommand(cmd, logger, nil, api)
	vc.cmd = cmd
	vc.setupFlags(cmd)
	return vc
}

func (vc *VersionCommand) setupFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&vc.short, "short", "s", false, "Show only version number")
}

func (vc *VersionCommand) runE(cmd *cobra.Command, args []string) error {
	actualVersion := getVersion()

	if vc.short {
		fmt.Println(actualVersion)
		return nil
	}

	fmt.Printf("ambros version %s\n", actualVersion)
	fmt.Printf("Git commit: %s\n", GitCommit)
	fmt.Printf("Build date: %s\n", BuildDate)
	fmt.Printf("Go version: %s\n", GoVersion)
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	vc.logger.Debug("Version information displayed",
		zap.String("version", actualVersion),
		zap.String("gitCommit", GitCommit),
		zap.String("buildDate", BuildDate),
		zap.String("goVersion", GoVersion),
	)

	return nil
}

func (vc *VersionCommand) Command() *cobra.Command {
	return vc.cmd
}
