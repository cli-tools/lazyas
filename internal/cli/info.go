package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"lazyas/internal/config"
	"lazyas/internal/manifest"
	"lazyas/internal/registry"
)

var infoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show details about a skill",
	Long: `Show detailed information about a skill.

Examples:
  lazyas info my-skill`,
	Args: cobra.ExactArgs(1),
	RunE: runInfo,
}

func runInfo(cmd *cobra.Command, args []string) error {
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

	// Fetch registry
	reg := registry.NewRegistry(cfg)
	if err := reg.Fetch(false); err != nil {
		// Continue anyway, might have local info
	}

	skill := reg.GetSkill(name)
	installed, isInstalled := mfst.GetInstalled(name)

	if skill == nil && !isInstalled {
		return fmt.Errorf("skill %s not found", name)
	}

	// Display info
	fmt.Printf("Name: %s\n", name)

	if skill != nil {
		if skill.Description != "" {
			fmt.Printf("Description: %s\n", skill.Description)
		}
		if skill.Author != "" {
			fmt.Printf("Author: %s\n", skill.Author)
		}
		fmt.Printf("Repository: %s\n", skill.Source.Repo)
		if skill.Source.Path != "" {
			fmt.Printf("Path: %s\n", skill.Source.Path)
		}
		version := skill.Source.Tag
		if version == "" {
			version = "latest"
		}
		fmt.Printf("Version: %s\n", version)
		if len(skill.Tags) > 0 {
			fmt.Printf("Tags: %v\n", skill.Tags)
		}
	}

	fmt.Println()
	if isInstalled {
		fmt.Println("Status: INSTALLED")
		fmt.Printf("  Installed version: %s\n", installed.Version)
		fmt.Printf("  Commit: %s\n", installed.Commit)
		fmt.Printf("  Installed at: %s\n", installed.InstalledAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Location: %s\n", mfst.GetSkillPath(name))
	} else {
		fmt.Println("Status: Not installed")
		fmt.Printf("\nInstall with: lazyas install %s\n", name)
	}

	return nil
}
