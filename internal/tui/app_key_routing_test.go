package tui

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"lazyas/internal/config"
	"lazyas/internal/tui/panels"
	ttesting "lazyas/internal/tui/testing"
)

func newAppForPageKeyRoutingTest(t *testing.T) *App {
	t.Helper()

	tmpDir := t.TempDir()
	cfg := &config.Config{
		Store:        ttesting.NewMockConfigStore(),
		SkillsDir:    filepath.Join(tmpDir, "skills"),
		ConfigDir:    filepath.Join(tmpDir, ".lazyas"),
		ConfigPath:   filepath.Join(tmpDir, ".lazyas", "config.toml"),
		ManifestPath: filepath.Join(tmpDir, ".lazyas", "manifest.yaml"),
		CachePath:    filepath.Join(tmpDir, ".lazyas", "cache.yaml"),
		ReposDir:     filepath.Join(tmpDir, "repos"),
		CacheTTL:     24,
	}

	app := NewApp(cfg)
	app.mode = ModeNormal
	app.layout.FocusLeft()

	app.skills = panels.NewSkillsPanel(ttesting.ManySkills(30), map[string]string{}, map[string]bool{})
	app.skills.SetSize(60, 10) // half-page jump = 5
	app.skills.SetFocused(true)

	app.detail = panels.NewDetailPanel()
	app.detail.SetFocused(false)

	return app
}

// Integration-style routing check: keys enter at App.Update and must reach SkillsPanel.
func TestApp_PageKeys_RoutedToSkillsPanel(t *testing.T) {
	app := newAppForPageKeyRoutingTest(t)

	if got := app.skills.Selected(); got != nil {
		t.Fatalf("expected initial selection on header (nil skill), got %q", got.Name)
	}

	app.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if got := app.skills.Selected(); got == nil {
		t.Fatal("expected PageDown routed through App.Update to select a skill")
	}

	app.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	if got := app.skills.Selected(); got != nil {
		t.Fatalf("expected PageUp routed through App.Update to return to top header, got %q", got.Name)
	}

	app.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if got := app.skills.Selected(); got == nil {
		t.Fatal("expected End routed through App.Update to select the bottom skill")
	}

	app.Update(tea.KeyMsg{Type: tea.KeyHome})
	if got := app.skills.Selected(); got != nil {
		t.Fatalf("expected Home routed through App.Update to return to top header, got %q", got.Name)
	}
}

