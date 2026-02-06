package registry

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// helper to create a directory with a SKILL.md file
func createSkill(t *testing.T, base string, parts ...string) {
	t.Helper()
	dir := filepath.Join(append([]string{base}, parts...)...)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# Test Skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScanForSkills_NestedSkills(t *testing.T) {
	tmp := t.TempDir()

	// Flat skill at root level: flat-skill/SKILL.md
	createSkill(t, tmp, "flat-skill")

	// Nested skill under a category: category/nested-skill/SKILL.md
	createSkill(t, tmp, "category", "nested-skill")

	r := &Registry{}
	skills, err := r.scanForSkills(tmp, "https://example.com/repo.git")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	sort.Strings(names)

	if len(names) != 2 {
		t.Fatalf("expected 2 skills, got %d: %v", len(names), names)
	}
	if names[0] != "flat-skill" || names[1] != "nested-skill" {
		t.Errorf("expected [flat-skill nested-skill], got %v", names)
	}

	// Verify paths
	pathMap := map[string]string{}
	for _, s := range skills {
		pathMap[s.Name] = s.Source.Path
	}
	if pathMap["flat-skill"] != "flat-skill" {
		t.Errorf("flat-skill path = %q, want %q", pathMap["flat-skill"], "flat-skill")
	}
	if pathMap["nested-skill"] != filepath.Join("category", "nested-skill") {
		t.Errorf("nested-skill path = %q, want %q", pathMap["nested-skill"], filepath.Join("category", "nested-skill"))
	}
}

func TestScanForSkills_SkipsHiddenDirs(t *testing.T) {
	tmp := t.TempDir()

	// Visible skill
	createSkill(t, tmp, "visible-skill")

	// Hidden top-level dir with a skill inside
	createSkill(t, tmp, ".hidden", "secret-skill")

	// Hidden nested dir inside a visible category
	createSkill(t, tmp, "category", ".private")

	r := &Registry{}
	skills, err := r.scanForSkills(tmp, "https://example.com/repo.git")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d: %v", len(skills), skills)
	}
	if skills[0].Name != "visible-skill" {
		t.Errorf("expected visible-skill, got %s", skills[0].Name)
	}
}

func TestScanForSkills_DeduplicatesAcrossSearchDirs(t *testing.T) {
	tmp := t.TempDir()

	// Skill inside skills/ subdir: the root scan finds "skills/" (no SKILL.md),
	// goes one level deeper and discovers "my-skill". The skills/ search dir
	// also finds "my-skill" directly. Without dedup this produces two entries.
	createSkill(t, tmp, "skills", "my-skill")

	r := &Registry{}
	skills, err := r.scanForSkills(tmp, "https://example.com/repo.git")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d: %v", len(skills), skillNames(skills))
	}
	if skills[0].Name != "my-skill" {
		t.Errorf("expected my-skill, got %s", skills[0].Name)
	}
	if skills[0].Source.Path != filepath.Join("skills", "my-skill") {
		t.Errorf("path = %q, want %q", skills[0].Source.Path, filepath.Join("skills", "my-skill"))
	}
}

func TestScanForSkills_DeduplicatesMultipleOverlaps(t *testing.T) {
	tmp := t.TempDir()

	// Two skills in skills/ — both would be found by root (nested) and skills/ (direct)
	createSkill(t, tmp, "skills", "alpha")
	createSkill(t, tmp, "skills", "beta")
	// A root-level skill that has no overlap
	createSkill(t, tmp, "root-skill")

	r := &Registry{}
	skills, err := r.scanForSkills(tmp, "https://example.com/repo.git")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	names := skillNames(skills)
	sort.Strings(names)

	if len(names) != 3 {
		t.Fatalf("expected 3 skills, got %d: %v", len(names), names)
	}
	want := []string{"alpha", "beta", "root-skill"}
	for i, w := range want {
		if names[i] != w {
			t.Errorf("names[%d] = %q, want %q", i, names[i], w)
		}
	}
}

func TestScanForSkills_EmptyRepo(t *testing.T) {
	tmp := t.TempDir()

	r := &Registry{}
	_, err := r.scanForSkills(tmp, "https://example.com/repo.git")
	if err == nil {
		t.Fatal("expected error for empty repo, got nil")
	}
}

func skillNames(skills []SkillEntry) []string {
	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	return names
}
