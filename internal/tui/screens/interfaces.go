package screens

import (
	"lazyas/internal/manifest"
	"lazyas/internal/registry"
)

// SkillRegistry defines the interface for registry operations used by screens
type SkillRegistry interface {
	ListSkills() []registry.SkillEntry
	SearchSkills(query string) []registry.SkillEntry
	GetSkill(name string) *registry.SkillEntry
}

// SkillManifest defines the interface for manifest operations used by screens
type SkillManifest interface {
	IsInstalled(name string) bool
	ListInstalled() map[string]manifest.InstalledSkill
	GetInstalled(name string) (manifest.InstalledSkill, bool)
	GetSkillPath(name string) string
	ScanLocalSkills() map[string]manifest.LocalSkill
}
