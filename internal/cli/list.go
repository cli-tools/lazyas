package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"lazyas/internal/config"
	"lazyas/internal/manifest"
	"lazyas/internal/registry"
)

var (
	listAvailable bool
	listAll       bool
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List skills",
	Long: `List installed skills or all available skills.

Examples:
  lazyas list              # List installed skills
  lazyas list --available  # List available skills from registry
  lazyas list --all        # List all skills with install status`,
	RunE: runList,
}

func init() {
	listCmd.Flags().BoolVarP(&listAvailable, "available", "a", false, "List available skills from registry")
	listCmd.Flags().BoolVar(&listAll, "all", false, "List all skills with install status")
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Load manifest
	mfst := manifest.NewManager(cfg)
	if err := mfst.Load(); err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	if listAvailable || listAll {
		return listFromRegistry(cfg, mfst, listAll)
	}

	return listInstalled(mfst)
}

func listInstalled(mfst *manifest.Manager) error {
	installed := mfst.ListInstalled()

	if len(installed) == 0 {
		fmt.Println("No skills installed")
		fmt.Println("\nUse 'lazyas browse' or 'lazyas list --available' to see available skills")
		return nil
	}

	// Sort by name
	names := make([]string, 0, len(installed))
	for name := range installed {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Println("Installed skills:")
	fmt.Println()

	for _, name := range names {
		info := installed[name]
		version := info.Version
		if version == "" {
			version = "latest"
		}
		fmt.Printf("  ● %s@%s\n", name, version)
		if info.Commit != "" {
			fmt.Printf("    commit: %s\n", truncateString(info.Commit, 7))
		}
	}

	return nil
}

func listFromRegistry(cfg *config.Config, mfst *manifest.Manager, showStatus bool) error {
	fmt.Println("Fetching skill index...")

	reg := registry.NewRegistry(cfg)
	if err := reg.Fetch(false); err != nil {
		return fmt.Errorf("failed to fetch index: %w", err)
	}

	skills := reg.ListSkills()
	if len(skills) == 0 {
		fmt.Println("No skills available in registry")
		return nil
	}

	if showStatus {
		fmt.Println("Skills (● installed, ○ available):")
	} else {
		fmt.Println("Available skills:")
	}
	fmt.Println()

	for _, skill := range skills {
		var status string
		if showStatus {
			if mfst.IsInstalled(skill.Name) {
				status = "● "
			} else {
				status = "○ "
			}
		} else {
			status = "  "
		}

		version := skill.Source.Tag
		if version == "" {
			version = "latest"
		}

		fmt.Printf("%s%s@%s\n", status, skill.Name, version)
		if skill.Description != "" {
			desc := skill.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Printf("    %s\n", desc)
		}
	}

	return nil
}

func truncateString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
