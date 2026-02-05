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

// StarterKitRepos are popular skill repositories offered on first run
var StarterKitRepos = []Repo{
	{Name: "anthropic-official", URL: "https://github.com/anthropics/skills"},
	{Name: "vercel-official", URL: "https://github.com/vercel-labs/agent-skills"},
	{Name: "context-engineering", URL: "https://github.com/muratcankoylan/Agent-Skills-for-Context-Engineering"},
	{Name: "antigravity", URL: "https://github.com/sickn33/antigravity-awesome-skills"},
	{Name: "ai-research", URL: "https://github.com/Orchestra-Research/AI-research-SKILLs"},
	{Name: "claude-skills", URL: "https://github.com/alirezarezvani/claude-skills"},
	{Name: "skillcreator", URL: "https://github.com/skillcreatorai/Ai-Agent-Skills"},
	{Name: "microsoft-official", URL: "https://github.com/microsoft/agent-skills"},
}

// KnownBackends are the built-in supported backends
var KnownBackends = []Backend{
	{Name: "claude", Path: "~/.claude/skills", Description: "Claude Code"},
	{Name: "codex", Path: "~/.codex/skills", Description: "OpenAI Codex"},
	{Name: "gemini", Path: "~/.gemini/skills", Description: "Gemini CLI"},
	{Name: "cursor", Path: "~/.cursor/skills", Description: "Cursor"},
	{Name: "copilot", Path: "~/.copilot/skills", Description: "GitHub Copilot"},
	{Name: "amp", Path: "$XDG_CONFIG_HOME/agents/skills", Description: "Amp"},
	{Name: "goose", Path: "$XDG_CONFIG_HOME/goose/skills", Description: "Goose"},
	{Name: "opencode", Path: "$XDG_CONFIG_HOME/opencode/skills", Description: "OpenCode"},
	{Name: "vibe", Path: "~/.vibe/skills", Description: "Mistral Vibe"},
}

// ConfigStore abstracts config persistence (serialization + I/O).
type ConfigStore interface {
	Save(cf *ConfigFile) error
	Load() (*ConfigFile, error)
}

// TOMLStore reads and writes ConfigFile as TOML on disk.
type TOMLStore struct {
	Path string
}

// Save writes cf to disk as TOML, creating parent directories as needed.
func (s *TOMLStore) Save(cf *ConfigFile) error {
	dir := filepath.Dir(s.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.Create(s.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cf)
}

// Load reads a TOML file from disk into a ConfigFile.
func (s *TOMLStore) Load() (*ConfigFile, error) {
	var cf ConfigFile
	if _, err := toml.DecodeFile(s.Path, &cf); err != nil {
		return nil, err
	}
	return &cf, nil
}

// ConfigFile represents the TOML config file structure
type ConfigFile struct {
	Repos               []Repo    `toml:"repos"`
	CacheTTL            int       `toml:"cache_ttl_hours,omitempty"`
	Viewer              string    `toml:"viewer,omitempty"`
	Backends            []Backend `toml:"backends,omitempty"`
	DismissedBackends   []string  `toml:"dismissed_backends,omitempty"`
	StarterKitDismissed bool      `toml:"starter_kit_dismissed,omitempty"`
	CollapsedGroups     []string  `toml:"collapsed_groups,omitempty"`
}

// Config holds the runtime configuration
type Config struct {
	Store               ConfigStore
	ConfigDir           string
	ConfigPath          string
	ManifestPath        string
	CachePath           string
	SkillsDir           string // Always ~/.lazyas/skills/ - the central skills directory
	ReposDir            string // Always ~/.lazyas/repos/ - per-repo sparse clones
	Repos               []Repo
	CacheTTL            int
	Viewer              string    // Command to view SKILL.md (e.g. "glow -t"); empty = auto-detect
	Backends            []Backend // Configured backends (symlink targets)
	DismissedBackends   []string  // Backend names dismissed from auto-show
	StarterKitDismissed bool      // Whether starter kit modal was dismissed
	CollapsedGroups     []string  // Group names that are collapsed in the TUI
}

// xdgConfigHome returns $XDG_CONFIG_HOME, falling back to ~/.config per spec.
func xdgConfigHome() (string, error) {
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config"), nil
}

// ExpandPath expands ~ and $XDG_CONFIG_HOME in a path
func ExpandPath(path string) (string, error) {
	const xdgPrefix = "$XDG_CONFIG_HOME"
	if strings.HasPrefix(path, xdgPrefix) {
		xdg, err := xdgConfigHome()
		if err != nil {
			return "", err
		}
		return filepath.Join(xdg, path[len(xdgPrefix):]), nil
	}
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
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
	reposDir := filepath.Join(configDir, "repos")

	// Initialize default backends from KnownBackends
	backends := make([]Backend, len(KnownBackends))
	copy(backends, KnownBackends)

	configPath := filepath.Join(configDir, ConfigFileName)

	cfg := &Config{
		Store:        &TOMLStore{Path: configPath},
		ConfigDir:    configDir,
		ConfigPath:   configPath,
		ManifestPath: filepath.Join(configDir, ManifestFileName),
		CachePath:    filepath.Join(configDir, CacheFileName),
		SkillsDir:    skillsDir,
		ReposDir:     reposDir,
		CacheTTL:     DefaultCacheTTLHours,
		Repos:        []Repo{},
		Backends:     backends,
	}

	// Try to load existing config
	if err := cfg.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return cfg, nil
}

// Load reads the config via the configured store
func (c *Config) Load() error {
	cf, err := c.Store.Load()
	if err != nil {
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

	c.Viewer = cf.Viewer
	c.DismissedBackends = cf.DismissedBackends
	c.StarterKitDismissed = cf.StarterKitDismissed
	c.CollapsedGroups = cf.CollapsedGroups

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

// Save writes the config via the configured store
func (c *Config) Save() error {
	if err := c.EnsureDirs(); err != nil {
		return err
	}

	cf := ConfigFile{
		Repos:               c.Repos,
		CacheTTL:            c.CacheTTL,
		Viewer:              c.Viewer,
		DismissedBackends:   c.DismissedBackends,
		StarterKitDismissed: c.StarterKitDismissed,
		CollapsedGroups:     c.CollapsedGroups,
	}

	// Only save backends that differ from known backends or are custom
	customBackends := filterCustomBackends(c.Backends)
	if len(customBackends) > 0 {
		cf.Backends = customBackends
	}

	return c.Store.Save(&cf)
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
	if err := os.MkdirAll(c.SkillsDir, 0755); err != nil {
		return err
	}
	return os.MkdirAll(c.ReposDir, 0755)
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

// GetBackend returns a backend by name
func (c *Config) GetBackend(name string) *Backend {
	for i := range c.Backends {
		if c.Backends[i].Name == name {
			return &c.Backends[i]
		}
	}
	return nil
}

// GetRepo returns a repo by name
func (c *Config) GetRepo(name string) *Repo {
	for i := range c.Repos {
		if c.Repos[i].Name == name {
			return &c.Repos[i]
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

// DismissBackend adds a backend name to the dismissed list
func (c *Config) DismissBackend(name string) {
	for _, d := range c.DismissedBackends {
		if d == name {
			return
		}
	}
	c.DismissedBackends = append(c.DismissedBackends, name)
}

// UndismissBackend removes a backend name from the dismissed list
func (c *Config) UndismissBackend(name string) {
	for i, d := range c.DismissedBackends {
		if d == name {
			c.DismissedBackends = append(c.DismissedBackends[:i], c.DismissedBackends[i+1:]...)
			return
		}
	}
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
