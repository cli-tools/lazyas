package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"lazyas/internal/config"
	"lazyas/internal/symlink"
)

var backendCmd = &cobra.Command{
	Use:   "backend",
	Short: "Manage AI agent backends",
	Long: `Manage symlinks between lazyas central skills directory
and AI agent backend skill directories.

lazyas manages skills in ~/.lazyas/skills/ and symlinks
backend directories (e.g., ~/.claude/skills/) to it.`,
}

var backendListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List configured backends and their link status",
	Aliases: []string{"ls"},
	RunE:    runBackendList,
}

var backendLinkCmd = &cobra.Command{
	Use:   "link [name]",
	Short: "Create symlink for a backend (or all backends)",
	Long: `Create a symlink from a backend's skills directory to the
central lazyas skills directory.

If no backend name is given, links all unlinked backends.

If the backend directory already exists with files, lazyas will
offer to migrate them to the central directory.

Examples:
  lazyas backend link           # Link all unlinked backends
  lazyas backend link claude    # Link specific backend`,
	RunE: runBackendLink,
}

var backendUnlinkCmd = &cobra.Command{
	Use:   "unlink <name>",
	Short: "Remove symlink for a backend",
	Long: `Remove the symlink from a backend's skills directory.
This does not delete any skills.

Examples:
  lazyas backend unlink claude`,
	Args: cobra.ExactArgs(1),
	RunE: runBackendUnlink,
}

var backendAddCmd = &cobra.Command{
	Use:   "add <name> <path>",
	Short: "Add a custom backend",
	Long: `Add a custom AI agent backend.

Examples:
  lazyas backend add myai ~/.myai/skills
  lazyas backend add work-tool ~/work/.ai/skills --description "Internal AI tool"`,
	Args: cobra.ExactArgs(2),
	RunE: runBackendAdd,
}

var backendRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a custom backend",
	Long: `Remove a custom backend from configuration.
Built-in backends (claude, codex) cannot be removed.

Examples:
  lazyas backend remove myai`,
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
	RunE:    runBackendRemove,
}

var backendDescription string

func init() {
	backendAddCmd.Flags().StringVar(&backendDescription, "description", "", "Human-readable description for the backend")

	backendCmd.AddCommand(backendListCmd)
	backendCmd.AddCommand(backendLinkCmd)
	backendCmd.AddCommand(backendUnlinkCmd)
	backendCmd.AddCommand(backendAddCmd)
	backendCmd.AddCommand(backendRemoveCmd)
}

func runBackendList(cmd *cobra.Command, args []string) error {
	cfg, err := config.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	statuses := symlink.CheckBackendLinks(cfg.Backends, cfg.SkillsDir)

	if len(statuses) == 0 {
		fmt.Println("No backends configured.")
		return nil
	}

	fmt.Println("Backends:")
	for _, s := range statuses {
		expandedPath, _ := config.ExpandPath(s.Backend.Path)
		status := "○ not linked"
		if s.Linked {
			status = "✓ linked"
		} else if s.HasFiles {
			status = "○ has files (run 'lazyas backend link' to migrate)"
		} else if s.Error != nil {
			status = fmt.Sprintf("✗ error: %v", s.Error)
		}

		desc := s.Backend.Description
		if desc == "" {
			desc = s.Backend.Name
		}

		fmt.Printf("  %-12s %-30s %s\n", s.Backend.Name, expandedPath, status)
		if s.Backend.Description != "" {
			fmt.Printf("  %-12s %s\n", "", s.Backend.Description)
		}
	}

	return nil
}

