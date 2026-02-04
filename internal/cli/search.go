package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"lazyas/internal/config"
	"lazyas/internal/manifest"
	"lazyas/internal/registry"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for skills by name, description, or tags",
	Long: `Search for skills in the registry by name, description, or tags.

Examples:
  lazyas search ros
  lazyas search robotics
  lazyas search cli`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func runSearch(cmd *cobra.Command, args []string) error {
	cfg, err := config.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	query := args[0]

	// Load manifest
	mfst := manifest.NewManager(cfg)
	if err := mfst.Load(); err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Fetch registry
	fmt.Println("Searching...")

	reg := registry.NewRegistry(cfg)
	if err := reg.Fetch(false); err != nil {
		return fmt.Errorf("failed to fetch index: %w", err)
	}

	// Search
	results := reg.SearchSkills(query)
	if len(results) == 0 {
		fmt.Printf("No skills matching '%s'\n", query)
		return nil
	}

	fmt.Printf("Found %d skill(s) matching '%s':\n\n", len(results), query)

	for _, skill := range results {
		var status string
		if mfst.IsInstalled(skill.Name) {
			status = "● "
		} else {
			status = "○ "
		}

		version := skill.Source.Tag
		if version == "" {
			version = "latest"
		}

		fmt.Printf("%s%s@%s\n", status, skill.Name, version)
		if skill.Description != "" {
			fmt.Printf("    %s\n", skill.Description)
		}
		if len(skill.Tags) > 0 {
			fmt.Printf("    tags: %v\n", skill.Tags)
		}
		fmt.Println()
	}

	return nil
}
