package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	DefaultCacheTTLHours = 24
	ConfigFileName       = "config.toml"
	ManifestFileName     = "manifest.yaml"
	CacheFileName        = "cache.yaml"
)

// Repo represents an upstream skills repository
type Repo struct {
	Name string `toml:"name"`
	URL  string `toml:"url"`
}

// Backend represents a target AI agent backend
type Backend struct {
	Name        string `toml:"name"`
	Path        string `toml:"path"`        // Expected symlink location (e.g., ~/.claude/skills)
	Description string `toml:"description"` // Human-readable name
	Linked      bool   `toml:"-"`           // Runtime: is symlink active?
}

// KnownBackends are the built-in supported backends
var KnownBackends = []Backend{
	{Name: "claude", Path: "~/.claude/skills", Description: "Claude Code"},
	{Name: "codex", Path: "~/.codex/skills", Description: "OpenAI Codex"},
}

// ConfigFile represents the TOML config file structure
type ConfigFile struct {
	Repos    []Repo    `toml:"repos"`
	CacheTTL int       `toml:"cache_ttl_hours,omitempty"`
	Backends []Backend `toml:"backends,omitempty"`
}

// Config holds the runtime configuration
type Config struct {
	ConfigDir    string
	ConfigPath   string
	ManifestPath string
	CachePath    string
	SkillsDir    string // Always ~/.lazyas/skills/ - the central skills directory
	Repos        []Repo
	CacheTTL     int
	Backends     []Backend // Configured backends (symlink targets)
}

// ExpandPath expands ~ to home directory in a path
func ExpandPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, path[1:]), nil
}

// DefaultConfig returns the default configuration
func DefaultConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Central lazyas directory is ~/.lazyas/
	configDir := filepath.Join(home, ".lazyas")
	skillsDir := filepath.Join(configDir, "skills")

	// Initialize default backends from KnownBackends
	backends := make([]Backend, len(KnownBackends))
	copy(backends, KnownBackends)

	cfg := &Config{
		ConfigDir:    configDir,
		ConfigPath:   filepath.Join(configDir, ConfigFileName),
		ManifestPath: filepath.Join(configDir, ManifestFileName),
		CachePath:    filepath.Join(configDir, CacheFileName),
		SkillsDir:    skillsDir,
		CacheTTL:     DefaultCacheTTLHours,
		Repos:        []Repo{},
		Backends:     backends,
	}

	// Try to load existing config
	if err := cfg.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// Migrate from old config location if needed
	if err := cfg.migrateOldConfig(); err != nil {
		// Non-fatal, just log and continue
	}

	return cfg, nil
}

// Load reads the config from disk
func (c *Config) Load() error {
	var cf ConfigFile
	if _, err := toml.DecodeFile(c.ConfigPath, &cf); err != nil {
		return err
	}

	if len(cf.Repos) > 0 {
		c.Repos = cf.Repos
	}
	if cf.CacheTTL > 0 {
		c.CacheTTL = cf.CacheTTL
	}
	// Merge backends from config file with known backends
	if len(cf.Backends) > 0 {
		c.Backends = mergeBackends(KnownBackends, cf.Backends)
	}

	return nil
}

// mergeBackends merges known backends with user-configured backends
func mergeBackends(known, configured []Backend) []Backend {
	result := make([]Backend, 0, len(known)+len(configured))
	seen := make(map[string]bool)

	// First, add configured backends (they take precedence)
	for _, b := range configured {
		result = append(result, b)
		seen[b.Name] = true
	}

	// Then add known backends that weren't in config
	for _, b := range known {
		if !seen[b.Name] {
			result = append(result, b)
		}
	}

	return result
}

// Save writes the config to disk
func (c *Config) Save() error {
	if err := c.EnsureDirs(); err != nil {
		return err
	}

	cf := ConfigFile{
		Repos:    c.Repos,
		CacheTTL: c.CacheTTL,
	}

	// Only save backends that differ from known backends or are custom
	customBackends := filterCustomBackends(c.Backends)
	if len(customBackends) > 0 {
		cf.Backends = customBackends
	}

	f, err := os.Create(c.ConfigPath)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	return encoder.Encode(cf)
}

// filterCustomBackends returns backends that are custom or modified from known defaults
func filterCustomBackends(backends []Backend) []Backend {
	knownByName := make(map[string]Backend)
	for _, b := range KnownBackends {
		knownByName[b.Name] = b
	}

	var custom []Backend
	for _, b := range backends {
		known, isKnown := knownByName[b.Name]
		if !isKnown {
			// Custom backend
			custom = append(custom, b)
		} else if b.Path != known.Path || b.Description != known.Description {
			// Modified known backend
			custom = append(custom, b)
		}
	}
	return custom
}

