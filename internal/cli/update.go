package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"lazyas/internal/config"
	"lazyas/internal/git"
	"lazyas/internal/manifest"
	"lazyas/internal/registry"
)

var (
	updateDryRun bool
	updateForce  bool
)

var updateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update installed skill(s)",
	Long: `Update one or all installed skills to their latest versions.

Skills with local modifications are skipped unless --force is used.
Use --dry-run to preview what would be updated.

Examples:
  lazyas update                # Update all skills
  lazyas update my-skill    # Update specific skill
  lazyas update --dry-run      # Preview updates
  lazyas update --force        # Update even modified skills`,
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().BoolVar(&updateDryRun, "dry-run", false, "Preview updates without making changes")
	updateCmd.Flags().BoolVarP(&updateForce, "force", "f", false, "Update even skills with local modifications")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	cfg, err := config.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Load manifest
	mfst := manifest.NewManager(cfg)
	if err := mfst.Load(); err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	installed := mfst.ListInstalled()
	if len(installed) == 0 {
		fmt.Println("No skills installed")
		return nil
	}

	// Fetch registry for version info
	fmt.Println("Fetching skill index...")
	reg := registry.NewRegistry(cfg)
	if err := reg.Fetch(true); err != nil {
		return fmt.Errorf("failed to fetch index: %w", err)
	}

	// Determine which skills to update
	var toUpdate []string
	if len(args) > 0 {
		name := args[0]
		if !mfst.IsInstalled(name) {
			return fmt.Errorf("skill %s is not installed", name)
		}
		toUpdate = []string{name}
	} else {
		for name := range installed {
			toUpdate = append(toUpdate, name)
		}
	}

	// Update each skill
	var updated, skipped, failed int
	for _, name := range toUpdate {
		info := installed[name]
		skill := reg.GetSkill(name)
		skillDir := mfst.GetSkillPath(name)

		// Check for local modifications
		modified, _ := git.IsModified(skillDir)
		if modified && !updateForce {
			if updateDryRun {
				fmt.Printf("  %s: has local changes (would skip)\n", name)
			} else {
				fmt.Printf("  %s: has local changes, skipping (use --force to overwrite)\n", name)
			}
			skipped++
			continue
		}

		// Determine target version
		targetTag := ""
		if skill != nil {
			targetTag = skill.Source.Tag
		}

		if updateDryRun {
			// Dry run mode - just show what would happen
			if skill == nil {
				fmt.Printf("  %s: not found in registry (would skip)\n", name)
				skipped++
				continue
			}

			currentVersion := info.Version
			if currentVersion == "" {
				currentVersion = "latest"
			}
			newVersion := targetTag
			if newVersion == "" {
				newVersion = "latest"
			}

			if modified {
				fmt.Printf("  %s: %s → %s (would force update)\n", name, currentVersion, newVersion)
			} else {
				fmt.Printf("  %s: %s → %s (would update)\n", name, currentVersion, newVersion)
			}
			updated++
			continue
		}

		fmt.Printf("Updating %s...\n", name)

		// If force and modified, reset changes first
		if modified && updateForce {
			fmt.Printf("  Discarding local changes...\n")
			if err := git.ResetChanges(skillDir); err != nil {
				fmt.Printf("  Failed to reset changes: %v\n", err)
				failed++
				continue
			}
		}

		result, err := git.Update(skillDir, targetTag)
		if err != nil {
			fmt.Printf("  Failed: %v\n", err)
			failed++
			continue
		}

		if result.Commit != info.Commit {
			sourceRepo := info.SourceRepo
			sourcePath := info.SourcePath
			if skill != nil {
				sourceRepo = skill.Source.Repo
				sourcePath = skill.Source.Path
			}
			mfst.AddSkill(name, targetTag, result.Commit, sourceRepo, sourcePath)
			fmt.Printf("  Updated to %s\n", truncateString(result.Commit, 7))
			updated++
		} else {
			fmt.Printf("  Already up to date\n")
			skipped++
		}
	}

	if updateDryRun {
		fmt.Printf("\nWould update: %d, Skip: %d\n", updated, skipped)
	} else {
		fmt.Printf("\nUpdated %d skill(s)", updated)
		if skipped > 0 {
			fmt.Printf(", %d skipped", skipped)
		}
		if failed > 0 {
			fmt.Printf(", %d failed", failed)
		}
		fmt.Println()
	}

	return nil
}
