package registry

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"lazyas/internal/config"
	"lazyas/internal/skillmd"
)

// Registry handles index operations
type Registry struct {
	cfg   *config.Config
	cache *CacheManager
	index *Index
}

// NewRegistry creates a new registry
func NewRegistry(cfg *config.Config) *Registry {
	return &Registry{
		cfg:   cfg,
		cache: NewCacheManager(cfg),
	}
}

// Fetch retrieves skills from all configured repositories
func (r *Registry) Fetch(forceRefresh bool) error {
	// Try cache first unless forced refresh
	if !forceRefresh {
		if err := r.cache.Load(); err == nil && r.cache.IsValid() {
			r.index = r.cache.Get()
			return nil
		}
	}

	// No repos configured
	if len(r.cfg.Repos) == 0 {
		r.index = &Index{}
		return fmt.Errorf("no repositories configured - add repos to %s", r.cfg.ConfigPath)
	}

	// Fetch from all configured repos
	var allSkills []SkillEntry
	var errors []string

	for _, repo := range r.cfg.Repos {
		skills, err := r.fetchRepo(repo.URL)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", repo.Name, err))
			continue
		}
		// Tag skills with their source repo name
		for i := range skills {
			if skills[i].Source.RepoName == "" {
				skills[i].Source.RepoName = repo.Name
			}
		}
		allSkills = append(allSkills, skills...)
	}

	r.index = &Index{Skills: allSkills}

	// Update cache
	if err := r.cache.Set(r.index); err != nil {
		// Non-fatal
		fmt.Fprintf(os.Stderr, "warning: failed to cache index: %v\n", err)
	}

	if len(errors) > 0 && len(allSkills) == 0 {
		return fmt.Errorf("failed to fetch from any repository:\n  %s", joinErrors(errors))
	}

	return nil
}

func (r *Registry) fetchRepo(repoURL string) ([]SkillEntry, error) {
	// Clone repo to temp dir
	tempDir, err := os.MkdirTemp("", "lazyas-index-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Shallow clone
	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, tempDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git clone failed: %s", string(output))
	}

	// Try index.yaml first (index repo)
	indexPath := filepath.Join(tempDir, "index.yaml")
	if data, err := os.ReadFile(indexPath); err == nil {
		var index Index
		if err := yaml.Unmarshal(data, &index); err != nil {
			return nil, fmt.Errorf("failed to parse index.yaml: %w", err)
		}
		return index.Skills, nil
	}

	// No index.yaml - scan for skills (skills repo)
	return r.scanForSkills(tempDir, repoURL)
}

// scanForSkills discovers skills by finding SKILL.md files
func (r *Registry) scanForSkills(repoDir, repoURL string) ([]SkillEntry, error) {
	var skills []SkillEntry
	seen := map[string]bool{}

	// Support single-skill repos where SKILL.md lives at the repo root.
	if _, err := os.Stat(filepath.Join(repoDir, "SKILL.md")); err == nil {
		rootName := inferRootSkillName(repoURL, repoDir)
		rootEntry := makeSkillEntry(rootName, repoDir, repoDir, repoURL)
		if !seen[rootEntry.Source.Path] {
			seen[rootEntry.Source.Path] = true
			skills = append(skills, rootEntry)
		}
	}

	// Common locations for skills
	searchDirs := []string{
		repoDir,                          // root
		filepath.Join(repoDir, "skills"), // skills/
		filepath.Join(repoDir, "external", "skills"), // external/skills/
	}

	for _, searchDir := range searchDirs {
		entries, err := os.ReadDir(searchDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			// Skip hidden directories
			if entry.Name()[0] == '.' {
				continue
			}

			skillDir := filepath.Join(searchDir, entry.Name())
			skillMdPath := filepath.Join(skillDir, "SKILL.md")

			if _, err := os.Stat(skillMdPath); err == nil {
				entry := makeSkillEntry(entry.Name(), skillDir, repoDir, repoURL)
				if !seen[entry.Source.Path] {
					seen[entry.Source.Path] = true
					skills = append(skills, entry)
				}
			} else {
				// Check one level deeper for category/skill-name layout
				subEntries, err := os.ReadDir(skillDir)
				if err != nil {
					continue
				}
				for _, sub := range subEntries {
					if !sub.IsDir() || sub.Name()[0] == '.' {
						continue
					}
					subDir := filepath.Join(skillDir, sub.Name())
					if _, err := os.Stat(filepath.Join(subDir, "SKILL.md")); err == nil {
						subEntry := makeSkillEntry(sub.Name(), subDir, repoDir, repoURL)
						if !seen[subEntry.Source.Path] {
							seen[subEntry.Source.Path] = true
							skills = append(skills, subEntry)
						}
					}
				}
			}
		}
	}

	if len(skills) == 0 {
		return nil, fmt.Errorf("no index.yaml and no skills found (looking for directories with SKILL.md)")
	}

	return skills, nil
}

func makeSkillEntry(name, skillDir, repoDir, repoURL string) SkillEntry {
	skill := SkillEntry{
		Name: name,
		Source: SkillSource{
			Repo: repoURL,
		},
	}
	relPath, _ := filepath.Rel(repoDir, skillDir)
	if relPath == "." {
		relPath = ""
	}
	skill.Source.Path = relPath
	if content, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md")); err == nil {
		skill.Description = skillmd.ExtractDescription(string(content))
	}
	return skill
}

func inferRootSkillName(repoURL, repoDir string) string {
	if parsed, err := url.Parse(repoURL); err == nil && parsed.Host != "" {
		p := strings.Trim(strings.TrimSuffix(parsed.Path, ".git"), "/")
		if p != "" {
			return filepath.Base(p)
		}
	}

	// Handle SCP-style git URLs: git@github.com:owner/repo.git
	if idx := strings.LastIndex(repoURL, ":"); idx != -1 {
		p := strings.Trim(strings.TrimSuffix(repoURL[idx+1:], ".git"), "/")
		if p != "" {
			return filepath.Base(p)
		}
	}

	if p := strings.Trim(strings.TrimSuffix(repoURL, ".git"), "/"); p != "" {
		if base := filepath.Base(p); base != "." && base != string(filepath.Separator) {
			return base
		}
	}
	return filepath.Base(repoDir)
}

func joinErrors(errors []string) string {
	result := ""
	for i, e := range errors {
		if i > 0 {
			result += "\n  "
		}
		result += e
	}
	return result
}

// GetIndex returns the current index
func (r *Registry) GetIndex() *Index {
	return r.index
}

// GetSkill finds a skill by name
func (r *Registry) GetSkill(name string) *SkillEntry {
	if r.index == nil {
		return nil
	}

	for i := range r.index.Skills {
		if r.index.Skills[i].Name == name {
			return &r.index.Skills[i]
		}
	}
	return nil
}

// SearchSkills searches for skills matching a query
func (r *Registry) SearchSkills(query string) []SkillEntry {
	if r.index == nil {
		return nil
	}

	var results []SkillEntry
	for _, skill := range r.index.Skills {
		if skill.MatchesQuery(query) {
			results = append(results, skill)
		}
	}
	return results
}

// ListSkills returns all skills
func (r *Registry) ListSkills() []SkillEntry {
	if r.index == nil {
		return nil
	}
	return r.index.Skills
}
