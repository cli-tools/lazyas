package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"lazyas/internal/config"
)

func TestApp_CtrlC_Quits(t *testing.T) {
	cfg := &config.Config{
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
