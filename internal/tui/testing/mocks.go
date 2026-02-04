package testing

import (
	"lazyas/internal/manifest"
	"lazyas/internal/registry"
)

// MockRegistry implements tui.SkillRegistry for testing
type MockRegistry struct {
	skills []registry.SkillEntry
}

// NewMockRegistry creates a new mock registry with the given skills
func NewMockRegistry(skills []registry.SkillEntry) *MockRegistry {
	return &MockRegistry{skills: skills}
}

// ListSkills returns all skills
func (m *MockRegistry) ListSkills() []registry.SkillEntry {
	return m.skills
}

// SearchSkills searches for skills matching a query
func (m *MockRegistry) SearchSkills(query string) []registry.SkillEntry {
	if query == "" {
		return m.skills
	}
	var results []registry.SkillEntry
	for _, skill := range m.skills {
		if skill.MatchesQuery(query) {
			results = append(results, skill)
		}
	}
	return results
}

// GetSkill finds a skill by name
func (m *MockRegistry) GetSkill(name string) *registry.SkillEntry {
	for i := range m.skills {
		if m.skills[i].Name == name {
			return &m.skills[i]
		}
	}
	return nil
}

// SetSkills updates the mock's skill list
func (m *MockRegistry) SetSkills(skills []registry.SkillEntry) {
	m.skills = skills
}

// MockManifest implements tui.SkillManifest for testing
type MockManifest struct {
	installed   map[string]manifest.InstalledSkill
	localSkills map[string]manifest.LocalSkill
	skillsPath  string
}

// NewMockManifest creates a new mock manifest
func NewMockManifest() *MockManifest {
	return &MockManifest{
		installed:   make(map[string]manifest.InstalledSkill),
		localSkills: make(map[string]manifest.LocalSkill),
		skillsPath:  "/mock/skills",
	}
}

// NewMockManifestWithInstalled creates a mock manifest with pre-installed skills
func NewMockManifestWithInstalled(installed map[string]manifest.InstalledSkill) *MockManifest {
	// Convert installed to local skills for IsInstalled to work
	localSkills := make(map[string]manifest.LocalSkill)
	for name := range installed {
		localSkills[name] = manifest.LocalSkill{Name: name}
	}
	return &MockManifest{
		installed:   installed,
		localSkills: localSkills,
		skillsPath:  "/mock/skills",
	}
}

// NewMockManifestWithLocalSkills creates a mock manifest with local skills
func NewMockManifestWithLocalSkills(localSkills map[string]manifest.LocalSkill) *MockManifest {
	return &MockManifest{
		installed:   make(map[string]manifest.InstalledSkill),
		localSkills: localSkills,
		skillsPath:  "/mock/skills",
	}
}

// IsInstalled checks if a skill is installed (exists locally)
func (m *MockManifest) IsInstalled(name string) bool {
	_, ok := m.localSkills[name]
	return ok
}

// ListInstalled returns all installed skills
func (m *MockManifest) ListInstalled() map[string]manifest.InstalledSkill {
	return m.installed
}

// GetInstalled returns info about an installed skill
func (m *MockManifest) GetInstalled(name string) (manifest.InstalledSkill, bool) {
	skill, ok := m.installed[name]
	return skill, ok
}

// GetSkillPath returns the path where a skill should be installed
func (m *MockManifest) GetSkillPath(name string) string {
	return m.skillsPath + "/" + name
}

// Install marks a skill as installed
func (m *MockManifest) Install(name string, skill manifest.InstalledSkill) {
	m.installed[name] = skill
	m.localSkills[name] = manifest.LocalSkill{Name: name}
}

// Remove uninstalls a skill
func (m *MockManifest) Remove(name string) {
	delete(m.installed, name)
	delete(m.localSkills, name)
}

// SetSkillsPath sets the base path for skills
func (m *MockManifest) SetSkillsPath(path string) {
	m.skillsPath = path
}

// ScanLocalSkills returns locally installed skills
func (m *MockManifest) ScanLocalSkills() map[string]manifest.LocalSkill {
	return m.localSkills
}

// AddLocalSkill adds a local skill (for testing)
func (m *MockManifest) AddLocalSkill(name, description string) {
	m.localSkills[name] = manifest.LocalSkill{
		Name:        name,
		Path:        m.skillsPath + "/" + name,
		Description: description,
	}
}
