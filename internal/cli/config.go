package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"lazyas/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage lazyas configuration",
	Long:  `View and modify lazyas configuration settings.`,
}

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage skill repositories",
	Long:  `Add, remove, and list skill repositories.`,
}

var repoAddCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add a skill repository",
	Long: `Add a skill repository to fetch skills from.

Examples:
  lazyas config repo add official https://github.com/anthropics/skills
  lazyas config repo add mycompany https://github.com/mycompany/skills`,
	Args: cobra.ExactArgs(2),
	RunE: runRepoAdd,
}

var repoRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a skill repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runRepoRemove,
}

var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured repositories",
	RunE:  runRepoList,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show all configuration",
	RunE:  runConfigShow,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show config file path",
	RunE:  runConfigPath,
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open config file in editor",
	RunE:  runConfigEdit,
}

func init() {
	repoCmd.AddCommand(repoAddCmd)
	repoCmd.AddCommand(repoRemoveCmd)
	repoCmd.AddCommand(repoListCmd)

	configCmd.AddCommand(repoCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configEditCmd)
}

func runRepoAdd(cmd *cobra.Command, args []string) error {
	cfg, err := config.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	name := args[0]
	url := args[1]

	if err := cfg.AddRepo(name, url); err != nil {
		return fmt.Errorf("failed to add repo: %w", err)
	}

	fmt.Printf("Added repository '%s': %s\n", name, url)
	return nil
}

func runRepoRemove(cmd *cobra.Command, args []string) error {
	cfg, err := config.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	name := args[0]

	if err := cfg.RemoveRepo(name); err != nil {
		return fmt.Errorf("failed to remove repo: %w", err)
	}

	fmt.Printf("Removed repository '%s'\n", name)
	return nil
}

func runRepoList(cmd *cobra.Command, args []string) error {
	cfg, err := config.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Repos) == 0 {
		fmt.Println("No repositories configured.")
		fmt.Println()
		fmt.Println("Add a repository with:")
		fmt.Println("  lazyas config repo add <name> <url>")
		return nil
	}

	fmt.Println("Configured repositories:")
	for _, repo := range cfg.Repos {
		fmt.Printf("  %s: %s\n", repo.Name, repo.URL)
	}

	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Configuration:")
	fmt.Printf("  config_file: %s\n", cfg.ConfigPath)
	fmt.Printf("  skills_dir:  %s\n", cfg.SkillsDir)
	fmt.Printf("  cache_ttl:   %d hours\n", cfg.CacheTTL)
	fmt.Println()

	if len(cfg.Repos) == 0 {
		fmt.Println("Repositories: (none)")
	} else {
		fmt.Println("Repositories:")
		for _, repo := range cfg.Repos {
			fmt.Printf("  %s: %s\n", repo.Name, repo.URL)
		}
	}
	fmt.Println()

	if len(cfg.Backends) == 0 {
		fmt.Println("Backends: (none)")
	} else {
		fmt.Println("Backends:")
		for _, b := range cfg.Backends {
			expandedPath, _ := config.ExpandPath(b.Path)
			desc := b.Description
			if desc == "" {
				desc = b.Name
			}
			fmt.Printf("  %s: %s (%s)\n", b.Name, expandedPath, desc)
		}
	}

	return nil
}

func runConfigPath(cmd *cobra.Command, args []string) error {
	cfg, err := config.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println(cfg.ConfigPath)
	return nil
}

func runConfigEdit(cmd *cobra.Command, args []string) error {
	cfg, err := config.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Ensure config file exists
	if err := cfg.EnsureDirs(); err != nil {
		return err
	}

	// Create default config if it doesn't exist
	if _, err := os.Stat(cfg.ConfigPath); os.IsNotExist(err) {
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
	}

	// Get editor from environment
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}

	// Open editor
	proc := os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}

	process, err := os.StartProcess("/usr/bin/env", []string{"env", editor, cfg.ConfigPath}, &proc)
	if err != nil {
		return fmt.Errorf("failed to start editor: %w", err)
	}

	_, err = process.Wait()
	return err
}
