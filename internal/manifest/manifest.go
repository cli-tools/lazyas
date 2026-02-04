package manifest

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
	"lazyas/internal/config"
)

// Manager handles manifest operations
type Manager struct {
	cfg      *config.Config
	manifest *Manifest
}

// NewManager creates a new manifest manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		cfg: cfg,
	}
}

// Load reads the manifest from disk
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.cfg.ManifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.manifest = NewManifest()
			return nil
		}
		return err
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return err
	}

	if manifest.Installed == nil {
		manifest.Installed = make(map[string]InstalledSkill)
	}

	m.manifest = &manifest
	return nil
}

// Save writes the manifest to disk
func (m *Manager) Save() error {
	if err := m.cfg.EnsureDirs(); err != nil {
		return err
	}

	data, err := yaml.Marshal(m.manifest)
	if err != nil {
		return err
	}

	return os.WriteFile(m.cfg.ManifestPath, data, 0644)
}

// Get returns the current manifest
func (m *Manager) Get() *Manifest {
	if m.manifest == nil {
		m.manifest = NewManifest()
	}
	return m.manifest
}

// AddSkill adds an installed skill to the manifest
func (m *Manager) AddSkill(name, version, commit, sourceRepo, sourcePath string) error {
	if m.manifest == nil {
		m.manifest = NewManifest()
	}

	m.manifest.Installed[name] = InstalledSkill{
		Version:     version,
		Commit:      commit,
		InstalledAt: time.Now(),
		SourceRepo:  sourceRepo,
		SourcePath:  sourcePath,
	}

	return m.Save()
}

// RemoveSkill removes a skill from the manifest
func (m *Manager) RemoveSkill(name string) error {
	if m.manifest == nil {
		return nil
	}

	delete(m.manifest.Installed, name)
	return m.Save()
}

// IsInstalled checks if a skill is installed (exists on disk with SKILL.md)
func (m *Manager) IsInstalled(name string) bool {
	skillPath := filepath.Join(m.cfg.SkillsDir, name)
	skillMdPath := filepath.Join(skillPath, "SKILL.md")
	_, err := os.Stat(skillMdPath)
	return err == nil
}

// GetInstalled returns info about an installed skill
func (m *Manager) GetInstalled(name string) (InstalledSkill, bool) {
	if m.manifest == nil {
		return InstalledSkill{}, false
	}
	skill, ok := m.manifest.Installed[name]
	return skill, ok
}

// ListInstalled returns all installed skills
func (m *Manager) ListInstalled() map[string]InstalledSkill {
	if m.manifest == nil {
		return make(map[string]InstalledSkill)
	}
	return m.manifest.Installed
}

// GetSkillPath returns the path where a skill should be installed
func (m *Manager) GetSkillPath(name string) string {
	return filepath.Join(m.cfg.SkillsDir, name)
}

// ScanLocalSkills scans the skills directory for locally installed skills
// Returns a map of skill name -> LocalSkill for each directory containing SKILL.md
func (m *Manager) ScanLocalSkills() map[string]LocalSkill {
	result := make(map[string]LocalSkill)

	entries, err := os.ReadDir(m.cfg.SkillsDir)
	if err != nil {
		return result
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip the .lazyas directory
		if entry.Name() == ".lazyas" {
			continue
		}

		skillPath := filepath.Join(m.cfg.SkillsDir, entry.Name())
		skillMdPath := filepath.Join(skillPath, "SKILL.md")

		if _, err := os.Stat(skillMdPath); err == nil {
			// Read SKILL.md to extract description
			description := ""
			if content, err := os.ReadFile(skillMdPath); err == nil {
				description = extractDescription(string(content))
			}

			// Check if it's a git repo and if it's modified
			isGitRepo := isGitRepository(skillPath)
			isModified := false
			if isGitRepo {
				isModified = hasLocalModifications(skillPath)
			}

			result[entry.Name()] = LocalSkill{
				Name:        entry.Name(),
				Path:        skillPath,
				Description: description,
				IsGitRepo:   isGitRepo,
				IsModified:  isModified,
			}
		}
	}

	return result
}

// isGitRepository checks if a path contains a .git directory
func isGitRepository(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// hasLocalModifications checks if a git repo has uncommitted changes
func hasLocalModifications(path string) bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(trimSpace(string(out))) > 0
}

// extractDescription extracts a brief description from SKILL.md content
func extractDescription(content string) string {
	lines := splitLines(content)
	inFrontmatter := false
	frontmatterCount := 0

	for _, line := range lines {
		trimmed := trimSpace(line)

		// Handle YAML frontmatter (between --- markers)
		if trimmed == "---" {
			frontmatterCount++
			inFrontmatter = frontmatterCount == 1
			if frontmatterCount == 2 {
				inFrontmatter = false
			}
			continue
		}

		// Look for description field in frontmatter
		if inFrontmatter {
			if len(trimmed) > 12 && trimmed[:12] == "description:" {
				desc := trimSpace(trimmed[12:])
				// Remove quotes if present
				if len(desc) >= 2 && (desc[0] == '"' || desc[0] == '\'') {
					desc = desc[1 : len(desc)-1]
				}
				if len(desc) > 100 {
					return desc[:97] + "..."
				}
				return desc
			}
			continue
		}

		if trimmed == "" {
			continue
		}

		// Skip headings
		if len(trimmed) > 0 && trimmed[0] == '#' {
			continue
		}

		// Skip code blocks and list markers
		if len(trimmed) >= 3 && trimmed[:3] == "```" {
			continue
		}
		if len(trimmed) > 0 && trimmed[0] == '-' {
			continue
		}

		// Return first content line (truncated)
		if len(trimmed) > 100 {
			return trimmed[:97] + "..."
		}
		return trimmed
	}
	return ""
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
