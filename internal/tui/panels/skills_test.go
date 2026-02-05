package panels

import (
	"strings"
	"testing"

	"lazyas/internal/registry"
)

// makeSkills creates n skills spread across repos so the flat list
// includes group headers + skill items.
func makeSkills(n int) []registry.SkillEntry {
	repos := []string{
		"https://github.com/repo-a/skills",
		"https://github.com/repo-b/skills",
	}
	skills := make([]registry.SkillEntry, n)
	for i := 0; i < n; i++ {
		skills[i] = registry.SkillEntry{
			Name: "skill-" + string(rune('A'+i%26)) + string(rune('0'+i/26)),
			Source: registry.SkillSource{
				Repo: repos[i%len(repos)],
			},
		}
	}
	return skills
}

func TestSkillsPanel_ViewFillsHeight(t *testing.T) {
	// Create 50 skills — more than enough to fill any screen.
	skills := makeSkills(50)
	installed := map[string]string{}
	modified := map[string]bool{}

	p := NewSkillsPanel(skills, installed, modified)

	// Simulate a tall terminal: set height to 40.
	p.SetSize(60, 40)

	view := p.View()
	lines := strings.Split(view, "\n")

	// The view should produce exactly p.height visible lines when
	// there are more flat items than the height.
	if len(lines) != 40 {
		t.Errorf("expected View to produce %d lines for height=%d, got %d",
			40, 40, len(lines))
		// Show the first and last few lines for debugging.
		if len(lines) > 5 {
			t.Logf("first 3 lines: %q", lines[:3])
			t.Logf("last 3 lines:  %q", lines[len(lines)-3:])
		}
	}
}

func TestSkillsPanel_ViewFillsHeight_FewItems(t *testing.T) {
	// Only 5 skills — fewer than the panel height.
	skills := makeSkills(5)
	installed := map[string]string{}
	modified := map[string]bool{}

	p := NewSkillsPanel(skills, installed, modified)
	p.SetSize(60, 40)

	view := p.View()
	lines := strings.Split(view, "\n")

	// With 5 skills in 2 repos → 2 headers + 5 items = 7 flat items.
	// View should only emit the items that exist, NOT pad to 40.
	flatCount := len(p.flatItems)
	if len(lines) != flatCount {
		t.Errorf("expected %d lines (one per flat item), got %d", flatCount, len(lines))
	}
}

func TestSkillsPanel_DefaultHeight_IsUsable(t *testing.T) {
	// Verify the default height (before any SetSize call) can show items.
	skills := makeSkills(30)
	installed := map[string]string{}
	modified := map[string]bool{}

	p := NewSkillsPanel(skills, installed, modified)
	// NOTE: no SetSize call — uses defaults (height=20, width=30)

	view := p.View()
	lines := strings.Split(view, "\n")

	if len(lines) < 10 {
		t.Errorf("default height should show at least 10 lines, got %d", len(lines))
	}

	if len(lines) > 20 {
		t.Errorf("default height=20 should cap at 20 lines, got %d", len(lines))
	}
}

func TestSkillsPanel_ScrollShowsAllItems(t *testing.T) {
	skills := makeSkills(50)
	installed := map[string]string{}
	modified := map[string]bool{}

	p := NewSkillsPanel(skills, installed, modified)
	p.SetSize(60, 10) // Small viewport

	// Collect all unique lines seen while scrolling through the entire list.
	seen := make(map[string]bool)
	for i := 0; i < len(p.flatItems)+5; i++ {
		view := p.View()
		for _, line := range strings.Split(view, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				seen[trimmed] = true
			}
		}
		p.moveDown()
	}

	// Every skill name should appear at least once.
	for _, s := range skills {
		found := false
		for text := range seen {
			if strings.Contains(text, s.Name) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("skill %q was never visible while scrolling", s.Name)
		}
	}
}
