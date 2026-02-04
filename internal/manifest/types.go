package manifest

import "time"

// Manifest represents the local manifest file structure
type Manifest struct {
	Version   int                       `yaml:"version"`
	Installed map[string]InstalledSkill `yaml:"installed"`
}

// InstalledSkill represents an installed skill tracked in manifest
type InstalledSkill struct {
	Version     string    `yaml:"version"`
	Commit      string    `yaml:"commit"`
	InstalledAt time.Time `yaml:"installed_at"`
	SourceRepo  string    `yaml:"source_repo"`
	SourcePath  string    `yaml:"source_path,omitempty"`
}

// LocalSkill represents a skill found on the local filesystem
type LocalSkill struct {
	Name        string
	Path        string
	Description string
	IsGitRepo   bool
	IsModified  bool
}

// NewManifest creates a new manifest with defaults
func NewManifest() *Manifest {
	return &Manifest{
		Version:   1,
		Installed: make(map[string]InstalledSkill),
	}
}
