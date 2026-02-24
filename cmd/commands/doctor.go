package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fatih/color"
	"github.com/gi4nks/ambros/v3/internal/plugins"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// DoctorCommand represents the diagnostic command
type DoctorCommand struct {
	*BaseCommand
}

// NewDoctorCommand creates a new doctor command
func NewDoctorCommand(logger *zap.Logger, repo RepositoryInterface, api plugins.CoreAPI) *DoctorCommand {
	dc := &DoctorCommand{}

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check system health and diagnostics",
		Long:  `Run a suite of health checks to ensure Ambros is correctly configured and working properly.`,
		RunE:  dc.runE,
	}

	dc.BaseCommand = NewBaseCommand(cmd, logger, repo, api)
	dc.cmd = cmd
	return dc
}

func (dc *DoctorCommand) runE(cmd *cobra.Command, args []string) error {
	color.Cyan("🧪 Running health checks for Ambros...")

	home, _ := os.UserHomeDir()
	ambrosDir := filepath.Join(home, ".ambros")

	checks := []struct {
		name string
		f    func() (bool, string)
	}{
		{"System OS", func() (bool, string) {
			return true, fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH)
		}},
		{"Configuration Directory", func() (bool, string) {
			info, err := os.Stat(ambrosDir)
			if err != nil {
				return false, "Directory not found"
			}
			if !info.IsDir() {
				return false, "Is not a directory"
			}
			return true, "OK"
		}},
		{"Database Access", func() (bool, string) {
			if dc.repository == nil {
				return false, "Not initialized"
			}
			_, err := dc.repository.GetLimitCommands(1)
			if err != nil {
				return false, fmt.Sprintf("Failed to read: %v", err)
			}
			return true, "OK"
		}},
		{"Shell Integration", func() (bool, string) {
			scriptsDir := filepath.Join(ambrosDir, ".ambros-integration.sh")
			if _, err := os.Stat(scriptsDir); err != nil {
				return false, "Not installed (run 'ambros integrate install')"
			}
			return true, "Installed"
		}},
		{"Plugin Registry", func() (bool, string) {
			reg := plugins.GetGlobalRegistry()
			if reg == nil {
				return false, "Not found"
			}
			count := len(reg.GetAllPlugins())
			return true, fmt.Sprintf("%d plugins loaded", count)
		}},
	}

	allOk := true
	for _, check := range checks {
		ok, msg := check.f()
		if ok {
			color.Green("  [✓] %-25s : %s", check.name, msg)
		} else {
			color.Red("  [✗] %-25s : %s", check.name, msg)
			allOk = false
		}
	}

	if allOk {
		fmt.Println("\nEverything looks good!")
	} else {
		fmt.Println("\nSome checks failed. Use 'ambros logs' for more details.")
	}

	return nil
}

func (dc *DoctorCommand) Command() *cobra.Command {
	return dc.cmd
}
