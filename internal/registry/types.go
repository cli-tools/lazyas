package registry

import "time"

// Index represents the registry index.yaml structure
type Index struct {
	Version  int           `yaml:"version"`
	Metadata IndexMetadata `yaml:"metadata"`
	Skills   []SkillEntry  `yaml:"skills"`
}

// IndexMetadata contains registry metadata
type IndexMetadata struct {
	Name      string    `yaml:"name"`
	UpdatedAt time.Time `yaml:"updated_at"`
}

// SkillEntry represents a skill in the registry
type SkillEntry struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Source      SkillSource `yaml:"source"`
	Author      string      `yaml:"author"`
	Tags        []string    `yaml:"tags"`
}

// SkillSource defines where to fetch the skill from
type SkillSource struct {
	Repo     string `yaml:"repo"`
	Path     string `yaml:"path"` // subdirectory within repo (optional)
	Tag      string `yaml:"tag"`  // version tag
	RepoName string `yaml:"-"`    // name of the config repo (not serialized)
}

// MatchesQuery checks if the skill matches a search query
func (s *SkillEntry) MatchesQuery(query string) bool {
	if query == "" {
		return true
	}

	// Check name
	if containsIgnoreCase(s.Name, query) {
		return true
	}

	// Check description
	if containsIgnoreCase(s.Description, query) {
		return true
	}

	// Check author
	if containsIgnoreCase(s.Author, query) {
		return true
	}

	// Check tags
	for _, tag := range s.Tags {
		if containsIgnoreCase(tag, query) {
			return true
		}
	}

	return false
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(substr) == 0 ||
			findIgnoreCase(s, substr))
}

func findIgnoreCase(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFoldAt(s, i, substr) {
			return true
		}
	}
	return false
}

func equalFoldAt(s string, start int, substr string) bool {
	for j := 0; j < len(substr); j++ {
		c1 := s[start+j]
		c2 := substr[j]
		if c1 == c2 {
			continue
		}
		// Simple ASCII case folding
		if c1 >= 'A' && c1 <= 'Z' {
			c1 += 'a' - 'A'
		}
		if c2 >= 'A' && c2 <= 'Z' {
			c2 += 'a' - 'A'
		}
		if c1 != c2 {
			return false
		}
	}
	return true
}