// EnsureDirs creates necessary directories if they don't exist
func (c *Config) EnsureDirs() error {
	if err := os.MkdirAll(c.ConfigDir, 0755); err != nil {
		return err
	}
	return os.MkdirAll(c.SkillsDir, 0755)
}

// AddRepo adds a new repository to the config
func (c *Config) AddRepo(name, url string) error {
	// Check if repo already exists
	for i, r := range c.Repos {
		if r.Name == name {
			c.Repos[i].URL = url
			return c.Save()
		}
	}
	c.Repos = append(c.Repos, Repo{Name: name, URL: url})
	return c.Save()
}

// RemoveRepo removes a repository from the config
func (c *Config) RemoveRepo(name string) error {
	for i, r := range c.Repos {
		if r.Name == name {
			c.Repos = append(c.Repos[:i], c.Repos[i+1:]...)
			return c.Save()
		}
	}
	return nil
}

// GetRepoURLs returns all configured repository URLs
func (c *Config) GetRepoURLs() []string {
	urls := make([]string, len(c.Repos))
	for i, r := range c.Repos {
		urls[i] = r.URL
	}
	return urls
}

// GetBackend returns a backend by name
func (c *Config) GetBackend(name string) *Backend {
	for i := range c.Backends {
		if c.Backends[i].Name == name {
			return &c.Backends[i]
		}
	}
	return nil
}

// AddBackend adds or updates a backend configuration
func (c *Config) AddBackend(name, path, description string) error {
	// Check if backend already exists
	for i := range c.Backends {
		if c.Backends[i].Name == name {
			c.Backends[i].Path = path
			if description != "" {
				c.Backends[i].Description = description
			}
			return c.Save()
		}
	}

	// Add new backend
	c.Backends = append(c.Backends, Backend{
		Name:        name,
		Path:        path,
		Description: description,
	})
	return c.Save()
}

// RemoveBackend removes a backend from the configuration
func (c *Config) RemoveBackend(name string) error {
	// Check if it's a known backend - can't remove those
	for _, known := range KnownBackends {
		if known.Name == name {
			return nil // Don't error, just don't remove known backends
		}
	}

	for i := range c.Backends {
		if c.Backends[i].Name == name {
			c.Backends = append(c.Backends[:i], c.Backends[i+1:]...)
			return c.Save()
		}
	}
	return nil
}

// migrateOldConfig migrates from old ~/.config/lazyas/ location to ~/.lazyas/
func (c *Config) migrateOldConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	oldConfigDir := filepath.Join(home, ".config", "lazyas")
	oldConfigPath := filepath.Join(oldConfigDir, ConfigFileName)
	oldManifestPath := filepath.Join(oldConfigDir, ManifestFileName)
	oldCachePath := filepath.Join(oldConfigDir, CacheFileName)

	// Check if old config exists but new doesn't
	if _, err := os.Stat(oldConfigPath); os.IsNotExist(err) {
		return nil // Nothing to migrate
	}

	if _, err := os.Stat(c.ConfigPath); err == nil {
		return nil // New config already exists, don't overwrite
	}

	// Ensure new directory exists
	if err := os.MkdirAll(c.ConfigDir, 0755); err != nil {
		return err
	}

	// Move files if they exist
	if _, err := os.Stat(oldConfigPath); err == nil {
		if err := os.Rename(oldConfigPath, c.ConfigPath); err != nil {
			// If rename fails (cross-device), try copy+delete
			if data, err := os.ReadFile(oldConfigPath); err == nil {
				os.WriteFile(c.ConfigPath, data, 0644)
			}
		}
	}

	if _, err := os.Stat(oldManifestPath); err == nil {
		if err := os.Rename(oldManifestPath, c.ManifestPath); err != nil {
			if data, err := os.ReadFile(oldManifestPath); err == nil {
				os.WriteFile(c.ManifestPath, data, 0644)
			}
		}
	}

	if _, err := os.Stat(oldCachePath); err == nil {
		if err := os.Rename(oldCachePath, c.CachePath); err != nil {
			if data, err := os.ReadFile(oldCachePath); err == nil {
				os.WriteFile(c.CachePath, data, 0644)
			}
		}
	}

	// Reload config after migration
	return c.Load()
}
