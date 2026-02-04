package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"lazyas/internal/config"
	"lazyas/internal/registry"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Force refresh the registry from all repositories",
	Long: `Force refresh the registry index from all configured repositories,
bypassing the cache TTL.

This is useful when you want to see the latest available skills
without waiting for the cache to expire.

Examples:
  lazyas sync`,
	RunE: runSync,
}

func runSync(cmd *cobra.Command, args []string) error {
	cfg, err := config.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Syncing repositories...")

	reg := registry.NewRegistry(cfg)
	if err := reg.Fetch(true); err != nil {
		return fmt.Errorf("failed to sync: %w", err)
	}

	skills := reg.ListSkills()
	fmt.Printf("Synced. %d skill(s) available.\n", len(skills))
	return nil
}
