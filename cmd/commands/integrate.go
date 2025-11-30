package commands

//go:generate sh -c "mkdir -p cmd/commands/scripts && cp scripts/.ambros-integration.sh cmd/commands/scripts/.ambros-integration.sh"

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/errors"
	"github.com/gi4nks/ambros/v3/internal/plugins" // New import
)

//go:embed scripts/.ambros-integration.sh
var embeddedFiles embed.FS

type IntegrateCommand struct {
	*BaseCommand
}

func NewIntegrateCommand(logger *zap.Logger, api plugins.CoreAPI) *IntegrateCommand {
	ic := &IntegrateCommand{}
	cmd := &cobra.Command{
		Use:   "integrate",
		Short: "Manage Ambros shell integration",
		Long:  "Install or uninstall the Ambros transparent shell integration script into your shell profiles.",
	}
	// Integrate command doesn't need repository, so pass nil for repo
	ic.BaseCommand = NewBaseCommand(cmd, logger, nil, api)
	ic.cmd = cmd
	return ic
}

func (c *IntegrateCommand) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integrate",
		Short: "Manage Ambros shell integration",
		Long:  "Install or uninstall the Ambros transparent shell integration script into your shell profiles.",
	}

	var shell string
	var yes bool

	install := &cobra.Command{
		Use:   "install",
		Short: "Install the Ambros integration script",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.install(cmd, args, shell, yes)
		},
	}
	install.Flags().StringVar(&shell, "shell", "", "Target shell rc file (e.g. ~/.zshrc). If empty, updates both ~/.bashrc and ~/.zshrc")
	install.Flags().BoolVar(&yes, "yes", false, "Assume yes for prompts and run non-interactively")

	uninstall := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the Ambros integration script",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.uninstall(cmd, args, shell, yes)
		},
	}
	uninstall.Flags().StringVar(&shell, "shell", "", "Target shell rc file (e.g. ~/.zshrc). If empty, updates both ~/.bashrc and ~/.zshrc")
	uninstall.Flags().BoolVar(&yes, "yes", false, "Assume yes for prompts and run non-interactively")

	cmd.AddCommand(install)
	cmd.AddCommand(uninstall)
	return cmd
}

func (c *IntegrateCommand) install(_ *cobra.Command, _ []string, shell string, yes bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	target := filepath.Join(home, ".ambros-integration.sh")
	content, err := embeddedFiles.ReadFile("scripts/.ambros-integration.sh")
	if err != nil {
		return err
	}

	// Idempotent write: only write if different
	existing, _ := os.ReadFile(target)
	if string(existing) == string(content) {
		c.Logger().Info("integration script already up-to-date", zap.String("target", target))
	} else {
		if err := os.WriteFile(target, content, 0755); err != nil {
			return err
		}
		fmt.Printf("Installed integration script to %s\n", target)
	}

	sourceLine := "source ~/.ambros-integration.sh"
	if shell != "" {
		rc := expandPath(shell)
		if !yes && !confirm(fmt.Sprintf("Add '%s' to %s?", sourceLine, rc)) {
			fmt.Println("skipping shell update")
			return nil
		}
		if err := addSourceIfMissing(rc, sourceLine); err != nil {
			return errors.NewError(errors.ErrInternalServer, fmt.Sprintf("failed to update %s", rc), err)
		}
		fmt.Printf("Updated %s\n", rc)
		return nil
	}

	// default: both bashrc and zshrc
	bashrc := filepath.Join(home, ".bashrc")
	zshrc := filepath.Join(home, ".zshrc")
	if !yes && !confirm(fmt.Sprintf("Add '%s' to %s and %s?", sourceLine, bashrc, zshrc)) {
		fmt.Println("skipping shell updates")
		return nil
	}
	if err := addSourceIfMissing(bashrc, sourceLine); err != nil {
		c.Logger().Warn("failed to update bashrc, continuing", zap.Error(err))
	}
	if err := addSourceIfMissing(zshrc, sourceLine); err != nil {
		c.logger.Warn("failed to update zshrc, continuing", zap.Error(err))
	}
	fmt.Printf("Updated %s and %s\n", bashrc, zshrc)
	return nil
}

func addSourceIfMissing(rcPath, line string) error {
	data, err := os.ReadFile(rcPath)
	if err != nil {
		if os.IsNotExist(err) {
			// file might not exist; create with the line
			return os.WriteFile(rcPath, []byte(line+"\n"), 0644)
		}
		return err
	}
	content := string(data)
	if !containsLine(content, line) {
		// Ensure file ends with a newline before appending
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += line + "\n"
		return os.WriteFile(rcPath, []byte(content), 0644)
	}
	return nil
}

func expandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, p[2:])
	}
	return p
}

func confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return false
	}
	resp := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return resp == "y" || resp == "yes"
}

func containsLine(content, line string) bool {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == strings.TrimSpace(line) {
			return true
		}
	}
	return false
}

func (c *IntegrateCommand) uninstall(_ *cobra.Command, _ []string, shell string, yes bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	target := filepath.Join(home, ".ambros-integration.sh")
	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		c.Logger().Warn("failed to remove integration script, continuing", zap.Error(err))
	}

	sourceLine := "source ~/.ambros-integration.sh"

	// If a specific shell rc file was provided, only operate on that
	if shell != "" {
		rc := expandPath(shell)
		if !yes && !confirm(fmt.Sprintf("Remove '%s' from %s?", sourceLine, rc)) {
			fmt.Println("skipping shell update")
			return nil
		}
		if err := removeSourceLine(rc, sourceLine); err != nil {
			return errors.NewError(errors.ErrInternalServer, fmt.Sprintf("failed to update %s", rc), err)
		}
		fmt.Printf("Updated %s\n", rc)
		fmt.Println("Uninstalled Ambros integration script")
		return nil
	}

	// default: both bashrc and zshrc
	bashrc := filepath.Join(home, ".bashrc")
	zshrc := filepath.Join(home, ".zshrc")
	if !yes && !confirm(fmt.Sprintf("Remove '%s' from %s and %s?", sourceLine, bashrc, zshrc)) {
		fmt.Println("skipping shell updates")
		return nil
	}
	if err := removeSourceLine(bashrc, sourceLine); err != nil {
		c.Logger().Warn("failed to update bashrc, continuing", zap.Error(err))
	}
	if err := removeSourceLine(zshrc, sourceLine); err != nil {
		c.logger.Warn("failed to update zshrc, continuing", zap.Error(err))
	}
	fmt.Printf("Updated %s and %s\n", bashrc, zshrc)
	fmt.Println("Uninstalled Ambros integration script")
	return nil
}

func removeSourceLine(rcPath, line string) error {
	data, err := os.ReadFile(rcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Nothing to do
		}
		return err
	}

	var out []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	var changed bool
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == strings.TrimSpace(line) {
			changed = true
			continue
		}
		out = append(out, scanner.Text())
	}

	if !changed {
		return nil // Line not found, nothing to do
	}

	return os.WriteFile(rcPath, []byte(strings.Join(out, "\n")+"\n"), 0644)
}