func runBackendLink(cmd *cobra.Command, args []string) error {
	cfg, err := config.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Ensure central directory exists
	if err := cfg.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	statuses := symlink.CheckBackendLinks(cfg.Backends, cfg.SkillsDir)

	var toLink []symlink.LinkStatus
	if len(args) > 0 {
		// Link specific backend
		name := args[0]
		found := false
		for _, s := range statuses {
			if s.Backend.Name == name {
				found = true
				if s.Linked {
					fmt.Printf("Backend '%s' is already linked.\n", name)
					return nil
				}
				toLink = append(toLink, s)
				break
			}
		}
		if !found {
			return fmt.Errorf("backend '%s' not found. Use 'lazyas backend list' to see configured backends", name)
		}
	} else {
		// Link all unlinked backends
		toLink = symlink.GetUnlinkedBackends(statuses)
		if len(toLink) == 0 {
			fmt.Println("All backends are already linked.")
			return nil
		}
	}

	for _, s := range toLink {
		expandedPath, _ := config.ExpandPath(s.Backend.Path)

		if s.Exists && s.HasFiles && !s.IsSymlink {
			// Directory exists with files - offer to migrate
			fmt.Printf("Backend '%s': %s exists with files.\n", s.Backend.Name, expandedPath)
			fmt.Printf("Move files to %s and create symlink? [y/N]: ", cfg.SkillsDir)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Printf("Skipping '%s'.\n", s.Backend.Name)
				continue
			}

			if err := symlink.MigrateExistingDir(s.Backend, cfg.SkillsDir); err != nil {
				fmt.Printf("Failed to migrate '%s': %v\n", s.Backend.Name, err)
				continue
			}
			fmt.Printf("Migrated and linked '%s' ✓\n", s.Backend.Name)
		} else if s.Exists && !s.IsSymlink {
			// Empty directory exists - remove and symlink
			if err := symlink.MigrateExistingDir(s.Backend, cfg.SkillsDir); err != nil {
				fmt.Printf("Failed to link '%s': %v\n", s.Backend.Name, err)
				continue
			}
			fmt.Printf("Linked '%s' ✓\n", s.Backend.Name)
		} else if !s.Exists {
			// Nothing exists - create symlink directly
			if err := symlink.CreateLink(s.Backend, cfg.SkillsDir); err != nil {
				fmt.Printf("Failed to link '%s': %v\n", s.Backend.Name, err)
				continue
			}
			fmt.Printf("Linked '%s': %s → %s ✓\n", s.Backend.Name, expandedPath, cfg.SkillsDir)
		}
	}

	return nil
}

func runBackendUnlink(cmd *cobra.Command, args []string) error {
	cfg, err := config.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	name := args[0]
	backend := cfg.GetBackend(name)
	if backend == nil {
		return fmt.Errorf("backend '%s' not found", name)
	}

	statuses := symlink.CheckBackendLinks([]config.Backend{*backend}, cfg.SkillsDir)
	if len(statuses) == 0 || !statuses[0].Linked {
		fmt.Printf("Backend '%s' is not linked.\n", name)
		return nil
	}

	if err := symlink.RemoveLink(*backend); err != nil {
		return fmt.Errorf("failed to unlink '%s': %w", name, err)
	}

	fmt.Printf("Unlinked '%s' ✓\n", name)
	return nil
}

func runBackendAdd(cmd *cobra.Command, args []string) error {
	cfg, err := config.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	name := args[0]
	path := args[1]

	if err := cfg.AddBackend(name, path, backendDescription); err != nil {
		return fmt.Errorf("failed to add backend: %w", err)
	}

	fmt.Printf("Added backend '%s': %s\n", name, path)
	fmt.Printf("Run 'lazyas backend link %s' to create the symlink.\n", name)
	return nil
}

func runBackendRemove(cmd *cobra.Command, args []string) error {
	cfg, err := config.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	name := args[0]

	// Check if it's a known backend
	for _, known := range config.KnownBackends {
		if known.Name == name {
			return fmt.Errorf("cannot remove built-in backend '%s'", name)
		}
	}

	// Unlink first if linked
	backend := cfg.GetBackend(name)
	if backend != nil {
		statuses := symlink.CheckBackendLinks([]config.Backend{*backend}, cfg.SkillsDir)
		if len(statuses) > 0 && statuses[0].Linked {
			if err := symlink.RemoveLink(*backend); err != nil {
				fmt.Printf("Warning: failed to remove symlink: %v\n", err)
			}
		}
	}

	if err := cfg.RemoveBackend(name); err != nil {
		return fmt.Errorf("failed to remove backend: %w", err)
	}

	fmt.Printf("Removed backend '%s'\n", name)
	return nil
}
