package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"lazyas/internal/config"
	"lazyas/internal/tui"
)

var browseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Launch the interactive TUI browser",
	Long:  `Browse available and installed skills using an interactive terminal UI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.DefaultConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		return tui.Run(cfg)
	},
}
