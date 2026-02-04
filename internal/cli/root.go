package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"lazyas/internal/config"
	"lazyas/internal/symlink"
	"lazyas/internal/tui"
)

var rootCmd = &cobra.Command{
	Use:   "lazyas",
	Short: "Lazy Agent Skills manager",
	Long: `lazyas is a package manager for Agent Skills.

It allows you to browse, install, remove, and update skills from
a centralized registry. Skills extend AI agent capabilities
with specialized knowledge and workflows.

Supports multiple AI agent backends through symlinks to a
central skills directory at ~/.lazyas/skills/.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip backend check for backend subcommands (they handle it themselves)
		if cmd.Parent() != nil && cmd.Parent().Name() == "backend" {
			return
		}
		// Also skip for the backend command itself
		if cmd.Name() == "backend" {
			return
		}

		checkBackendLinks()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Launch browse as default action when no subcommand given
		cfg, err := config.DefaultConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		return tui.Run(cfg)
	},
}

// checkBackendLinks checks if any backends need linking and prints a hint
func checkBackendLinks() {
	cfg, err := config.DefaultConfig()
	if err != nil {
		return
	}

	statuses := symlink.CheckBackendLinks(cfg.Backends, cfg.SkillsDir)
	unlinked := symlink.GetUnlinkedBackends(statuses)
	if len(unlinked) > 0 {
		fmt.Printf("Hint: %d backend(s) not linked. Run 'lazyas backend link' to connect them.\n\n", len(unlinked))
	}
}

// SetVersion sets the version string for the CLI
func SetVersion(v string) {
	rootCmd.Version = v
}

// Execute runs the CLI
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(browseCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(backendCmd)
	rootCmd.AddCommand(syncCmd)
}
