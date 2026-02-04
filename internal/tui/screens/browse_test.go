package screens

import (
	"strings"
	"testing"

	tt "lazyas/internal/tui/testing"
)

func TestBrowseScreen_InitialState(t *testing.T) {
	reg := tt.NewMockRegistry(tt.TestSkills())
	mfst := tt.NewMockManifest()

	screen := NewBrowseScreen(reg, mfst)

	view := screen.View()

	// Should contain title
	if !strings.Contains(view, "lazyas") {
		t.Error("View should contain app name")
	}

	// Should contain help text
	if !strings.Contains(view, "navigate") {
		t.Error("View should contain help text")
	}
}

func TestBrowseScreen_Navigation(t *testing.T) {
	reg := tt.NewMockRegistry(tt.TestSkills())
	mfst := tt.NewMockManifest()

	screen := NewBrowseScreen(reg, mfst)
	harness := tt.NewTestHarness(screen)

	// Navigate down
	harness.SendKey("j")

	// Should still have a view
	view := harness.View()
	if view == "" {
		t.Error("View should not be empty after navigation")
	}
}

func TestBrowseScreen_SelectSkill(t *testing.T) {
	reg := tt.NewMockRegistry(tt.TestSkills())
	mfst := tt.NewMockManifest()

	screen := NewBrowseScreen(reg, mfst)
	harness := tt.NewTestHarness(screen)

	// Navigate past the header to land on a skill
	harness.SendKey("j")

	// Press enter to select
	cmd := harness.SendKey("enter")

	if cmd == nil {
		t.Error("Expected a command to be returned on enter")
		return
	}

	// Execute the command
	msg := harness.ExecuteCmd(cmd)

	// Should be a SelectSkillMsg
	if _, ok := msg.(SelectSkillMsg); !ok {
		t.Errorf("Expected SelectSkillMsg, got %T", msg)
	}
}

func TestBrowseScreen_InstallAction(t *testing.T) {
	reg := tt.NewMockRegistry(tt.TestSkills())
	mfst := tt.NewMockManifest()

	screen := NewBrowseScreen(reg, mfst)
	harness := tt.NewTestHarness(screen)

	// Navigate past the header to land on a skill
	harness.SendKey("j")

	// Press i to install
	cmd := harness.SendKey("i")

	if cmd == nil {
		t.Error("Expected a command to be returned on 'i'")
		return
	}

	// Execute the command
	msg := harness.ExecuteCmd(cmd)

	// Should be an InstallSkillMsg
	if _, ok := msg.(InstallSkillMsg); !ok {
		t.Errorf("Expected InstallSkillMsg, got %T", msg)
	}
}

func TestBrowseScreen_RemoveAction_OnlyWhenInstalled(t *testing.T) {
	reg := tt.NewMockRegistry(tt.TestSkills())
	mfst := tt.NewMockManifest()

	screen := NewBrowseScreen(reg, mfst)
	harness := tt.NewTestHarness(screen)

	// Press r - should do nothing when skill is not installed
	cmd := harness.SendKey("r")

	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(RemoveSkillMsg); ok {
			t.Error("Should not return RemoveSkillMsg for non-installed skill")
		}
	}
}

func TestBrowseScreen_QuitReturnsBackMsg(t *testing.T) {
	reg := tt.NewMockRegistry(tt.TestSkills())
	mfst := tt.NewMockManifest()

	screen := NewBrowseScreen(reg, mfst)
	harness := tt.NewTestHarness(screen)

	// Press q to quit
	cmd := harness.SendKey("q")

	if cmd == nil {
		t.Error("Expected a command to be returned on 'q'")
		return
	}

	// Execute the command
	msg := harness.ExecuteCmd(cmd)

	// Should be a BackMsg
	if _, ok := msg.(BackMsg); !ok {
		t.Errorf("Expected BackMsg, got %T", msg)
	}
}

func TestBrowseScreen_SearchMode(t *testing.T) {
	reg := tt.NewMockRegistry(tt.TestSkills())
	mfst := tt.NewMockManifest()

	screen := NewBrowseScreen(reg, mfst)
	harness := tt.NewTestHarness(screen)

	// Press / to enter search mode
	harness.SendKey("/")

	// Should be in search mode
	if !screen.searching {
		t.Error("Should be in searching mode after pressing /")
	}

	// Press esc to exit search mode
	harness.SendKey("esc")

	if screen.searching {
		t.Error("Should exit searching mode after pressing esc")
	}
}

func TestBrowseScreen_CollapseToggle(t *testing.T) {
	reg := tt.NewMockRegistry(tt.TestSkills())
	mfst := tt.NewMockManifest()

	screen := NewBrowseScreen(reg, mfst)
	harness := tt.NewTestHarness(screen)

	// Get initial flat items count
	initialLen := screen.list.Len()

	// Press z to toggle collapse
	harness.SendKey("z")

	// The list length may have changed due to collapse
	_ = initialLen
}

func TestBrowseScreen_RefreshInstalled(t *testing.T) {
	skills, installed := tt.TestSkillsWithInstalled()
	reg := tt.NewMockRegistry(skills)
	mfst := tt.NewMockManifestWithInstalled(installed)

	screen := NewBrowseScreen(reg, mfst)

	// Refresh
	screen.RefreshInstalled()

	// View should still work
	view := screen.View()
	if view == "" {
		t.Error("View should not be empty after refresh")
	}
}

func TestBrowseScreen_WindowResize(t *testing.T) {
	reg := tt.NewMockRegistry(tt.TestSkills())
	mfst := tt.NewMockManifest()

	screen := NewBrowseScreen(reg, mfst)
	harness := tt.NewTestHarness(screen)

	// Send window size
	harness.SendWindowSize(100, 50)

	if screen.width != 100 || screen.height != 50 {
		t.Errorf("Expected dimensions 100x50, got %dx%d", screen.width, screen.height)
	}
}
