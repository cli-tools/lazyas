package git

import (
	"os"
	"path/filepath"
)

// ValidateSkill checks if a skill directory is valid
func ValidateSkill(skillPath string) error {
	// Check for SKILL.md
	skillMD := filepath.Join(skillPath, "SKILL.md")
	if _, err := os.Stat(skillMD); os.IsNotExist(err) {
		return &ValidationError{
			Path:    skillPath,
			Message: "SKILL.md not found",
		}
	}
	return nil
}

// ValidationError represents a skill validation error
type ValidationError struct {
	Path    string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
