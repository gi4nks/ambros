package commands

import (
	"os"

	"github.com/gi4nks/ambros/v3/internal/plugins"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// NewCompletionCommand creates a new completion command
func NewCompletionCommand(logger *zap.Logger, repo RepositoryInterface, api plugins.CoreAPI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

  $ source <(ambros completion bash)

  # To load completions for each session, add to your .bashrc:
  # [ -f /usr/local/bin/ambros ] && source <(ambros completion bash)

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ ambros completion zsh > "${fpath[1]}/_ambros"

Fish:

  $ ambros completion fish | source

  # To load completions for each session, execute once:
  $ ambros completion fish > ~/.config/fish/completions/ambros.fish

PowerShell:

  PS> ambros completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> ambros completion powershell > ambros.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}

	return cmd
}
