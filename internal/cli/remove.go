package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"lazyas/internal/config"
	"lazyas/internal/manifest"
)

var (
	removeForce bool
)

var removeCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"rm", "uninstall"},
	Short:   "Remove an installed skill",
	Long: `Remove an installed skill from the local system.

Examples:
  lazyas remove my-skill
  lazyas rm my-skill`,
	Args: cobra.ExactArgs(1),
	RunE: runRemove,
}

func init() {
	removeCmd.Flags().BoolVarP(&removeForce, "force", "f", false, "Force removal without confirmation")
}

func runRemove(cmd *cobra.Command, args []string) error {
	cfg, err := config.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	name := args[0]

	// Load manifest
	mfst := manifest.NewManager(cfg)
	if err := mfst.Load(); err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Check if installed
	if !mfst.IsInstalled(name) {
		return fmt.Errorf("skill %s is not installed", name)
	}

	// Confirm unless forced
	if !removeForce {
		fmt.Printf("Remove skill %s? [y/N]: ", name)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	fmt.Printf("Removing %s...\n", name)

	// Remove directory
	skillDir := mfst.GetSkillPath(name)
	if err := os.RemoveAll(skillDir); err != nil {
		return fmt.Errorf("failed to remove skill directory: %w", err)
	}

	// Update manifest
	if err := mfst.RemoveSkill(name); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	fmt.Printf("Successfully removed %s\n", name)
	return nil
}
