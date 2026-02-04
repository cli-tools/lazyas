package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"lazyas/internal/config"
	"lazyas/internal/git"
	"lazyas/internal/manifest"
	"lazyas/internal/registry"
)

var installForce bool

var installCmd = &cobra.Command{
	Use:   "install <name>[@version]",
	Short: "Install a skill from the registry",
	Long: `Install a skill from the registry.

If the skill already exists and has local modifications, you'll be
prompted to confirm overwrite. Use --force to skip confirmation.

Examples:
  lazyas install my-skill
  lazyas install my-skill@v1.2.0
  lazyas install --force my-skill`,
	Args: cobra.ExactArgs(1),
	RunE: runInstall,
}

func init() {
	installCmd.Flags().BoolVarP(&installForce, "force", "f", false, "Force install, overwriting local modifications")
}

func runInstall(cmd *cobra.Command, args []string) error {
	cfg, err := config.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Parse name@version
	name, version := parseSkillArg(args[0])

	// Load manifest
	mfst := manifest.NewManager(cfg)
	if err := mfst.Load(); err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Check if already installed
	if mfst.IsInstalled(name) {
		// Check for local modifications
		skillPath := mfst.GetSkillPath(name)
		modified, _ := git.IsModified(skillPath)
		if modified && !installForce {
			fmt.Printf("Skill %s has local modifications.\n", name)
			modFiles, _ := git.GetModifiedFiles(skillPath)
			if len(modFiles) > 0 {
				fmt.Println("Modified files:")
				for _, f := range modFiles {
					fmt.Printf("  %s\n", f)
				}
			}
			fmt.Printf("Overwrite? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Cancelled")
				return nil
			}
		} else if !modified && !installForce {
			return fmt.Errorf("skill %s is already installed (use 'lazyas update' to update)", name)
		}

		// Remove existing to reinstall
		os.RemoveAll(skillPath)
	}

	// Fetch registry
	fmt.Println("Fetching skill index...")
	reg := registry.NewRegistry(cfg)
	if err := reg.Fetch(false); err != nil {
		return fmt.Errorf("failed to fetch index: %w", err)
	}

	// Find skill
	skill := reg.GetSkill(name)
	if skill == nil {
		return fmt.Errorf("skill %s not found in registry", name)
	}

	// Use specified version or default
	skillVersion := skill.Source.Tag
	if version != "" {
		skillVersion = version
	}

	fmt.Printf("Installing %s", name)
	if skillVersion != "" {
		fmt.Printf("@%s", skillVersion)
	}
	fmt.Println("...")

	// Clone skill
	targetDir := mfst.GetSkillPath(name)
	result, err := git.Clone(git.CloneOptions{
		Repo:      skill.Source.Repo,
		Path:      skill.Source.Path,
		Tag:       skillVersion,
		TargetDir: targetDir,
	})
	if err != nil {
		return fmt.Errorf("failed to clone skill: %w", err)
	}

	// Validate skill
	if err := git.ValidateSkill(targetDir); err != nil {
		os.RemoveAll(targetDir)
		return fmt.Errorf("skill validation failed: %w", err)
	}

	// Update manifest
	if err := mfst.AddSkill(
		name,
		skillVersion,
		result.Commit,
		skill.Source.Repo,
		skill.Source.Path,
	); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	fmt.Printf("Successfully installed %s\n", name)
	return nil
}

func parseSkillArg(arg string) (name, version string) {
	parts := strings.SplitN(arg, "@", 2)
	name = parts[0]
	if len(parts) > 1 {
		version = parts[1]
	}
	return
}
