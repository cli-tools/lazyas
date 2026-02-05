package tui

import (
	"path/filepath"
	"sort"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"lazyas/internal/config"
	"lazyas/internal/manifest"
	"lazyas/internal/registry"
	"lazyas/internal/tui/panels"
	ttesting "lazyas/internal/tui/testing"
)

func TestApp_CtrlC_Quits(t *testing.T) {
	cfg := &config.Config{
		Store:        ttesting.NewMockConfigStore(),
		SkillsDir:    "/tmp/test",
		ConfigDir:    "/tmp/test/.lazyas",
		ConfigPath:   "/tmp/test/.lazyas/config.toml",
		ManifestPath: "/tmp/test/.lazyas/manifest.yaml",
		CachePath:    "/tmp/test/.lazyas/cache.yaml",
		CacheTTL:     24,
	}

	app := NewApp(cfg)

	// Send Ctrl+C
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	// Should return tea.Quit
	if cmd == nil {
		t.Fatal("Expected a command to be returned")
	}

	msg := cmd()
	if msg != tea.Quit() {
		t.Error("Ctrl+C should return tea.Quit")
	}
}

func TestApp_WindowSizeMsg_UpdatesDimensions(t *testing.T) {
	cfg := &config.Config{
		Store:        ttesting.NewMockConfigStore(),
		SkillsDir:    "/tmp/test",
		ConfigDir:    "/tmp/test/.lazyas",
		ConfigPath:   "/tmp/test/.lazyas/config.toml",
		ManifestPath: "/tmp/test/.lazyas/manifest.yaml",
		CachePath:    "/tmp/test/.lazyas/cache.yaml",
		CacheTTL:     24,
	}

	app := NewApp(cfg)

	// Send window size
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if app.width != 120 || app.height != 40 {
		t.Errorf("Expected 120x40, got %dx%d", app.width, app.height)
	}
}

func TestApp_InitialMode_IsLoading(t *testing.T) {
	cfg := &config.Config{
		Store:        ttesting.NewMockConfigStore(),
		SkillsDir:    "/tmp/test",
		ConfigDir:    "/tmp/test/.lazyas",
		ConfigPath:   "/tmp/test/.lazyas/config.toml",
		ManifestPath: "/tmp/test/.lazyas/manifest.yaml",
		CachePath:    "/tmp/test/.lazyas/cache.yaml",
		CacheTTL:     24,
	}

	app := NewApp(cfg)

	if app.mode != ModeLoading {
		t.Errorf("Expected initial mode to be ModeLoading, got %v", app.mode)
	}
}

func TestApp_ConfirmMode_YCancels(t *testing.T) {
	cfg := &config.Config{
		Store:        ttesting.NewMockConfigStore(),
		SkillsDir:    "/tmp/test",
		ConfigDir:    "/tmp/test/.lazyas",
		ConfigPath:   "/tmp/test/.lazyas/config.toml",
		ManifestPath: "/tmp/test/.lazyas/manifest.yaml",
		CachePath:    "/tmp/test/.lazyas/cache.yaml",
		CacheTTL:     24,
	}

	app := NewApp(cfg)
	app.mode = ModeConfirm
	app.confirmSel = 1 // No selected

	// Press n to cancel
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	if app.mode != ModeNormal {
		t.Errorf("Expected mode to be ModeNormal after cancel, got %v", app.mode)
	}
}

func TestApp_SaveCollapseState(t *testing.T) {
	mockStore := ttesting.NewMockConfigStore()
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Store:        mockStore,
		SkillsDir:    filepath.Join(tmpDir, "skills"),
		ConfigDir:    tmpDir,
		ConfigPath:   filepath.Join(tmpDir, "config.toml"),
		ManifestPath: filepath.Join(tmpDir, "manifest.yaml"),
		CachePath:    filepath.Join(tmpDir, "cache.yaml"),
		ReposDir:     filepath.Join(tmpDir, "repos"),
		CacheTTL:     24,
	}

	skills := []registry.SkillEntry{
		{Name: "skill-a", Source: registry.SkillSource{Repo: "https://github.com/org/repo-a"}},
		{Name: "skill-b", Source: registry.SkillSource{Repo: "https://github.com/org/repo-b"}},
	}
	sp := panels.NewSkillsPanel(skills, nil, nil)
	sp.SetCollapseMap(map[string]bool{
		"github.com/org/repo-a": true,
		"github.com/org/repo-b": false,
	})

	app := NewApp(cfg)
	app.skills = sp

	app.saveCollapseState()

	if mockStore.SaveCount != 1 {
		t.Fatalf("Expected SaveCount == 1, got %d", mockStore.SaveCount)
	}
	if mockStore.Data == nil {
		t.Fatal("Expected Data to be non-nil after save")
	}

	got := mockStore.Data.CollapsedGroups
	if len(got) != 1 || got[0] != "github.com/org/repo-a" {
		t.Errorf("Expected CollapsedGroups == [github.com/org/repo-a], got %v", got)
	}
}

func TestApp_InitPanels_RestoresCollapseFromConfig(t *testing.T) {
	mockStore := ttesting.NewMockConfigStore()
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Store:           mockStore,
		SkillsDir:       filepath.Join(tmpDir, "skills"),
		ConfigDir:       tmpDir,
		ConfigPath:      filepath.Join(tmpDir, "config.toml"),
		ManifestPath:    filepath.Join(tmpDir, "manifest.yaml"),
		CachePath:       filepath.Join(tmpDir, "cache.yaml"),
		ReposDir:        filepath.Join(tmpDir, "repos"),
		CacheTTL:        24,
		CollapsedGroups: []string{"group-a", "group-b"},
	}

	app := NewApp(cfg)
	// Ensure registry and manifest are usable (empty results are fine)
	app.registry = registry.NewRegistry(cfg)
	app.manifest = manifest.NewManager(cfg)

	app.initPanels()

	cm := app.skills.GetCollapseMap()
	if !cm["group-a"] {
		t.Error("Expected group-a to be collapsed")
	}
	if !cm["group-b"] {
		t.Error("Expected group-b to be collapsed")
	}
}

func TestApp_InitPanels_PreservesCollapseAcrossRebuild(t *testing.T) {
	mockStore := ttesting.NewMockConfigStore()
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Store:        mockStore,
		SkillsDir:    filepath.Join(tmpDir, "skills"),
		ConfigDir:    tmpDir,
		ConfigPath:   filepath.Join(tmpDir, "config.toml"),
		ManifestPath: filepath.Join(tmpDir, "manifest.yaml"),
		CachePath:    filepath.Join(tmpDir, "cache.yaml"),
		ReposDir:     filepath.Join(tmpDir, "repos"),
		CacheTTL:     24,
	}

	app := NewApp(cfg)
	app.registry = registry.NewRegistry(cfg)
	app.manifest = manifest.NewManager(cfg)

	// Build panels the first time
	app.initPanels()

	// Simulate collapsing a group on the existing panel
	app.skills.SetCollapseMap(map[string]bool{
		"my-group": true,
	})

	// Rebuild panels (simulates e.g. adding a starter kit repo)
	app.initPanels()

	cm := app.skills.GetCollapseMap()
	if !cm["my-group"] {
		t.Error("Expected my-group to remain collapsed after panel rebuild")
	}

	// Verify only collapsed groups survive (no stale entries with false values)
	var collapsed []string
	for name, v := range cm {
		if v {
			collapsed = append(collapsed, name)
		}
	}
	sort.Strings(collapsed)
	if len(collapsed) != 1 || collapsed[0] != "my-group" {
		t.Errorf("Expected exactly [my-group] collapsed, got %v", collapsed)
	}
}
