package testing

import (
	"fmt"
	"time"

	"lazyas/internal/config"
	"lazyas/internal/manifest"
	"lazyas/internal/registry"
)

// TestSkills returns a set of test skill entries with varied properties
func TestSkills() []registry.SkillEntry {
	return []registry.SkillEntry{
		{
			Name:        "test-skill-1",
			Description: "A test skill for unit testing",
			Author:      "test-author",
			Tags:        []string{"testing", "example"},
			Source: registry.SkillSource{
				Repo: "https://github.com/example/skills",
				Path: "test-skill-1",
				Tag:  "v1.0.0",
			},
		},
		{
			Name:        "test-skill-2",
			Description: "Another test skill",
			Author:      "test-author",
			Tags:        []string{"testing"},
			Source: registry.SkillSource{
				Repo: "https://github.com/example/skills",
				Path: "test-skill-2",
				Tag:  "v2.0.0",
			},
		},
		{
			Name:        "different-repo-skill",
			Description: "A skill from a different repo",
			Author:      "other-author",
			Tags:        []string{"utility"},
			Source: registry.SkillSource{
				Repo: "https://github.com/other/repo",
				Path: "",
				Tag:  "v1.0.0",
			},
		},
		{
			Name:        "third-skill",
			Description: "Third test skill from example repo",
			Author:      "test-author",
			Tags:        []string{"testing", "advanced"},
			Source: registry.SkillSource{
				Repo: "https://github.com/example/skills",
				Path: "third-skill",
				Tag:  "v1.5.0",
			},
		},
		{
			Name:        "standalone-skill",
			Description: "A standalone skill",
			Author:      "solo-author",
			Tags:        []string{"standalone"},
			Source: registry.SkillSource{
				Repo: "https://github.com/solo/skill",
				Path: "",
				Tag:  "v0.1.0",
			},
		},
	}
}

// TestSkillsWithInstalled returns test skills plus a map of which are installed
func TestSkillsWithInstalled() ([]registry.SkillEntry, map[string]manifest.InstalledSkill) {
	skills := TestSkills()
	installed := map[string]manifest.InstalledSkill{
		"test-skill-1": {
			Version:     "v1.0.0",
			Commit:      "abc123",
			InstalledAt: time.Now().Add(-24 * time.Hour),
			SourceRepo:  "https://github.com/example/skills",
			SourcePath:  "test-skill-1",
		},
		"different-repo-skill": {
			Version:     "v1.0.0",
			Commit:      "def456",
			InstalledAt: time.Now().Add(-48 * time.Hour),
			SourceRepo:  "https://github.com/other/repo",
			SourcePath:  "",
		},
	}
	return skills, installed
}

// TestConfig returns a test configuration
func TestConfig() *config.Config {
	return &config.Config{
		ConfigDir:    "/tmp/lazyas-test/.lazyas",
		ConfigPath:   "/tmp/lazyas-test/.lazyas/config.toml",
		SkillsDir:    "/tmp/lazyas-test/skills",
		ManifestPath: "/tmp/lazyas-test/.lazyas/manifest.yaml",
		CachePath:    "/tmp/lazyas-test/.lazyas/cache.yaml",
		CacheTTL:     24,
	}
}

// SingleSkill returns a single test skill
func SingleSkill() registry.SkillEntry {
	return registry.SkillEntry{
		Name:        "single-skill",
		Description: "A single test skill",
		Author:      "test-author",
		Tags:        []string{"single"},
		Source: registry.SkillSource{
			Repo: "https://github.com/test/single",
			Path: "",
			Tag:  "v1.0.0",
		},
	}
}

// ManySkills returns a larger set of skills for pagination testing
func ManySkills(count int) []registry.SkillEntry {
	skills := make([]registry.SkillEntry, count)
	repos := []string{
		"https://github.com/repo-a/skills",
		"https://github.com/repo-b/skills",
		"https://github.com/repo-c/skills",
	}
	for i := 0; i < count; i++ {
		skills[i] = registry.SkillEntry{
			Name:        fmt.Sprintf("skill-%03d", i+1),
			Description: fmt.Sprintf("Test skill number %d", i+1),
			Author:      "test-author",
			Tags:        []string{"bulk", "testing"},
			Source: registry.SkillSource{
				Repo: repos[i%len(repos)],
				Path: fmt.Sprintf("skill-%03d", i+1),
				Tag:  "v1.0.0",
			},
		}
	}
	return skills
}
