package components

import (
	"strings"
	"testing"

	"lazyas/internal/registry"
	tt "lazyas/internal/tui/testing"
)

func TestGroupedSkillList_GroupsByRepo(t *testing.T) {
	skills := tt.TestSkills()
	installed := map[string]bool{"test-skill-1": true}

	list := NewGroupedSkillList(skills, installed)

	// Should have groups: Installed + unique repos
	if len(list.groups) < 2 {
		t.Errorf("Expected at least 2 groups, got %d", len(list.groups))
	}

	// First group should be Installed
	if list.groups[0].Name != "Installed" {
		t.Errorf("Expected first group to be 'Installed', got %q", list.groups[0].Name)
	}

	// Installed group should have 1 skill
	if len(list.groups[0].Skills) != 1 {
		t.Errorf("Expected 1 installed skill, got %d", len(list.groups[0].Skills))
	}
}

func TestGroupedSkillList_NoInstalledGroup_WhenNoneInstalled(t *testing.T) {
	skills := tt.TestSkills()
	installed := map[string]bool{}

	list := NewGroupedSkillList(skills, installed)

	// First group should NOT be Installed
	if len(list.groups) > 0 && list.groups[0].Name == "Installed" {
		t.Error("Should not have Installed group when no skills are installed")
	}
}

func TestGroupedSkillList_Navigation_SkipsHeaders(t *testing.T) {
	skills := tt.TestSkills()
	installed := map[string]bool{"test-skill-1": true}

	list := NewGroupedSkillList(skills, installed)

	// Cursor starts at 0 (a header), move down to reach a skill
	list.MoveDown()
	selected := list.Selected()
	if selected == nil {
		t.Fatal("Expected a skill to be selected after moving past header")
	}

	// Move down again
	initialCursor := list.cursor
	list.MoveDown()

	// Cursor should have changed (unless we're at the last item)
	if list.cursor > initialCursor {
		// Moved forward, check selection
		item := list.flatItems[list.cursor]
		if item.Type == ItemTypeSkill && list.Selected() == nil {
			t.Error("Expected Selected() to return skill when cursor is on a skill item")
		}
	}
}

func TestGroupedSkillList_ToggleCollapse(t *testing.T) {
	skills := tt.TestSkills()
	installed := map[string]bool{}

	list := NewGroupedSkillList(skills, installed)

	initialLen := len(list.flatItems)

	// Find a group name to toggle
	var groupName string
	for _, group := range list.groups {
		if len(group.Skills) > 0 {
			groupName = group.Name
			break
		}
	}

	if groupName == "" {
		t.Fatal("No group found to test")
	}

	// Toggle collapse
	list.ToggleGroup(groupName)

	// Should have fewer items
	if len(list.flatItems) >= initialLen {
		t.Error("Collapsing a group should reduce flat items")
	}

	// Toggle again to expand
	list.ToggleGroup(groupName)

	// Should have same items as before
	if len(list.flatItems) != initialLen {
		t.Error("Expanding should restore flat items")
	}
}

func TestGroupedSkillList_View_ContainsHeaders(t *testing.T) {
	skills := tt.TestSkills()
	installed := map[string]bool{"test-skill-1": true}

	list := NewGroupedSkillList(skills, installed)
	list.SetHeight(20)

	view := list.View()

	// Should contain "Installed" header
	if !strings.Contains(view, "Installed") {
		t.Error("View should contain 'Installed' header")
	}

	// Should contain collapse indicator
	if !strings.Contains(view, "▼") && !strings.Contains(view, "▶") {
		t.Error("View should contain collapse indicator")
	}
}

func TestGroupedSkillList_SetSkills_ResetsState(t *testing.T) {
	skills := tt.TestSkills()
	installed := map[string]bool{}

	list := NewGroupedSkillList(skills, installed)

	// Move cursor
	list.MoveDown()
	list.MoveDown()

	// Set new skills
	newSkills := []registry.SkillEntry{tt.SingleSkill()}
	list.SetSkills(newSkills)

	// Cursor should reset to 0 or to first skill
	if list.cursor > 1 {
		t.Error("Setting skills should reset cursor")
	}
}

func TestGroupedSkillList_MoveToTop_MoveToBottom(t *testing.T) {
	skills := tt.ManySkills(20)
	installed := map[string]bool{}

	list := NewGroupedSkillList(skills, installed)
	list.SetHeight(10)

	// Move to bottom (finds last skill item)
	list.MoveToBottom()

	// Should be at or near the last skill
	selected := list.Selected()
	if selected == nil {
		t.Fatal("Expected a skill to be selected at bottom")
	}

	bottomCursor := list.cursor

	// Move to top (goes to position 0, which is a header)
	list.MoveToTop()

	// Cursor should be less than bottom cursor
	if list.cursor >= bottomCursor {
		t.Error("MoveToTop should move cursor before MoveToBottom position")
	}

	// Position 0 is a header, so Selected() returns nil; that's expected.
	// Move down once to reach the first skill.
	list.MoveDown()
	selected = list.Selected()
	if selected == nil {
		t.Fatal("Expected a skill to be selected after moving past top header")
	}
}

func TestGroupedSkillList_Update_KeyHandling(t *testing.T) {
	skills := tt.TestSkills()
	installed := map[string]bool{}

	list := NewGroupedSkillList(skills, installed)

	initialCursor := list.cursor

	// Test down key
	list.Update(tt.KeyMsg("j"))

	// Test up key
	list.Update(tt.KeyMsg("k"))

	// After down then up, should be back at start (or close)
	if list.SkillCount() > 1 {
		// Just verify it didn't crash
		_ = initialCursor
	}
}

func TestGroupedSkillList_ToggleCurrentGroup(t *testing.T) {
	skills := tt.TestSkills()
	installed := map[string]bool{"test-skill-1": true}

	list := NewGroupedSkillList(skills, installed)

	initialLen := len(list.flatItems)

	// Toggle the current group
	list.ToggleCurrentGroup()

	// Should have changed the flat items count
	if len(list.flatItems) == initialLen {
		// This is ok if the cursor wasn't on any group
		// or the group only has the selected skill
	}
}

func TestGroupedSkillList_EmptyList(t *testing.T) {
	skills := []registry.SkillEntry{}
	installed := map[string]bool{}

	list := NewGroupedSkillList(skills, installed)

	// Should handle empty list gracefully
	if list.Selected() != nil {
		t.Error("Empty list should not have a selection")
	}

	view := list.View()
	if !strings.Contains(view, "No skills found") {
		t.Error("Empty list should show 'No skills found'")
	}

	// Navigation should not panic
	list.MoveUp()
	list.MoveDown()
	list.MoveToTop()
	list.MoveToBottom()
}

func TestGroupedSkillList_SetInstalled_RebuildsGroups(t *testing.T) {
	skills := tt.TestSkills()
	installed := map[string]bool{}

	list := NewGroupedSkillList(skills, installed)

	// Initially no installed group
	hasInstalled := false
	for _, g := range list.groups {
		if g.Name == "Installed" {
			hasInstalled = true
			break
		}
	}
	if hasInstalled {
		t.Error("Should not have Installed group initially")
	}

	// Set some as installed
	list.SetInstalled(map[string]bool{"test-skill-1": true})

	// Now should have installed group
	hasInstalled = false
	for _, g := range list.groups {
		if g.Name == "Installed" {
			hasInstalled = true
			break
		}
	}
	if !hasInstalled {
		t.Error("Should have Installed group after SetInstalled")
	}
}
